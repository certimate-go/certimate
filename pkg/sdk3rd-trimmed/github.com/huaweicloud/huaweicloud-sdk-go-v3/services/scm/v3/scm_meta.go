package v3

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	scm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3"
)

func GenReqDefForExportCertificate() *def.HttpRequestDef {
	return scm.GenReqDefForExportCertificate()
}

func GenReqDefForImportCertificate() *def.HttpRequestDef {
	return scm.GenReqDefForImportCertificate()
}

func GenReqDefForListCertificates() *def.HttpRequestDef {
	return scm.GenReqDefForListCertificates()
}
