package common

import (
	"fmt"
	"regexp"
	"strings"
)

type EndpointVariant int

const (
	EndpointVariantNone EndpointVariant = iota
	EndpointVariantFIPS
	EndpointVariantDualStack
)

type EndpointPartition struct {
	ID          string
	RegionRegex *regexp.Regexp
	Templates   map[EndpointVariant]string
}

var endpointPartitions = []EndpointPartition{
	{
		ID:          "aws",
		RegionRegex: regexp.MustCompile(`^(us|eu|ap|sa|ca|me|af|il|mx)-\w+-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone:                            "https://{service}.{region}.amazonaws.com",
			EndpointVariantFIPS:                            "https://{service}-fips.{region}.amazonaws.com",
			EndpointVariantDualStack:                       "https://{service}.{region}.api.aws",
			EndpointVariantFIPS | EndpointVariantDualStack: "https://{service}-fips.{region}.api.aws",
		},
	},
	{
		ID:          "aws-cn",
		RegionRegex: regexp.MustCompile(`^cn-\w+-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone:                            "https://{service}.{region}.amazonaws.com.cn",
			EndpointVariantFIPS:                            "https://{service}-fips.{region}.amazonaws.com.cn",
			EndpointVariantDualStack:                       "https://{service}.{region}.api.amazonwebservices.com.cn",
			EndpointVariantFIPS | EndpointVariantDualStack: "https://{service}-fips.{region}.api.amazonwebservices.com.cn",
		},
	},
	{
		ID:          "aws-eusc",
		RegionRegex: regexp.MustCompile(`^eusc\-(de)\-\w+\-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone:                            "https://{service}.{region}.amazonaws.eu",
			EndpointVariantFIPS:                            "https://{service}-fips.{region}.amazonaws.eu",
			EndpointVariantDualStack:                       "https://{service}.{region}.api.amazonwebservices.eu",
			EndpointVariantFIPS | EndpointVariantDualStack: "https://{service}-fips.{region}.api.amazonwebservices.eu",
		},
	},
	{
		ID:          "aws-iso",
		RegionRegex: regexp.MustCompile(`^us\-iso\-\w+\-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone: "https://{service}.{region}.c2s.ic.gov",
			EndpointVariantFIPS: "https://{service}-fips.{region}.c2s.ic.gov",
		},
	},
	{
		ID:          "aws-iso-b",
		RegionRegex: regexp.MustCompile(`^us\-isob\-\w+\-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone: "https://{service}.{region}.sc2s.sgov.gov",
			EndpointVariantFIPS: "https://{service}-fips.{region}.sc2s.sgov.gov",
		},
	},
	{
		ID:          "aws-iso-e",
		RegionRegex: regexp.MustCompile(`^us\-isoe\-\w+\-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone: "https://{service}.{region}.cloud.adc-e.uk",
			EndpointVariantFIPS: "https://{service}-fips.{region}.cloud.adc-e.uk",
		},
	},
	{
		ID:          "aws-iso-f",
		RegionRegex: regexp.MustCompile(`^us\-isof\-\w+\-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone: "https://{service}.{region}.csp.hci.ic.gov",
			EndpointVariantFIPS: "https://{service}-fips.{region}.csp.hci.ic.gov",
		},
	},
	{
		ID:          "aws-us-gov",
		RegionRegex: regexp.MustCompile(`^us-gov-\w+-\d+$`),
		Templates: map[EndpointVariant]string{
			EndpointVariantNone:      "https://{service}.{region}.amazonaws.com",
			EndpointVariantFIPS:      "https://{service}-fips.{region}.amazonaws.com",
			EndpointVariantDualStack: "https://{service}.{region}.api.aws",
		},
	},
}

var deprecatedFIPSRegions = map[string]string{
	"us-east-1-fips":    "us-east-1",
	"us-east-2-fips":    "us-east-2",
	"us-west-1-fips":    "us-west-1",
	"us-west-2-fips":    "us-west-2",
	"ca-central-1-fips": "ca-central-1",
	"ca-west-1-fips":    "ca-west-1",
}

func ResolveBaseEndpoint(service, region string, variant EndpointVariant) (string, error) {
	if service == "" {
		return "", fmt.Errorf("service could not be empty")
	}

	if r, ok := deprecatedFIPSRegions[region]; ok {
		region = r
		variant |= EndpointVariantFIPS
	}

	for _, p := range endpointPartitions {
		if p.RegionRegex.MatchString(region) {
			tpl, ok := p.Templates[variant]
			if ok {
				endpoint := tpl
				endpoint = strings.Replace(endpoint, "{service}", service, -1)
				if region == "" {
					endpoint = strings.Replace(endpoint, ".{region}.", ".", -1)
				} else {
					endpoint = strings.Replace(endpoint, "{region}", region, -1)
				}
				return endpoint, nil
			}
		}
	}

	return "", fmt.Errorf("unable to resolve endpoint for service %s in region %s", service, region)
}
