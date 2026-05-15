package cdn

import (
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

type DescribeCertificatesRequest struct {
	*common.BaseRequest

	CertificateIds   []string `json:"certificateIds,omitempty"`
	CertificateLabel string   `json:"certificateLabel,omitempty"`
	San              string   `json:"san,omitempty"`
	ResourceGroupId  string   `json:"resourceGroupId,omitempty"`
	Expired          *bool    `json:"expired,omitempty"`
	PageSize         int      `json:"pageSize,omitempty"`
	PageNum          int      `json:"pageNum,omitempty"`
}

type DescribeCertificatesResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response *struct {
		RequestId string `json:"requestId,omitempty"`

		TotalCount int                `json:"totalCount,omitempty"`
		DataSet    []*CertificateInfo `json:"dataSet,omitempty"`
	} `json:"response"`
}

type CertificateInfo struct {
	CertificateId    string   `json:"certificateId,omitempty"`
	CertificateLabel string   `json:"certificateLabel,omitempty"`
	Common           string   `json:"common,omitempty"`
	Fingerprint      string   `json:"fingerprint,omitempty"`
	Issuer           string   `json:"issuer,omitempty"`
	Sans             []string `json:"sans,omitempty"`
	Algorithm        string   `json:"algorithm,omitempty"`
	CreateTime       string   `json:"createTime,omitempty"`
	StartTime        string   `json:"startTime,omitempty"`
	EndTime          string   `json:"endTime,omitempty"`
	Expired          bool     `json:"expired,omitempty"`
	ResourceGroupId  string   `json:"resourceGroupId,omitempty"`
}

type CreateCertificateRequest struct {
	*common.BaseRequest

	CertificateContent string `json:"certificateContent,omitempty"`
	CertificateKey     string `json:"certificateKey,omitempty"`
	CertificateLabel   string `json:"certificateLabel,omitempty"`
	ResourceGroupId    string `json:"resourceGroupId,omitempty"`
}

type CreateCertificateResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response *struct {
		RequestId string `json:"requestId,omitempty"`

		CertificateId string `json:"certificateId,omitempty"`
	} `json:"response"`
}

type ModifyCertificateRequest struct {
	*common.BaseRequest

	CertificateId      string `json:"certificateId,omitempty"`
	CertificateContent string `json:"certificateContent,omitempty"`
	CertificateKey     string `json:"certificateKey,omitempty"`
}

type ModifyCertificateResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response struct {
		RequestId     string `json:"requestId,omitempty"`
		CertificateId string `json:"certificateId,omitempty"`
	} `json:"response"`
}

type DeleteCertificateRequest struct {
	*common.BaseRequest

	CertificateId string `json:"certificateId,omitempty"`
}

type DeleteCertificateResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response struct {
		RequestId string `json:"requestId,omitempty"`
	} `json:"response"`
}

type DescribeDomainsRequest struct {
	*common.BaseRequest

	DomainIds            []string `json:"domainIds,omitempty"`
	DomainName           string   `json:"domainName,omitempty"`
	DomainStatus         string   `json:"domainStatus,omitempty"`
	ConfigStatus         string   `json:"configStatus,omitempty"`
	AccelerationRegionId string   `json:"accelerationRegionId,omitempty"`
	BusinessTypeId       string   `json:"businessTypeId,omitempty"`
	ResourceGroupId      string   `json:"resourceGroupId,omitempty"`
	PageSize             int      `json:"pageSize,omitempty"`
	PageNum              int      `json:"pageNum,omitempty"`
}

type DescribeDomainsResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response *struct {
		RequestId string `json:"requestId,omitempty"`

		TotalCount int           `json:"totalCount,omitempty"`
		DataSet    []*DomainInfo `json:"dataSet,omitempty"`
	} `json:"response"`
}

type DomainInfo struct {
	DomainId        string `json:"domainId,omitempty"`
	BusinessType    string `json:"businessType,omitempty"`
	DomainName      string `json:"domainName,omitempty"`
	DomainStatus    string `json:"domainStatus,omitempty"`
	ConfigStatus    string `json:"configStatus,omitempty"`
	Cname           string `json:"cname,omitempty"`
	CreateTime      string `json:"createTime,omitempty"`
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
}

type DescribeDomainCertificateRequest struct {
	*common.BaseRequest

	DomainId string `json:"domainId,omitempty"`
}

type DescribeDomainCertificateResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response struct {
		RequestId string `json:"requestId,omitempty"`

		Certificate *CertificateInfo `json:"certificate,omitempty"`
	} `json:"response"`
}

type ModifyDomainCertificateRequest struct {
	*common.BaseRequest

	DomainId      string `json:"domainId,omitempty"`
	CertificateId string `json:"certificateId,omitempty"`
}

type ModifyDomainCertificateResponse struct {
	*common.BaseResponse

	RequestId string `json:"requestId,omitempty"`

	Response struct {
		RequestId string `json:"requestId,omitempty"`
	} `json:"response"`
}
