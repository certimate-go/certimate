import { z } from "zod";

import { get as httpGet, post as httpPost } from "./_api";

export const ACCESS_TEST_REASON_CODES = [
  "auth_rejected",
  "connect_failed",
  "tls_failed",
  "timeout",
  "canceled",
  "probe_forbidden",
  "remote_error",
  "no_checker",
  "plugin_missing",
  "plugin_unavailable",
  "plugin_protocol_old",
  "config_invalid",
  "internal_error",
  "concurrency_limited",
  "access_not_found",
] as const;
export type AccessTestReasonCode = (typeof ACCESS_TEST_REASON_CODES)[number];

export type AccessTestRequest = { source: "stored"; accessId: string } | { source: "draft"; provider: string; config: Record<string, unknown> };

const accessTestCapabilitiesSchema = z.object({
  provider: z.string(),
  supported: z.boolean(),
  source: z.string(),
  ownerProviderType: z.string(),
  scope: z.array(z.string()),
  reasonCode: z.string(),
});

const accessTestResultSchema = z.object({
  status: z.enum(["passed", "failed", "inconclusive", "unsupported", "unavailable"]),
  reasonCode: z.string(),
  items: z
    .array(
      z.object({
        id: z.string(),
        status: z.enum(["passed", "failed", "skipped"]),
      })
    )
    .default([]),
  diagnostics: z.array(z.string()).default([]),
});

export type AccessTestCapabilities = z.infer<typeof accessTestCapabilitiesSchema>;
export type AccessTestResult = z.infer<typeof accessTestResultSchema>;

export const fetchAccessTestCapabilities = async (provider: string): Promise<AccessTestCapabilities> => {
  const resp = await httpGet<unknown>({
    url: `/api/accesses/test-capabilities?${new URLSearchParams({ provider })}`,
  });
  return accessTestCapabilitiesSchema.parse(resp.data);
};

export const testAccess = async (request: AccessTestRequest): Promise<AccessTestResult> => {
  const resp = await httpPost<unknown>({
    url: "/api/accesses/test",
    body: request,
  });
  return accessTestResultSchema.parse(resp.data);
};
