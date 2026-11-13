import { ClientResponseError } from "pocketbase";
import { z } from "zod";

import { getPocketBase } from "@/repository/_pocketbase";
import { get as httpGet, post as httpPost } from "./_api";

const pb = getPocketBase();

export const accessDependencyErrorSchema = z.object({
  reason: z.literal("access_owner_in_use"),
  providerType: z.string(),
  accessType: z.string(),
  dependents: z.array(z.string()).nonempty(),
});

const marketSourceSchema = z.object({ id: z.string(), repository: z.string() });
const marketSourceStatusSchema = marketSourceSchema.extend({
  official: z.boolean(),
  syncedAt: z.string().optional(),
  error: z.string().optional(),
  stale: z.boolean(),
  refreshing: z.boolean(),
});
export type MarketSourceStatus = z.infer<typeof marketSourceStatusSchema>;

export const marketConflictSchema = z.object({
  reason: z.enum(["install_in_progress", "ownership_conflict", "source_busy", "duplicate_source", "operation_in_progress"]),
});

const marketEntrySchema = z.object({
  provider_type: z.string(),
  display_name_key: z.string(),
  deploy_category: z.string(),
  capability: z.string().optional(),
  version: z.string(),
  description: z.string(),
  icon_url: z.string().optional(),
  status: z.enum(["not_installed", "installed", "update_available", "installed_manual", "unsupported_platform", "ownership_conflict"]),
  source: marketSourceSchema,
  installedOrigin: marketSourceSchema.optional(),
  official: z.boolean(),
  sourceAvailable: z.boolean(),
  listed: z.boolean(),
  canInstall: z.boolean(),
  canUpdate: z.boolean(),
  canDelete: z.boolean(),
});
export type MarketEntry = z.infer<typeof marketEntryWithJobSchema>;

export type MarketFilter = "all" | "installed" | "not_installed";

export type MarketCapabilityFilter = "all" | "deploy" | "dns" | "notify";

export interface MarketQuery {
  page: number;
  perPage: number;
  keyword: string;
  status: MarketFilter;
  capability: MarketCapabilityFilter;
  locale: string;
  refresh: boolean;
  source: string;
}

const reloadResultSchema = z.object({
  added: z.array(z.string()).nullable(),
  changed: z.array(z.string()).nullable(),
  removed: z.array(z.string()).nullable(),
  errors: z.array(z.string()).nullable(),
});
export type ReloadResult = z.infer<typeof reloadResultSchema>;

const installJobSchema = z.object({
  id: z.string(),
  source: marketSourceSchema,
  providerType: z.string(),
  state: z.enum(["queued", "downloading", "verifying", "extracting", "reloading", "installed", "failed"]),
  stage: z.string().optional(),
  error: z.string().optional(),
  downloaded: z.number().optional(),
  total: z.number().optional(),
  result: reloadResultSchema.optional(),
});
export type InstallJobStatus = z.infer<typeof installJobSchema>;
export type InstallJobState = InstallJobStatus["state"];

const marketEntryWithJobSchema = marketEntrySchema.extend({ job: installJobSchema.optional() });
const marketPageSchema = z.object({
  items: z.array(marketEntryWithJobSchema),
  page: z.number(),
  perPage: z.number(),
  totalItems: z.number(),
  totalPages: z.number(),
  syncedAt: z.string().optional(),
  refreshFailed: z.boolean(),
  refreshing: z.boolean(),
  activeJobs: z.array(installJobSchema),
  sources: z.array(marketSourceStatusSchema),
});
export type MarketPage = z.infer<typeof marketPageSchema>;

export const isTerminalState = (state: InstallJobState): boolean => state === "installed" || state === "failed";

export const fetchMarketListing = async (query: MarketQuery): Promise<MarketPage> => {
  const params = new URLSearchParams({
    page: String(query.page),
    perPage: String(query.perPage),
    keyword: query.keyword,
    status: query.status,
    capability: query.capability,
    locale: query.locale,
    refresh: String(query.refresh),
    source: query.source,
  });
  const resp = await httpGet<unknown>({ url: `/api/plugin/market?${params}` });
  return marketPageSchema.parse(resp.data);
};

export const installPlugin = async ({ providerType, source }: PluginTarget): Promise<InstallJobStatus> => {
  const resp = await httpPost<unknown>({
    url: "/api/plugin/market/install",
    body: { providerType, source },
  });
  return installJobSchema.parse(resp.data);
};

export const getInstallStatus = async (providerType: string): Promise<InstallJobStatus> => {
  const resp = await httpGet<unknown>({
    url: `/api/plugin/market/install/status/${providerType}`,
  });
  return installJobSchema.parse(resp.data);
};

const deleteRequest = async (url: string): Promise<unknown> => {
  const resp = await pb.send<{ code: number; msg: string; data: unknown }>(url, { method: "DELETE" });
  if (resp.code !== 0) {
    throw new ClientResponseError({ status: resp.code, response: { ...resp, message: resp.msg }, data: {} });
  }
  return resp.data;
};

export type PluginTarget = { providerType: string; source: string };

export const deletePlugin = async ({ providerType, source }: PluginTarget): Promise<ReloadResult> => {
  const params = new URLSearchParams({ source });
  return reloadResultSchema.parse(await deleteRequest(`/api/plugin/market/${encodeURIComponent(providerType)}?${params}`));
};

export const updatePlugin = async ({ providerType, source }: PluginTarget): Promise<ReloadResult> => {
  const resp = await httpPost<unknown>({
    url: `/api/plugin/market/update/${encodeURIComponent(providerType)}`,
    body: { source },
  });
  return reloadResultSchema.parse(resp.data);
};

export const addMarketSource = async (repository: string) =>
  marketSourceStatusSchema.parse((await httpPost<unknown>({ url: "/api/plugin/market/sources", body: { repository } })).data);

export const refreshMarketSource = async (source: string) =>
  marketSourceStatusSchema.parse((await httpPost<unknown>({ url: `/api/plugin/market/sources/${encodeURIComponent(source)}/refresh` })).data);

export const removeMarketSource = async (source: string) => {
  await deleteRequest(`/api/plugin/market/sources/${encodeURIComponent(source)}`);
};
