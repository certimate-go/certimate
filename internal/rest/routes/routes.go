package routes

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/certificate"
	"github.com/certimate-go/certimate/internal/notify"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/handlers"
	"github.com/certimate-go/certimate/internal/statistics"
	"github.com/certimate-go/certimate/internal/workflow"
)

var (
	certificateSvc *certificate.CertificateService
	workflowSvc    *workflow.WorkflowService
	statisticsSvc  *statistics.StatisticsService
	notifySvc      *notify.NotifyService
)

func BindRouter(router *router.Router[*core.RequestEvent]) {
	accessRepo := repository.NewAccessRepository()
	workflowRepo := repository.NewWorkflowRepository()
	workflowRunRepo := repository.NewWorkflowRunRepository()
	acmeAccountRepo := repository.NewACMEAccountRepository()
	certificateRepo := repository.NewCertificateRepository()
	statisticsRepo := repository.NewStatisticsRepository()

	certificateSvc = certificate.NewCertificateService(acmeAccountRepo, certificateRepo)
	workflowSvc = workflow.NewWorkflowService(workflowRepo, workflowRunRepo)
	statisticsSvc = statistics.NewStatisticsService(statisticsRepo)
	notifySvc = notify.NewNotifyService(accessRepo)

	group := router.Group("/api")
	group.Bind(apis.RequireSuperuserAuth())
	handlers.NewCertificatesHandler(group, certificateSvc)
	handlers.NewWorkflowsHandler(group, workflowSvc)
	handlers.NewStatisticsHandler(group, statisticsSvc)
	handlers.NewNotificationsHandler(group, notifySvc)
}
