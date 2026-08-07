package deployers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

func TestPilot_AliyunCDN_SchemaMatchesStaticForm(t *testing.T) {
	schema, err := providerschema.Registries.Get(string(domain.DeploymentProviderTypeAliyunCDN))
	if err != nil {
		t.Fatalf("aliyun-cdn schema not registered: %v", err)
	}
	env := providerschema.Emit(schema)

	got := fieldNames(env)
	want := []string{"region", "domainMatchPattern", "domain"}
	if !equalSlices(got, want) {
		t.Fatalf("field order/names = %v, want %v", got, want)
	}

	region := env.Schema.Columns[0]
	if region.ValueType != providerschema.ValueTypeAutocomplete || len(region.Options) != 2 || !region.TooltipHtml {
		t.Fatalf("region column not as expected: %+v", region)
	}

	dmp := env.Schema.Columns[1]
	if dmp.ValueType != providerschema.ValueTypeRadio || len(dmp.Options) != 3 || dmp.Default != "exact" || !dmp.Required {
		t.Fatalf("domainMatchPattern column not as expected: %+v", dmp)
	}

	domain := env.Schema.Columns[2]
	if len(domain.VisibleWhen) != 1 || domain.VisibleWhen[0].Op != providerschema.OpNotIn {
		t.Fatalf("domain visibleWhen not as expected: %+v", domain.VisibleWhen)
	}
	if len(domain.RequiredWhen) != 1 || domain.RequiredWhen[0].Op != providerschema.OpIn {
		t.Fatalf("domain requiredWhen not as expected: %+v", domain.RequiredWhen)
	}
	if len(domain.ValidateWhen) != 1 || domain.ValidateWhen[0].Name != providerschema.ValidatorDomain {
		t.Fatalf("domain validateWhen not as expected: %+v", domain.ValidateWhen)
	}
	if allow, _ := domain.ValidateWhen[0].Params["allowWildcard"].(bool); !allow {
		t.Fatalf("domain validator should allow wildcards: %+v", domain.ValidateWhen[0].Params)
	}
}

func TestPilot_AliyunCDN_DeploySideDriftGuard(t *testing.T) {
	schema, err := providerschema.Registries.Get(string(domain.DeploymentProviderTypeAliyunCDN))
	if err != nil {
		t.Fatalf("aliyun-cdn schema not registered: %v", err)
	}
	schemaFields := make(map[string]struct{}, len(schema.Fields))
	for _, f := range schema.Fields {
		schemaFields[f.Name] = struct{}{}
	}

	readKeys, err := extendedConfigKeys("sp_aliyun_cdn.go")
	if err != nil {
		t.Fatalf("parse sp_aliyun_cdn.go: %v", err)
	}

	for key := range readKeys {
		if _, ok := schemaFields[key]; !ok {
			t.Errorf("deployer reads extended-config key %q but schema has no such field", key)
		}
	}
	for name := range schemaFields {
		if _, ok := readKeys[name]; !ok {
			t.Errorf("schema field %q has no matching xmaps.Get* reader in sp_aliyun_cdn.go", name)
		}
	}
}

func fieldNames(env *providerschema.Envelope) []string {
	out := make([]string, len(env.Schema.Columns))
	for i, c := range env.Schema.Columns {
		out[i] = c.Name
	}
	return out
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func extendedConfigKeys(filename string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]struct{})
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "xmaps" {
			return true
		}
		if len(call.Args) < 2 {
			return true
		}
		src, ok := call.Args[0].(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if src.Sel == nil || src.Sel.Name != "ProviderExtendedConfig" {
			return true
		}
		lit, ok := call.Args[1].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		keys[strings.Trim(lit.Value, `"`)] = struct{}{}
		return true
	})
	return keys, nil
}
