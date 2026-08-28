import { type ACMEAccountImportPayload, type ACMEAccountView } from "@/domain/acmeAccount";

import { get as httpGet, post as httpPost } from "./_api";

export type ListACMEAccountsParams = {
  page?: number;
  perPage?: number;
  ca?: string;
};

export const list = (params?: ListACMEAccountsParams) => {
  type RespData = {
    items: ACMEAccountView[];
    totalItems: number;
    page: number;
    perPage: number;
  };

  const qs = new URLSearchParams();
  if (params?.page != null) qs.set("page", String(params.page));
  if (params?.perPage != null) qs.set("perPage", String(params.perPage));
  if (params?.ca) qs.set("ca", params.ca);
  const query = qs.toString();

  return httpGet<RespData>({
    url: query ? `/api/acme-accounts?${query}` : "/api/acme-accounts",
  });
};

export const importAccount = (payload: ACMEAccountImportPayload) => {
  type RespData = {
    item: ACMEAccountView;
  };

  return httpPost<RespData>({
    url: "/api/acme-accounts/import",
    body: payload,
  });
};

export const exportAccount = (accountId: string) => {
  type RespData = {
    privateKeyPem: string;
  };

  return httpPost<RespData>({
    url: `/api/acme-accounts/${encodeURIComponent(accountId)}/export`,
  });
};

export const rotateAccount = (accountId: string) => {
  type RespData = {
    item: ACMEAccountView;
  };

  return httpPost<RespData>({
    url: `/api/acme-accounts/${encodeURIComponent(accountId)}/rotate`,
  });
};
