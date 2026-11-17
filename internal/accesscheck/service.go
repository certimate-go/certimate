package accesscheck

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/pkg/core"
)

const (
	checkBudget            = core.AccessCheckDefaultBudget
	globalConcurrencyLimit = 8
)

type accessRepository interface {
	GetById(ctx context.Context, id string) (*domain.Access, error)
}

type Service struct {
	registry   *Registry
	accessRepo accessRepository
	limiter    *concurrencyLimiter
	logger     *slog.Logger
}

func NewService(accessRepo accessRepository) *Service {
	return NewServiceWithRegistry(DefaultRegistry(), accessRepo)
}

func NewServiceWithRegistry(registry *Registry, accessRepo accessRepository) *Service {
	return &Service{
		registry:   registry,
		accessRepo: accessRepo,
		limiter:    newConcurrencyLimiter(globalConcurrencyLimit),
		logger:     slog.Default(),
	}
}

func (s *Service) SetLogger(logger *slog.Logger) {
	if logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	} else {
		s.logger = logger
	}
}

func (s *Service) TestCapabilities(providerType string) *dtos.AccessTestCapabilitiesResp {
	entry, ok := s.registry.Get(providerType)
	if !ok {
		return &dtos.AccessTestCapabilitiesResp{
			Provider:   providerType,
			Supported:  false,
			Scope:      []string{},
			ReasonCode: string(dtos.AccessTestReasonCodeNoChecker),
		}
	}

	resp := &dtos.AccessTestCapabilitiesResp{
		Provider:          providerType,
		Source:            entry.Source,
		OwnerProviderType: entry.OwnerProviderType,
		Scope:             []string{},
	}
	if entry.Checker != nil {
		resp.Scope = entry.Checker.Scope()
	}
	if entry.UnavailableReason != "" {
		resp.ReasonCode = string(entry.UnavailableReason)
		return resp
	}

	resp.Supported = true
	return resp
}

func (s *Service) Test(ctx context.Context, userId string, req *dtos.AccessTestReq) (*dtos.AccessTestResp, error) {
	providerType, config, err := s.resolveRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	entry, ok := s.registry.Get(providerType)
	if !ok {
		return nil, domain.NewError(400, string(dtos.AccessTestReasonCodeNoChecker))
	}
	if entry.UnavailableReason != "" {
		return unavailableEntryResult(entry.UnavailableReason), nil
	}

	if !s.limiter.acquire(userId) {
		return nil, domain.NewError(429, string(dtos.AccessTestReasonCodeConcurrencyLimited))
	}
	defer s.limiter.release(userId)

	checkCtx, cancel := context.WithTimeout(ctx, checkBudget)
	defer cancel()

	startedAt := time.Now()
	result, err := entry.Checker.Check(checkCtx, config)
	if err != nil {
		if errors.Is(err, core.ErrAccessCheckConfigInvalid) {
			s.logger.Warn("access check config rejected", slog.String("provider", providerType), slog.String("error", err.Error()))
			return nil, domain.NewError(400, string(dtos.AccessTestReasonCodeConfigInvalid))
		}
		s.logger.Error("access check failed internally", slog.String("provider", providerType), slog.String("error", err.Error()))
		return nil, domain.NewError(500, string(dtos.AccessTestReasonCodeInternalError))
	}

	s.logger.Info("access check finished",
		slog.String("provider", providerType),
		slog.String("status", result.Status),
		slog.String("reasonCode", result.ReasonCode),
		slog.Duration("elapsed", time.Since(startedAt)),
	)
	return toAccessTestResp(result), nil
}

func toAccessTestResp(res *core.AccessCheckResult) *dtos.AccessTestResp {
	items := make([]dtos.AccessTestItemResp, 0, len(res.Items))
	for _, item := range res.Items {
		items = append(items, dtos.AccessTestItemResp{Id: item.Id, Status: dtos.AccessTestItemStatus(item.Status)})
	}
	diagnostics := res.Diagnostics
	if diagnostics == nil {
		diagnostics = []string{}
	}
	return &dtos.AccessTestResp{
		Status:      dtos.AccessTestStatus(res.Status),
		ReasonCode:  res.ReasonCode,
		Items:       items,
		Diagnostics: diagnostics,
	}
}

func (s *Service) resolveRequest(ctx context.Context, req *dtos.AccessTestReq) (string, map[string]any, error) {
	switch req.Source {
	case dtos.AccessTestSourceStored:
		access, err := s.accessRepo.GetById(ctx, req.AccessId)
		if err != nil {
			return "", nil, domain.NewError(404, string(dtos.AccessTestReasonCodeAccessNotFound))
		}
		return access.Provider, access.Config, nil

	case dtos.AccessTestSourceDraft:
		if !s.registry.IsKnownProvider(req.Provider) {
			return "", nil, domain.NewError(400, fmt.Sprintf("unknown provider %q", req.Provider))
		}
		return req.Provider, req.Config, nil

	default:
		return "", nil, domain.ErrInvalidParams
	}
}

func unavailableEntryResult(reason dtos.AccessTestReasonCode) *dtos.AccessTestResp {
	status := lo.Switch[dtos.AccessTestReasonCode, dtos.AccessTestStatus](reason).
		Case(dtos.AccessTestReasonCodePluginProtocolOld, dtos.AccessTestStatusUnsupported).
		Case(dtos.AccessTestReasonCodeNoChecker, dtos.AccessTestStatusUnsupported).
		Default(dtos.AccessTestStatusUnavailable)

	return &dtos.AccessTestResp{
		Status:      status,
		ReasonCode:  string(reason),
		Items:       []dtos.AccessTestItemResp{},
		Diagnostics: []string{},
	}
}

type concurrencyLimiter struct {
	mu        sync.Mutex
	perUser   map[string]int
	global    int
	maxGlobal int
}

func newConcurrencyLimiter(maxGlobal int) *concurrencyLimiter {
	return &concurrencyLimiter{
		perUser:   make(map[string]int),
		maxGlobal: maxGlobal,
	}
}

func (l *concurrencyLimiter) acquire(userId string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.perUser[userId] >= 1 || l.global >= l.maxGlobal {
		return false
	}

	l.perUser[userId]++
	l.global++
	return true
}

func (l *concurrencyLimiter) release(userId string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.perUser[userId] <= 0 {
		return
	}

	l.perUser[userId]--
	if l.perUser[userId] == 0 {
		delete(l.perUser, userId)
	}
	l.global--
}
