package k8ssecret

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	k8score "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
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
	// Kubernetes Secret 中用于存放证书的 Key。
	SecretDataKeyForCrt string `json:"secretDataKeyForCrt,omitempty"`
	// Kubernetes Secret 中用于存放私钥的 Key。
	SecretDataKeyForKey string `json:"secretDataKeyForKey,omitempty"`
	// Kubernetes Secret 中用于存放CA证书。
	SecretDataKeyForIssuer string `json:"secretDataKeyForIssuer,omitempty"`
	// Kubernetes Secret 注解。
	SecretAnnotations map[string]string `json:"secretAnnotations,omitempty"`
	// Kubernetes Secret 标签。
	SecretLabels map[string]string `json:"secretLabels,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM, issuerCertificatePEM string) (*deployer.DeployResult, error) {
	if d.config.Namespace == "" {
		return nil, errors.New("config `namespace` is required")
	}
	if d.config.SecretName == "" {
		return nil, errors.New("config `secretName` is required")
	}
	if d.config.SecretType == "" {
		return nil, errors.New("config `secretType` is required")
	}
	if d.config.SecretDataKeyForCrt == "" {
		return nil, errors.New("config `secretDataKeyForCrt` is required")
	}
	if d.config.SecretDataKeyForKey == "" {
		return nil, errors.New("config `secretDataKeyForKey` is required")
	}

	certX509, err := xcert.ParseCertificateFromPEM(certPEM)
	if err != nil {
		return nil, err
	}

	// 连接
	client, err := createK8sClient(d.config.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	var secretPayload *k8score.Secret
	secretAnnotations := map[string]string{
		"certimate/common-name":       certX509.Subject.CommonName,
		"certimate/subject-sn":        certX509.Subject.SerialNumber,
		"certimate/subject-alt-names": strings.Join(certX509.DNSNames, ","),
		"certimate/issuer-sn":         certX509.Issuer.SerialNumber,
		"certimate/issuer-org":        strings.Join(certX509.Issuer.Organization, ","),
		"certimate/not-before":        certX509.NotBefore.Format("2006-01-02T15:04:05Z"),
		"certimate/not-after":         certX509.NotAfter.Format("2006-01-02T15:04:05Z"),
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

	// 获取 Secret 实例，如果不存在则创建
	secretPayload, err = client.CoreV1().Secrets(d.config.Namespace).Get(ctx, d.config.SecretName, k8smeta.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get kubernetes secret: %w", err)
		}

		secretPayload = &k8score.Secret{
			TypeMeta: k8smeta.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: k8smeta.ObjectMeta{
				Name:        d.config.SecretName,
				Annotations: secretAnnotations,
				Labels:      secretLabels,
			},
			Type: k8score.SecretType(d.config.SecretType),
		}
		secretPayload.Data = make(map[string][]byte)
		secretPayload.Data[d.config.SecretDataKeyForCrt] = []byte(certPEM)
		secretPayload.Data[d.config.SecretDataKeyForKey] = []byte(privkeyPEM)
		if d.config.SecretDataKeyForIssuer != "" {
			secretPayload.Data[d.config.SecretDataKeyForIssuer] = []byte(issuerCertificatePEM)
		}

		secretPayload, err = client.CoreV1().Secrets(d.config.Namespace).Create(ctx, secretPayload, k8smeta.CreateOptions{})
		d.logger.Debug("kubernetes operate 'Secrets.Create'", slog.String("namespace", d.config.Namespace), slog.Any("secret", secretPayload))
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes secret: %w", err)
		} else {
			return &deployer.DeployResult{}, nil
		}
	}

	// 更新 Secret 实例
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
	secretPayload.Data[d.config.SecretDataKeyForCrt] = []byte(certPEM)
	secretPayload.Data[d.config.SecretDataKeyForKey] = []byte(privkeyPEM)
	if d.config.SecretDataKeyForIssuer != "" {
		secretPayload.Data[d.config.SecretDataKeyForIssuer] = []byte(issuerCertificatePEM)
	}
	secretPayload, err = client.CoreV1().Secrets(d.config.Namespace).Update(ctx, secretPayload, k8smeta.UpdateOptions{})
	d.logger.Debug("kubernetes operate 'Secrets.Update'", slog.String("namespace", d.config.Namespace), slog.Any("secret", secretPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to update kubernetes secret: %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createK8sClient(kubeConfig string) (*kubernetes.Clientset, error) {
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

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
