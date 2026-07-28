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

const evalVisibilityFor = (col: SchemaColumn, discName: string, value: string): boolean => {
  for (const cond of col.visibleWhen ?? []) {
    if (cond.field !== discName) continue;
    if (!matchCondition(cond, value)) return false;
  }
  return true;
};

const matchesDisc = (col: SchemaColumn, clause: "requiredWhen" | "validateWhen", discName: string, value: string): boolean => {
  const list = (col[clause] ?? []) as SchemaCondition[];
  for (const cond of list) {
    if (cond.field !== discName) continue;
    if (matchCondition(cond, value)) return true;
  }
  return false;
};

const runValidator = (name: string, params: Record<string, unknown> | undefined, value: unknown): boolean => {
  const s = String(value);
  switch (name) {
    case "domain":
      return isDomain(s, { allowWildcard: !!params?.allowWildcard });
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

const fieldBaseType = (col: SchemaColumn): z.ZodType => {
  switch (col.valueType) {
    case "switch":
      return z.boolean();
    case "number": {
      let s = z.coerce.number();
      if (col.min != null) s = s.min(col.min);
      if (col.max != null) s = s.max(col.max);
      return s;
    }
    default:
      return z.string();
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
  let schema = fieldBaseType(col);
  if (col.valueType !== "switch" && col.valueType !== "number") {
    schema = required ? (schema as z.ZodString).min(1) : schema.nullish();
  } else if (!required) {
    schema = schema.nullish();
  }
  return applyRules(schema, rules, !required);
};

const detectDiscriminator = (columns: SchemaColumn[]): SchemaColumn | null => {
  const referenced = new Set<string>();
  for (const col of columns) {
    for (const clause of ["visibleWhen", "requiredWhen", "validateWhen"] as const) {
      for (const cond of (col[clause] ?? []) as SchemaCondition[]) {
        referenced.add(cond.field);
      }
    }
  }
  if (referenced.size !== 1) return null;
  const name = [...referenced][0];
  const col = columns.find((c) => c.name === name);
  if (!col || !col.options || col.options.length === 0) return null;
  return col;
};

const buildDiscriminated = (disc: SchemaColumn, columns: SchemaColumn[]): z.ZodType => {
  const variants = disc.options!.map((opt) => {
    const shape: Record<string, z.ZodType> = {
      [disc.name]: z.literal(opt.value),
    };
    for (const col of columns) {
      if (col.name === disc.name) continue;
      if (!evalVisibilityFor(col, disc.name, opt.value)) continue;
      const required = col.required === true || matchesDisc(col, "requiredWhen", disc.name, opt.value);
      const rules = (col.validateWhen ?? []).filter((r) => r.field === disc.name && matchCondition(r, opt.value));
      shape[col.name] = buildFieldSchema(col, required, rules);
    }
    return z.object(shape);
  });
  return z.discriminatedUnion(disc.name, variants as unknown as Parameters<typeof z.discriminatedUnion>[1]);
};

const buildPlainObject = (columns: SchemaColumn[]): z.ZodType => {
  const shape: Record<string, z.ZodType> = {};
  for (const col of columns) {
    shape[col.name] = buildFieldSchema(col, col.required === true, []);
  }
  const base = z.object(shape);
  return base.superRefine((values, ctx) => {
    const record = values as Record<string, unknown>;
    for (const col of columns) {
      const value = record[col.name];
      for (const cond of col.requiredWhen ?? []) {
        if (!matchCondition(cond, record[cond.field])) continue;
        if (isEmpty(value)) {
          ctx.addIssue({ code: "custom", message: i18n.t("common.errmsg.required"), path: [col.name] });
        }
      }
      for (const rule of col.validateWhen ?? []) {
        if (!matchCondition(rule, record[rule.field])) continue;
        if (!isEmpty(value) && !runValidator(rule.validator, rule.params, value)) {
          ctx.addIssue({ code: "custom", message: i18n.t(`common.errmsg.${rule.validator}_invalid`), path: [col.name] });
        }
      }
    }
  });
};

export const deriveZodSchema = (envelope: ProviderSchemaEnvelope): z.ZodType => {
  const columns = envelope.schema?.columns ?? [];
  const disc = detectDiscriminator(columns);
  if (disc) {
    return buildDiscriminated(disc, columns);
  }
  return buildPlainObject(columns);
};
