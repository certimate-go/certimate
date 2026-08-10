package huaweiibmc

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
)

type GetManagersResponse struct {
	sdkResponseBase

	Members []*Entity `json:"Members"`
}

type ListManagersResponse struct {
	sdkResponseBase

	Members []*Entity `json:"Members"`
}

func (c *Client) ListManagers() (*ListManagersResponse, error) {
	return c.ListManagersWithContext(context.Background())
}

func (c *Client) ListManagersWithContext(ctx context.Context) (*ListManagersResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/redfish/v1/Managers")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ListManagersResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	} else {
		for _, member := range result.Members {
			if member.ODataID == "" && member.ID != "" {
				member.ODataID = "/redfish/v1/Managers/" + url.PathEscape(member.ID)
			}
			if member.ID == "" && member.ODataID != "" {
				re := regexp.MustCompile(`/redfish/v1/Managers/([^/]+)$`)
				matches := re.FindStringSubmatch(member.ODataID)
				if len(matches) == 2 {
					member.ID = matches[1]
				}
			}
		}
	}

	return result, nil
}
