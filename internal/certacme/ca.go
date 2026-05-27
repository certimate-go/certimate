package certacme

import (
	"fmt"
	"strings"

	"github.com/certimate-go/certimate/internal/domain"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

var caDirUrls = map[string]string{
	domain.CAProviderTypeLetsEncrypt.String():         "https://acme-v02.api.letsencrypt.org/directory",
	domain.CAProviderTypeLetsEncryptStaging.String():  "https://acme-staging-v02.api.letsencrypt.org/directory",
	domain.CAProviderTypeActalisSSL.String():          "https://acme-api.actalis.com/acme/directory",
	domain.CAProviderTypeDigiCert.String():            "https://acme.digicert.com/v2/acme/directory",
	domain.CAProviderTypeGlobalSignAtlas.String():     "https://emea.acme.atlas.globalsign.com/directory",
	domain.CAProviderTypeGoogleTrustServices.String(): "https://dv.acme-v02.api.pki.goog/directory",
	domain.CAProviderTypeLiteSSL.String():             "https://acme.litessl.com/acme/v2/directory",
	domain.CAProviderTypeSSLCom.String():              "https://acme.ssl.com/sslcom-dv-rsa",
	domain.CAProviderTypeSSLCom.String() + "RSA":      "https://acme.ssl.com/sslcom-dv-rsa",
	domain.CAProviderTypeSSLCom.String() + "ECC":      "https://acme.ssl.com/sslcom-dv-ecc",
	domain.CAProviderTypeSectigo.String():             "https://acme.sectigo.com/v2/DV",
	domain.CAProviderTypeSectigo.String() + "DV":      "https://acme.sectigo.com/v2/DV",
	domain.CAProviderTypeSectigo.String() + "OV":      "https://acme.sectigo.com/v2/OV",
	domain.CAProviderTypeSectigo.String() + "EV":      "https://acme.sectigo.com/v2/EV",
	domain.CAProviderTypeZeroSSL.String():             "https://acme.zerossl.com/v2/DV90",
}

func getCADirUrl(providerType domain.CAProviderType, providerAccessConfig map[string]any, keyAlgorithm domain.CertificateKeyAlgorithmType) (string, error) {
	switch providerType {
	case domain.CAProviderTypeSectigo:
		credentials := &domain.AccessConfigForGlobalSectigo{}
		if err := xmaps.Populate(providerAccessConfig, &credentials); err != nil {
			return "", err
		} else if strings.EqualFold(credentials.ValidationType, "DV") {
			return caDirUrls[domain.CAProviderTypeSectigo.String()+"DV"], nil
		} else if strings.EqualFold(credentials.ValidationType, "OV") {
			return caDirUrls[domain.CAProviderTypeSectigo.String()+"OV"], nil
		} else if strings.EqualFold(credentials.ValidationType, "EV") {
			return caDirUrls[domain.CAProviderTypeSectigo.String()+"EV"], nil
		} else {
			return caDirUrls[domain.CAProviderTypeSectigo.String()], nil
		}

	case domain.CAProviderTypeSSLCom:
		if strings.HasPrefix(keyAlgorithm.String(), "RSA") {
			return caDirUrls[domain.CAProviderTypeSSLCom.String()+"RSA"], nil
		} else if strings.HasPrefix(keyAlgorithm.String(), "EC") {
			return caDirUrls[domain.CAProviderTypeSSLCom.String()+"ECC"], nil
		} else {
			return caDirUrls[domain.CAProviderTypeSSLCom.String()], nil
		}

	case domain.CAProviderTypeACMECA:
		credentials := &domain.AccessConfigForACMECA{}
		if err := xmaps.Populate(providerAccessConfig, &credentials); err != nil {
			return "", err
		} else if credentials.Endpoint == "" {
			return "", fmt.Errorf("the endpoint of custom ACME CA is empty")
		}
		return credentials.Endpoint, nil

	default:
		endpoint := caDirUrls[providerType.String()]
		if endpoint == "" {
			return "", fmt.Errorf("the endpoint of the ACME CA provider is empty")
		}
		return endpoint, nil
	}
}
