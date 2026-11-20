package routes

import (
	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"

	"github.com/certimate-go/certimate/internal/accesscheck"
	"github.com/certimate-go/certimate/internal/acmeaccount"
	"github.com/certimate-go/certimate/internal/certificate"
	"github.com/certimate-go/certimate/internal/installer"
	"github.com/certimate-go/certimate/internal/notify"
	"github.com/certimate-go/certimate/internal/pluginhost"
	"github.com/certimate-go/certimate/internal/providerschema"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/rest/handlers"
	"github.com/certimate-go/certimate/internal/statistics"
	"github.com/certimate-go/certimate/internal/workflow"
	"github.com/certimate-go/certimate/internal/workflowtemplate"
)

var (
	installerSvc      *installer.InstallerService
	certificateSvc    *certificate.CertificateService
	workflowSvc       *workflow.WorkflowService
	statisticsSvc     *statistics.StatisticsService
	notifySvc         *notify.NotifyService
	providerSchemaSvc *providerschema.ProviderSchemaService
)

func BindRouter(router *router.Router[*core.RequestEvent]) {
	accessRepo := repository.NewAccessRepository()
	workflowRepo := repository.NewWorkflowRepository()
	workflowRunRepo := repository.NewWorkflowRunRepository()
	acmeAccountRepo := repository.NewACMEAccountRepository()
	certificateRepo := repository.NewCertificateRepository()
	statisticsRepo := repository.NewStatisticsRepository()
	providerSchemaRepo := repository.NewProviderSchemaRepository()

	installerSvc = installer.SingletonInstallerService()
	certificateSvc = certificate.NewCertificateService(acmeAccountRepo, certificateRepo)
	workflowSvc = workflow.NewWorkflowService(workflowRepo, workflowRunRepo)
	statisticsSvc = statistics.NewStatisticsService(statisticsRepo)
	notifySvc = notify.NewNotifyService(accessRepo)
	providerSchemaSvc = providerschema.NewProviderSchemaService(providerSchemaRepo)
	acmeAccountSvc := acmeaccount.NewService(acmeAccountRepo)
	accessCheckSvc := accesscheck.NewService(accessRepo)
	workflowTemplateRepo := repository.NewWorkflowTemplateRepository()
	workflowTemplateSvc := workflowtemplate.NewWorkflowTemplateService(workflowTemplateRepo, workflowRepo, accessRepo, providerSchemaSvc)

	groupWithoutAuth := router.Group("/api")
	handlers.NewInstallerHandler(groupWithoutAuth, installerSvc)

	groupWithAuth := router.Group("/api")
	groupWithAuth.Bind(apis.RequireSuperuserAuth())
	handlers.NewCertificatesHandler(groupWithAuth, certificateSvc)
	handlers.NewWorkflowsHandler(groupWithAuth, workflowSvc)
	handlers.NewWorkflowTemplatesHandler(groupWithAuth, workflowTemplateSvc)
	handlers.NewStatisticsHandler(groupWithAuth, statisticsSvc)
	handlers.NewNotificationsHandler(groupWithAuth, notifySvc)
	handlers.NewProviderSchemaHandler(groupWithAuth, providerSchemaSvc)
	handlers.NewPluginCatalogHandler(groupWithAuth, pluginhost.GlobalCatalog())
	handlers.NewPluginAdminHandler(groupWithAuth)
	handlers.NewPluginMarketHandler(groupWithAuth, pluginhost.GlobalMarketService())
	handlers.NewACMEAccountsHandler(groupWithAuth, acmeAccountSvc)
	handlers.NewAccessChecksHandler(groupWithAuth, accessCheckSvc)
}
