package workflowtemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/certimate-go/certimate/internal/certmgmt/deployers"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/providerschema"
)

type WorkflowTemplateService struct {
	templateRepo      workflowTemplateRepository
	workflowRepo      workflowRepository
	accessRepo        accessRepository
	providerSchemaSvc providerSchemaService
}

func NewWorkflowTemplateService(
	templateRepo workflowTemplateRepository,
	workflowRepo workflowRepository,
	accessRepo accessRepository,
	providerSchemaSvc providerSchemaService,
) *WorkflowTemplateService {
	return &WorkflowTemplateService{
		templateRepo:      templateRepo,
		workflowRepo:      workflowRepo,
		accessRepo:        accessRepo,
		providerSchemaSvc: providerSchemaSvc,
	}
}

func (s *WorkflowTemplateService) GetById(ctx context.Context, id string) (*domain.WorkflowTemplate, error) {
	return s.templateRepo.GetById(ctx, id)
}

func (s *WorkflowTemplateService) List(ctx context.Context, keyword string) ([]*domain.WorkflowTemplate, error) {
	return s.templateRepo.List(ctx, strings.TrimSpace(keyword))
}

func (s *WorkflowTemplateService) Create(ctx context.Context, req *dtos.WorkflowTemplateCreateReq) (*domain.WorkflowTemplate, error) {
	template := &domain.WorkflowTemplate{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Graph:       req.Graph,
		Params:      req.Params.Sorted(),
	}
	if err := s.validateForSave(ctx, template); err != nil {
		return nil, err
	}

	template.FormatVersion = domain.WorkflowTemplateFormatVersion
	template.Revision = 1

	template, err := s.templateRepo.Save(ctx, template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (s *WorkflowTemplateService) Update(ctx context.Context, id string, req *dtos.WorkflowTemplateUpdateReq) (*domain.WorkflowTemplate, error) {
	template, err := s.templateRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ExpectedRevision != template.Revision {
		return nil, domain.ErrRevisionConflict
	}

	template.Name = strings.TrimSpace(req.Name)
	template.Description = strings.TrimSpace(req.Description)
	template.Graph = req.Graph
	template.Params = req.Params.Sorted()
	if err := s.validateForSave(ctx, template); err != nil {
		return nil, err
	}

	template.Revision++

	template, err = s.templateRepo.Save(ctx, template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (s *WorkflowTemplateService) Delete(ctx context.Context, id string) error {
	return s.templateRepo.Delete(ctx, id)
}

func (s *WorkflowTemplateService) Instantiate(ctx context.Context, id string, req *dtos.WorkflowTemplateInstantiateReq) (*dtos.WorkflowTemplateInstantiateResp, error) {
	template, err := s.templateRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if template.FormatVersion != domain.WorkflowTemplateFormatVersion {
		return nil, domain.NewError(400, "template format version mismatch")
	}

	if req.Revision != template.Revision {
		return nil, domain.ErrRevisionConflict
	}

	workflowName := strings.TrimSpace(req.Name)
	if workflowName == "" {
		return nil, domain.NewError(400, "the workflow name cannot be blank")
	}

	if err := template.Params.Verify(template.Graph); err != nil {
		return nil, domain.NewError(400, fmt.Sprintf("template is broken: %s", err.Error()))
	}

	graph, err := cloneGraph(template.Graph)
	if err != nil {
		return nil, err
	}

	submittedIds := make(map[string]bool, len(req.Params))
	for paramId := range req.Params {
		if _, _, ok := findParam(template.Params, paramId); !ok {
			return nil, domain.NewError(400, fmt.Sprintf("unknown parameter #%s", paramId))
		}
		submittedIds[paramId] = true
	}

	paramErrs := &dtos.WorkflowTemplateParamErrors{Items: make([]dtos.WorkflowTemplateParamError, 0)}
	for _, param := range template.Params.Sorted() {
		value := req.Params[param.Id]

		switch param.Kind {
		case domain.WorkflowTemplateParamKindDomains:
			if err := s.applyDomainsParam(graph, param, value, paramErrs); err != nil {
				return nil, err
			}

		case domain.WorkflowTemplateParamKindDeployment:
			if err := s.applyDeploymentParam(ctx, graph, param, value, paramErrs); err != nil {
				return nil, err
			}
		}
	}

	if len(paramErrs.Items) > 0 {
		return nil, paramErrs
	}

	if err := graph.Verify(); err != nil {
		return nil, domain.NewError(400, fmt.Sprintf("invalid workflow graph: %s", err.Error()))
	}

	workflow := &domain.Workflow{
		Name:        workflowName,
		Description: strings.TrimSpace(req.Description),
		Enabled:     false,
		GraphDraft:  graph,
		HasDraft:    true,
		HasContent:  false,
	}
	if _, err := s.workflowRepo.Save(ctx, workflow); err != nil {
		return nil, err
	}

	return &dtos.WorkflowTemplateInstantiateResp{WorkflowId: workflow.Id}, nil
}

func (s *WorkflowTemplateService) validateForSave(ctx context.Context, template *domain.WorkflowTemplate) error {
	if template.Name == "" {
		return domain.NewError(400, "the template name cannot be blank")
	}

	if template.Graph == nil {
		return domain.NewError(400, "the template graph cannot be blank")
	}
	if err := template.Graph.Verify(); err != nil {
		return domain.NewError(400, fmt.Sprintf("invalid workflow graph: %s", err.Error()))
	}

	if err := template.Params.Verify(template.Graph); err != nil {
		return domain.NewError(400, err.Error())
	}

	if err := s.sanitizeGraph(ctx, template.Graph, template.Params); err != nil {
		return err
	}

	return nil
}

func (s *WorkflowTemplateService) sanitizeGraph(ctx context.Context, graph *domain.WorkflowGraph, params domain.WorkflowTemplateParams) error {
	for _, param := range params {
		switch param.Kind {
		case domain.WorkflowTemplateParamKindDomains:
			{
				for _, nodeId := range param.ApplyNodeIds {
					node, _ := graph.GetNodeById(nodeId)
					if node == nil {
						continue
					}

					node.Data.Config["domains"] = ""
				}
			}

		case domain.WorkflowTemplateParamKindDeployment:
			{
				node, _ := graph.GetNodeById(param.DeployNodeId)
				if node == nil {
					continue
				}

				node.Data.Config["providerAccessId"] = ""

				provider, _ := node.Data.Config["provider"].(string)
				cfg, ok := node.Data.Config["providerConfig"].(map[string]any)
				if !ok {
					continue
				}

				envelope := s.getDeployEnvelope(ctx, provider)
				if envelope == nil {
					delete(node.Data.Config, "providerConfig")
					continue
				}

				for key := range cfg {
					if col := findColumn(envelope, key); col == nil || col.Secret {
						delete(cfg, key)
					}
				}
			}
		}
	}

	return nil
}

func (s *WorkflowTemplateService) getDeployEnvelope(ctx context.Context, providerType string) *providerschema.Envelope {
	if strings.TrimSpace(providerType) == "" {
		return nil
	}

	envelope, err := s.providerSchemaSvc.GetByProviderType(ctx, providerType, providerschema.CategoryDeploy)
	if err != nil || envelope == nil {
		return nil
	}
	if envelope.SchemaVersion != providerschema.SchemaVersion {
		return nil
	}

	return envelope
}

func (s *WorkflowTemplateService) applyDomainsParam(graph *domain.WorkflowGraph, param *domain.WorkflowTemplateParam, value dtos.WorkflowTemplateParamValue, paramErrs *dtos.WorkflowTemplateParamErrors) error {
	domains := make([]string, 0)
	for _, domain := range value.Domains {
		if domain = strings.TrimSpace(domain); domain != "" {
			domains = append(domains, domain)
		}
	}

	if len(domains) == 0 {
		paramErrs.Items = append(paramErrs.Items, dtos.WorkflowTemplateParamError{
			ParamId: param.Id,
			Code:    ParamErrCodeRequired,
			Message: "domains cannot be blank",
		})
		return nil
	}

	for _, domain := range domains {
		if !isValidDomainName(domain, true) {
			paramErrs.Items = append(paramErrs.Items, dtos.WorkflowTemplateParamError{
				ParamId: param.Id,
				Code:    ParamErrCodeInvalidDomain,
				Message: fmt.Sprintf("invalid domain: %q", domain),
			})
			return nil
		}
	}

	joined := strings.Join(domains, ";")
	for _, nodeId := range param.ApplyNodeIds {
		node, _ := graph.GetNodeById(nodeId)
		if node == nil {
			continue
		}

		node.Data.Config["domains"] = joined
	}

	return nil
}

func (s *WorkflowTemplateService) applyDeploymentParam(ctx context.Context, graph *domain.WorkflowGraph, param *domain.WorkflowTemplateParam, value dtos.WorkflowTemplateParamValue, paramErrs *dtos.WorkflowTemplateParamErrors) error {
	addErr := func(code string, message string) {
		paramErrs.Items = append(paramErrs.Items, dtos.WorkflowTemplateParamError{ParamId: param.Id, Code: code, Message: message})
	}

	node, _ := graph.GetNodeById(param.DeployNodeId)
	if node == nil {
		addErr(ParamErrCodeBrokenTemplate, "the bound deploy node no longer exists in the template")
		return nil
	}

	provider, _ := node.Data.Config["provider"].(string)
	if _, err := deployers.Registries.Get(domain.DeploymentProviderType(provider)); err != nil {
		addErr(ParamErrCodeProviderUnavailable, fmt.Sprintf("the deploy provider %q is unavailable", provider))
		return nil
	}

	accessId := strings.TrimSpace(value.ProviderAccessId)
	if providerNeedsAccess(provider) {
		if accessId == "" {
			addErr(ParamErrCodeAccessRequired, "provider access is required")
			return nil
		}

		access, err := s.accessRepo.GetById(ctx, accessId)
		if err != nil {
			addErr(ParamErrCodeAccessNotFound, fmt.Sprintf("the access #%s record does not exist", accessId))
			return nil
		}

		if expected := deriveAccessProviderType(provider); access.Provider != expected {
			addErr(ParamErrCodeAccessProviderMismatch, fmt.Sprintf("the access #%s record does not belong to the provider %q", accessId, provider))
			return nil
		}
	}

	merged := make(map[string]any)
	if defaults, ok := node.Data.Config["providerConfig"].(map[string]any); ok {
		for k, v := range defaults {
			merged[k] = v
		}
	}
	for k, v := range value.ProviderConfig {
		if v == nil {
			continue
		}
		merged[k] = v
	}

	if envelope := s.getDeployEnvelope(ctx, provider); envelope != nil {
		filtered := make(map[string]any)
		for k, v := range merged {
			if col := findColumn(envelope, k); col != nil {
				filtered[k] = v
			}
		}

		if colErr := validateProviderConfig(envelope, filtered); colErr != nil {
			addErr(ParamErrCodeConfigInvalid, colErr.Error())
			return nil
		}

		merged = filtered
	}

	node.Data.Config["providerAccessId"] = accessId
	node.Data.Config["providerConfig"] = merged

	return nil
}

func cloneGraph(source *domain.WorkflowGraph) (*domain.WorkflowGraph, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	target := &domain.WorkflowGraph{}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, err
	}

	return target, nil
}

func findParam(params domain.WorkflowTemplateParams, paramId string) (*domain.WorkflowTemplateParam, int, bool) {
	for i, param := range params {
		if param.Id == paramId {
			return param, i, true
		}
	}
	return nil, 0, false
}

func deriveAccessProviderType(providerType string) string {
	return strings.SplitN(providerType, "-", 2)[0]
}

func providerNeedsAccess(providerType string) bool {
	switch deriveAccessProviderType(providerType) {
	case "local":
		return false
	default:
		return true
	}
}
