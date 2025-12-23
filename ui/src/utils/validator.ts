import { z } from "zod";

import { validateCronExpression } from "./cron";

export const isCron = (value: string) => {
  return validateCronExpression(value);
};

export const isDomain = (value: string, { allowWildcard = false }: { allowWildcard?: boolean } = {}) => {
  const re = allowWildcard
    ? /^(?:\*\.)?(?!-)[A-Za-z0-9-]{1,}(?<!-)(\.[A-Za-z0-9-]{1,}(?<!-)){0,}(?<![-0-9])$/
    : /^(?!-)[A-Za-z0-9-]{1,}(?<!-)(\.[A-Za-z0-9-]{1,}(?<!-)){0,}(?<![-0-9])$/;
  return re.test(value);
};

export const isEmail = (value: string) => {
  return z.email().safeParse(value).success;
};

export const isHostname = (value: string) => {
  return isDomain(value, { allowWildcard: false }) || isIPv4(value) || isIPv6(value);
};

export const isIPv4 = (value: string) => {
  return z.ipv4().safeParse(value).success;
};

export const isIPv6 = (value: string) => {
  return z.ipv6().safeParse(value).success;
};

export const isJsonObject = (value: string) => {
  try {
    const obj = JSON.parse(value);
    return typeof obj === "object" && !Array.isArray(obj);
  } catch {
    return false;
  }
};

export const isPortNumber = (value: string | number) => {
  return z.coerce.number().int().min(1).max(65535).safeParse(value).success;
};

export const isUrlWithHttp = (value: string) => {
  return z.url().startsWith("http://").safeParse(value).success;
};

export const isUrlWithHttps = (value: string) => {
  return z.url().startsWith("https://").safeParse(value).success;
};

export const isUrlWithHttpOrHttps = (value: string) => {
  return isUrlWithHttp(value) || isUrlWithHttps(value);
};
