package workflowtemplate

import (
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/providerschema"
)

func buildTestGraph() *domain.WorkflowGraph {
	return &domain.WorkflowGraph{
		Nodes: []*domain.WorkflowNode{
			{Id: "start1", Type: domain.WorkflowNodeTypeStart, Data: domain.WorkflowNodeData{Name: "start"}},
			{
				Id:   "apply1",
				Type: domain.WorkflowNodeTypeBizApply,
				Data: domain.WorkflowNodeData{Name: "apply", Config: domain.WorkflowNodeConfig{"identifier": "domain", "domains": "old.example.com"}},
			},
			{
				Id:   "deploy1",
				Type: domain.WorkflowNodeTypeBizDeploy,
				Data: domain.WorkflowNodeData{Name: "deploy", Config: domain.WorkflowNodeConfig{"provider": "ssh", "providerAccessId": "acc1"}},
			},
			{Id: "end1", Type: domain.WorkflowNodeTypeEnd, Data: domain.WorkflowNodeData{Name: "end"}},
		},
	}
}

func TestWorkflowTemplateParams_Verify(t *testing.T) {
	graph := buildTestGraph()

	valid := domain.WorkflowTemplateParams{
		{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "域名", Order: 0, ApplyNodeIds: []string{"apply1"}},
		{Id: "p2", Kind: domain.WorkflowTemplateParamKindDeployment, Title: "部署", Order: 1, DeployNodeId: "deploy1"},
	}
	if err := valid.Verify(graph); err != nil {
		t.Errorf("expected valid params to pass, got %v", err)
	}

	testcases := []struct {
		name   string
		params domain.WorkflowTemplateParams
	}{
		{"blank param id", domain.WorkflowTemplateParams{{Id: "", Kind: domain.WorkflowTemplateParamKindDomains, Title: "t", ApplyNodeIds: []string{"apply1"}}}},
		{"blank title", domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "", ApplyNodeIds: []string{"apply1"}}}},
		{"duplicated param id", domain.WorkflowTemplateParams{
			{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "t", ApplyNodeIds: []string{"apply1"}},
			{Id: "p1", Kind: domain.WorkflowTemplateParamKindDeployment, Title: "t", DeployNodeId: "deploy1"},
		}},
		{"unknown kind", domain.WorkflowTemplateParams{{Id: "p1", Kind: "bogus", Title: "t"}}},
		{"domains bound to unknown node", domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "t", ApplyNodeIds: []string{"ghost"}}}},
		{"domains bound to end node", domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "t", ApplyNodeIds: []string{"end1"}}}},
		{"domains without bindings", domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, Title: "t"}}},
		{"deployment bound to apply node", domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDeployment, Title: "t", DeployNodeId: "apply1"}}},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.params.Verify(graph); err == nil {
				t.Errorf("expected Verify to fail")
			}
		})
	}

	t.Run("deployment without provider", func(t *testing.T) {
		g := buildTestGraph()
		for _, node := range g.Nodes {
			if node.Id == "deploy1" {
				node.Data.Config = domain.WorkflowNodeConfig{}
			}
		}
		params := domain.WorkflowTemplateParams{{Id: "p1", Kind: domain.WorkflowTemplateParamKindDeployment, Title: "t", DeployNodeId: "deploy1"}}
		if err := params.Verify(g); err == nil {
			t.Errorf("expected Verify to fail")
		}
	})
}

func TestIsValidDomainName(t *testing.T) {
	testcases := []struct {
		value         string
		allowWildcard bool
		expected      bool
	}{
		{"example.com", false, true},
		{"a-b.example.com", false, true},
		{"*.example.com", true, true},
		{"*.example.com", false, false},
		{"https://example.com", false, false},
		{"-abc.example.com", false, false},
		{"abc-.example.com", false, false},
		{"ex ample.com", false, false},
		{"", false, false},
		{"a..b.com", false, false},
	}
	for _, tc := range testcases {
		if got := isValidDomainName(tc.value, tc.allowWildcard); got != tc.expected {
			t.Errorf("isValidDomainName(%q, %v) = %v, expected %v", tc.value, tc.allowWildcard, got, tc.expected)
		}
	}
}

func buildTestEnvelope() *providerschema.Envelope {
	return &providerschema.Envelope{
		SchemaVersion: "form/v1",
		Provider:      "aliyun-cdn",
		Category:      providerschema.CategoryDeploy,
		Schema: providerschema.EnvelopeSchema{
			Columns: []providerschema.Column{
				{Name: "region", ValueType: providerschema.ValueTypeText},
				{
					Name:      "domainMatchPattern",
					ValueType: providerschema.ValueTypeSelect,
					Required:  true,
					Default:   "exact",
					Secret:    false,
					Options:   []providerschema.Option{{Value: "exact"}, {Value: "wildcard"}, {Value: "certsan"}},
				},
				{
					Name:         "domain",
					ValueType:    providerschema.ValueTypeText,
					RequiredWhen: []providerschema.Condition{{Field: "domainMatchPattern", Op: providerschema.OpIn, Values: []string{"exact", "wildcard"}}},
					ValidateWhen: []providerschema.ValidationRule{
						{
							Condition: providerschema.Condition{Field: "domainMatchPattern", Op: providerschema.OpIn, Values: []string{"exact", "wildcard"}},
							Name:      providerschema.ValidatorDomain,
							Params:    map[string]any{"allowWildcard": true},
						},
					},
				},
				{
					Name:        "accelerateDomain",
					ValueType:   providerschema.ValueTypeText,
					Required:    true,
					VisibleWhen: []providerschema.Condition{{Field: "mode", Op: providerschema.OpEquals, Values: []string{"accelerated"}}},
				},
				{Name: "secretToken", ValueType: providerschema.ValueTypeText, Secret: true},
			},
		},
	}
}

func TestValidateProviderConfig(t *testing.T) {
	envelope := buildTestEnvelope()

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "wildcard", "domain": "*.ok.example.com"}); err != nil {
		t.Errorf("expected valid config to pass, got %v", err)
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "wildcard", "domain": "https://bad"}); err == nil {
		t.Errorf("expected invalid domain to fail")
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "exact"}); err == nil {
		t.Errorf("expected missing required conditional domain to fail")
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "bogus", "domain": "x.com"}); err == nil {
		t.Errorf("expected disallowed select option to fail")
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "certsan"}); err != nil {
		t.Errorf("expected certsan branch without domain to pass, got %v", err)
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "exact", "mode": "plain"}); err == nil {
		t.Errorf("expected hidden required column to be skipped only when invisible; this config should fail on missing domain")
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "exact", "domain": "x.com", "mode": "plain", "accelerateDomain": ""}); err != nil {
		t.Errorf("expected invisible required column to be skipped, got %v", err)
	}

	if err := validateProviderConfig(envelope, map[string]any{"region": "cn", "domainMatchPattern": "exact", "domain": "x.com", "secretToken": "should-not-be-required-but-may-exist"}); err != nil {
		t.Errorf("expected non-required secret field to pass, got %v", err)
	}
}

func TestApplyDomainsParam_ReportsParamErrors(t *testing.T) {
	svc := &WorkflowTemplateService{}
	graph := buildTestGraph()
	param := &domain.WorkflowTemplateParam{Id: "p1", Kind: domain.WorkflowTemplateParamKindDomains, ApplyNodeIds: []string{"apply1"}}
	paramErrs := &dtos.WorkflowTemplateParamErrors{}

	if err := svc.applyDomainsParam(graph, param, dtos.WorkflowTemplateParamValue{Domains: []string{" "}}, paramErrs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paramErrs.Items) != 1 || paramErrs.Items[0].Code != ParamErrCodeRequired {
		t.Errorf("expected required error, got %+v", paramErrs.Items)
	}

	paramErrs = &dtos.WorkflowTemplateParamErrors{}
	if err := svc.applyDomainsParam(graph, param, dtos.WorkflowTemplateParamValue{Domains: []string{"*.ok.com"}}, paramErrs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paramErrs.Items) != 0 {
		t.Errorf("expected no errors, got %+v", paramErrs.Items)
	}
	if node, ok := graph.GetNodeById("apply1"); !ok || node.Data.Config["domains"] != "*.ok.com" {
		t.Errorf("expected domains written into apply node, got %+v", node.Data.Config["domains"])
	}
}

func TestProviderNeedsAccess(t *testing.T) {
	if providerNeedsAccess("ssh") != true {
		t.Errorf("expected ssh to need access")
	}
	if providerNeedsAccess("aliyun-cdn") != true {
		t.Errorf("expected aliyun-cdn to need access")
	}
	if providerNeedsAccess("local") != false {
		t.Errorf("expected local to not need access")
	}
}

func TestDeriveAccessProviderType(t *testing.T) {
	testcases := map[string]string{
		"ssh":              "ssh",
		"aliyun-cdn":       "aliyun",
		"tencentcloud-clb": "tencentcloud",
	}
	for provider, expected := range testcases {
		if got := deriveAccessProviderType(provider); got != expected {
			t.Errorf("deriveAccessProviderType(%q) = %q, expected %q", provider, got, expected)
		}
	}
}
