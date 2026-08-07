package gcore

import (
	"net/http"

	"github.com/G-Core/gcorelabscdn-go/gcore"
)

type AuthRequestSigner struct {
	apiToken string
}

var _ gcore.RequestSigner = (*AuthRequestSigner)(nil)

func NewAuthRequestSigner(apiToken string) *AuthRequestSigner {
	return &AuthRequestSigner{
		apiToken: apiToken,
	}
}

func (s *AuthRequestSigner) Sign(req *http.Request) error {
	req.Header.Set("Authorization", "APIKey "+s.apiToken)
	return nil
}
