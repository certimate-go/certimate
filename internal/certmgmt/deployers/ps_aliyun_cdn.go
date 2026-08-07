package deployers

import (
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

func init() {
	schema, err := providerschema.New(string(domain.DeploymentProviderTypeAliyunCDN), providerschema.CategoryDeploy).
		Field("region", providerschema.ValueTypeAutocomplete,
			providerschema.LabelKey("workflow_node.deploy.form.aliyun_cdn_region.label"),
			providerschema.PlaceholderKey("workflow_node.deploy.form.aliyun_cdn_region.placeholder"),
			providerschema.TooltipKey("workflow_node.deploy.form.aliyun_cdn_region.tooltip"),
			providerschema.TooltipHtml(),
			providerschema.Options("cn-hangzhou", "ap-southeast-1"),
			providerschema.WithFilterMode(providerschema.FilterModePrefix),
		).
		Field("domainMatchPattern", providerschema.ValueTypeRadio,
			providerschema.LabelKey("workflow_node.deploy.form.shared_domain_match_pattern.label"),
			providerschema.OptionsWith(
				providerschema.Option{Value: "exact", LabelKey: "workflow_node.deploy.form.shared_domain_match_pattern.option.exact.label"},
				providerschema.Option{Value: "wildcard", LabelKey: "workflow_node.deploy.form.shared_domain_match_pattern.option.wildcard.label"},
				providerschema.Option{Value: "certsan", LabelKey: "workflow_node.deploy.form.shared_domain_match_pattern.option.certsan.label"},
			),
			providerschema.Default("exact"),
			providerschema.Required(),
		).
		Field("domain", providerschema.ValueTypeText,
			providerschema.LabelKey("workflow_node.deploy.form.aliyun_cdn_domain.label"),
			providerschema.PlaceholderKey("workflow_node.deploy.form.aliyun_cdn_domain.placeholder"),
			providerschema.VisibleWhen("domainMatchPattern").NotIn("certsan"),
			providerschema.RequiredWhen("domainMatchPattern").In("exact", "wildcard"),
			providerschema.ValidateWhen("domainMatchPattern").In("exact", "wildcard").Validator(providerschema.ValidatorDomain, providerschema.WithParam("allowWildcard", true)),
		).
		Build()
	if err != nil {
		panic(err)
	}

	providerschema.Registries.MustRegister(string(domain.DeploymentProviderTypeAliyunCDN), schema)
}
