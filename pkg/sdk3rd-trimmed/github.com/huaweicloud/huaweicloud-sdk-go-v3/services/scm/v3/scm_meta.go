package v3

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	hwscm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3"
)

func GenReqDefForExportCertificate() *def.HttpRequestDef {
	return hwscm.GenReqDefForExportCertificate()
}

func GenReqDefForImportCertificate() *def.HttpRequestDef {
	return hwscm.GenReqDefForImportCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return hwscm.GenReqDefForListCertificates()
}
