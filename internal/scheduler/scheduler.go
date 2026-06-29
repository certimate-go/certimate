package scheduler

import (
	"log/slog"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/certificate"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/workflow"
)

func Setup() {
	workflowRepo := repository.NewWorkflowRepository()
	workflowRunRepo := repository.NewWorkflowRunRepository()
	acmeAccountRepo := repository.NewACMEAccountRepository()
	certificateRepo := repository.NewCertificateRepository()
	settingsRepo := repository.NewSettingsRepository()

	workflowSvc := workflow.NewWorkflowService(workflowRepo, workflowRunRepo, settingsRepo)
	certificateSvc := certificate.NewCertificateService(acmeAccountRepo, certificateRepo, settingsRepo)

	if err := initWorkflowScheduler(workflowSvc); err != nil {
		app.GetLogger().Error("failed to init workflow scheduler", slog.Any("error", err))
	}

	if err := initCertificateScheduler(certificateSvc); err != nil {
		app.GetLogger().Error("failed to init certificate scheduler", slog.Any("error", err))
	}
}
