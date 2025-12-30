package nginxproxymanager

import (
	"encoding/json"
	"fmt"
)

type apiResponse interface {
	GetError() string
}

type apiResponseBase struct {
	Error json.RawMessage `json:"error"`
}

func (r *apiResponseBase) GetError() string {
	if len(r.Error) == 0 {
		return ""
	}

	var errStr string
	if err := json.Unmarshal(r.Error, &errStr); err == nil {
		return errStr
	}

	type errObjType struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	var errObj errObjType
	if err := json.Unmarshal(r.Error, &errObj); err == nil && errObj.Message != "" {
		if errObj.Code != 0 {
			return fmt.Sprintf("%d %s", errObj.Code, errObj.Message)
		}
		return errObj.Message
	}

	var errMap map[string]interface{}
	if err := json.Unmarshal(r.Error, &errMap); err == nil {
		if message, ok := errMap["message"].(string); ok {
			return message
		}
	}

	return ""
}

var _ apiResponse = (*apiResponseBase)(nil)

type CertificateRecord struct {
	Id          int64           `json:"id"`
	CreatedOn   string          `json:"created_on"`
	ModifiedOn  string          `json:"modified_on"`
	Provider    string          `json:"provider"`
	NiceName    string          `json:"nice_name"`
	DomainNames []string        `json:"domain_names"`
	ExpiresOn   string          `json:"expires_on"`
	Meta        CertificateMeta `json:"meta"`
}

type CertificateMeta struct {
	Certificate             string `json:"certificate"`
	CertificateKey          string `json:"certificate_key"`
	IntermediateCertificate string `json:"intermediate_certificate"`
}

type HostRecord struct {
	Id            int64    `json:"id"`
	CreatedOn     string   `json:"created_on"`
	ModifiedOn    string   `json:"modified_on"`
	DomainNames   []string `json:"domain_names"`
	CertificateId int64    `json:"certificate_id"`
	Meta          HostMeta `json:"meta"`
	Enabled       bool     `json:"enabled"`
}

type HostMeta struct {
	NginxOnline bool `json:"nginx_online"`
	NginxErr    any  `json:"nginx_err"`
}

type ProxyHostRecord struct {
	HostRecord
	ForwardScheme  string `json:"forward_scheme"`
	ForwardHost    string `json:"forward_host"`
	ForwardPort    int32  `json:"forward_port"`
	SslForced      bool   `json:"ssl_forced"`
	Http2Support   bool   `json:"http2_support"`
	HstsEnabled    bool   `json:"hsts_enabled"`
	HstsSubdomains bool   `json:"hsts_subdomains"`
}

type RedirectionHostRecord struct {
	HostRecord
	ForwardScheme     string `json:"forward_scheme"`
	ForwardDomainName string `json:"forward_domain_name"`
	ForwardHttpCode   int32  `json:"forward_http_code"`
	SslForced         bool   `json:"ssl_forced"`
	Http2Support      bool   `json:"http2_support"`
	HstsEnabled       bool   `json:"hsts_enabled"`
	HstsSubdomains    bool   `json:"hsts_subdomains"`
}

type StreamHostRecord struct {
	HostRecord
	ForwardingHost string `json:"forwarding_host"`
	ForwardingPort int32  `json:"forwarding_port"`
	IncomingPort   int32  `json:"incoming_port"`
	TcpForwarding  bool   `json:"tcp_forwarding"`
	UdpForwarding  bool   `json:"udp_forwarding"`
}

type DeadHostRecord struct {
	HostRecord
	SslForced      bool `json:"ssl_forced"`
	Http2Support   bool `json:"http2_support"`
	HstsEnabled    bool `json:"hsts_enabled"`
	HstsSubdomains bool `json:"hsts_subdomains"`
}
