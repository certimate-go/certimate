import { get as httpGet } from "./_api";

export type SchemaCategory = "deploy" | "access" | "apply" | "notify";

export type SchemaFilterMode = "fuzzy" | "prefix" | "none";

export type SchemaConditionOp = "eq" | "ne" | "in" | "notIn";

export interface SchemaOption {
  value: string;
  labelKey?: string;
}

export interface SchemaCondition {
  field: string;
  op: SchemaConditionOp;
  values: string[];
}

export interface SchemaValidationRule extends SchemaCondition {
  validator: string;
  params?: Record<string, unknown>;
}

export interface SchemaColumn {
  name: string;
  valueType: string;
  numericValue?: "integer" | "positiveInteger";
  inputType?: "number";
  allowClear?: boolean;
  separator?: string;
  modalTitleKey?: string;
  modalPlaceholderKey?: string;
  labelKey?: string;
  placeholderKey?: string;
  tooltipKey?: string;
  extraKey?: string;
  tooltipHtml?: boolean;
  extraHtml?: boolean;
  extraWhen?: SchemaCondition[];
  default?: unknown;
  defaultOnMissing?: boolean;
  secret?: boolean;
  required?: boolean;
  options?: SchemaOption[];
  allowCustomValue?: boolean;
  min?: number;
  max?: number;
  language?: string;
  unitKey?: string;
  filterMode?: SchemaFilterMode;
  span?: number;
  dependencies?: string[];
  visibleWhen?: SchemaCondition[];
  requiredWhen?: SchemaCondition[];
  validateWhen?: SchemaValidationRule[];
}

export interface ProviderSchemaEnvelope {
  schemaVersion: string;
  provider: string;
  category: SchemaCategory;
  schema: {
    noticeKey?: string;
    noticeHtml?: boolean;
    columns: SchemaColumn[];
  };
}

const CURRENT_SCHEMA_VERSION = "form/v1";

const cache = new Map<string, ProviderSchemaEnvelope>();
const pending = new Map<string, Promise<ProviderSchemaEnvelope | null>>();

let cacheGeneration = 0;

const cacheKey = (provider: string, schemaVersion: string, category?: SchemaCategory) => `${schemaVersion}::${category ?? ""}::${provider}`;

export const getProviderSchema = (provider: string, category?: SchemaCategory): Promise<ProviderSchemaEnvelope | null> => {
  const key = cacheKey(provider, CURRENT_SCHEMA_VERSION, category);
  const generation = cacheGeneration;
  const cached = cache.get(key);
  if (cached) {
    return Promise.resolve(cached);
  }
  const existing = pending.get(key);
  if (existing) {
    return existing;
  }

  const request = (async (): Promise<ProviderSchemaEnvelope | null> => {
    try {
      const resp = await httpGet<ProviderSchemaEnvelope>({
        url: `/api/provider-schemas/${encodeURIComponent(provider)}${category ? `?category=${encodeURIComponent(category)}` : ""}`,
      });
      if (generation !== cacheGeneration) {
        return getProviderSchema(provider, category);
      }
      const envelope = resp.data;
      if (envelope && envelope.schemaVersion === CURRENT_SCHEMA_VERSION && (!category || envelope.category === category)) {
        cache.set(key, envelope);
      }
      return envelope && (!category || envelope.category === category) ? envelope : null;
    } catch {
      if (generation !== cacheGeneration) {
        return getProviderSchema(provider, category);
      }
      return null;
    } finally {
      if (generation === cacheGeneration) {
        pending.delete(key);
      }
    }
  })();
  pending.set(key, request);
  return request;
};

export const listProviderSchemas = async (): Promise<ProviderSchemaEnvelope[]> => {
  try {
    const resp = await httpGet<ProviderSchemaEnvelope[]>({
      url: `/api/provider-schemas`,
    });
    return resp.data ?? [];
  } catch {
    return [];
  }
};

export const clearProviderSchemaCache = () => {
  cacheGeneration++;
  cache.clear();
  pending.clear();
};
