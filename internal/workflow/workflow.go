package workflow

import (
	"context"
)

func Setup() {
	registerWorkflowRecordEvents()
}

func Teardown() {
	thisSvcInst().Shutdown(context.Background())
}
