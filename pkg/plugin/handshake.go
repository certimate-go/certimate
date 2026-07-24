package plugin

import (
	githubplugin "github.com/hashicorp/go-plugin"
)

const ProtocolVersion = 1

const MagicCookieKey = "CERTIMATE_PLUGIN"

const MagicCookieValue = "certimate-deployer-plugin"

var HandshakeConfig = githubplugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieValue,
}
