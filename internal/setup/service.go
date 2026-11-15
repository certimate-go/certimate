package setup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	validation "ganigeorgiev.com/pocketbase/ozzo-validation/v4"
	"ganigeorgiev.com/pocketbase/ozzo-validation/v4/is"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
	"github.com/certimate-go/certimate/internal/repository"
)

var (
	errInvalidInput = domain.NewError(400, "Invalid administrator email or password. Passwords must match, contain 10–71 characters, and use at most 72 UTF-8 bytes.")
	errInvalidToken = domain.NewError(403, "Initialization token is invalid or expired. Restart Certimate to obtain a new token.")
	errInitialized  = domain.NewError(409, "Administrator already configured. Please log in.")
	errInternal     = domain.NewError(500, "Unable to initialize administrator.")
)

type Service struct {
	repo      *repository.SuperuserRepository
	mu        sync.Mutex
	tokenHash [32]byte
	expiresAt time.Time
	now       func() time.Time
}

func NewService(app core.App) *Service {
	return &Service{repo: repository.NewSuperuserRepository(app), now: time.Now}
}

func (s *Service) Prepare(username, password string, output io.Writer) error {
	if err := s.repo.PrepareInstaller(); err != nil {
		return errors.New("unable to secure legacy installer")
	}
	initialized, err := s.repo.Initialized()
	if err != nil {
		return errInternal
	}
	if initialized {
		return nil
	}
	email := strings.TrimSpace(username)
	if email != "" || password != "" {
		if email == "" || password == "" {
			return errors.New("CERTIMATE_ADMIN_USERNAME and CERTIMATE_ADMIN_PASSWORD must both be provided")
		}
		if err := validateCredentials(email, password, password); err != nil {
			return fmt.Errorf("invalid CERTIMATE_ADMIN_USERNAME or CERTIMATE_ADMIN_PASSWORD: %w", err)
		}
		err := s.repo.CreateFirst(context.Background(), email, password, nil)
		if errors.Is(err, repository.ErrAlreadyInitialized) {
			return nil
		}
		return creationError(err)
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return errors.New("unable to generate initialization token")
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	s.tokenHash = sha256.Sum256([]byte(token))
	s.expiresAt = s.now().Add(30 * time.Minute)
	_, err = fmt.Fprintf(output, "[CERTIMATE] Initialization token (valid for 30 minutes): %s\n", token)
	return err
}

func (s *Service) Status() (*dtos.SetupStatusResp, error) {
	initialized, err := s.repo.Initialized()
	if err != nil {
		return nil, errInternal
	}
	return &dtos.SetupStatusResp{Initialized: initialized}, nil
}

func (s *Service) Create(ctx context.Context, req *dtos.SetupReq) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	status, err := s.Status()
	if err != nil {
		return err
	}
	if status.Initialized {
		return errInitialized
	}
	if err := s.authorize(req.Token); err != nil {
		return err
	}
	email := strings.TrimSpace(req.Email)
	if err := validateCredentials(email, req.Password, req.PasswordConfirm); err != nil {
		return err
	}
	err = s.repo.CreateFirst(ctx, email, req.Password, func() error { return s.authorize(req.Token) })
	if err != nil {
		return creationError(err)
	}
	s.tokenHash = [32]byte{}
	s.expiresAt = time.Time{}
	return nil
}

func (s *Service) authorize(token string) error {
	hash := sha256.Sum256([]byte(token))
	if subtle.ConstantTimeCompare(hash[:], s.tokenHash[:]) != 1 || !s.now().Before(s.expiresAt) {
		return errInvalidToken
	}
	return nil
}

func validateCredentials(email, password, confirm string) error {
	length := utf8.RuneCountInString(password)
	if strings.EqualFold(email, core.DefaultInstallerEmail) || validation.Validate(email, validation.Required, is.EmailFormat) != nil ||
		password != confirm || !utf8.ValidString(password) || length < 10 || length > 71 || len(password) > 72 {
		return errInvalidInput
	}
	return nil
}

func creationError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrAlreadyInitialized) {
		return errInitialized
	}
	if errors.Is(err, errInvalidToken) {
		return errInvalidToken
	}
	var validationErrors validation.Errors
	if errors.As(err, &validationErrors) {
		return errInvalidInput
	}
	return errInternal
}
