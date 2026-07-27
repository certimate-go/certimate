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

export const installPlugin = async (providerType: string): Promise<unknown> => {
  const resp = await httpPost<unknown>({
    url: "/api/plugin/market/install",
    body: { providerType },
  });
  return resp.data;
};

export const deletePlugin = async (providerType: string): Promise<unknown> => {
  const resp = await pb.send<{ code: number; data: unknown }>(`/api/plugin/market/${providerType}`, {
    method: "DELETE",
  });
  return resp.data;
};

export const updatePlugin = async (providerType: string): Promise<unknown> => {
  const resp = await httpPost<unknown>({
    url: `/api/plugin/market/update/${providerType}`,
    body: { providerType },
  });
  return resp.data;
};
