export interface ACMEAccountView {
  id: string;
  ca: string;
  email: string;
  acmeDirUrl: string;
  acmeAcctUrl: string;
  created: string;
  updated: string;
}

export interface ACMEAccountImportPayload {
  privateKeyPem: string;
  ca: string;
  acmeDirUrl?: string;
  email?: string;
}
