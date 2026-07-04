package eomakers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/app"
)

func doRequest(ctx context.Context, apiToken string, reqBody any, respBody any) error {
	if ctx == nil {
		ctx = context.Background()
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", APIBaseURL, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", app.AppUserAgent)
	req.Header.Set("Authorization", strings.Join([]string{"Bearer", apiToken}, " "))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sdkErr: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, respBody)
}
