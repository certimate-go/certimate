package plugin

import (
	"context"
	"encoding/json"
	"testing"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

func jsonMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
