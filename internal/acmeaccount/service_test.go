package acmeaccount

import (
	"context"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domain/dtos"
)

type mockRepo struct {
	accounts []*domain.ACMEAccount
	byID     map[string]*domain.ACMEAccount
	page     int
	perPage  int
	ca       string
}

func (m *mockRepo) List(ctx context.Context, page, perPage int, ca string) ([]*domain.ACMEAccount, int64, error) {
	m.page = page
	m.perPage = perPage
	m.ca = ca
	filtered := make([]*domain.ACMEAccount, 0)
	for _, a := range m.accounts {
		if ca != "" && a.CA != ca {
			continue
		}
		filtered = append(filtered, a)
	}
	total := int64(len(filtered))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}
	start := (page - 1) * perPage
	if start >= len(filtered) {
		return []*domain.ACMEAccount{}, total, nil
	}
	end := start + perPage
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (m *mockRepo) GetById(ctx context.Context, id string) (*domain.ACMEAccount, error) {
	if a, ok := m.byID[id]; ok {
		return a, nil
	}
	return nil, domain.ErrRecordNotFound
}

func TestList_OmitsPrivateKeyAndPaginates(t *testing.T) {
	repo := &mockRepo{
		accounts: []*domain.ACMEAccount{
			{
				Meta:             domain.Meta{Id: "a1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				CA:               "letsencrypt",
				Email:            "a@example.com",
				PrivateKey:       "-----BEGIN EC PRIVATE KEY-----\nSECRET\n-----END EC PRIVATE KEY-----",
				ACMEDirectoryUrl: "https://acme-v02.api.letsencrypt.org/directory",
				ACMEAccountUrl:   "https://acme-v02.api.letsencrypt.org/acme/acct/1",
			},
			{
				Meta:             domain.Meta{Id: "a2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				CA:               "zerossl",
				Email:            "b@example.com",
				PrivateKey:       "SECRET2",
				ACMEDirectoryUrl: "https://acme.zerossl.com/v2/DV90",
				ACMEAccountUrl:   "https://acme.zerossl.com/v2/DV90/account/2",
			},
		},
	}
	svc := NewService(repo)
	resp, err := svc.List(context.Background(), &dtos.ACMEAccountListReq{Page: 1, PerPage: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.TotalItems != 2 {
		t.Fatalf("totalItems = %d", resp.TotalItems)
	}
	if resp.Page != 1 || resp.PerPage != 1 {
		t.Fatalf("page/perPage = %d/%d", resp.Page, resp.PerPage)
	}
	if resp.Items[0].Email != "a@example.com" {
		t.Fatalf("email = %q", resp.Items[0].Email)
	}
	if resp.Items[0].Id != "a1" || resp.Items[0].ACMEAccountUrl == "" {
		t.Fatalf("unexpected view: %+v", resp.Items[0])
	}
}

func TestList_FilterByCA(t *testing.T) {
	repo := &mockRepo{
		accounts: []*domain.ACMEAccount{
			{Meta: domain.Meta{Id: "a1"}, CA: "letsencrypt", Email: "a@example.com"},
			{Meta: domain.Meta{Id: "a2"}, CA: "zerossl", Email: "b@example.com"},
		},
	}
	svc := NewService(repo)
	resp, err := svc.List(context.Background(), &dtos.ACMEAccountListReq{Page: 1, PerPage: 10, CA: "zerossl"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.TotalItems != 1 || resp.Items[0].CA != "zerossl" {
		t.Fatalf("unexpected: %+v total=%d", resp.Items, resp.TotalItems)
	}
	if repo.ca != "zerossl" {
		t.Fatalf("repo ca = %q", repo.ca)
	}
}

func TestExport_ReturnsPEM(t *testing.T) {
	pem := "-----BEGIN EC PRIVATE KEY-----\nTEST\n-----END EC PRIVATE KEY-----"
	repo := &mockRepo{
		byID: map[string]*domain.ACMEAccount{
			"a1": {
				Meta:       domain.Meta{Id: "a1"},
				PrivateKey: pem,
			},
		},
	}
	svc := NewService(repo)
	resp, err := svc.Export(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.PrivateKeyPem != pem {
		t.Fatalf("pem mismatch")
	}
}

func TestExport_NotFound(t *testing.T) {
	svc := NewService(&mockRepo{byID: map[string]*domain.ACMEAccount{}})
	_, err := svc.Export(context.Background(), "missing")
	if !domain.IsRecordNotFoundError(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestToView_Nil(t *testing.T) {
	if toView(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestImport_NilRequest(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.Import(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	_ = &dtos.ACMEAccountImportReq{}
}
