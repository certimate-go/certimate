package routes

import (
	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/accesscheck"
	"github.com/certimate-go/certimate/internal/acmeaccount"
	"github.com/certimate-go/certimate/internal/certificate"
	"github.com/certimate-go/certimate/internal/notify"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/handlers"
	"github.com/certimate-go/certimate/internal/setup"
	"github.com/certimate-go/certimate/internal/statistics"
	"github.com/certimate-go/certimate/internal/workflow"
	"github.com/certimate-go/certimate/internal/workflowtemplate"
)

var (
	certificateSvc    *certificate.CertificateService
	workflowSvc       *workflow.WorkflowService
	statisticsSvc     *statistics.StatisticsService
	notifySvc         *notify.NotifyService
	providerSchemaSvc *providerschema.ProviderSchemaService
)

func BindRouter(router *router.Router[*core.RequestEvent], setupService *setup.Service) {
	handlers.NewSetupHandler(router.Group("/api"), setupService)
	accessRepo := repository.NewAccessRepository()
	workflowRepo := repository.NewWorkflowRepository()
	workflowRunRepo := repository.NewWorkflowRunRepository()
	acmeAccountRepo := repository.NewACMEAccountRepository()
	certificateRepo := repository.NewCertificateRepository()
	statisticsRepo := repository.NewStatisticsRepository()
	providerSchemaRepo := repository.NewProviderSchemaRepository()

	certificateSvc = certificate.NewCertificateService(acmeAccountRepo, certificateRepo)
	workflowSvc = workflow.NewWorkflowService(workflowRepo, workflowRunRepo)
	statisticsSvc = statistics.NewStatisticsService(statisticsRepo)
	notifySvc = notify.NewNotifyService(accessRepo)
	providerSchemaSvc = providerschema.NewProviderSchemaService(providerSchemaRepo)
	acmeAccountSvc := acmeaccount.NewService(acmeAccountRepo)
	accessCheckSvc := accesscheck.NewService(accessRepo)
	workflowTemplateRepo := repository.NewWorkflowTemplateRepository()
	workflowTemplateSvc := workflowtemplate.NewWorkflowTemplateService(workflowTemplateRepo, workflowRepo, accessRepo, providerSchemaSvc)

	group := router.Group("/api")
	group.Bind(apis.RequireSuperuserAuth())
	handlers.NewCertificatesHandler(group, certificateSvc)
	handlers.NewWorkflowsHandler(group, workflowSvc)
	handlers.NewWorkflowTemplatesHandler(group, workflowTemplateSvc)
	handlers.NewStatisticsHandler(group, statisticsSvc)
	handlers.NewNotificationsHandler(group, notifySvc)
	handlers.NewProviderSchemaHandler(group, providerSchemaSvc)
	handlers.NewPluginCatalogHandler(group, pluginhost.GlobalCatalog())
	handlers.NewPluginAdminHandler(group)
	handlers.NewPluginMarketHandler(group, pluginhost.GlobalMarketService())
	handlers.NewACMEAccountsHandler(group, acmeAccountSvc)
	handlers.NewAccessChecksHandler(group, accessCheckSvc)
}
