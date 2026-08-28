import type { SchemaColumn } from "@/api/providerschema";

import { matchCondition } from "./schemaValidation";

export const mapValueType = (valueType: string): string => {
  switch (valueType) {
    case "text":
      return "text";
    case "textarea":
      return "textarea";
    case "select":
      return "select";
    case "radio":
      return "radio";
    case "switch":
      return "switch";
    case "number":
      return "digit";
    case "secret":
      return "password";
    case "code":
      return "code";
    case "autocomplete":
      return "text";
    default:
      return "text";
  }
};

export const buildValueEnum = (col: SchemaColumn, t: (key: string) => string): Record<string, { text: string }> | undefined => {
  if (!col.options) return undefined;
  if (col.valueType !== "select" && col.valueType !== "radio") return undefined;
  const valueEnum: Record<string, { text: string }> = {};
  for (const opt of col.options) {
    valueEnum[opt.value] = { text: opt.labelKey ? t(opt.labelKey) : opt.value };
  }
  return valueEnum;
};

export const isColumnVisible = (col: SchemaColumn, getValue: (field: string) => unknown): boolean => {
  for (const cond of col.visibleWhen ?? []) {
    if (!matchCondition(cond, getValue(cond.field))) return false;
  }
  return true;
};

export const discriminatorFields = (columns: SchemaColumn[]): string[] => {
  const seen = new Set<string>();
  for (const col of columns) {
    for (const dep of col.dependencies ?? []) {
      seen.add(dep);
    }
  }
  return [...seen];
};

export const unitFieldProps = (
  mappedValueType: string,
  unitKey: string | undefined,
  t: (key: string) => string
): Record<string, unknown> => {
  if (!unitKey) return {};
  const text = t(unitKey);
  return mappedValueType === "digit" ? { addonAfter: text } : { suffix: text };
};
