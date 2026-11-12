package plugin

import (
	githubplugin "github.com/hashicorp/go-plugin"
)

const ProtocolVersion = 3

const MagicCookieKey = "CERTIMATE_PLUGIN"

const MagicCookieValue = "certimate-plugin"

var HandshakeConfig = githubplugin.HandshakeConfig{
	ProtocolVersion:  ProtocolVersion,
	MagicCookieKey:   MagicCookieKey,
	MagicCookieValue: MagicCookieValue,
}
