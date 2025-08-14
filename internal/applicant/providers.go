package applicant

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/certimate-go/certimate/internal/domain"
	pACMEDNS "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/acmedns"
	pACMEHttpReq "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/acmehttpreq"
	pAliyun "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/aliyun"
	pAliyunESA "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/aliyun-esa"
	pAWSRoute53 "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/aws-route53"
	pAzureDNS "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/azure-dns"
	pBaiduCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/baiducloud"
	pBunny "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/bunny"
	pCloudflare "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/cloudflare"
	pClouDNS "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/cloudns"
	pCMCCCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/cmcccloud"
	pConstellix "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/constellix"
	pCTCCCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/ctcccloud"
	pDeSEC "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/desec"
	pDigitalOcean "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/digitalocean"
	pDNSLA "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/dnsla"
	pDuckDNS "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/duckdns"
	pDynv6 "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/dynv6"
	pGcore "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/gcore"
	pGname "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/gname"
	pGoDaddy "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/godaddy"
	pHetzner "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/hetzner"
	pHuaweiCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/huaweicloud"
	pJDCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/jdcloud"
	pNamecheap "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/namecheap"
	pNameDotCom "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/namedotcom"
	pNameSilo "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/namesilo"
	pNetcup "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/netcup"
	pNetlify "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/netlify"
	pNS1 "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/ns1"
	pPorkbun "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/porkbun"
	pPowerDNS "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/powerdns"
	pRainYun "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/rainyun"
	pSpaceship "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/spaceship"
	pTencentCloud "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/tencentcloud"
	pTencentCloudEO "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/tencentcloud-eo"
	pUCloudUDNR "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/ucloud-udnr"
	pVercel "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/vercel"
	pVolcEngine "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/volcengine"
	pWestcn "github.com/certimate-go/certimate/pkg/core/ssl-applicator/acme-dns01/providers/westcn"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type applicantProviderOptions struct {
	Domains                 []string
	ContactEmail            string
	Provider                domain.ACMEDns01ProviderType
	ProviderAccessConfig    map[string]any
	ProviderServiceConfig   map[string]any
	CAProvider              domain.CAProviderType
	CAProviderAccessId      string
	CAProviderAccessConfig  map[string]any
	CAProviderServiceConfig map[string]any
	KeyAlgorithm            string
	Nameservers             []string
	DnsPropagationWait      int32
	DnsPropagationTimeout   int32
	DnsTTL                  int32
	ACMEProfile             string
	DisableFollowCNAME      bool
	ARIReplaceAcct          string
	ARIReplaceCert          string
}

func createApplicantProvider(options *applicantProviderOptions) (challenge.Provider, error) {
	/*
	  注意：如果追加新的常量值，请保持以 ASCII 排序。
	  NOTICE: If you add new constant, please keep ASCII order.
	*/
	switch options.Provider {
	case domain.ACMEDns01ProviderTypeACMEDNS:
		{
			access := domain.AccessConfigForACMEDNS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pACMEDNS.NewChallengeProvider(&pACMEDNS.ChallengeProviderConfig{
				ApiBase:        access.ApiBase,
				StorageBaseUrl: access.StorageBaseUrl,
				StoragePath:    access.StoragePath,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeACMEHttpReq:
		{
			access := domain.AccessConfigForACMEHttpReq{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pACMEHttpReq.NewChallengeProvider(&pACMEHttpReq.ChallengeProviderConfig{
				Endpoint:              access.Endpoint,
				Mode:                  access.Mode,
				Username:              access.Username,
				Password:              access.Password,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeAliyun, domain.ACMEDns01ProviderTypeAliyunDNS, domain.ACMEDns01ProviderTypeAliyunESA:
		{
			access := domain.AccessConfigForAliyun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.ACMEDns01ProviderTypeAliyun, domain.ACMEDns01ProviderTypeAliyunDNS:
				applicant, err := pAliyun.NewChallengeProvider(&pAliyun.ChallengeProviderConfig{
					AccessKeyId:           access.AccessKeyId,
					AccessKeySecret:       access.AccessKeySecret,
					DnsPropagationTimeout: options.DnsPropagationTimeout,
					DnsTTL:                options.DnsTTL,
				})
				return applicant, err

			case domain.ACMEDns01ProviderTypeAliyunESA:
				applicant, err := pAliyunESA.NewChallengeProvider(&pAliyunESA.ChallengeProviderConfig{
					AccessKeyId:           access.AccessKeyId,
					AccessKeySecret:       access.AccessKeySecret,
					Region:                xmaps.GetString(options.ProviderServiceConfig, "region"),
					DnsPropagationTimeout: options.DnsPropagationTimeout,
					DnsTTL:                options.DnsTTL,
				})
				return applicant, err

			default:
				break
			}
		}

	case domain.ACMEDns01ProviderTypeAWS, domain.ACMEDns01ProviderTypeAWSRoute53:
		{
			access := domain.AccessConfigForAWS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pAWSRoute53.NewChallengeProvider(&pAWSRoute53.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				SecretAccessKey:       access.SecretAccessKey,
				Region:                xmaps.GetString(options.ProviderServiceConfig, "region"),
				HostedZoneId:          xmaps.GetString(options.ProviderServiceConfig, "hostedZoneId"),
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeAzure, domain.ACMEDns01ProviderTypeAzureDNS:
		{
			access := domain.AccessConfigForAzure{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pAzureDNS.NewChallengeProvider(&pAzureDNS.ChallengeProviderConfig{
				TenantId:              access.TenantId,
				ClientId:              access.ClientId,
				ClientSecret:          access.ClientSecret,
				CloudName:             access.CloudName,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeBaiduCloud, domain.ACMEDns01ProviderTypeBaiduCloudDNS:
		{
			access := domain.AccessConfigForBaiduCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pBaiduCloud.NewChallengeProvider(&pBaiduCloud.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				SecretAccessKey:       access.SecretAccessKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeBunny:
		{
			access := domain.AccessConfigForBunny{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pBunny.NewChallengeProvider(&pBunny.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeCloudflare:
		{
			access := domain.AccessConfigForCloudflare{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pCloudflare.NewChallengeProvider(&pCloudflare.ChallengeProviderConfig{
				DnsApiToken:           access.DnsApiToken,
				ZoneApiToken:          access.ZoneApiToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeClouDNS:
		{
			access := domain.AccessConfigForClouDNS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pClouDNS.NewChallengeProvider(&pClouDNS.ChallengeProviderConfig{
				AuthId:                access.AuthId,
				AuthPassword:          access.AuthPassword,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeCMCCCloud, domain.ACMEDns01ProviderTypeCMCCCloudDNS:
		{
			access := domain.AccessConfigForCMCCCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pCMCCCloud.NewChallengeProvider(&pCMCCCloud.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				AccessKeySecret:       access.AccessKeySecret,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeConstellix:
		{
			access := domain.AccessConfigForConstellix{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pConstellix.NewChallengeProvider(&pConstellix.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				SecretKey:             access.SecretKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeCTCCCloud, domain.ACMEDns01ProviderTypeCTCCCloudSmartDNS:
		{
			access := domain.AccessConfigForCTCCCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pCTCCCloud.NewChallengeProvider(&pCTCCCloud.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				SecretAccessKey:       access.SecretAccessKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeDeSEC:
		{
			access := domain.AccessConfigForDeSEC{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pDeSEC.NewChallengeProvider(&pDeSEC.ChallengeProviderConfig{
				Token:                 access.Token,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeDigitalOcean:
		{
			access := domain.AccessConfigForDigitalOcean{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pDigitalOcean.NewChallengeProvider(&pDigitalOcean.ChallengeProviderConfig{
				AccessToken:           access.AccessToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeDNSLA:
		{
			access := domain.AccessConfigForDNSLA{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pDNSLA.NewChallengeProvider(&pDNSLA.ChallengeProviderConfig{
				ApiId:                 access.ApiId,
				ApiSecret:             access.ApiSecret,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeDuckDNS:
		{
			access := domain.AccessConfigForDuckDNS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pDuckDNS.NewChallengeProvider(&pDuckDNS.ChallengeProviderConfig{
				Token:                 access.Token,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeDynv6:
		{
			access := domain.AccessConfigForDynv6{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pDynv6.NewChallengeProvider(&pDynv6.ChallengeProviderConfig{
				HttpToken:             access.HttpToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeGcore:
		{
			access := domain.AccessConfigForGcore{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pGcore.NewChallengeProvider(&pGcore.ChallengeProviderConfig{
				ApiToken:              access.ApiToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeGname:
		{
			access := domain.AccessConfigForGname{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pGname.NewChallengeProvider(&pGname.ChallengeProviderConfig{
				AppId:                 access.AppId,
				AppKey:                access.AppKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeGoDaddy:
		{
			access := domain.AccessConfigForGoDaddy{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pGoDaddy.NewChallengeProvider(&pGoDaddy.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				ApiSecret:             access.ApiSecret,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeHetzner:
		{
			access := domain.AccessConfigForHetzner{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pHetzner.NewChallengeProvider(&pHetzner.ChallengeProviderConfig{
				ApiToken:              access.ApiToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeHuaweiCloud, domain.ACMEDns01ProviderTypeHuaweiCloudDNS:
		{
			access := domain.AccessConfigForHuaweiCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pHuaweiCloud.NewChallengeProvider(&pHuaweiCloud.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				SecretAccessKey:       access.SecretAccessKey,
				Region:                xmaps.GetString(options.ProviderServiceConfig, "region"),
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeJDCloud, domain.ACMEDns01ProviderTypeJDCloudDNS:
		{
			access := domain.AccessConfigForJDCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pJDCloud.NewChallengeProvider(&pJDCloud.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				AccessKeySecret:       access.AccessKeySecret,
				RegionId:              xmaps.GetString(options.ProviderServiceConfig, "regionId"),
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNamecheap:
		{
			access := domain.AccessConfigForNamecheap{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNamecheap.NewChallengeProvider(&pNamecheap.ChallengeProviderConfig{
				Username:              access.Username,
				ApiKey:                access.ApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNameDotCom:
		{
			access := domain.AccessConfigForNameDotCom{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNameDotCom.NewChallengeProvider(&pNameDotCom.ChallengeProviderConfig{
				Username:              access.Username,
				ApiToken:              access.ApiToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNameSilo:
		{
			access := domain.AccessConfigForNameSilo{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNameSilo.NewChallengeProvider(&pNameSilo.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNetcup:
		{
			access := domain.AccessConfigForNetcup{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNetcup.NewChallengeProvider(&pNetcup.ChallengeProviderConfig{
				CustomerNumber:        access.CustomerNumber,
				ApiKey:                access.ApiKey,
				ApiPassword:           access.ApiPassword,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNetlify:
		{
			access := domain.AccessConfigForNetlify{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNetlify.NewChallengeProvider(&pNetlify.ChallengeProviderConfig{
				ApiToken:              access.ApiToken,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeNS1:
		{
			access := domain.AccessConfigForNS1{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pNS1.NewChallengeProvider(&pNS1.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypePorkbun:
		{
			access := domain.AccessConfigForPorkbun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pPorkbun.NewChallengeProvider(&pPorkbun.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				SecretApiKey:          access.SecretApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypePowerDNS:
		{
			access := domain.AccessConfigForPowerDNS{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pPowerDNS.NewChallengeProvider(&pPowerDNS.ChallengeProviderConfig{
				ServerUrl:                access.ServerUrl,
				ApiKey:                   access.ApiKey,
				AllowInsecureConnections: access.AllowInsecureConnections,
				DnsPropagationTimeout:    options.DnsPropagationTimeout,
				DnsTTL:                   options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeRainYun:
		{
			access := domain.AccessConfigForRainYun{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pRainYun.NewChallengeProvider(&pRainYun.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeSpaceship:
		{
			access := domain.AccessConfigForSpaceship{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pSpaceship.NewChallengeProvider(&pSpaceship.ChallengeProviderConfig{
				ApiKey:                access.ApiKey,
				ApiSecret:             access.ApiSecret,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeTencentCloud, domain.ACMEDns01ProviderTypeTencentCloudDNS, domain.ACMEDns01ProviderTypeTencentCloudEO:
		{
			access := domain.AccessConfigForTencentCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			switch options.Provider {
			case domain.ACMEDns01ProviderTypeTencentCloud, domain.ACMEDns01ProviderTypeTencentCloudDNS:
				applicant, err := pTencentCloud.NewChallengeProvider(&pTencentCloud.ChallengeProviderConfig{
					SecretId:              access.SecretId,
					SecretKey:             access.SecretKey,
					DnsPropagationTimeout: options.DnsPropagationTimeout,
					DnsTTL:                options.DnsTTL,
				})
				return applicant, err

			case domain.ACMEDns01ProviderTypeTencentCloudEO:
				applicant, err := pTencentCloudEO.NewChallengeProvider(&pTencentCloudEO.ChallengeProviderConfig{
					SecretId:              access.SecretId,
					SecretKey:             access.SecretKey,
					ZoneId:                xmaps.GetString(options.ProviderServiceConfig, "zoneId"),
					DnsPropagationTimeout: options.DnsPropagationTimeout,
					DnsTTL:                options.DnsTTL,
				})
				return applicant, err

			default:
				break
			}
		}

	case domain.ACMEDns01ProviderTypeUCloudUDNR:
		{
			access := domain.AccessConfigForUCloud{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pUCloudUDNR.NewChallengeProvider(&pUCloudUDNR.ChallengeProviderConfig{
				PrivateKey:            access.PrivateKey,
				PublicKey:             access.PublicKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeVercel:
		{
			access := domain.AccessConfigForVercel{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pVercel.NewChallengeProvider(&pVercel.ChallengeProviderConfig{
				ApiAccessToken:        access.ApiAccessToken,
				TeamId:                access.TeamId,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeVolcEngine, domain.ACMEDns01ProviderTypeVolcEngineDNS:
		{
			access := domain.AccessConfigForVolcEngine{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pVolcEngine.NewChallengeProvider(&pVolcEngine.ChallengeProviderConfig{
				AccessKeyId:           access.AccessKeyId,
				SecretAccessKey:       access.SecretAccessKey,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}

	case domain.ACMEDns01ProviderTypeWestcn:
		{
			access := domain.AccessConfigForWestcn{}
			if err := xmaps.Populate(options.ProviderAccessConfig, &access); err != nil {
				return nil, fmt.Errorf("failed to populate provider access config: %w", err)
			}

			applicant, err := pWestcn.NewChallengeProvider(&pWestcn.ChallengeProviderConfig{
				Username:              access.Username,
				ApiPassword:           access.ApiPassword,
				DnsPropagationTimeout: options.DnsPropagationTimeout,
				DnsTTL:                options.DnsTTL,
			})
			return applicant, err
		}
	}

	return nil, fmt.Errorf("unsupported applicant provider '%s'", string(options.Provider))
}
