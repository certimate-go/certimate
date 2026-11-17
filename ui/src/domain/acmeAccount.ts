import { z } from "zod";

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
export const dnsPersistRecordSchema = z.object({
  domain: z.string().min(1),
  name: z.string().min(1),
  type: z.literal("TXT"),
  value: z.string().min(1),
});

export const dnsPersistPreparationSchema = z.object({
  accountId: z.string().min(1),
  accountUrl: z.url(),
  records: z.array(dnsPersistRecordSchema).min(1),
});

export const dnsPersistCheckSchema = z.object({
  ready: z.boolean(),
  records: z.array(
    dnsPersistRecordSchema.extend({
      status: z.enum(["ready", "missing", "mismatch", "lookup_failed"]),
      message: z.string().optional(),
    })
  ),
});

export type DNSPersistRecord = z.infer<typeof dnsPersistRecordSchema>;
export type DNSPersistPreparation = z.infer<typeof dnsPersistPreparationSchema>;
export type DNSPersistCheck = z.infer<typeof dnsPersistCheckSchema>;

export type DNSPersistPreparePayload = {
  domains: string[];
  contactEmail: string;
  caProvider?: string;
  caProviderAccessId?: string;
  caProviderConfig?: Record<string, unknown>;
  nameservers?: string[];
  keyAlgorithm?: string;
  acmeProfile?: string;
};
