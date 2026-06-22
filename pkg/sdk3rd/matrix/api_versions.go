package matrix

import (
	"fmt"
)

// probeVersions checks that the homeserver exposes the Client-Server API.
// Проверяет доступность homeserver (GET /_matrix/client/versions).
// REF: https://spec.matrix.org/latest/client-server-api/#get_matrixclientversions
func (c *Client) probeVersions() error {
	_, err := c.rc.R().Get("/_matrix/client/versions")
	if err != nil {
		return fmt.Errorf("sdkerr: failed to probe Matrix Client API versions: %w", err)
	}

	return nil
}
