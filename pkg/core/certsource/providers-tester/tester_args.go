//go:build tester

package tester

import (
	"github.com/certimate-go/certimate/pkg/core/internal/testing"
)

func InitArgs(prefix string) testing.ArgsParser {
	return testing.InitArgs(prefix)
}
