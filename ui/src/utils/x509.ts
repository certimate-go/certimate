import { ECPrivateKey } from "@peculiar/asn1-ecc";
import { PrivateKeyInfo } from "@peculiar/asn1-pkcs8";
import { RSAPrivateKey } from "@peculiar/asn1-rsa";
import { AsnParser } from "@peculiar/asn1-schema";
import { PemConverter, SubjectAlternativeNameExtension, X509Certificate } from "@peculiar/x509";

export const parseCertificate = (certPEM: string): X509Certificate => {
  if (!X509Certificate.isAsnEncoded(certPEM)) {
    throw new Error("Could not parse X.509 certificate. Maybe it is not in PEM format?");
  }

  try {
    const cert = new X509Certificate(certPEM);
    if (cert == null) {
      throw new Error("Parse PEM certificate failed: result is null");
    }

    return cert;
  } catch (err) {
    throw new Error("Could not parse X.509 certificate", { cause: err });
  }
};

export const getCertificateSubjectAltNames = (certificate: string | X509Certificate): string[] => {
  try {
    const certX509 = certificate instanceof X509Certificate ? certificate : parseCertificate(certificate);
    if (certX509 == null) return [];

    const sanExt = certX509.getExtension(SubjectAlternativeNameExtension);
    return sanExt?.names?.items?.map((san) => san.value) || [];
  } catch {
    return [];
  }
};

export const parsePKCS1PrivateKey = (keyPEM: string): RSAPrivateKey => {
  try {
    const PEM_BLOCK_TYPE = "RSA PRIVATE KEY";
    const pemBlock = PemConverter.decodeWithHeaders(keyPEM)[0];
    if (!pemBlock || pemBlock.type !== PEM_BLOCK_TYPE) {
      throw new Error(`PEM block is not of type '${PEM_BLOCK_TYPE}'`);
    }

    const key = AsnParser.parse(pemBlock.rawData, RSAPrivateKey);
    if (key == null) {
      throw new Error("Read private key failed: result is null");
    }

    return key;
  } catch (err) {
    throw new Error("Could not parse PKCS#1 RSA private key", { cause: err });
  }
};

export const parsePKCS8PrivateKey = (keyPEM: string): PrivateKeyInfo => {
  try {
    const PEM_BLOCK_TYPE = "PRIVATE KEY";
    const pemBlock = PemConverter.decodeWithHeaders(keyPEM)[0];
    if (!pemBlock || pemBlock.type !== PEM_BLOCK_TYPE) {
      throw new Error(`PEM block is not of type '${PEM_BLOCK_TYPE}'`);
    }

    const key = AsnParser.parse(pemBlock.rawData, PrivateKeyInfo);
    if (key == null) {
      throw new Error("Read private key failed: result is null");
    }

    return key;
  } catch (err) {
    throw new Error("Could not parse PKCS#8 private key", { cause: err });
  }
};

export const parseECPrivateKey = (keyPEM: string): ECPrivateKey => {
  try {
    const PEM_BLOCK_TYPE = "EC PRIVATE KEY";
    const pemBlock = PemConverter.decodeWithHeaders(keyPEM)[0];
    if (!pemBlock || pemBlock.type !== PEM_BLOCK_TYPE) {
      throw new Error(`PEM block is not of type '${PEM_BLOCK_TYPE}'`);
    }

    const key = AsnParser.parse(pemBlock.rawData, ECPrivateKey);
    if (key == null) {
      throw new Error("Read private key failed: result is null");
    }

    return key;
  } catch (err) {
    throw new Error("Could not parse EC private key", { cause: err });
  }
};

export const parsePrivateKey = (keyPEM: string): RSAPrivateKey | ECPrivateKey | PrivateKeyInfo => {
  try {
    return parsePKCS1PrivateKey(keyPEM);
  } catch {
    try {
      return parseECPrivateKey(keyPEM);
    } catch {
      return parsePKCS8PrivateKey(keyPEM);
    }
  }
};

export const getPrivateKeyAlgorithm = (keyPEM: string): { algorithm?: "RSA" | "EC"; keySize?: number } => {
  try {
    const key = parsePrivateKey(keyPEM);

    if (key instanceof RSAPrivateKey) {
      return { algorithm: "RSA", keySize: (key.modulus.byteLength - 1) * 8 };
    }

    if (key instanceof ECPrivateKey) {
      return { algorithm: "EC", keySize: key.privateKey.byteLength * 8 };
    }

    if (key instanceof PrivateKeyInfo) {
      const OLD_PUBKEY_RSA = "1.2.840.113549.1.1.1";
      const OLD_PUBKEY_ECDSA = "1.2.840.10045.2.1";
      if (key.privateKeyAlgorithm.algorithm === OLD_PUBKEY_RSA) {
        const rsaKey = AsnParser.parse(key.privateKey, RSAPrivateKey);
        return { algorithm: "RSA", keySize: (rsaKey.modulus.byteLength - 1) * 8 };
      }
      if (key.privateKeyAlgorithm.algorithm === OLD_PUBKEY_ECDSA) {
        const ecKey = AsnParser.parse(key.privateKey, ECPrivateKey);
        return { algorithm: "RSA", keySize: ecKey.privateKey.byteLength * 8 };
      }
    }

    return {};
  } catch {
    return {};
  }
};

export const validatePEMCertificate = (certPEM: string): boolean => {
  try {
    const cert = parseCertificate(certPEM);
    return !!cert.getExtension(SubjectAlternativeNameExtension);
  } catch {
    return false;
  }
};

export const validatePEMPrivateKey = (keyPEM: string): boolean => {
  try {
    parsePrivateKey(keyPEM);
    return true;
  } catch {
    return false;
  }
};
