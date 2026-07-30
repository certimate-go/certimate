package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
)

type DeleteDomainEntryRequest = lightsail.DeleteDomainEntryInput

type DeleteDomainEntryResponse = lightsail.DeleteDomainEntryOutput

func (c *Client) DeleteDomainEntry(req *DeleteDomainEntryRequest) (*DeleteDomainEntryResponse, error) {
	return c.DeleteDomainEntryWithContext(context.Background(), req)
}

func (c *Client) DeleteDomainEntryWithContext(ctx context.Context, req *DeleteDomainEntryRequest) (*DeleteDomainEntryResponse, error) {
	httpreq, err := c.newRequest(buildAmzTarget("DeleteDomainEntry"))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DeleteDomainEntryResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
