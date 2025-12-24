import { SubjectAlternativeNameExtension, X509Certificate } from "@peculiar/x509";

export const parseX509Certificate = (certPEM: string): X509Certificate => {
  const certX509 = new X509Certificate(certPEM);
  if (certX509 == null) throw new Error("Could not parse X.509 certificate. Maybe it is not in PEM format?");
  return certX509;
};

export const getSubjectAltNames = (cert: string | X509Certificate): string[] => {
  const subjectAltNames: string[] = [];

  try {
    const certX509 = X509Certificate.isAsnEncoded(cert) ? parseX509Certificate(cert) : cert;
    if (certX509 == null) return [];

    const sanExt = certX509.getExtension(SubjectAlternativeNameExtension);
    subjectAltNames.push(...(sanExt?.names?.items?.map((san) => san.value) || []));
  } catch (err) {
    console.error(err);
  }

  return subjectAltNames;
};
