package onepanel

const (
	// 资源类型：替换指定网站的证书。
	RESOURCE_TYPE_WEBSITE = "website"
	// 资源类型：替换指定证书。
	RESOURCE_TYPE_CERTIFICATE = "certificate"
)

const (
	// 匹配模式：指定 ID。
	WEBSITE_MATCH_PATTERN_SPECIFIED = "specified"
	// 匹配模式：证书 SAN 匹配。
	WEBSITE_MATCH_PATTERN_CERTSAN = "certsan"
)
