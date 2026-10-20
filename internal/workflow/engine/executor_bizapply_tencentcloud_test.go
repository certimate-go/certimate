package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log/slog"
	"maps"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/certapply"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	_ "github.com/certimate-go/certimate/migrations"
)

type nativeTestAccessRepo struct{ provider string }

func (r *nativeTestAccessRepo) GetById(context.Context, string) (*domain.Access, error) {
	return &domain.Access{Provider: r.provider, Config: map[string]any{"secretId": "test-id", "secretKey": "test-secret", "projectId": 42}}, nil
}
func (r *nativeTestAccessRepo) Save(_ context.Context, access *domain.Access) (*domain.Access, error) {
	return access, nil
}

type nativeTestCertificateRepo struct {
	certificate *domain.Certificate
	saveErr     error
}

func (r *nativeTestCertificateRepo) GetById(context.Context, string) (*domain.Certificate, error) {
	return r.certificate, nil
}
func (r *nativeTestCertificateRepo) GetByWorkflowRunIdAndNodeId(_ context.Context, runId, _ string) (*domain.Certificate, error) {
	if r.certificate == nil || r.certificate.WorkflowRunId != runId {
		return nil, domain.ErrRecordNotFound
	}
	return r.certificate, nil
}
func (r *nativeTestCertificateRepo) Save(_ context.Context, cert *domain.Certificate) (*domain.Certificate, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	cert.Id = "saved-certificate"
	r.certificate = cert
	return cert, nil
}

type nativeTestOutputRepo struct {
	last    *domain.WorkflowOutput
	saves   int
	saveErr error
}

func (r *nativeTestOutputRepo) GetByWorkflowIdAndNodeId(context.Context, string, string) (*domain.WorkflowOutput, error) {
	if r.last == nil {
		return nil, domain.ErrRecordNotFound
	}
	return r.last, nil
}
func (r *nativeTestOutputRepo) Save(_ context.Context, output *domain.WorkflowOutput) (*domain.WorkflowOutput, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	copied := *output
	copied.NodeConfig = maps.Clone(output.NodeConfig)
	copied.Id = "pending-output"
	r.last = &copied
	r.saves++
	return &copied, nil
}

type nativeTestIssuer func(context.Context, certapply.TencentCloudRequest, func(string) error) (*certapply.TencentCloudResult, error)

func (f nativeTestIssuer) Obtain(ctx context.Context, req certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
	return f(ctx, req, created)
}

func nativeTestNodeConfig() domain.WorkflowNodeConfig {
	return domain.WorkflowNodeConfig{
		"source": "tencentcloud", "identifier": "domain", "domains": "test.example.com", "tencentcloudAccessId": "tc-access",
		"keySource": "auto", "keyAlgorithm": "EC256", "contactEmail": "test@example.com", "skipBeforeExpiryDays": 30,
	}
}

func nativeTestContext(config domain.WorkflowNodeConfig, runId string) *NodeExecutionContext {
	return (&NodeExecutionContext{}).SetContext(context.Background()).SetExecutingWorkflow("workflow", runId, nil).
		SetExecutingNode(&Node{Id: "apply-node", Type: NodeTypeBizApply, Data: domain.WorkflowNodeData{Config: config}})
}

func nativeTestExecutor(t *testing.T, issuer nativeTestIssuer) (*bizApplyNodeExecutor, *nativeTestOutputRepo, *nativeTestCertificateRepo) {
	t.Helper()
	outputs := &nativeTestOutputRepo{}
	certificates := &nativeTestCertificateRepo{}
	executor := &bizApplyNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.New(slog.NewTextHandler(io.Discard, nil))},
		accessRepo:   &nativeTestAccessRepo{provider: "tencentcloud"}, certificateRepo: certificates, wfoutputRepo: outputs,
		newTencentCloudClient: func(config certapply.TencentCloudConfig, _ *slog.Logger) (tencentCloudIssuer, error) {
			if config.SecretId != "test-id" || config.SecretKey != "test-secret" || config.ProjectId != 42 {
				t.Fatal("credentials were not mapped correctly")
			}
			return issuer, nil
		},
	}
	return executor, outputs, certificates
}

func nativeTestPEM(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test.example.com"}, DNSNames: []string{"test.example.com"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(90 * 24 * time.Hour)}
	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "Test CA"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(365 * 24 * time.Hour)}
	issuerDER, err := x509.CreateCertificate(rand.Reader, issuer, issuer, issuerKey.Public(), issuerKey)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuer, key.Public(), issuerKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	bundle := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: issuerDER})...)
	return string(bundle), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
}

type nativeTestTransport func(*http.Request) (*http.Response, error)

func (f nativeTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func nativeTestDatabase(t *testing.T, dataDir string) core.App {
	t.Helper()
	if dataDir == "" {
		t.Fatal("isolated test database directory is required")
	}
	originalArgs := os.Args
	os.Args = []string{os.Args[0], "--dir", dataDir}
	t.Cleanup(func() { os.Args = originalArgs })
	pb := app.GetApp()
	if pb.DataDir() != dataDir {
		t.Fatal("refusing to bootstrap outside the isolated test database")
	}
	if err := pb.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pb.ResetBootstrapState(); err != nil {
			t.Error(err)
		}
	})
	if err := pb.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	return pb
}

func nativeTestRecord(t *testing.T, pb core.App, collectionName, id string, fields map[string]any) {
	t.Helper()
	collection, err := pb.FindCollectionByNameOrId(collectionName)
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Id = id
	for name, value := range fields {
		record.Set(name, value)
	}
	if err := pb.Save(record); err != nil {
		t.Fatal(err)
	}
}

func TestNativeTencentCloudDatabaseRecovery(t *testing.T) {
	phase := os.Getenv("CERTIMATE_TENCENTCLOUD_TEST_PHASE")
	if phase == "" {
		dataDir := t.TempDir()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, step := range []string{"create", "resume", "recover-saved", "renew"} {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			command := exec.CommandContext(ctx, executable, "-test.run=^TestNativeTencentCloudDatabaseRecovery$", "-test.v")
			command.Env = append(os.Environ(), "CERTIMATE_TENCENTCLOUD_TEST_PHASE="+step, "CERTIMATE_TENCENTCLOUD_TEST_DIR="+dataDir)
			output, err := command.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("database recovery phase %s failed: %v\n%s", step, err, output)
			}
			t.Logf("database recovery phase %s passed in a separate process", step)
		}
		return
	}

	pb := nativeTestDatabase(t, os.Getenv("CERTIMATE_TENCENTCLOUD_TEST_DIR"))

	const workflowId, accessId = "111111111111111", "222222222222222"
	const cloudId = "persisted-cloud-order"
	config := nativeTestNodeConfig()
	config["tencentcloudAccessId"] = accessId
	outputs := repository.NewWorkflowOutputRepository()
	certificates := repository.NewCertificateRepository()
	ctx := context.Background()
	runId := "333333333333333"
	if phase == "create" {
		nativeTestRecord(t, pb, domain.CollectionNameWorkflow, workflowId, map[string]any{"name": "Native database recovery"})
		nativeTestRecord(t, pb, domain.CollectionNameAccess, accessId, map[string]any{
			"name": "Mock Tencent credential", "provider": "tencentcloud",
			"config": map[string]any{"secretId": "test-id", "secretKey": "test-secret", "projectId": 42},
		})
	} else {
		last, err := outputs.GetByWorkflowIdAndNodeId(ctx, workflowId, "apply-node")
		if err != nil {
			t.Fatal(err)
		}
		config = last.NodeConfig
		nodeConfig := config.AsBizApply()
		if last.Succeeded || tencentCloudPendingCertificate(&nodeConfig, last) != cloudId {
			t.Fatal("order ID or native configuration did not survive process restart")
		}
		runId = "444444444444444"
		if phase == "recover-saved" {
			runId = "555555555555555"
		}
	}
	requestedCloudId := cloudId
	if phase == "renew" {
		runId = "666666666666666"
		requestedCloudId = "renewed-cloud-order"
		previous, err := certificates.GetByWorkflowRunIdAndNodeId(ctx, "444444444444444", "apply-node")
		if err != nil {
			t.Fatal(err)
		}
		previous.ValidityNotAfter = time.Now().Add(20 * 24 * time.Hour)
		if _, err := certificates.Save(ctx, previous); err != nil {
			t.Fatal(err)
		}
	}
	nativeTestRecord(t, pb, domain.CollectionNameWorkflowRun, runId, map[string]any{"workflowRef": workflowId, "status": "processing", "trigger": "manual"})

	certificatePEM, privateKeyPEM := nativeTestPEM(t)
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for name, content := range map[string]string{"Nginx/test.example.com_bundle.crt": certificatePEM, "Nginx/test.example.com.key": privateKeyPEM} {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(file, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	requestCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var actions []string
	transport := nativeTestTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Scheme != "https" || request.URL.Host != "ssl.tencentcloudapi.com" || request.Method != http.MethodPost {
			t.Fatal("refusing unexpected cloud request")
		}
		if !strings.HasPrefix(request.Header.Get("Authorization"), "TC3-HMAC-SHA256 Credential=test-id/") {
			t.Fatal("expected SDK request signed with the database credential")
		}
		if phase == "recover-saved" {
			t.Fatal("saved certificate must be reused without another cloud request")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		action := request.Header.Get("X-TC-Action")
		actions = append(actions, action)
		if action != "ApplyCertificate" && body["CertificateId"] != requestedCloudId {
			t.Fatal("cloud request did not use the persisted order ID")
		}
		response := map[string]any{"RequestId": "local-test-request"}
		switch action {
		case "ApplyCertificate":
			if (phase != "create" && phase != "renew") || len(actions) != 1 {
				t.Fatal("recovery attempted to create a duplicate cloud order")
			}
			for field, expected := range map[string]any{
				"DomainName": "test.example.com", "DvAuthMethod": "DNS_AUTO", "PackageType": "83", "ValidityPeriod": "3",
				"ProjectId": float64(42), "ContactEmail": "test@example.com", "DeleteDnsAutoRecord": true,
				"CsrEncryptAlgo": "ECC", "CsrKeyParameter": "prime256v1",
			} {
				if body[field] != expected {
					t.Fatalf("incorrect application field %s", field)
				}
			}
			response["CertificateId"] = requestedCloudId
		case "DescribeCertificateDetail":
			pending, err := outputs.GetByWorkflowIdAndNodeId(ctx, workflowId, "apply-node")
			if err != nil {
				t.Fatal(err)
			}
			nodeConfig := pending.NodeConfig.AsBizApply()
			if pending.RunId != runId || tencentCloudPendingCertificate(&nodeConfig, pending) != requestedCloudId {
				t.Fatal("cloud order must be committed to SQLite before polling")
			}
			response["Status"] = 1
			if phase == "create" {
				response["Status"] = 0
			}
		case "CompleteCertificate":
			if phase != "create" {
				t.Fatal("unexpected validation request after issuance")
			}
			response["CertificateId"] = cloudId
			cancel()
		case "DownloadCertificate":
			if phase != "resume" && phase != "renew" {
				t.Fatal("unexpected certificate download before recovery")
			}
			response["ContentType"] = "application/zip"
			response["Content"] = base64.StdEncoding.EncodeToString(archive.Bytes())
		default:
			t.Fatalf("unexpected cloud action: %s", action)
		}
		data, err := json.Marshal(map[string]any{"Response": response})
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(data))}, nil
	})
	originalTransport, originalClient := http.DefaultTransport, common.DefaultHttpClient
	http.DefaultTransport = transport
	common.DefaultHttpClient = nil
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
		common.DefaultHttpClient = originalClient
	})

	executor := &bizApplyNodeExecutor{
		nodeExecutor:          nodeExecutor{logger: slog.New(slog.NewTextHandler(io.Discard, nil))},
		accessRepo:            repository.NewAccessRepository(),
		certificateRepo:       certificates,
		wfoutputRepo:          outputs,
		newTencentCloudClient: newTencentCloudIssuer,
	}
	execCtx := nativeTestContext(config, runId).SetContext(requestCtx)
	execCtx.WorkflowId = workflowId
	result, err := executor.Execute(execCtx)
	expectedActions := map[string][]string{
		"create":        {"ApplyCertificate", "DescribeCertificateDetail", "CompleteCertificate"},
		"resume":        {"DescribeCertificateDetail", "DownloadCertificate"},
		"recover-saved": {},
		"renew":         {"ApplyCertificate", "DescribeCertificateDetail", "DownloadCertificate"},
	}
	if !slices.Equal(actions, expectedActions[phase]) {
		t.Fatalf("unexpected SDK action sequence for %s: %v", phase, actions)
	}
	if phase == "create" {
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation after order persistence, got %v", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	certificateRunId := "444444444444444"
	if phase == "renew" {
		certificateRunId = runId
	}
	stored, err := certificates.GetByWorkflowRunIdAndNodeId(ctx, certificateRunId, "apply-node")
	if err != nil {
		t.Fatal(err)
	}
	if stored.CA != "tencentcloud" || stored.ACMEAccountUrl != "" || stored.ACMECertificateUrl != "" || stored.SubjectAltNames != "test.example.com" {
		t.Fatal("native certificate metadata did not round-trip through SQLite")
	}
	if phase != "recover-saved" && (stored.Certificate != certificatePEM || stored.PrivateKey != privateKeyPEM) {
		t.Fatal("certificate PEM did not round-trip through SQLite")
	}
	foundReference := false
	for _, output := range result.Outputs {
		if output.Name == "certificate" && output.Value == "certificate#"+stored.Id {
			foundReference = true
		}
	}
	if !foundReference {
		t.Fatal("database-backed certificate output is missing")
	}
	expectedCount := 1
	if phase == "renew" {
		expectedCount = 2
	}
	records, err := pb.FindAllRecords(domain.CollectionNameCertificate)
	if err != nil || len(records) != expectedCount {
		t.Fatalf("expected %d stored certificates after %s, got %d: %v", expectedCount, phase, len(records), err)
	}
}

func TestNativeTencentCloudWorkflowDeployment(t *testing.T) {
	phase := os.Getenv("CERTIMATE_TENCENTCLOUD_TEST_PHASE")
	if phase == "" {
		dataDir := t.TempDir()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, step := range []string{"issue", "reuse"} {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			command := exec.CommandContext(ctx, executable, "-test.run=^TestNativeTencentCloudWorkflowDeployment$", "-test.v")
			command.Env = append(os.Environ(), "CERTIMATE_TENCENTCLOUD_TEST_PHASE="+step, "CERTIMATE_TENCENTCLOUD_TEST_DIR="+dataDir)
			output, err := command.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("workflow deployment phase %s failed: %v\n%s", step, err, output)
			}
			t.Logf("workflow deployment phase %s passed in a separate process", step)
		}
		return
	}

	pb := nativeTestDatabase(t, os.Getenv("CERTIMATE_TENCENTCLOUD_TEST_DIR"))
	const workflowId, accessId, issuedRunId = "111111111111111", "222222222222222", "333333333333333"
	const cloudId = "workflow-cloud-order"
	config := nativeTestNodeConfig()
	config["tencentcloudAccessId"] = accessId
	ctx := context.Background()
	outputs := repository.NewWorkflowOutputRepository()
	certificates := repository.NewCertificateRepository()
	runId := issuedRunId
	if phase == "issue" {
		nativeTestRecord(t, pb, domain.CollectionNameWorkflow, workflowId, map[string]any{"name": "Native workflow deployment"})
		nativeTestRecord(t, pb, domain.CollectionNameAccess, accessId, map[string]any{
			"name": "Mock Tencent credential", "provider": "tencentcloud",
			"config": map[string]any{"secretId": "test-id", "secretKey": "test-secret", "projectId": 42},
		})
	} else {
		runId = "444444444444444"
		last, err := outputs.GetByWorkflowIdAndNodeId(ctx, workflowId, "apply-node")
		if err != nil {
			t.Fatal(err)
		}
		if !last.Succeeded {
			t.Fatal("successful application output did not survive restart")
		}
		config = last.NodeConfig
	}
	nativeTestRecord(t, pb, domain.CollectionNameWorkflowRun, runId, map[string]any{"workflowRef": workflowId, "status": "processing", "trigger": "manual"})

	certificatePEM, privateKeyPEM := nativeTestPEM(t)
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for name, content := range map[string]string{"Nginx/test.example.com_bundle.crt": certificatePEM, "Nginx/test.example.com.key": privateKeyPEM} {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(file, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	var actions []string
	transport := nativeTestTransport(func(request *http.Request) (*http.Response, error) {
		if phase != "issue" {
			t.Fatal("reusing a valid certificate must not make cloud requests")
		}
		if request.URL.Scheme != "https" || request.URL.Host != "ssl.tencentcloudapi.com" || !strings.HasPrefix(request.Header.Get("Authorization"), "TC3-HMAC-SHA256 Credential=test-id/") {
			t.Fatal("refusing unexpected or unsigned cloud request")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		action := request.Header.Get("X-TC-Action")
		actions = append(actions, action)
		if action != "ApplyCertificate" && body["CertificateId"] != cloudId {
			t.Fatal("incorrect cloud order ID")
		}
		response := map[string]any{"RequestId": "local-test-request"}
		switch action {
		case "ApplyCertificate":
			if len(actions) != 1 || body["DomainName"] != "test.example.com" || body["DvAuthMethod"] != "DNS_AUTO" || body["PackageType"] != "83" {
				t.Fatal("incorrect native workflow application")
			}
			response["CertificateId"] = cloudId
		case "DescribeCertificateDetail":
			response["Status"] = 1
		case "DownloadCertificate":
			response["ContentType"] = "application/zip"
			response["Content"] = base64.StdEncoding.EncodeToString(archive.Bytes())
		default:
			t.Fatalf("unexpected cloud action: %s", action)
		}
		data, err := json.Marshal(map[string]any{"Response": response})
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(data))}, nil
	})
	originalTransport, originalClient := http.DefaultTransport, common.DefaultHttpClient
	http.DefaultTransport = transport
	common.DefaultHttpClient = nil
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
		common.DefaultHttpClient = originalClient
	})

	deployDir := t.TempDir()
	certificatePath, keyPath := filepath.Join(deployDir, "certificate.pem"), filepath.Join(deployDir, "private.key")
	graph := &Graph{Nodes: []*Node{
		{Id: "apply-node", Type: NodeTypeBizApply, Data: domain.WorkflowNodeData{Config: config}},
		{Id: "deploy-node", Type: NodeTypeBizDeploy, Data: domain.WorkflowNodeData{Config: domain.WorkflowNodeConfig{
			"certificateOutputNodeId": "apply-node", "provider": "local",
			"providerConfig": map[string]any{"fileFormat": "PEM", "filePathForCrt": certificatePath, "filePathForKey": keyPath},
		}}},
	}}
	engine := NewWorkflowEngine()
	completed := make(map[string]*NodeExecutionResult)
	engine.OnError(func(_ context.Context, err error) error {
		t.Errorf("workflow failed: %v", err)
		return nil
	})
	engine.OnNodeEnd(func(_ context.Context, node *Node, result *NodeExecutionResult) error {
		completed[node.Id] = result
		return nil
	})
	if err := engine.Invoke(ctx, WorkflowExecution{WorkflowId: workflowId, RunId: runId, Graph: graph}); err != nil {
		t.Fatal(err)
	}
	if len(completed) != 2 || completed["apply-node"] == nil || completed["deploy-node"] == nil {
		t.Fatal("application and deployment must both complete")
	}
	foundSkipped := false
	for _, variable := range completed["apply-node"].Variables {
		if variable.Scope == "apply-node" && variable.Key == stateVarKeyNodeSkipped {
			foundSkipped = variable.Value == (phase == "reuse")
		}
	}
	if !foundSkipped {
		t.Fatal("incorrect renewal skip state")
	}
	if phase == "issue" && !slices.Equal(actions, []string{"ApplyCertificate", "DescribeCertificateDetail", "DownloadCertificate"}) {
		t.Fatalf("incorrect cloud action sequence: %v", actions)
	}
	stored, err := certificates.GetByWorkflowRunIdAndNodeId(ctx, issuedRunId, "apply-node")
	if err != nil {
		t.Fatal(err)
	}
	if stored.CA != "tencentcloud" || stored.ACMEAccountUrl != "" || stored.ACMECertificateUrl != "" {
		t.Fatal("native certificate metadata was lost")
	}
	if phase == "issue" && (stored.Certificate != certificatePEM || stored.PrivateKey != privateKeyPEM) {
		t.Fatal("issued certificate or private key changed before storage")
	}
	for filePath, expected := range map[string]string{certificatePath: stored.Certificate, keyPath: stored.PrivateKey} {
		contents, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if string(contents) != expected {
			t.Fatal("deployed file differs from the stored native certificate material")
		}
	}
	for nodeId, expectedRunId := range map[string]string{"apply-node": issuedRunId, "deploy-node": runId} {
		output, err := outputs.GetByWorkflowIdAndNodeId(ctx, workflowId, nodeId)
		if err != nil {
			t.Fatal(err)
		}
		if !output.Succeeded || output.RunId != expectedRunId {
			t.Fatalf("successful output for %s was not persisted", nodeId)
		}
		if nodeId == "apply-node" {
			values := make(map[string]string)
			for _, entry := range output.Outputs {
				values[entry.Name] = entry.Value
			}
			if values["certificate"] != "certificate#"+stored.Id || values["tencentcloudCertificateId"] != cloudId || values[tencentCloudPendingCertificateKey] != "" {
				t.Fatal("engine did not persist the final certificate reference and cloud order ID")
			}
		}
	}
	records, err := pb.FindAllRecords(domain.CollectionNameCertificate)
	if err != nil || len(records) != 1 {
		t.Fatalf("expected one certificate after %s, got %d: %v", phase, len(records), err)
	}
}

func TestNativeTencentCloudWorkflowSaveAndOutputs(t *testing.T) {
	cert, key := nativeTestPEM(t)
	executor, outputs, certificates := nativeTestExecutor(t, func(ctx context.Context, request certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
		if request.CertificateId != "" || request.Domains[0] != "test.example.com" || request.KeyAlgorithm != "EC256" {
			t.Fatal("incorrect native request")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("native issuance must have a deadline")
		}
		if err := created("cloud-certificate"); err != nil {
			return nil, err
		}
		return &certapply.TencentCloudResult{CertificateId: "cloud-certificate", Certificate: cert, PrivateKey: key}, nil
	})
	result, err := executor.Execute(nativeTestContext(nativeTestNodeConfig(), "run-1"))
	if err != nil {
		t.Fatal(err)
	}
	if outputs.saves != 1 || outputs.last.Succeeded || outputs.last.Outputs[0].Value != "cloud-certificate" {
		t.Fatal("pending order was not saved before obtaining PEM")
	}
	saved := certificates.certificate
	if saved == nil || saved.CA != "tencentcloud" || saved.Source != domain.CertificateSourceTypeRequest || saved.Certificate != cert || saved.PrivateKey != key {
		t.Fatalf("incorrect stored certificate: %v", saved != nil)
	}
	if saved.ACMEAccountUrl != "" || saved.ACMECertificateUrl != "" || saved.SubjectAltNames != "test.example.com" || saved.KeyAlgorithm != domain.CertificateKeyAlgorithmTypeEC256 {
		t.Fatal("native certificate was not parsed or incorrectly used ACME metadata")
	}
	if saved.WorkflowId != "workflow" || saved.WorkflowRunId != "run-1" || saved.WorkflowNodeId != "apply-node" {
		t.Fatal("incorrect workflow certificate references")
	}
	foundCertificate, foundCloudId := false, false
	for _, output := range result.Outputs {
		if output.Name == "certificate" && output.Value == "certificate#saved-certificate" && output.Persistent {
			foundCertificate = true
		}
		if output.Name == "tencentcloudCertificateId" && output.Value == "cloud-certificate" && output.Persistent {
			foundCloudId = true
		}
	}
	if !foundCertificate || !foundCloudId {
		t.Fatal("deployment certificate reference or cloud ID missing")
	}
	if len(result.Variables) == 0 {
		t.Fatal("certificate workflow variables missing")
	}
}

func TestNativeTencentCloudWorkflowResumesAfterTimeout(t *testing.T) {
	attempts := 0
	executor, outputs, certificates := nativeTestExecutor(t, func(_ context.Context, request certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
		attempts++
		if attempts == 1 {
			if request.CertificateId != "" {
				t.Fatal("unexpected initial order")
			}
			if err := created("pending-certificate"); err != nil {
				return nil, err
			}
		} else if request.CertificateId != "pending-certificate" {
			t.Fatal("retry did not resume the pending order")
		}
		return nil, context.DeadlineExceeded
	})
	for _, runId := range []string{"run-1", "run-2"} {
		_, err := executor.Execute(nativeTestContext(nativeTestNodeConfig(), runId))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected timeout, got %v", err)
		}
	}
	if attempts != 2 || outputs.saves != 2 || certificates.certificate != nil || outputs.last.RunId != "run-2" {
		t.Fatal("incorrect retry persistence")
	}
}

func TestNativeTencentCloudWorkflowStatusRecovery(t *testing.T) {
	for _, status := range []uint64{2, 3, 5, 7, 8, 10, 12, 14, 999} {
		terminal := status == 2 || status == 3 || status == 7 || status == 10 || status == 14
		attempts := 0
		executor, outputs, certificates := nativeTestExecutor(t, func(_ context.Context, request certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
			attempts++
			if attempts == 2 && (request.CertificateId == "") != terminal {
				t.Fatalf("incorrect recovery for status %d: %q", status, request.CertificateId)
			}
			if request.CertificateId == "" {
				if err := created("status-order"); err != nil {
					return nil, err
				}
			}
			return nil, &certapply.TencentCloudStatusError{CertificateId: "status-order", Status: status}
		})
		for _, runId := range []string{"run-1", "run-2"} {
			_, err := executor.Execute(nativeTestContext(nativeTestNodeConfig(), runId))
			if err == nil || (len(outputs.last.Outputs) == 0) != terminal || certificates.certificate != nil {
				t.Fatalf("incorrect order persistence for status %d", status)
			}
		}
	}
}

func TestNativeTencentCloudWorkflowSaveFailureRetainsOrder(t *testing.T) {
	cert, key := nativeTestPEM(t)
	executor, outputs, certificates := nativeTestExecutor(t, func(_ context.Context, _ certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
		if err := created("issued-order"); err != nil {
			return nil, err
		}
		return &certapply.TencentCloudResult{CertificateId: "issued-order", Certificate: cert, PrivateKey: key}, nil
	})
	certificates.saveErr = errors.New("database unavailable")
	_, err := executor.Execute(nativeTestContext(nativeTestNodeConfig(), "run-1"))
	if err == nil || outputs.last.Outputs[0].Value != "issued-order" {
		t.Fatal("issued order was lost on database failure")
	}
}

func TestNativeTencentCloudWorkflowRenewsAfterSavedRecovery(t *testing.T) {
	for _, reason := range []string{"expiring", "revoked"} {
		t.Run(reason, func(t *testing.T) {
			cert, key := nativeTestPEM(t)
			executor, outputs, certificates := nativeTestExecutor(t, func(_ context.Context, request certapply.TencentCloudRequest, created func(string) error) (*certapply.TencentCloudResult, error) {
				if request.CertificateId != "" {
					t.Fatalf("renewal must not resume an already saved certificate order: %s", request.CertificateId)
				}
				if err := created("renewed-order"); err != nil {
					return nil, err
				}
				return &certapply.TencentCloudResult{CertificateId: "renewed-order", Certificate: cert, PrivateKey: key}, nil
			})
			config := nativeTestNodeConfig()
			outputs.last = &domain.WorkflowOutput{
				WorkflowId: "workflow", RunId: "old-run", NodeId: "apply-node", NodeConfig: maps.Clone(config),
				Outputs: []*domain.WorkflowOutputEntry{{Name: tencentCloudPendingCertificateKey, Value: "fulfilled-order"}},
			}
			certificates.certificate = &domain.Certificate{
				WorkflowId: "workflow", WorkflowRunId: "old-run", WorkflowNodeId: "apply-node", CA: "tencentcloud",
				ValidityNotAfter: time.Now().Add(20 * 24 * time.Hour),
			}
			if reason == "revoked" {
				certificates.certificate.IsRevoked = true
				certificates.certificate.ValidityNotAfter = time.Now().Add(80 * 24 * time.Hour)
			}
			if _, err := executor.Execute(nativeTestContext(config, "new-run")); err != nil {
				t.Fatal(err)
			}
			if certificates.certificate.WorkflowRunId != "new-run" || outputs.last.Outputs[0].Value != "renewed-order" {
				t.Fatal("renewal did not save a new order and certificate")
			}
		})
	}
}

func TestNativeTencentCloudWorkflowValidation(t *testing.T) {
	for field, value := range map[string]any{"source": "unknown", "domains": "*.example.com", "keySource": "reuse", "keyAlgorithm": "RSA4096", "tencentcloudAccessId": "", "challengeType": "http-01", "dnsPropagationWait": -1} {
		config := nativeTestNodeConfig()
		config[field] = value
		executor, _, _ := nativeTestExecutor(t, func(context.Context, certapply.TencentCloudRequest, func(string) error) (*certapply.TencentCloudResult, error) {
			t.Fatal("invalid config reached issuer")
			return nil, nil
		})
		if _, err := executor.Execute(nativeTestContext(config, "run-1")); err == nil {
			t.Fatalf("accepted invalid field %s", field)
		}
	}
	executor, _, _ := nativeTestExecutor(t, nil)
	executor.accessRepo = &nativeTestAccessRepo{provider: "aliyun"}
	if _, err := executor.Execute(nativeTestContext(nativeTestNodeConfig(), "run-1")); err == nil {
		t.Fatal("accepted wrong credential provider")
	}
}

func TestNativeTencentCloudPendingConfigurationMatch(t *testing.T) {
	config := nativeTestNodeConfig()
	last := &domain.WorkflowOutput{NodeConfig: config, Outputs: []*domain.WorkflowOutputEntry{{Name: tencentCloudPendingCertificateKey, Value: "pending-id"}}}
	current := config.AsBizApply()
	if got := tencentCloudPendingCertificate(&current, last); got != "pending-id" {
		t.Fatalf("got %q", got)
	}
	for field, value := range map[string]any{"source": "acme", "domains": "other.example.com", "ipaddrs": "127.0.0.1", "keyAlgorithm": "RSA2048", "tencentcloudAccessId": "different", "contactEmail": "other@example.com"} {
		changed := maps.Clone(config)
		changed[field] = value
		last.NodeConfig = changed
		if got := tencentCloudPendingCertificate(&current, last); got != "" {
			t.Fatalf("reused order after changing %s", field)
		}
	}
	last.NodeConfig, last.Succeeded = config, true
	if got := tencentCloudPendingCertificate(&current, last); got != "" {
		t.Fatal("resumed completed order")
	}
}

func TestNativeTencentCloudRenewalSkipAndSourceChange(t *testing.T) {
	config := nativeTestNodeConfig()
	last := &domain.WorkflowOutput{Succeeded: true, NodeConfig: maps.Clone(config)}
	certificate := &domain.Certificate{ValidityNotAfter: time.Now().Add(80 * 24 * time.Hour)}
	executor := &bizApplyNodeExecutor{}
	if skip, _ := executor.checkCanSkip(nativeTestContext(config, "run"), last, certificate); !skip {
		t.Fatal("valid certificate should skip renewal")
	}
	certificate.ValidityNotAfter = time.Now().Add(20 * 24 * time.Hour)
	if skip, _ := executor.checkCanSkip(nativeTestContext(config, "run"), last, certificate); skip {
		t.Fatal("expiring certificate should renew")
	}
	certificate.ValidityNotAfter = time.Now().Add(80 * 24 * time.Hour)
	for field, value := range map[string]any{"source": "acme", "tencentcloudAccessId": "other-access", "domains": "other.example.com", "keyAlgorithm": "RSA2048"} {
		changed := maps.Clone(config)
		changed[field] = value
		if skip, _ := executor.checkCanSkip(nativeTestContext(changed, "run"), last, certificate); skip {
			t.Fatalf("should not skip when %s changes", field)
		}
	}
	legacy := domain.WorkflowNodeConfig{}.AsBizApply()
	if legacy.Source != "acme" {
		t.Fatalf("legacy source should default to acme, got %q", legacy.Source)
	}
}
