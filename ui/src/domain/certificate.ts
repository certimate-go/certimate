import { type WorkflowModel } from "./workflow";

export interface CertificateModel extends BaseModel {
  source: string;
  subjectAltNames: string;
  serialNumber: string;
  certificate: string;
  issuerCertificate: string;
  privateKey: string;
  issuerOrg: string;
  keyAlgorithm: string;
  validityNotBefore: ISO8601String;
  validityNotAfter: ISO8601String;
  isRenewed: boolean;
  isRevoked: boolean;
  workflowRef: string;
  expand?: {
    workflowRef?: Pick<WorkflowModel, "id" | "name" | "description">;
  };
}

export const CERTIFICATE_SOURCES = Object.freeze({
  REQUEST: "request",
  UPLOAD: "upload",
} as const);

export type CertificateSourceType = (typeof CERTIFICATE_SOURCES)[keyof typeof CERTIFICATE_SOURCES];

export const CERTIFICATE_FORMATS = Object.freeze({
  PEM: "PEM",
  PFX: "PFX",
  JKS: "JKS",
} as const);

export type CertificateFormatType = (typeof CERTIFICATE_FORMATS)[keyof typeof CERTIFICATE_FORMATS];
