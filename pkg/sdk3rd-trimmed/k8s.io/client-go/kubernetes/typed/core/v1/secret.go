package v1

import (
	corev1 "k8s.io/api/core/v1"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	gentype "k8s.io/client-go/gentype"
	scheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type SecretsGetter = typedcorev1.SecretsGetter

type SecretInterface = typedcorev1.SecretInterface

type secrets struct {
	*gentype.ClientWithListAndApply[*corev1.Secret, *corev1.SecretList, *applyconfigurationscorev1.SecretApplyConfiguration]
}

func newSecrets(c *CoreV1Client, namespace string) *secrets {
	return &secrets{
		gentype.NewClientWithListAndApply[*corev1.Secret, *corev1.SecretList, *applyconfigurationscorev1.SecretApplyConfiguration](
			"secrets",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *corev1.Secret { return &corev1.Secret{} },
			func() *corev1.SecretList { return &corev1.SecretList{} },
			gentype.PrefersProtobuf[*corev1.Secret](),
		),
	}
}
