package routes

import (
	"context"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/certificate"
	"github.com/certimate-go/certimate/internal/notify"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/handlers"
	"github.com/certimate-go/certimate/internal/statistics"
	"github.com/certimate-go/certimate/internal/system"
	"github.com/certimate-go/certimate/internal/workflow"
)

var (
	certificateSvc *certificate.CertificateService
	workflowSvc    *workflow.WorkflowService
	statisticsSvc  *statistics.StatisticsService
	notifySvc      *notify.NotifyService
	systemSvc      *system.EnvironmentService
)

func Register(router *router.Router[*core.RequestEvent]) {
	accessRepo := repository.NewAccessRepository()
	workflowRepo := repository.NewWorkflowRepository()
	workflowRunRepo := repository.NewWorkflowRunRepository()
	certificateRepo := repository.NewCertificateRepository()
	settingsRepo := repository.NewSettingsRepository()
	statisticsRepo := repository.NewStatisticsRepository()

	certificateSvc = certificate.NewCertificateService(certificateRepo, settingsRepo)
	workflowSvc = workflow.NewWorkflowService(workflowRepo, workflowRunRepo, settingsRepo)
	statisticsSvc = statistics.NewStatisticsService(statisticsRepo)
	notifySvc = notify.NewNotifyService(accessRepo)
	systemSvc = system.NewEnvironmentService(nil)

	group := router.Group("/api")
	group.Bind(apis.RequireSuperuserAuth())
	handlers.NewCertificateHandler(group, certificateSvc)
	handlers.NewWorkflowHandler(group, workflowSvc)
	handlers.NewStatisticsHandler(group, statisticsSvc)
	handlers.NewSystemHandler(group, systemSvc)
	handlers.NewNotifyHandler(group, notifySvc)
}

func Unregister() {
	if workflowSvc != nil {
		workflowSvc.Shutdown(context.Background())
	}
}
