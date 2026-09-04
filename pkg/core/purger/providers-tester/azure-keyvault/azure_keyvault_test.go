//go:build tester

package azurekeyvault_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tester "github.com/certimate-go/certimate/pkg/core/purger/providers-tester"
	impl "github.com/certimate-go/certimate/pkg/core/purger/providers/azure-keyvault"
)

var (
	fp            = tester.InitArgs("AZUREKEYVAULT_")
	fTenantId     string
	fClientId     string
	fClientSecret string
	fCloudName    string
	fKeyVaultName string
)

func init() {
	fp.DefineString(&fTenantId, "TENANTID")
	fp.DefineString(&fClientId, "CLIENTID")
	fp.DefineString(&fClientSecret, "CLIENTSECRET")
	fp.DefineString(&fCloudName, "CLOUDNAME")
	fp.DefineString(&fKeyVaultName, "KEYVAULTNAME")
}

/*
Shell command to run this test:

	go test -tags=tester -v ./azure_keyvault_test.go -args \
	--AZUREKEYVAULT_TENANTID="your-tenant-id" \
	--AZUREKEYVAULT_CLIENTID="your-app-registration-client-id" \
	--AZUREKEYVAULT_CLIENTSECRET="your-app-registration-client-secret" \
	--AZUREKEYVAULT_CLOUDNAME="china" \
	--AZUREKEYVAULT_KEYVAULTNAME="your-keyvault-name"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Purge", func(t *testing.T) {
		provider, err := impl.NewPurger(&impl.PurgerConfig{
			TenantId:     fTenantId,
			ClientId:     fClientId,
			ClientSecret: fClientSecret,
			CloudName:    fCloudName,
			KeyVaultName: fKeyVaultName,
		})
		require.NoError(t, err)

		tester.Purge(t, provider, tester.PurgeInput{})
	})
}
