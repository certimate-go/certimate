import type { FormInstance } from "antd";

export const TRIM_EXEMPT_FIELD_NAMES = new Set([
  "accesskey",
  "accesskeyid",
  "accesskeypassword",
  "accesskeysecret",
  "accesstoken",
  "accesstokenid",
  "apiaccesstoken",
  "apikey",
  "apikeysecret",
  "apipassword",
  "apisecret",
  "apitoken",
  "apitokenforzone",
  "apitokensecret",
  "appkey",
  "applicationkey",
  "applicationsecret",
  "applicationtoken",
  "authpassword",
  "bottoken",
  "clientsecret",
  "clienttoken",
  "confirmpassword",
  "consumerkey",
  "credentials",
  "eabhmackey",
  "httptoken",
  "jkskeypass",
  "jksstorepass",
  "key",
  "keypassphrase",
  "mtlsprivatekey",
  "newpassword",
  "oldpassword",
  "password",
  "personalaccesstoken",
  "pfxpassword",
  "privatekey",
  "privatekeypassphrase",
  "secret",
  "secretaccesskey",
  "secretapikey",
  "secretid",
  "secretkey",
  "serviceaccountkey",
  "token",
  "totpsecret",
  "tsiggsspassword",
  "tsigkey",
  "tsigsecret",
]);

const isPlainObject = (value: unknown): value is Record<string, unknown> => {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const proto = Object.getPrototypeOf(value);
  return proto === Object.prototype || proto === null;
};

const trimValue = (value: unknown, key?: string): unknown => {
  if (typeof value === "string") {
    return key != null && TRIM_EXEMPT_FIELD_NAMES.has(key.toLowerCase()) ? value : value.trim();
  }

  if (Array.isArray(value)) {
    return value.map((item) => trimValue(item));
  }

  if (isPlainObject(value)) {
    const obj: Record<string, unknown> = {};
    Object.keys(value).forEach((k) => {
      obj[k] = trimValue(value[k], k);
    });
    return obj;
  }

  return value;
};

export const trimFormValues = <T>(values: T): T => {
  return trimValue(values) as T;
};

export const applyTrimmedFormValues = (formInst: FormInstance): void => {
  formInst.setFieldsValue(trimFormValues(formInst.getFieldsValue(true)));
};
