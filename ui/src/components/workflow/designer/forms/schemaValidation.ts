import { z } from "zod";

import type { ProviderSchemaEnvelope, SchemaColumn, SchemaCondition, SchemaValidationRule } from "@/api/providerschema";
import i18n from "@/i18n";
import {
  isCron,
  isDomain,
  isHostname,
  isIPv4,
  isIPv6,
  isJsonObject,
  isPortNumber,
  isUrlWithHttp,
  isUrlWithHttpOrHttps,
  isUrlWithHttps,
} from "@/utils/validator";

const isEmpty = (v: unknown) => v == null || v === "";

export const matchCondition = (cond: SchemaCondition, value: unknown): boolean => {
  const present = (cond.values ?? []).includes(String(value));
  switch (cond.op) {
    case "eq":
    case "in":
      return present;
    case "ne":
    case "notIn":
      return !present;
    default:
      return true;
  }
};

const runValidator = (name: string, params: Record<string, unknown> | undefined, value: unknown): boolean => {
  const s = String(value);
  switch (name) {
    case "positiveInteger":
      return z.coerce.number().int().positive().safeParse(value).success;
    case "domain":
      return isDomain(s, { allowWildcard: !!params?.allowWildcard });
    case "hostname_or_url":
      return isHostname(s) || isUrlWithHttpOrHttps(s);
    case "hostname":
      return isHostname(s);
    case "ipv4":
      return isIPv4(s);
    case "ipv6":
      return isIPv6(s);
    case "port":
      return isPortNumber(value as string | number);
    case "url":
      return isUrlWithHttpOrHttps(s);
    case "url_http":
      return isUrlWithHttp(s);
    case "url_https":
      return isUrlWithHttps(s);
    case "json_object":
      return isJsonObject(s);
    case "cron":
      return isCron(s);
    case "regex": {
      const pattern = params?.pattern;
      if (typeof pattern !== "string") return true;
      return new RegExp(pattern).test(s);
    }
    default:
      return true;
  }
};

const fieldBaseType = (col: SchemaColumn, required: boolean): z.ZodType => {
  switch (col.valueType) {
    case "switch":
      return z.boolean();
    case "multipleSplit":
      return z.string().refine((value) => (!required && value === "") || value.split(col.separator ?? ";").every((part) => part.trim().length > 0), {
        message: i18n.t("common.errmsg.multiple_values_invalid"),
      });
    case "select":
    case "radio":
      if (col.options?.length && !col.allowCustomValue) return z.enum(col.options.map((option) => option.value));
      return required ? z.string().min(1) : z.string();
    case "number": {
      let s = z.coerce.number();
      if (col.min != null) s = s.min(col.min);
      if (col.max != null) s = s.max(col.max);
      return s;
    }
    default: {
      const text = required ? z.string().min(1) : z.string();
      if (col.valueType === "text" && col.numericValue) {
        const integer = col.numericValue === "positiveInteger" ? z.int().positive() : z.int();
        return z.union([text, integer]);
      }
      return text;
    }
  }
};

const applyRules = (schema: z.ZodType, rules: SchemaValidationRule[], allowEmpty: boolean): z.ZodType => {
  for (const rule of rules) {
    const ok = (v: unknown) => (allowEmpty && isEmpty(v)) || runValidator(rule.validator, rule.params, v);
    schema = schema.refine(ok, { message: i18n.t(`common.errmsg.${rule.validator}_invalid`) });
  }
  return schema;
};

const buildFieldSchema = (col: SchemaColumn, required: boolean, rules: SchemaValidationRule[]): z.ZodType => {
  let schema = fieldBaseType(col, required);
  if (!required) {
    schema = schema.nullish();
  }
  return applyRules(schema, rules, !required);
};

export const deriveZodSchema = (envelope: ProviderSchemaEnvelope): z.ZodType => {
  const columns = envelope.schema?.columns ?? [];
  const shape = Object.fromEntries(columns.map((col) => [col.name, z.unknown().optional()]));
  const defaults = new Map(columns.filter((col) => col.defaultOnMissing).map((col) => [col.name, col.default]));
  return z.object(shape).transform((values, ctx) => {
    const getValue = (field: string): unknown => (values[field] === undefined ? defaults.get(field) : values[field]);
    const parsed: Record<string, unknown> = {};
    for (const col of columns) {
      if (!(col.visibleWhen ?? []).every((condition) => matchCondition(condition, getValue(condition.field)))) continue;
      const required = col.required === true || (col.requiredWhen ?? []).some((condition) => matchCondition(condition, getValue(condition.field)));
      const rules = (col.validateWhen ?? []).filter((rule) => matchCondition(rule, getValue(rule.field)));
      const result = buildFieldSchema(col, required, rules).safeParse(getValue(col.name));
      if (!result.success) {
        for (const issue of result.error.issues) {
          ctx.addIssue({ ...issue, path: [col.name, ...issue.path] });
        }
      } else if (col.name in values || defaults.has(col.name)) {
        parsed[col.name] = result.data;
      }
    }
    return parsed;
  });
};
