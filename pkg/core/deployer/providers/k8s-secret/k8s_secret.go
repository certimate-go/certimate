package k8ssecret

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	k8score "k8s.io/api/core/v1"
	k8serrs "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// kubeconfig 文件内容。
	KubeConfig string `json:"kubeConfig,omitempty"`
	// Kubernetes 命名空间。
	Namespace string `json:"namespace,omitempty"`
	// Kubernetes Secret 名称。
	SecretName string `json:"secretName"`
	// Kubernetes Secret 类型。
	SecretType string `json:"secretType"`
	// Kubernetes Secret 中用于存放私钥的键。
	SecretDataKeyForKey string `json:"secretDataKeyForKey,omitempty"`
	// Kubernetes Secret 中用于存放证书的键。
	SecretDataKeyForCrt string `json:"secretDataKeyForCrt,omitempty"`
	// Kubernetes Secret 中用于存放证书（仅含服务器证书）的键。
	// 选填。
	SecretDataKeyForCrtOnlyServer string `json:"secretDataKeyForCrtOnlyServer,omitempty"`
	// Kubernetes Secret 中用于存放证书（仅含中间证书）的键。
	// 选填。
	SecretDataKeyForCrtOnlyIntermedia string `json:"secretDataKeyForCrtOnlyIntermedia,omitempty"`
	// Kubernetes Secret 注解。
	SecretAnnotations map[string]string `json:"secretAnnotations,omitempty"`
	// Kubernetes Secret 标签。
	SecretLabels map[string]string `json:"secretLabels,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	return &Deployer{
		logger: slog.Default(),
		config: config,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	if d.config.Namespace == "" {
		return nil, fmt.Errorf("config `namespace` is required")
	}
	if d.config.SecretName == "" {
		return nil, fmt.Errorf("config `secretName` is required")
	}
	if d.config.SecretType == "" {
		return nil, fmt.Errorf("config `secretType` is required")
	}

	// 解析证书内容
	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 提取服务器证书和中间证书
	serverCertPEM, issuerCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 连接到 Kubernetes
	client, err := createK8sClient(d.config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// 获取 Secret 实例
	secretPayload := &k8score.Secret{}
	secretIsNew := false
	secretGetResp := client.Get().
		Namespace(d.config.Namespace).
		Resource("secrets").
		Name(d.config.SecretName).
		VersionedParams(&meta.GetOptions{}, meta.ParameterCodec).
		Do(ctx)
	if err := secretGetResp.Error(); err != nil {
		if !k8serrs.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get kubernetes secret: %w", err)
		}

		secretPayload = &k8score.Secret{
			Type: k8score.SecretType(d.config.SecretType),
			TypeMeta: meta.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: d.config.SecretName,
			},
		}
		secretIsNew = true
	} else if err := secretGetResp.Into(secretPayload); err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes secret: %w", err)
	}
	d.logger.Debug("kubernetes operate 'Secrets.Get'", slog.String("namespace", d.config.Namespace), slog.Any("secret", d.config.SecretName))

	// 生成 Secret 注解和标签
	secretAnnotations := map[string]string{
		"certimate/common-name":       certX509.Subject.CommonName,
		"certimate/subject-sn":        certX509.Subject.SerialNumber,
		"certimate/subject-alt-names": strings.Join(certX509.DNSNames, ","),
		"certimate/issuer-sn":         certX509.Issuer.SerialNumber,
		"certimate/issuer-org":        strings.Join(certX509.Issuer.Organization, ","),
	}
	secretLabels := map[string]string{}
	if d.config.SecretAnnotations != nil {
		for k, v := range d.config.SecretAnnotations {
			secretAnnotations[k] = v
		}
	}
	if d.config.SecretLabels != nil {
		for k, v := range d.config.SecretLabels {
			secretLabels[k] = v
		}
	}

	// 赋值 Secret 实例
	secretPayload.Type = k8score.SecretType(d.config.SecretType)
	if secretPayload.ObjectMeta.Annotations == nil {
		secretPayload.ObjectMeta.Annotations = secretAnnotations
	} else {
		for k, v := range secretAnnotations {
			secretPayload.ObjectMeta.Annotations[k] = v
		}
	}
	if secretPayload.ObjectMeta.Labels == nil {
		secretPayload.ObjectMeta.Labels = secretLabels
	} else {
		for k, v := range secretLabels {
			secretPayload.ObjectMeta.Labels[k] = v
		}
	}
	if secretPayload.Data == nil {
		secretPayload.Data = make(map[string][]byte)
	}
	if d.config.SecretDataKeyForKey != "" {
		secretPayload.Data[d.config.SecretDataKeyForKey] = []byte(privkeyPEM)
	}
	if d.config.SecretDataKeyForCrt != "" {
		secretPayload.Data[d.config.SecretDataKeyForCrt] = []byte(certPEM)
	}
	if d.config.SecretDataKeyForCrtOnlyServer != "" {
		secretPayload.Data[d.config.SecretDataKeyForCrtOnlyServer] = []byte(serverCertPEM)
	}
	if d.config.SecretDataKeyForCrtOnlyIntermedia != "" {
		secretPayload.Data[d.config.SecretDataKeyForCrtOnlyIntermedia] = []byte(issuerCertPEM)
	}

	// 创建或更新 Secret 实例
	if secretIsNew {
		secretPostResp := client.Post().
			Namespace(d.config.Namespace).
			Resource("secrets").
			Name(d.config.SecretName).
			VersionedParams(&meta.GetOptions{}, meta.ParameterCodec).
			Body(secretPayload).
			Do(ctx)
		d.logger.Debug("kubernetes operate 'Secrets.Post'", slog.String("namespace", d.config.Namespace), slog.Any("secret", d.config.SecretName))
		if err := secretPostResp.Error(); err != nil {
			return nil, fmt.Errorf("failed to create kubernetes secret: %w", err)
		}
	} else {
		secretPutResp := client.Put().
			Namespace(d.config.Namespace).
			Resource("secrets").
			Name(d.config.SecretName).
			VersionedParams(&meta.GetOptions{}, meta.ParameterCodec).
			Body(secretPayload).
			Do(ctx)
		d.logger.Debug("kubernetes operate 'Secrets.Put'", slog.String("namespace", d.config.Namespace), slog.Any("secret", d.config.SecretName))
		if err := secretPutResp.Error(); err != nil {
			return nil, fmt.Errorf("failed to update kubernetes secret: %w", err)
		}
	}

	return &DeployResult{}, nil
}

func createK8sClient(kubeConfig string) (*rest.RESTClient, error) {
	var config *rest.Config
	var err error
	if kubeConfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		kubeConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeConfig))
		if err != nil {
			return nil, err
		}
		config, err = kubeConfig.ClientConfig()
	}
	if err != nil {
		return nil, err
	}

	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
