package azure

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

func IsPublicEnv(env string) bool {
	switch strings.ToLower(env) {
	case "", "default", "public", "azurecloud":
		return true
	default:
		return false
	}
}

func IsUSGovernmentEnv(env string) bool {
	switch strings.ToLower(env) {
	case "usgovernment", "government", "azureusgovernment", "azuregovernment":
		return true
	default:
		return false
	}
}

func IsChinaEnv(env string) bool {
	switch strings.ToLower(env) {
	case "china", "chinacloud", "azurechina", "azurechinacloud":
		return true
	default:
		return false
	}
}

func GetCloudEnvConfiguration(env string) (cloud.Configuration, error) {
	if IsPublicEnv(env) {
		return cloud.AzurePublic, nil
	} else if IsUSGovernmentEnv(env) {
		return cloud.AzureGovernment, nil
	} else if IsChinaEnv(env) {
		return cloud.AzureChina, nil
	}

	return cloud.Configuration{}, fmt.Errorf("unknown azure cloud environment %s", env)
}
