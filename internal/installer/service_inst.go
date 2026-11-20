package installer

import (
	"sync"

	"github.com/certimate-go/certimate/internal/repository"
)

var (
	thisSvc     *InstallerService
	thisSvcOnce sync.Once
)

func thisSvcInst() *InstallerService {
	thisSvcOnce.Do(func() {
		thisSvc = NewInstallerService(
			repository.NewSuperuserRepository(),
		)
	})
	return thisSvc
}

func SingletonInstallerService() *InstallerService {
	return thisSvcInst()
}
