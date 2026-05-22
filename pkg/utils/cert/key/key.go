package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
)

type KeyAlgorithm = x509.PublicKeyAlgorithm

func DetectPublicKeyAlgorithm(pubkey crypto.PublicKey) (_algorithm KeyAlgorithm, _size int, _error error) {
	switch t := pubkey.(type) {
	case *rsa.PublicKey:
		size := t.N.BitLen()
		return x509.RSA, size, nil

	case *ecdsa.PublicKey:
		size := t.Curve.Params().BitSize
		return x509.ECDSA, size, nil

	case ed25519.PublicKey:
		return x509.Ed25519, 256, nil
	}

	return x509.UnknownPublicKeyAlgorithm, 0, fmt.Errorf("unknown public key type")
}

func DetectPrivateKeyAlgorithm(privkey crypto.PrivateKey) (_algorithm KeyAlgorithm, _size int, _error error) {
	switch t := privkey.(type) {
	case *rsa.PrivateKey:
		size := t.N.BitLen()
		return x509.RSA, size, nil

	case *ecdsa.PrivateKey:
		size := t.Curve.Params().BitSize
		return x509.ECDSA, size, nil

	case ed25519.PrivateKey:
		return x509.Ed25519, 512, nil
	}

	return x509.UnknownPublicKeyAlgorithm, 0, fmt.Errorf("unknown private key type")
}
