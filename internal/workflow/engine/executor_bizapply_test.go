package engine

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/certimate-go/certimate/internal/domain"
)

type mockCertificateRepo struct {
	byId              map[string]*domain.Certificate
	byWorkflowAndNode map[string]*domain.Certificate
	byRunAndNode      map[string]*domain.Certificate
}

func (m *mockCertificateRepo) GetById(ctx context.Context, id string) (*domain.Certificate, error) {
	if cert, ok := m.byId[id]; ok {
		return cert, nil
	}
	return nil, domain.ErrRecordNotFound
}

func (m *mockCertificateRepo) GetByWorkflowIdAndNodeId(ctx context.Context, workflowId string, workflowNodeId string) (*domain.Certificate, error) {
	if cert, ok := m.byWorkflowAndNode[workflowId+"|"+workflowNodeId]; ok {
		return cert, nil
	}
	return nil, domain.ErrRecordNotFound
}

func (m *mockCertificateRepo) GetByWorkflowRunIdAndNodeId(ctx context.Context, workflowRunId string, workflowNodeId string) (*domain.Certificate, error) {
	if cert, ok := m.byRunAndNode[workflowRunId+"|"+workflowNodeId]; ok {
		return cert, nil
	}
	return nil, domain.ErrRecordNotFound
}

func (m *mockCertificateRepo) Save(ctx context.Context, certificate *domain.Certificate) (*domain.Certificate, error) {
	if m.byId == nil {
		m.byId = make(map[string]*domain.Certificate)
	}
	m.byId[certificate.Id] = certificate
	return certificate, nil
}

type mockWorkflowOutputRepo struct {
	byWorkflowAndNode map[string]*domain.WorkflowOutput
}

func (m *mockWorkflowOutputRepo) GetByWorkflowIdAndNodeId(ctx context.Context, workflowId string, workflowNodeId string) (*domain.WorkflowOutput, error) {
	if out, ok := m.byWorkflowAndNode[workflowId+"|"+workflowNodeId]; ok {
		return out, nil
	}
	return nil, domain.ErrRecordNotFound
}

func (m *mockWorkflowOutputRepo) Save(ctx context.Context, workflowOutput *domain.WorkflowOutput) (*domain.WorkflowOutput, error) {
	if m.byWorkflowAndNode == nil {
		m.byWorkflowAndNode = make(map[string]*domain.WorkflowOutput)
	}
	m.byWorkflowAndNode[workflowOutput.WorkflowId+"|"+workflowOutput.NodeId] = workflowOutput
	return workflowOutput, nil
}

func newTestApplyExecutor(certRepo certificateRepository, outputRepo workflowOutputRepository) *bizApplyNodeExecutor {
	ne := &bizApplyNodeExecutor{
		accessRepo:      nil,
		certificateRepo: certRepo,
		wfoutputRepo:    outputRepo,
	}
	ne.SetLogger(slog.Default())
	return ne
}

func newTestApplyExecCtx(workflowId, runId, nodeId string, cfg domain.WorkflowNodeConfig) *NodeExecutionContext {
	node := &Node{
		Id:   nodeId,
		Type: NodeTypeBizApply,
		Data: domain.WorkflowNodeData{
			Name:   "apply",
			Config: cfg,
		},
	}
	return (&NodeExecutionContext{}).
		SetExecutingWorkflow(workflowId, runId, nil).
		SetExecutingNode(node).
		SetContext(context.Background())
}

func TestBizApply_resolveLastCertificate_fallbackWhenOutputMissing(t *testing.T) {
	// Reproduces #1413: workflow_output cascade-deleted with run history cleanup,
	// but the certificate still exists and should keep skip-before-expiry working.
	cert := &domain.Certificate{
		Meta:             domain.Meta{Id: "cert-alive"},
		WorkflowId:       "wf1",
		WorkflowRunId:    "run-old",
		WorkflowNodeId:   "node-apply",
		ValidityNotAfter: time.Now().Add(60 * 24 * time.Hour),
	}
	certRepo := &mockCertificateRepo{
		byWorkflowAndNode: map[string]*domain.Certificate{
			"wf1|node-apply": cert,
		},
	}
	outputRepo := &mockWorkflowOutputRepo{} // no outputs

	ne := newTestApplyExecutor(certRepo, outputRepo)
	execCtx := newTestApplyExecCtx("wf1", "run-new", "node-apply", domain.WorkflowNodeConfig{
		"domains":              []any{"example.com"},
		"skipBeforeExpiryDays": 30,
	})

	lastOutput, lastCertificate, err := ne.getLastOutputArtifacts(execCtx)
	require.NoError(t, err)
	assert.Nil(t, lastOutput)
	require.NotNil(t, lastCertificate)
	assert.Equal(t, "cert-alive", lastCertificate.Id)

	skippable, reason := ne.checkCanSkip(execCtx, lastOutput, lastCertificate)
	assert.True(t, skippable, "expected skip when remaining validity is well above threshold; reason=%q", reason)
	assert.Contains(t, reason, "renewal will be performed when the remaining validity is less than")
}

func TestBizApply_resolveLastCertificate_prefersOutputRefOverRunId(t *testing.T) {
	// After a skip-only run, workflow_output.runRef is the skip run, while the
	// certificate still points at the original apply run. Resolve via certificate#id ref.
	cert := &domain.Certificate{
		Meta:             domain.Meta{Id: "cert-from-ref"},
		WorkflowId:       "wf1",
		WorkflowRunId:    "run-apply",
		WorkflowNodeId:   "node-apply",
		ValidityNotAfter: time.Now().Add(60 * 24 * time.Hour),
	}
	wrongCert := &domain.Certificate{
		Meta: domain.Meta{Id: "should-not-use"},
	}
	certRepo := &mockCertificateRepo{
		byId: map[string]*domain.Certificate{
			"cert-from-ref": cert,
		},
		byRunAndNode: map[string]*domain.Certificate{
			// Would be used if we only looked up by lastOutput.RunId
			"run-skip|node-apply": wrongCert,
		},
	}
	outputRepo := &mockWorkflowOutputRepo{
		byWorkflowAndNode: map[string]*domain.WorkflowOutput{
			"wf1|node-apply": {
				WorkflowId: "wf1",
				RunId:      "run-skip",
				NodeId:     "node-apply",
				Succeeded:  true,
				Outputs: []*domain.WorkflowOutputEntry{
					{
						Type:      stateIOTypeRef,
						Name:      "certificate",
						Value:     domain.CollectionNameCertificate + "#cert-from-ref",
						ValueType: stateValTypeString,
					},
				},
			},
		},
	}

	ne := newTestApplyExecutor(certRepo, outputRepo)
	execCtx := newTestApplyExecCtx("wf1", "run-next", "node-apply", domain.WorkflowNodeConfig{
		"domains":              []any{"example.com"},
		"skipBeforeExpiryDays": 30,
	})

	_, lastCertificate, err := ne.getLastOutputArtifacts(execCtx)
	require.NoError(t, err)
	require.NotNil(t, lastCertificate)
	assert.Equal(t, "cert-from-ref", lastCertificate.Id)
}

func TestBizApply_resolveLastCertificate_legacyRunIdLookup(t *testing.T) {
	cert := &domain.Certificate{
		Meta:             domain.Meta{Id: "cert-by-run"},
		WorkflowId:       "wf1",
		WorkflowRunId:    "run-apply",
		WorkflowNodeId:   "node-apply",
		ValidityNotAfter: time.Now().Add(60 * 24 * time.Hour),
	}
	certRepo := &mockCertificateRepo{
		byRunAndNode: map[string]*domain.Certificate{
			"run-apply|node-apply": cert,
		},
	}
	outputRepo := &mockWorkflowOutputRepo{
		byWorkflowAndNode: map[string]*domain.WorkflowOutput{
			"wf1|node-apply": {
				WorkflowId: "wf1",
				RunId:      "run-apply",
				NodeId:     "node-apply",
				Succeeded:  true,
				// No certificate ref in outputs — legacy path uses runId
			},
		},
	}

	ne := newTestApplyExecutor(certRepo, outputRepo)
	execCtx := newTestApplyExecCtx("wf1", "run-next", "node-apply", domain.WorkflowNodeConfig{
		"skipBeforeExpiryDays": 30,
	})

	_, lastCertificate, err := ne.getLastOutputArtifacts(execCtx)
	require.NoError(t, err)
	require.NotNil(t, lastCertificate)
	assert.Equal(t, "cert-by-run", lastCertificate.Id)
}

func TestBizApply_Execute_skipPersistsCertificateOutput(t *testing.T) {
	cert := &domain.Certificate{
		Meta:             domain.Meta{Id: "cert-skip"},
		WorkflowId:       "wf1",
		WorkflowRunId:    "run-apply",
		WorkflowNodeId:   "node-apply",
		SubjectAltNames:  "example.com",
		ValidityNotAfter: time.Now().Add(60 * 24 * time.Hour),
	}
	certRepo := &mockCertificateRepo{
		byWorkflowAndNode: map[string]*domain.Certificate{
			"wf1|node-apply": cert,
		},
		byId: map[string]*domain.Certificate{
			"cert-skip": cert,
		},
	}
	outputRepo := &mockWorkflowOutputRepo{}

	ne := newTestApplyExecutor(certRepo, outputRepo)
	execCtx := newTestApplyExecCtx("wf1", "run-skip", "node-apply", domain.WorkflowNodeConfig{
		"domains":              []any{"example.com"},
		"skipBeforeExpiryDays": 30,
	})

	execRes, err := ne.Execute(execCtx)
	require.NoError(t, err)
	require.NotNil(t, execRes)

	var skipped bool
	for _, v := range execRes.Variables {
		if v.Scope == "node-apply" && v.Key == stateVarKeyNodeSkipped {
			skipped = v.Value.(bool)
		}
	}
	assert.True(t, skipped, "node should be skipped")

	require.NotEmpty(t, execRes.Outputs)
	foundPersistent := false
	for _, out := range execRes.Outputs {
		if out.Name == "certificate" {
			assert.True(t, out.Persistent, "skipped apply must persist certificate output to survive history cleanup")
			assert.Equal(t, domain.CollectionNameCertificate+"#cert-skip", out.Value)
			foundPersistent = true
		}
	}
	assert.True(t, foundPersistent, "expected a certificate output on skip")
}

func TestBizApply_checkCanSkip_reapplyWhenNearExpiry(t *testing.T) {
	cert := &domain.Certificate{
		Meta:             domain.Meta{Id: "cert-near"},
		ValidityNotAfter: time.Now().Add(10 * 24 * time.Hour),
	}
	ne := newTestApplyExecutor(&mockCertificateRepo{}, &mockWorkflowOutputRepo{})
	execCtx := newTestApplyExecCtx("wf1", "run1", "node-apply", domain.WorkflowNodeConfig{
		"skipBeforeExpiryDays": 30,
	})

	skippable, reason := ne.checkCanSkip(execCtx, nil, cert)
	assert.False(t, skippable)
	assert.Contains(t, reason, "renewal window period has been reached")
}
