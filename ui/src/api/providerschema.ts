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
  labelKey?: string;
  placeholderKey?: string;
  tooltipKey?: string;
  extraKey?: string;
  tooltipHtml?: boolean;
  extraHtml?: boolean;
  default?: unknown;
  secret?: boolean;
  required?: boolean;
  options?: SchemaOption[];
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
    columns: SchemaColumn[];
  };
}

const CURRENT_SCHEMA_VERSION = "form/v1";

const cache = new Map<string, ProviderSchemaEnvelope>();

const cacheKey = (provider: string, schemaVersion: string) => `${schemaVersion}::${provider}`;

export const getProviderSchema = async (provider: string): Promise<ProviderSchemaEnvelope | null> => {
  const key = cacheKey(provider, CURRENT_SCHEMA_VERSION);
  const cached = cache.get(key);
  if (cached) {
    return cached;
  }

  try {
    const resp = await httpGet<ProviderSchemaEnvelope>({
      url: `/api/provider-schemas/${encodeURIComponent(provider)}`,
    });
    const envelope = resp.data;
    if (envelope && envelope.schemaVersion === CURRENT_SCHEMA_VERSION) {
      cache.set(key, envelope);
    }
    return envelope ?? null;
  } catch {
    return null;
  }
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
  cache.clear();
};

export const isProviderSchemaCached = (provider: string) => {
  return cache.has(cacheKey(provider, CURRENT_SCHEMA_VERSION));
};
