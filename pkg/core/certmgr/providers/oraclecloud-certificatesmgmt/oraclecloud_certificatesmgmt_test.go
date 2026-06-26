package oraclecloudcertificatesmgmt_test

import (
	"testing"

	"github.com/certimate-go/certimate/pkg/core/certmgr/internal/tester"
	impl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/oraclecloud-certificatesmgmt"
)

var (
	fp                    = tester.Args("ORACLECLOUDCERTIFICATESMGMT_")
	fTestCertPath         string
	fTestKeyPath          string
	fPrivateKey           string
	fPrivateKeyPassphrase string
	fPublicKeyFingerprint string
	fTenancyOcid          string
	fUserOcid             string
	fCompartmentOcid      string
)

func init() {
	fp.DefineString(&fTestCertPath, "TESTCERTPATH")
	fp.DefineString(&fTestKeyPath, "TESTKEYPATH")
	fp.DefineString(&fPrivateKey, "PRIVATEKEY")
	fp.DefineString(&fPrivateKeyPassphrase, "PRIVATEKEYPASSPHRASE")
	fp.DefineString(&fPublicKeyFingerprint, "PUBLICKEYFINGERPRINT")
	fp.DefineString(&fTenancyOcid, "TENANCYOCID")
	fp.DefineString(&fUserOcid, "USEROCID")
	fp.DefineString(&fCompartmentOcid, "COMPARTMENTOCID")
}

/*
Shell command to run this test:

	go test -v ./oraclecloud_certificatesmgmt_test.go -args \
	--ORACLECLOUDCERTIFICATESMGMT_TESTCERTPATH="/path/to/your-test-cert.pem" \
	--ORACLECLOUDCERTIFICATESMGMT_TESTKEYPATH="/path/to/your-test-key.pem" \
	--ORACLECLOUDCERTIFICATESMGMT_PRIVATEKEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----" \
	--ORACLECLOUDCERTIFICATESMGMT_PRIVATEKEYPASSPHRASE="" \
	--ORACLECLOUDCERTIFICATESMGMT_PUBLICKEYFINGERPRINT="00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00" \
	--ORACLECLOUDCERTIFICATESMGMT_TENANCYOCID="ocid1.tenancy.oc1..secret" \
	--ORACLECLOUDCERTIFICATESMGMT_USEROCID="ocid1.user.oc1..secret" \
	--ORACLECLOUDCERTIFICATESMGMT_COMPARTMENTOCID="ocid1.tenancy.oc1..secret"
*/
func TestProvider(t *testing.T) {
	fp.Parse()

	t.Run("Upload", func(t *testing.T) {
		provider, err := impl.NewCertmgr(&impl.CertmgrConfig{
			AuthMethod:           impl.AUTH_METHOD_APIKEY,
			PrivateKey:           fPrivateKey,
			PrivateKeyPassphrase: fPrivateKeyPassphrase,
			PublicKeyFingerprint: fPublicKeyFingerprint,
			TenancyOcid:          fTenancyOcid,
			UserOcid:             fUserOcid,
			CompartmentOcid:      fCompartmentOcid,
		})
		if err != nil {
			t.Errorf("err: %+v", err)
			return
		}

		tester.TestUpload(t, provider, tester.TestUploadArgs{CertPath: fTestCertPath, KeyPath: fTestKeyPath})
	})
}
