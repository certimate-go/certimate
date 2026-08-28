import { ClientResponseError } from "pocketbase";

import { get as httpGet, post as httpPost } from "./_api";
import { getPocketBase } from "@/repository/_pocketbase";

const pb = getPocketBase();

export interface MarketEntry {
  provider_type: string;
  access_provider_type: string;
  display_name_key: string;
  deploy_category: string;
  version: string;
  description: string;
  icon: string;
  protocol_version: number;
  status: "not_installed" | "installed" | "update_available" | "installed_manual" | "unsupported_platform";
  release?: {
    repo: string;
    tag: string;
    assets: Record<string, string>;
  };
}

export type InstallJobState = "queued" | "downloading" | "verifying" | "extracting" | "reloading" | "installed" | "failed";

export interface InstallJobStatus {
  providerType: string;
  state: InstallJobState;
  stage?: string;
  error?: string;
  downloaded?: number;
  total?: number;
  result?: { added: string[]; changed: string[]; removed: string[]; errors: string[] };
}

export const isTerminalState = (state: InstallJobState): boolean => state === "installed" || state === "failed";

export const fetchMarketListing = async (): Promise<MarketEntry[]> => {
  try {
    const resp = await httpGet<MarketEntry[]>({
      url: "/api/plugin/market",
    });
    return resp.data ?? [];
  } catch {
    return [];
  }
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

export const deletePlugin = async (providerType: string): Promise<unknown> => {
  const resp = await pb.send<{ code: number; data: unknown }>(`/api/plugin/market/${providerType}`, {
    method: "DELETE",
  });
  if (resp.code !== 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }
  return resp.data;
};

export const updatePlugin = async (providerType: string): Promise<unknown> => {
  const resp = await httpPost<unknown>({
    url: `/api/plugin/market/update/${providerType}`,
    body: { providerType },
  });
  return resp.data;
};
