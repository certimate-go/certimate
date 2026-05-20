package azurekeyvault_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/azure-keyvault"
)

var (
	fp            = tester.Args("AZUREKEYVAULT_")
	fTestCertPath string
	fTestKeyPath  string
	fTenantId     string
	fClientId     string
	fClientSecret string
	fCloudName    string
	fKeyVaultName string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fTenantId, "TENANTID")
	fp.DefineString(&fClientId, "CLIENTID")
	fp.DefineString(&fClientSecret, "CLIENTSECRET")
	fp.DefineString(&fCloudName, "CLOUDNAME")
	fp.DefineString(&fKeyVaultName, "KEYVAULTNAME")
}

/*
Shell command to run this test:

	go test -v ./azure_keyvault_test.go -args \
	--AZUREKEYVAULT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--AZUREKEYVAULT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--AZUREKEYVAULT_TENANTID="your-tenant-id" \
	--AZUREKEYVAULT_CLIENTID="your-app-registration-client-id" \
	--AZUREKEYVAULT_CLIENTSECRET="your-app-registration-client-secret" \
	--AZUREKEYVAULT_CLOUDNAME="china" \
	--AZUREKEYVAULT_KEYVAULTNAME="your-keyvault-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			TenantId:     fTenantId,
			ClientId:     fClientId,
			ClientSecret: fClientSecret,
			CloudName:    fCloudName,
			KeyVaultName: fKeyVaultName,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
