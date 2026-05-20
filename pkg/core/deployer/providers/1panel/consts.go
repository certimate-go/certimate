package onepanel

const (
	// 部署目标：替换指定网站的证书。
	DEPLOY_TARGET_WEBSITE = "website"
	// 部署目标：替换指定证书。
	DEPLOY_TARGET_CERTIFICATE = "certificate"
)

const (
	// 匹配模式：指定 ID。
	WEBSITE_MATCH_PATTERN_SPECIFIED = "specified"
	// 匹配模式：证书 SAN 匹配。
	WEBSITE_MATCH_PATTERN_CERTSAN = "certsan"
)
