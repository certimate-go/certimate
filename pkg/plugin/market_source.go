package plugin

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type MarketSource struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
}

var githubRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*/[a-zA-Z0-9_.-]+$`)

func ValidateGitHubRepo(repo string) error {
	if !githubRepoPattern.MatchString(repo) || strings.HasSuffix(repo, "/.") || strings.HasSuffix(repo, "/..") {
		return fmt.Errorf("plugin: invalid GitHub repository %q", repo)
	}
	return nil
}

func OfficialMarketSource() MarketSource {
	return MarketSource{ID: "official", Repository: "certimate-go/plugins"}
}

func ParseMarketSource(raw string) (MarketSource, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") || strings.Contains(u.EscapedPath(), "%") {
		return MarketSource{}, fmt.Errorf("market: expected https://github.com/owner/repository")
	}
	repo := strings.ToLower(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), "/"), ".git"))
	if strings.HasSuffix(repo, ".git") {
		return MarketSource{}, fmt.Errorf("market: invalid GitHub repository %q", repo)
	}
	if err := ValidateGitHubRepo(repo); err != nil {
		return MarketSource{}, err
	}
	if repo == "certimate-go/plugins" {
		return OfficialMarketSource(), nil
	}
	return MarketSource{ID: fmt.Sprintf("github-%x", sha256.Sum256([]byte(repo+"@main"))), Repository: repo}, nil
}

func (s MarketSource) IndexURL() string {
	return "https://raw.githubusercontent.com/" + s.Repository + "/refs/heads/main/index.json"
}
