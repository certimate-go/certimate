package volcenginevod

const (
	// 匹配模式：精确匹配。
	DOMAIN_MATCH_PATTERN_EXACT = "exact"
	// 匹配模式：通配符匹配。
	DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard"
	// 匹配模式：证书 SAN 匹配。
	DOMAIN_MATCH_PATTERN_CERTSAN = "certsan"
)

const (
	// 域名类型：点播加速域名。
	DOMAIN_TYPE_PLAY = "play"
	// 域名类型：封面加速域名。
	DOMAIN_TYPE_IMAGE = "image"
)
