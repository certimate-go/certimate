import { ClientResponseError } from "pocketbase";

import { getPocketBase } from "@/repository/_pocketbase";
import { get as httpGet, post as httpPost } from "./_api";

const pb = getPocketBase();

export interface MarketEntry {
  provider_type: string;
  display_name_key: string;
  deploy_category: string;
  capability?: string;
  version: string;
  description: string;
  icon_url?: string;
  status: "not_installed" | "installed" | "update_available" | "installed_manual" | "unsupported_platform";
  job?: InstallJobStatus;
}

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
}

export interface MarketPage {
  items: MarketEntry[];
  page: number;
  perPage: number;
  totalItems: number;
  totalPages: number;
  syncedAt?: string;
  refreshFailed: boolean;
  refreshing: boolean;
  activeJobs: InstallJobStatus[];
}

export type InstallJobState = "queued" | "downloading" | "verifying" | "extracting" | "reloading" | "installed" | "failed";

export interface ReloadResult {
  added: string[] | null;
  changed: string[] | null;
  removed: string[] | null;
  errors: string[] | null;
}

export interface InstallJobStatus {
  providerType: string;
  state: InstallJobState;
  stage?: string;
  error?: string;
  downloaded?: number;
  total?: number;
  result?: ReloadResult;
}

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
  });
  const resp = await httpGet<MarketPage>({ url: `/api/plugin/market?${params}` });
  return resp.data;
};

export const installPlugin = async (providerType: string): Promise<InstallJobStatus> => {
  const resp = await httpPost<InstallJobStatus>({
    url: "/api/plugin/market/install",
    body: { providerType },
  });
  return resp.data;
};

export const getInstallStatus = async (providerType: string): Promise<InstallJobStatus> => {
  const resp = await httpGet<InstallJobStatus>({
    url: `/api/plugin/market/install/status/${providerType}`,
  });
  return resp.data;
};

export const deletePlugin = async (providerType: string): Promise<ReloadResult> => {
  const resp = await pb.send<{ code: number; data: ReloadResult }>(`/api/plugin/market/${providerType}`, {
    method: "DELETE",
  });
  if (resp.code !== 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }
  return resp.data;
};

export const updatePlugin = async (providerType: string): Promise<ReloadResult> => {
  const resp = await httpPost<ReloadResult>({
    url: `/api/plugin/market/update/${providerType}`,
    body: { providerType },
  });
  return resp.data;
};
