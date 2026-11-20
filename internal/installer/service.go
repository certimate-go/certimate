package installer

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	validation "ganigeorgiev.com/pocketbase/ozzo-validation/v4"
	"ganigeorgiev.com/pocketbase/ozzo-validation/v4/is"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/security"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
)

type InstallerService struct {
	superuserRepo superuserRepository

	token          string
	tokenExpiredAt time.Time

	mu sync.Mutex
}

func NewInstallerService(superuserRepo superuserRepository) *InstallerService {
	srv := &InstallerService{
		superuserRepo: superuserRepo,
		mu:            sync.Mutex{},
	}
	return srv
}

func (s *InstallerService) Status(ctx context.Context) (*dtos.InstallerStatusResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	isEmpty, err := s.superuserRepo.IsEmpty(ctx)
	if err != nil {
		return nil, err
	}

	return &dtos.InstallerStatusResp{Installed: !isEmpty}, nil
}

func (s *InstallerService) Prepare(ctx context.Context) (*dtos.InstallerPrepareResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	isEmpty, err := s.superuserRepo.IsEmpty(ctx)
	if err != nil {
		return nil, err
	} else if !isEmpty {
		return &dtos.InstallerPrepareResp{
			Installed: true,
		}, nil
	}

	if s.token == "" || s.tokenExpiredAt.Before(time.Now()) {
		s.token = security.RandomString(32)
		s.tokenExpiredAt = time.Now().Add(30 * time.Minute)
	}

	return &dtos.InstallerPrepareResp{
		Installed:      false,
		Token:          s.token,
		TokenExpiredAt: s.tokenExpiredAt.Unix(),
	}, nil
}

func (s *InstallerService) InitializeSuperuser(ctx context.Context, req *dtos.InstallerInitializeSuperuserReq) (*dtos.InstallerInitializeSuperuserResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !security.Equal(req.Token, s.token) || s.token == "" || s.tokenExpiredAt.Before(time.Now()) {
		return nil, domain.NewError(http.StatusBadRequest, "invalid or expired installer token")
	}

	if strings.EqualFold(req.Email, core.DefaultInstallerEmail) ||
		validation.Validate(req.Email, validation.Required, is.EmailFormat) != nil ||
		validation.Validate(req.Password, validation.Required) != nil {
		return nil, domain.ErrInvalidParams
	}

	isEmpty, err := s.superuserRepo.IsEmpty(ctx)
	if err != nil {
		return nil, err
	} else if !isEmpty {
		return nil, domain.NewError(http.StatusConflict, "superuser already initialized")
	}

	err = s.superuserRepo.Create(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	s.token = ""
	s.tokenExpiredAt = time.Time{}

	return &dtos.InstallerInitializeSuperuserResp{}, nil
}
