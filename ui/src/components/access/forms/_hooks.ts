import { useMemo } from "react";

import { ACCESS_USAGES, type AccessProvider } from "@/domain/provider";

export const useProviderFilterByUsage = (usage?: "dns" | "hosting" | "dns-hosting" | "ca" | "notification" | "certsource") => {
  return useMemo(() => {
    if (usage == null) return;

    switch (usage) {
      case "dns":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.DNS);
      case "hosting":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.HOSTING);
      case "dns-hosting":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.DNS) || option.usages.includes(ACCESS_USAGES.HOSTING);
      case "ca":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.CA);
      case "notification":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.NOTIFICATION);
      case "certsource":
        return (_: string, option: AccessProvider) => option.usages.includes(ACCESS_USAGES.CERTSOURCE);
      default:
        console.warn(`[certimate] unsupported provider usage: '${usage}'`);
    }
  }, [usage]);
};
