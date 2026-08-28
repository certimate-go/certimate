package accessschema

import (
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/providerschema"
)

func init() {
	schema, err := providerschema.New(string(domain.AccessProviderTypeAliyun), providerschema.CategoryAccess).
		Field("accessKeyId", providerschema.ValueTypeText,
			providerschema.LabelKey("access.form.aliyun_access_key_id.label"),
			providerschema.PlaceholderKey("access.form.aliyun_access_key_id.placeholder"),
			providerschema.TooltipKey("access.form.aliyun_access_key_id.tooltip"),
			providerschema.TooltipHtml(),
			providerschema.Secret(),
			providerschema.Required(),
		).
		Field("accessKeySecret", providerschema.ValueTypeText,
			providerschema.LabelKey("access.form.aliyun_access_key_secret.label"),
			providerschema.PlaceholderKey("access.form.aliyun_access_key_secret.placeholder"),
			providerschema.TooltipKey("access.form.aliyun_access_key_secret.tooltip"),
			providerschema.TooltipHtml(),
			providerschema.Secret(),
			providerschema.Required(),
		).
		Field("resourceGroupId", providerschema.ValueTypeText,
			providerschema.LabelKey("access.form.aliyun_resource_group_id.label"),
			providerschema.PlaceholderKey("access.form.aliyun_resource_group_id.placeholder"),
			providerschema.TooltipKey("access.form.aliyun_resource_group_id.tooltip"),
			providerschema.TooltipHtml(),
		).
		Build()
	if err != nil {
		panic(err)
	}

	providerschema.Registries.MustRegister(string(domain.AccessProviderTypeAliyun), schema)
}
