package v1

import (
	http "net/http"

	corev1 "k8s.io/api/core/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

type CoreV1Interface interface {
	RESTClient() rest.Interface
	SecretsGetter
}

type CoreV1Client struct {
	restClient rest.Interface
}

func (c *CoreV1Client) Secrets(namespace string) SecretInterface {
	return newSecrets(c, namespace)
}

func NewForConfig(c *rest.Config) (*CoreV1Client, error) {
	config := *c
	setConfigDefaults(&config)
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

func NewForConfigAndClient(c *rest.Config, h *http.Client) (*CoreV1Client, error) {
	config := *c
	setConfigDefaults(&config)
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &CoreV1Client{client}, nil
}

func setConfigDefaults(config *rest.Config) {
	gv := corev1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.NegotiatedSerializer = rest.CodecFactoryForGeneratedClient(scheme.Scheme, scheme.Codecs).WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}

func (c *CoreV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
