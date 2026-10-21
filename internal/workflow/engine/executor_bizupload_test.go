package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
)

type uploadTestCertificateRepo struct {
	previous, saved *domain.Certificate
}

func (r *uploadTestCertificateRepo) GetById(context.Context, string) (*domain.Certificate, error) {
	return r.previous, nil
}

func (r *uploadTestCertificateRepo) GetByWorkflowRunIdAndNodeId(context.Context, string, string) (*domain.Certificate, error) {
	return r.previous, nil
}

func (r *uploadTestCertificateRepo) Save(_ context.Context, certificate *domain.Certificate) (*domain.Certificate, error) {
	certificate.Id = "new-certificate"
	certificate.CreatedAt = time.Now()
	r.saved = certificate
	return certificate, nil
}

type uploadTestOutputRepo struct{ config domain.WorkflowNodeConfig }

func (r *uploadTestOutputRepo) GetByWorkflowIdAndNodeId(context.Context, string, string) (*domain.WorkflowOutput, error) {
	return &domain.WorkflowOutput{RunId: "previous-run", NodeId: "upload", NodeConfig: r.config, Succeeded: true}, nil
}

func (r *uploadTestOutputRepo) Save(context.Context, *domain.WorkflowOutput) (*domain.WorkflowOutput, error) {
	return nil, fmt.Errorf("unexpected output save")
}

func uploadTestChains(t *testing.T) (previous, changed, key string) {
	t.Helper()
	newKey := func() (ed25519.PublicKey, ed25519.PrivateKey) {
		t.Helper()
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		return publicKey, privateKey
	}
	issue := func(template, parent *x509.Certificate, publicKey ed25519.PublicKey, signer ed25519.PrivateKey) (*x509.Certificate, string) {
		t.Helper()
		der, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, signer)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return cert, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	}
	template := func(serial int64, ca bool) *x509.Certificate {
		return &x509.Certificate{
			SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "test.example.com"},
			NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
			IsCA: ca, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		}
	}
	rootPublic, rootKey := newKey()
	rootTemplate := template(1, true)
	root, rootPEM := issue(rootTemplate, rootTemplate, rootPublic, rootKey)
	intermediatePublic, intermediateKey := newKey()
	intermediate, intermediatePEM := issue(template(2, true), root, intermediatePublic, rootKey)
	_, alternatePEM := issue(template(3, true), root, intermediatePublic, rootKey)
	leafPublic, leafKey := newKey()
	_, leafPEM := issue(template(4, false), intermediate, leafPublic, intermediateKey)
	keyDER, err := x509.MarshalPKCS8PrivateKey(leafKey)
	if err != nil {
		t.Fatal(err)
	}
	return leafPEM + intermediatePEM + rootPEM, leafPEM + alternatePEM + rootPEM,
		string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
}

func TestBizUploadCertificateChainChanges(t *testing.T) {
	previous, changed, key := uploadTestChains(t)
	for _, source := range []string{BizUploadSourceForm, BizUploadSourceLocal, BizUploadSourceURL} {
		for _, update := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/changed=%v", source, update), func(t *testing.T) {
				certificate := previous
				if update {
					certificate = changed
				} else {
					// Exercise the secondary comparison even for form uploads.
					certificate = strings.ReplaceAll(previous, "\n", "\r\n")
				}
				config := domain.WorkflowNodeConfig{"source": source, "certificate": certificate, "privateKey": key}
				lastConfig := domain.WorkflowNodeConfig{"source": source, "certificate": previous, "privateKey": key}
				if source == BizUploadSourceLocal {
					dir := t.TempDir()
					for name, content := range map[string]string{"certificate": certificate, "privateKey": key} {
						path := filepath.Join(dir, name)
						if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
							t.Fatal(err)
						}
						config[name] = path
					}
				} else if source == BizUploadSourceURL {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path == "/certificate" {
							_, _ = io.WriteString(w, certificate)
						} else {
							_, _ = io.WriteString(w, key)
						}
					}))
					defer server.Close()
					config["certificate"], config["privateKey"] = server.URL+"/certificate", server.URL+"/key"
				}
				repo := &uploadTestCertificateRepo{previous: (&domain.Certificate{WorkflowRunId: "previous-run"}).PopulateFromPEM(previous, key)}
				repo.previous.Id = "previous-certificate"
				executor := &bizUploadNodeExecutor{
					nodeExecutor:    nodeExecutor{logger: slog.New(slog.NewTextHandler(io.Discard, nil))},
					certificateRepo: repo, wfoutputRepo: &uploadTestOutputRepo{config: lastConfig},
				}
				ctx := (&NodeExecutionContext{}).SetContext(context.Background()).SetExecutingWorkflow("workflow", "new-run", nil).
					SetExecutingNode(&Node{Id: "upload", Type: NodeTypeBizUpload, Data: domain.WorkflowNodeData{Config: config}})
				result, err := executor.Execute(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if (repo.saved != nil) != update {
					t.Fatalf("saved certificate = %v, want chain changed = %v", repo.saved != nil, update)
				}
				if update {
					if repo.saved.Certificate != certificate {
						t.Fatal("saved certificate must contain the replacement chain")
					}
					found := false
					for _, output := range result.Outputs {
						found = found || (output.Persistent && output.Value == "certificate#new-certificate")
					}
					if !found {
						t.Fatal("downstream nodes must receive the new certificate record")
					}
				} else {
					for _, output := range result.Outputs {
						if output.Persistent {
							t.Fatal("unchanged chains must not create a new persistent output")
						}
					}
					if len(result.Variables) == 0 || result.Variables[len(result.Variables)-1].Value != true {
						t.Fatal("unchanged chains must mark the upload as skipped")
					}
				}
			})
		}
	}
}

func TestBizUploadValidatesChangedPrivateKeyBeforeSkipping(t *testing.T) {
	previous, _, key := uploadTestChains(t)
	_, _, wrongKey := uploadTestChains(t)
	repo := &uploadTestCertificateRepo{previous: (&domain.Certificate{}).PopulateFromPEM(previous, key)}
	config := domain.WorkflowNodeConfig{"source": BizUploadSourceForm, "certificate": previous, "privateKey": wrongKey}
	executor := &bizUploadNodeExecutor{
		nodeExecutor:    nodeExecutor{logger: slog.New(slog.NewTextHandler(io.Discard, nil))},
		certificateRepo: repo,
		wfoutputRepo: &uploadTestOutputRepo{config: domain.WorkflowNodeConfig{
			"source": BizUploadSourceForm, "certificate": previous, "privateKey": key,
		}},
	}
	ctx := (&NodeExecutionContext{}).SetContext(context.Background()).SetExecutingWorkflow("workflow", "new-run", nil).
		SetExecutingNode(&Node{Id: "upload", Type: NodeTypeBizUpload, Data: domain.WorkflowNodeData{Config: config}})
	_, err := executor.Execute(ctx)
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected mismatched private key error, got %v", err)
	}
	if repo.saved != nil {
		t.Fatal("mismatched private key must not be saved")
	}
}
