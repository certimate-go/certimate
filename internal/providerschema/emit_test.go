package providerschema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmit_PreservesFieldOrderAndEnvelope(t *testing.T) {
	s, err := New("aliyun-cdn", CategoryDeploy).
		Field("region", ValueTypeText, LabelKey("k.region")).
		Field("domainMatchPattern", ValueTypeRadio, Options("exact", "wildcard"), Default("exact")).
		Field("domain", ValueTypeText, VisibleWhen("domainMatchPattern").NotIn("certsan")).
		Build()
	if err != nil {
		t.Fatal(err)
	}

	env := Emit(s)
	if env.SchemaVersion != "form/v1" {
		t.Fatalf("schemaVersion = %q", env.SchemaVersion)
	}
	if env.Provider != "aliyun-cdn" || env.Category != CategoryDeploy {
		t.Fatalf("provider/category = %q/%q", env.Provider, env.Category)
	}
	if len(env.Schema.Columns) != 3 {
		t.Fatalf("column count = %d", len(env.Schema.Columns))
	}
	if env.Schema.Columns[0].Name != "region" || env.Schema.Columns[2].Name != "domain" {
		t.Fatalf("field order not preserved: %v", columnNames(env))
	}
}

func TestEmit_OptionsAndSecret(t *testing.T) {
	s, _ := New("p", CategoryDeploy).
		Field("fmt", ValueTypeSelect, Options("a", "b")).
		Field("token", ValueTypeSecret, Secret()).
		Build()
	env := Emit(s)

	if len(env.Schema.Columns[0].Options) != 2 {
		t.Fatalf("select options not emitted: %+v", env.Schema.Columns[0])
	}
	if env.Schema.Columns[0].Options[0].Value != "a" {
		t.Fatalf("option value not emitted: %+v", env.Schema.Columns[0].Options[0])
	}
	if !env.Schema.Columns[1].Secret || env.Schema.Columns[1].ValueType != ValueTypeSecret {
		t.Fatalf("secret field not emitted as secret: %+v", env.Schema.Columns[1])
	}

	textOnly, _ := New("p2", CategoryDeploy).Field("x", ValueTypeText).Build()
	if len(Emit(textOnly).Schema.Columns[0].Options) != 0 {
		t.Fatal("text field should not emit options")
	}
}

func TestEmit_ConditionalsAndDependencies(t *testing.T) {
	s, _ := New("ssh", CategoryDeploy).
		Field("useSCP", ValueTypeSwitch).
		Field("fileFormat", ValueTypeSelect, Options("PEM", "PFX")).
		Field(
			"pfxPassword", ValueTypeSecret,
			VisibleWhen("fileFormat").Equals("PFX"),
			RequiredWhen("fileFormat").Equals("PFX"),
			VisibleWhen("useSCP").Equals("true"),
			ValidateWhen("fileFormat").Equals("PFX").Validator(ValidatorDomain),
		).
		Build()
	env := Emit(s)
	col := env.Schema.Columns[2]

	if len(col.VisibleWhen) != 2 {
		t.Fatalf("visibleWhen count = %d, want 2", len(col.VisibleWhen))
	}
	if len(col.RequiredWhen) != 1 || col.RequiredWhen[0].Field != "fileFormat" {
		t.Fatalf("requiredWhen not emitted: %+v", col.RequiredWhen)
	}
	if len(col.ValidateWhen) != 1 || col.ValidateWhen[0].Name != ValidatorDomain {
		t.Fatalf("validateWhen not emitted: %+v", col.ValidateWhen)
	}

	deps := col.Dependencies
	if len(deps) != 2 {
		t.Fatalf("dependencies count = %d, want 2 (fileFormat+useSCP): %v", len(deps), deps)
	}
	hasFileFormat, hasUseSCP := false, false
	for _, d := range deps {
		if d == "fileFormat" {
			hasFileFormat = true
		}
		if d == "useSCP" {
			hasUseSCP = true
		}
	}
	if !hasFileFormat || !hasUseSCP {
		t.Fatalf("dependencies missing discriminators: %v", deps)
	}
}

func TestEmit_JSONRoundTrip(t *testing.T) {
	s, _ := New("aliyun-cdn", CategoryDeploy).
		Field("domainMatchPattern", ValueTypeRadio, Options("exact", "wildcard", "certsan"), Default("exact")).
		Field("domain", ValueTypeText, RequiredWhen("domainMatchPattern").In("exact", "wildcard")).
		Build()
	env := Emit(s)

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	str := string(data)
	if !strings.Contains(str, `"schemaVersion":"form/v1"`) {
		t.Fatalf("envelope missing schemaVersion: %s", str)
	}
	if !strings.Contains(str, `"columns"`) {
		t.Fatalf("envelope missing columns: %s", str)
	}
	if !strings.Contains(str, `"requiredWhen"`) {
		t.Fatalf("envelope missing requiredWhen: %s", str)
	}
}

func columnNames(env *Envelope) []string {
	out := make([]string, len(env.Schema.Columns))
	for i, c := range env.Schema.Columns {
		out[i] = c.Name
	}
	return out
}
