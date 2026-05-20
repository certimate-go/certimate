package nginxproxymanager

const (
	AUTH_METHOD_PASSWORD = "password"
	AUTH_METHOD_TOKEN    = "token"
)

const (
	// 部署目标：替换指定主机的证书。
	DEPLOY_TARGET_HOST = "host"
	// 部署目标：替换指定证书。
	DEPLOY_TARGET_CERTIFICATE = "certificate"
)

const (
	// 匹配模式：指定 ID。
	HOST_MATCH_PATTERN_SPECIFIED = "specified"
	// 匹配模式：证书 SAN 匹配。
	HOST_MATCH_PATTERN_CERTSAN = "certsan"
)

const (
	HOST_TYPE_PROXY       = "proxy"
	HOST_TYPE_REDIRECTION = "redirection"
	HOST_TYPE_STREAM      = "stream"
	HOST_TYPE_DEAD        = "dead"
)
