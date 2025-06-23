import { z } from "zod/v4";

import { validCronExpression as _validCronExpression } from "./cron";

export const validCronExpression = (value: string) => {
  return _validCronExpression(value);
};

export const validDomainName = (value: string, { allowWildcard = false }: { allowWildcard?: boolean } = {}) => {
  const re = allowWildcard ? /^(?:\*\.)?([a-zA-Z0-9-]+\.)+[a-zA-Z]{1,}$/ : /^(?!-)[A-Za-z0-9-]{1,}(?<!-)(\.[A-Za-z0-9-]{1,}(?<!-)){0,}$/;
  return re.test(value);
};

export const validEmailAddress = (value: string) => {
  try {
    z.email().parse(value);
    return true;
  } catch (_) {
    return false;
  }
};

export const validIPv4Address = (value: string) => {
  try {
    z.ipv4().parse(value);
    return true;
  } catch (_) {
    return false;
  }
};

export const validIPv6Address = (value: string) => {
  try {
    z.ipv6().parse(value);
    return true;
  } catch (_) {
    return false;
  }
};

export const validHttpOrHttpsUrl = (value: string) => {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
};

export const validPortNumber = (value: string | number) => {
  return parseInt(value + "") === +value && String(+value) === String(value) && +value >= 1 && +value <= 65535;
};
