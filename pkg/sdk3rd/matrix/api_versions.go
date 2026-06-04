package matrix

import "fmt"

// probeVersions checks that the homeserver exposes the Client-Server API.
// Проверяет доступность homeserver (GET /_matrix/client/versions).
// REF: https://spec.matrix.org/latest/client-server-api/#get_matrixclientversions
func (c *Client) probeVersions() error {
	base, err := c.ResolveBaseURL()
	if err != nil {
		return err
	}
	r, err := c.http.R().Get(base + "/_matrix/client/versions")
	if err != nil {
		return fmt.Errorf("sdkerr: cannot reach homeserver at %s: %w", base, err)
	}
	if r.IsError() {
		return fmt.Errorf("sdkerr: homeserver at %s returned HTTP %d", base, r.StatusCode())
	}
	return nil
}
