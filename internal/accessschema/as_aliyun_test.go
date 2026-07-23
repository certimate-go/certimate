package accessschema

import (
	"reflect"
	"strings"
	"testing"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

func TestAccessSchema_DriftGuard_Aliyun(t *testing.T) {
	schema, err := providerschema.Registries.Get(string(domain.AccessProviderTypeAliyun))
	if err != nil {
		t.Fatalf("aliyun access schema not registered: %v", err)
	}

	schemaFields := make(map[string]struct{}, len(schema.Fields))
	for _, f := range schema.Fields {
		schemaFields[f.Name] = struct{}{}
	}

	structTags := jsonFieldNames(reflect.TypeFor[domain.AccessConfigForAliyun]())

	for name := range schemaFields {
		if _, ok := structTags[name]; !ok {
			t.Errorf("schema field %q has no matching json tag on AccessConfigForAliyun", name)
		}
	}
	for name := range structTags {
		if _, ok := schemaFields[name]; !ok {
			t.Errorf("AccessConfigForAliyun json tag %q has no matching schema field", name)
		}
	}

	for _, f := range schema.Fields {
		if (f.Name == "accessKeyId" || f.Name == "accessKeySecret") && !f.Secret {
			t.Errorf("access field %q should be marked secret", f.Name)
		}
	}
}

func jsonFieldNames(t reflect.Type) map[string]struct{} {
	names := make(map[string]struct{})
	walkStruct(t, names)
	return names
}

func walkStruct(t reflect.Type, names map[string]struct{}) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			walkStruct(field.Type, names)
			continue
		}
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if name == "" {
			name = field.Name
		}
		if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct) {
			names[name] = struct{}{}
			if field.Type.Kind() == reflect.Struct {
				walkStruct(field.Type, names)
			}
			continue
		}
		names[name] = struct{}{}
	}
}
