package cloudflaressl

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func deployRequestField(environment string) (*string, error) {
	env := strings.ToLower(strings.TrimSpace(environment))
	switch env {
	case "", "production":
		return nil, nil
	case "staging":
		return lo.ToPtr("staging"), nil
	default:
		return nil, fmt.Errorf("config `environment` must be empty, \"production\", or \"staging\", got %q", environment)
	}
}
