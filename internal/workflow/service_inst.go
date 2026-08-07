package workflow

import (
	"sync"

	"github.com/certimate-go/certimate/internal/repository"
)

var (
	thisSvc     *WorkflowService
	thisSvcOnce sync.Once
)

func thisSvcInst() *WorkflowService {
	thisSvcOnce.Do(func() {
		thisSvc = NewWorkflowService(
			repository.NewWorkflowRepository(),
			repository.NewWorkflowRunRepository(),
		)
	})
	return thisSvc
}
