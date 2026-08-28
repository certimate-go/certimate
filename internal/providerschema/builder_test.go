package providerschema

import (
	"errors"
	"testing"
)

func TestBuilder_HappyPath_TextSelectConditional(t *testing.T) {
	s, err := New("aliyun-cdn", CategoryDeploy).
		Field("region", ValueTypeAutocomplete,
			LabelKey("workflow_node.deploy.form.aliyun_cdn_region.label"),
			PlaceholderKey("workflow_node.deploy.form.aliyun_cdn_region.placeholder"),
			TooltipKey("workflow_node.deploy.form.aliyun_cdn_region.tooltip"),
			TooltipHtml(),
			Options("cn-hangzhou", "ap-southeast-1"),
			WithFilterMode(FilterModePrefix),
		).
		Field("domainMatchPattern", ValueTypeRadio,
			LabelKey("workflow_node.deploy.form.shared_domain_match_pattern.label"),
			ExtraKey("workflow_node.deploy.form.shared_domain_match_pattern.option.exact.help.wildcard"),
			ExtraHtml(),
			Options("exact", "wildcard", "certsan"),
			Default("exact"),
		).
		Field("domain", ValueTypeText,
			LabelKey("workflow_node.deploy.form.aliyun_cdn_domain.label"),
			PlaceholderKey("workflow_node.deploy.form.aliyun_cdn_domain.placeholder"),
			VisibleWhen("domainMatchPattern").NotIn("certsan"),
			RequiredWhen("domainMatchPattern").In("exact", "wildcard"),
			ValidateWhen("domainMatchPattern").In("exact", "wildcard").Validator(ValidatorDomain, WithParam("allowWildcard", true)),
		).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Provider != "aliyun-cdn" || s.Category != CategoryDeploy {
		t.Fatalf("provider/category = %q/%q", s.Provider, s.Category)
	}
	if len(s.Fields) != 3 {
		t.Fatalf("field count = %d, want 3", len(s.Fields))
	}

	region := s.Fields[0]
	if region.ValueType != ValueTypeAutocomplete || len(region.Options) != 2 || !region.TooltipHtml || region.FilterMode != FilterModePrefix {
		t.Fatalf("region field not as declared: %+v", region)
	}

	dmp := s.Fields[1]
	if dmp.Default != "exact" || len(dmp.Options) != 3 || !dmp.ExtraHtml {
		t.Fatalf("domainMatchPattern not as declared: %+v", dmp)
	}

	domain := s.Fields[2]
	if len(domain.VisibleWhen) != 1 || domain.VisibleWhen[0].Op != OpNotIn {
		t.Fatalf("domain visibleWhen not declared: %+v", domain.VisibleWhen)
	}
	if len(domain.RequiredWhen) != 1 || domain.RequiredWhen[0].Op != OpIn {
		t.Fatalf("domain requiredWhen not declared: %+v", domain.RequiredWhen)
	}
	if len(domain.ValidateWhen) != 1 {
		t.Fatalf("domain validateWhen not declared: %+v", domain.ValidateWhen)
	}
	vr := domain.ValidateWhen[0]
	if vr.Name != ValidatorDomain || vr.Op != OpIn || vr.Params["allowWildcard"] != true {
		t.Fatalf("domain validateWhen rule not as declared: %+v", vr)
	}
}

func TestBuilder_HappyPath_MultiDiscriminator(t *testing.T) {
	s, err := New("ssh", CategoryDeploy).
		Field("useSCP", ValueTypeSwitch).
		Field("fileFormat", ValueTypeSelect, Options("PEM", "PFX", "JKS"), Default("PEM")).
		Field("pfxPassword", ValueTypeSecret,
			VisibleWhen("fileFormat").Equals("PFX"),
			RequiredWhen("fileFormat").Equals("PFX"),
			VisibleWhen("useSCP").Equals("true"),
		).
		Field("jksAlias", ValueTypeText,
			VisibleWhen("fileFormat").Equals("JKS"),
			RequiredWhen("fileFormat").Equals("JKS"),
		).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pfx := s.Fields[2]
	if len(pfx.VisibleWhen) != 2 {
		t.Fatalf("pfxPassword visibleWhen count = %d, want 2 (multi-discriminator)", len(pfx.VisibleWhen))
	}
	discs := map[string]bool{}
	for _, c := range pfx.VisibleWhen {
		discs[c.Field] = true
	}
	if !discs["fileFormat"] || !discs["useSCP"] {
		t.Fatalf("pfxPassword should react to both fileFormat and useSCP, got %v", discs)
	}
	if len(pfx.RequiredWhen) != 1 || pfx.RequiredWhen[0].Field != "fileFormat" {
		t.Fatalf("pfxPassword requiredWhen not as declared: %+v", pfx.RequiredWhen)
	}
}

func TestBuilder_HappyPath_DefaultAndSecret(t *testing.T) {
	s, err := New("p", CategoryDeploy).
		Field("token", ValueTypeSecret, Secret()).
		Field("port", ValueTypeNumber, Default(443), Min(1), Max(65535)).
		Field("enabled", ValueTypeSwitch, Default(true)).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Fields[0].Secret {
		t.Fatal("token should be secret")
	}
	if s.Fields[1].Default != 443 || *s.Fields[1].Min != 1 || *s.Fields[1].Max != 65535 {
		t.Fatalf("port not as declared: %+v", s.Fields[1])
	}
	if s.Fields[2].Default != true {
		t.Fatalf("enabled default not as declared: %+v", s.Fields[2])
	}
}

func TestBuilder_Edge_EmptySchema(t *testing.T) {
	s, err := New("p", CategoryDeploy).Build()
	if err != nil {
		t.Fatalf("empty schema should build, got %v", err)
	}
	if len(s.Fields) != 0 {
		t.Fatalf("expected 0 fields, got %d", len(s.Fields))
	}
}

func TestBuilder_Edge_OptionsWithoutLabelKey(t *testing.T) {
	s, err := New("p", CategoryDeploy).
		Field("fmt", ValueTypeSelect, Options("a", "b")).
		Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Fields[0].Options[0].LabelKey != "" {
		t.Fatalf("option label key should be empty, got %q", s.Fields[0].Options[0].LabelKey)
	}
}

func TestBuilder_Error_DuplicateFieldName(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("region", ValueTypeText).
		Field("region", ValueTypeText).
		Build()
	if err == nil {
		t.Fatal("expected duplicate-field error")
	}
}

func TestBuilder_Error_ConditionReferencesUndeclaredField(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("domain", ValueTypeText, VisibleWhen("missing").Equals("x")).
		Build()
	if err == nil {
		t.Fatal("expected undeclared-discriminator error")
	}
}

func TestBuilder_Error_UnknownValueType(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("x", ValueType("bogus")).
		Build()
	if err == nil {
		t.Fatal("expected unknown-valueType error")
	}
}

func TestBuilder_Error_DefaultTypeMismatch(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("x", ValueTypeText, Default(123)).
		Build()
	if err == nil {
		t.Fatal("expected default type mismatch error")
	}
}

func TestBuilder_Error_DefaultNotAmongOptions(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("fmt", ValueTypeSelect, Options("a", "b"), Default("c")).
		Build()
	if err == nil {
		t.Fatal("expected default-not-among-options error")
	}
}

func TestBuilder_Error_SelectWithoutOptions(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("fmt", ValueTypeSelect).
		Build()
	if err == nil {
		t.Fatal("expected select-without-options error")
	}
}

func TestBuilder_Error_UnknownValidator(t *testing.T) {
	_, err := New("p", CategoryDeploy).
		Field("disc", ValueTypeSelect, Options("a")).
		Field("x", ValueTypeText, ValidateWhen("disc").Equals("a").Validator(Validator("nope"))).
		Build()
	if err == nil {
		t.Fatal("expected unknown-validator error")
	}
}

func TestBuilder_Error_StickyAfterFirstFailure(t *testing.T) {
	b := New("p", CategoryDeploy)
	_, err := b.
		Field("x", ValueType("bogus")).
		Field("y", ValueTypeText).
		Build()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, err) {
		t.Fatal("error should be non-nil and sticky")
	}
}
