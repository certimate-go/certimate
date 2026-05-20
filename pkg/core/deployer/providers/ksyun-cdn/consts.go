package ksyuncdn

const (
	// 部署目标：替换指定域名的证书。
	DEPLOY_TARGET_DOMAIN = "domain"
	// 部署目标：替换指定证书。
	DEPLOY_TARGET_CERTIFICATE = "certificate"
)

const (
	// 匹配模式：精确匹配。
	DOMAIN_MATCH_PATTERN_EXACT = "exact"
	// 匹配模式：通配符匹配。
	DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard"
	// 匹配模式：证书 SAN 匹配。
	DOMAIN_MATCH_PATTERN_CERTSAN = "certsan"
)
