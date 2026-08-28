package providerschema

import "testing"

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := newRegistry[string]()
	s, err := New("aliyun-cdn", CategoryDeploy).
		Field("region", ValueTypeText).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Register("aliyun-cdn", s); err != nil {
		t.Fatal(err)
	}
	got, err := r.Get("aliyun-cdn")
	if err != nil {
		t.Fatal(err)
	}
	if got != s {
		t.Fatal("Get should return the registered schema")
	}
}

func TestRegistry_GetUnregistered(t *testing.T) {
	r := newRegistry[string]()
	if _, err := r.Get("missing"); err == nil {
		t.Fatal("expected not-found error")
	}
	if r.Has("missing") {
		t.Fatal("Has should be false for unregistered")
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := newRegistry[string]()
	s, _ := New("p", CategoryDeploy).Field("a", ValueTypeText).Build()
	r.MustRegister("p", s)
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate MustRegister")
		}
	}()
	r.MustRegister("p", s)
}

func TestRegistry_NilSchemaRejected(t *testing.T) {
	r := newRegistry[string]()
	if err := r.Register("p", nil); err == nil {
		t.Fatal("expected error for nil schema")
	}
}

func TestRegistry_ListIsSorted(t *testing.T) {
	r := newRegistry[string]()
	for _, name := range []string{"c", "a", "b"} {
		s, _ := New(name, CategoryDeploy).Field("f", ValueTypeText).Build()
		r.MustRegister(name, s)
	}
	got := r.List()
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("List not sorted: %v", got)
	}
}

func TestRegistry_SingletonTypedByDeploymentProviderType(t *testing.T) {
	if Registries == nil {
		t.Fatal("Registries singleton should be initialized")
	}
	if Registries.Has("test-provider-not-registered") {
		t.Fatal("Has should be false for an unregistered provider type")
	}
	if _, err := Registries.Get("test-provider-not-registered"); err == nil {
		t.Fatal("Get should error for an unregistered provider type")
	}
}
