package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
)

type CreateDomainEntryRequest = lightsail.CreateDomainEntryInput

type CreateDomainEntryResponse = lightsail.CreateDomainEntryOutput

func (c *Client) CreateDomainEntry(req *CreateDomainEntryRequest) (*CreateDomainEntryResponse, error) {
	return c.CreateDomainEntryWithContext(context.Background(), req)
}

func (c *Client) CreateDomainEntryWithContext(ctx context.Context, req *CreateDomainEntryRequest) (*CreateDomainEntryResponse, error) {
	httpreq, err := c.newRequest(buildAmzTarget("CreateDomainEntry"))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &CreateDomainEntryResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
