package markettest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/certimate-go/certimate/pkg/plugin"
)

func WithReleaseDetails(next http.Handler) http.Handler {
	var details sync.Map
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/market/details/") {
			if detail, ok := details.Load(r.URL.Path); ok {
				_, _ = w.Write(detail.([]byte))
				return
			}
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/" && r.URL.Path != "/index.json" {
			next.ServeHTTP(w, r)
			return
		}
		response := httptest.NewRecorder()
		next.ServeHTTP(response, r)
		body := response.Body.Bytes()
		if response.Code == http.StatusOK {
			var index struct {
				Plugins []*plugin.MarketManifest `json:"plugins"`
			}
			if err := json.Unmarshal(body, &index); err == nil {
				for _, entry := range index.Plugins {
					if entry == nil || entry.Release == nil {
						continue
					}
					release, err := json.Marshal(entry)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					digest := fmt.Sprintf("%x", sha256.Sum256(release))
					details.Store("/market/details/"+digest+".json", release)
					ref := plugin.MarketDetailRef{SHA256: digest, Repo: entry.Release.Repo}
					for platform := range entry.Release.Assets {
						ref.Platforms = append(ref.Platforms, platform)
					}
					if len(ref.Platforms) > 0 {
						entry.Detail = &ref
					}
					entry.Release = nil
				}
				var err error
				body, err = json.Marshal(index)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		for key, values := range response.Header() {
			w.Header()[key] = values
		}
		w.WriteHeader(response.Code)
		_, _ = w.Write(body)
	})
}
