package zenlayercdn

const (
	// 资源类型：替换指定域名的证书。
	RESOURCE_TYPE_DOMAIN = "domain"
	// 资源类型：替换指定证书。
	RESOURCE_TYPE_CERTIFICATE = "certificate"
)

const (
	// 匹配模式：精确匹配。
	DOMAIN_MATCH_PATTERN_EXACT = "exact"
	// 匹配模式：通配符匹配。
	DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard"
	// 匹配模式：证书 SAN 匹配。
	DOMAIN_MATCH_PATTERN_CERTSAN = "certsan"
)
