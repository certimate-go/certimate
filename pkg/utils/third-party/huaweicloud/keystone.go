package huaweicloud

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	hwiam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	hwiammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	hwiamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	"github.com/samber/lo"
)

func GetKeystoneProjectIDWithRegion(auth auth.ICredential, region string) (string, error) {
	fallbackRegion := "cn-north-4" // IAM 服务默认区域：华北北京四
	if region == "" {
		region = fallbackRegion
	}

	hcRegion, err := hwiamregion.SafeValueOf(region)
	if err != nil {
		return "", err
	}

	hcClient, err := hwiam.IamClientBuilder().
		WithRegion(hcRegion).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return "", err
	}

	client := hwiam.NewIamClient(hcClient)

	page := 1
	pageSize := 1000
	fallbackId := ""
	for {
		request := &hwiammodel.KeystoneListProjectsRequest{
			Name:    &region,
			Enabled: lo.ToPtr(true),
			Page:    lo.ToPtr(int32(page)),
			PerPage: lo.ToPtr(int32(pageSize)),
		}
		response, err := client.KeystoneListProjects(request)
		if err != nil {
			return "", err
		}

		for _, project := range *response.Projects {
			if project.DomainId == project.ParentId {
				if project.Name == region {
					return project.Id, nil
				}
				if project.Name == fallbackRegion && fallbackId == "" {
					fallbackId = project.Id
				}
			}
		}

		if len(*response.Projects) < pageSize {
			break
		}
		page++
	}

	if fallbackId == "" {
		return "", fmt.Errorf("unable to find huaweicloud project")
	}
	return fallbackId, nil
}
