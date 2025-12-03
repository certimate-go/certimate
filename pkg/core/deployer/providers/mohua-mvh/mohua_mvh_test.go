package mohuamvh_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	provider "github.com/certimate-go/certimate/pkg/core/deployer/providers/mohua-mvh"
)

var (
	fInputCertPath string
	fInputKeyPath  string
	fUsername     string
	fApiPassword     string
	fHostID        string
	fDomainID      string
)

func init() {
	argsPrefix := "MOHUAMVH_"

	flag.StringVar(&fInputCertPath, argsPrefix+"INPUTCERTPATH", "", "Path to certificate PEM file")
	flag.StringVar(&fInputKeyPath, argsPrefix+"INPUTKEYPATH", "", "Path to private key PEM file")
	flag.StringVar(&fUsername, argsPrefix+"USERNAME", "", "Mohua Cloud Access Key")
	flag.StringVar(&fApiPassword, argsPrefix+"APIPASSWORD", "", "Mohua Cloud Secret Key")
	flag.StringVar(&fHostID, argsPrefix+"HOSTID", "", "Virtual Host ID")
	flag.StringVar(&fDomainID, argsPrefix+"DOMAINID", "", "Domain ID (integer)")
}

/*
Shell command to run this test:

	go test -v ./mohuamvh_test.go -args \
	--MOHUAMVH_INPUTCERTPATH="/path/to/your-input-cert.pem" \
	--MOHUAMVH_INPUTKEYPATH="/path/to/your-input-key.pem" \
	--MOHUAMVH_USERNAME="your-access-key" \
	--MOHUAMVH_APIPASSWORD="your-secret-key" \
	--MOHUAMVH_HOSTID="your-virtual-host-id" \
	--MOHUAMVH_DOMAINID="123"  # Domain ID should be an integer
*/
func TestDeploy(t *testing.T) {
	flag.Parse()

	t.Run("Deploy", func(t *testing.T) {
		t.Log(strings.Join([]string{
			"args:",
			fmt.Sprintf("INPUTCERTPATH: %v", fInputCertPath),
			fmt.Sprintf("INPUTKEYPATH: %v", fInputKeyPath),
			fmt.Sprintf("USERNAME: %v", fUsername),
			fmt.Sprintf("APIPASSWORD: %v", strings.Repeat("*", len(fApiPassword))), // Hide secret key for security
			fmt.Sprintf("HOSTID: %v", fHostID),
			fmt.Sprintf("DOMAINID: %v", fDomainID),
		}, "\n"))

		// Validate required parameters
		if fUsername == "" {
			t.Skip("Skipping test: MOHUAMVH_USERNAME is required")
		}
		if fApiPassword == "" {
			t.Skip("Skipping test: MOHUAMVH_APIPASSWORD is required")
		}
		if fHostID == "" {
			t.Skip("Skipping test: MOHUAMVH_HOSTID is required")
		}
		if fDomainID == "" {
			t.Skip("Skipping test: MOHUAMVH_DOMAINID is required")
		}
		if fInputCertPath == "" {
			t.Skip("Skipping test: MOHUAMVH_INPUTCERTPATH is required")
		}
		if fInputKeyPath == "" {
			t.Skip("Skipping test: MOHUAMVH_INPUTKEYPATH is required")
		}

		// Create provider
		deployer, err := provider.NewDeployer(&provider.DeployerConfig{
			Username: fUsername,
			ApiPassword: fApiPassword,
			HostID:    fHostID,
			DomainID:  fDomainID,
		})
		if err != nil {
			t.Errorf("Failed to create deployer: %+v", err)
			return
		}

		// Read certificate and private key files
		certData, err := os.ReadFile(fInputCertPath)
		if err != nil {
			t.Errorf("Failed to read certificate file: %+v", err)
			return
		}

		keyData, err := os.ReadFile(fInputKeyPath)
		if err != nil {
			t.Errorf("Failed to read private key file: %+v", err)
			return
		}

		// Deploy certificate
		res, err := deployer.Deploy(context.Background(), string(certData), string(keyData))
		if err != nil {
			t.Errorf("Failed to deploy certificate: %+v", err)
			return
		}

		t.Logf("Deployment successful: %v", res)
	})
}

// TestNewDeployerWithInvalidConfig 测试无效配置
func TestNewDeployerWithInvalidConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config *provider.DeployerConfig
		expect string
	}{
		{
			name:   "nil config",
			config: nil,
			expect: "nil",
		},
		{
			name: "empty access key",
			config: &provider.DeployerConfig{
				Username: "",
				ApiPassword: "secret",
				HostID:    "host123",
				DomainID:  "456",
			},
			expect: "username",
		},
		{
			name: "empty secret key",
			config: &provider.DeployerConfig{
				Username: "access",
				ApiPassword: "",
				HostID:    "host123",
				DomainID:  "456",
			},
			expect: "apiPassword",
		},
		{
			name: "empty host ID",
			config: &provider.DeployerConfig{
				Username: "access",
				ApiPassword: "secret",
				HostID:    "",
				DomainID:  "456",
			},
			expect: "hostID",
		},
		{
			name: "empty domain ID",
			config: &provider.DeployerConfig{
				Username: "access",
				ApiPassword: "secret",
				HostID:    "host123",
				DomainID:  "",
			},
			expect: "domainID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.NewDeployer(tc.config)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			} else {
				t.Logf("Expected error: %v", err)
			}
		})
	}
}