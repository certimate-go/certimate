package accesscheck

import (
	"github.com/certimate-go/certimate/pkg/core"
	cfimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/cloudflare"
	sshimpl "github.com/certimate-go/certimate/pkg/core/accesschecker/providers/ssh"
)

func init() {
	registerCoreChecker(sshimpl.New())
	registerCoreChecker(cfimpl.New())
}

func registerCoreChecker(checker core.AccessChecker) {
	if err := DefaultRegistry().Register(&Registration{
		Source:  SourceCore,
		Checker: checker,
	}); err != nil {
		panic(err)
	}
}
