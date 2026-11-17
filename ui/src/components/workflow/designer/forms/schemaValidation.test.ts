import { describe, expect, it } from "vitest";

import type { ProviderSchemaEnvelope, SchemaConditionOp } from "@/api/providerschema";

import { deriveZodSchema, matchCondition } from "./schemaValidation";

const envelope = (columns: ProviderSchemaEnvelope["schema"]["columns"]): ProviderSchemaEnvelope => ({
  schemaVersion: "form/v1",
  provider: "test",
  category: "deploy",
  schema: { columns },
});

describe("deriveZodSchema - plain object", () => {
  it("validates required vs optional fields", () => {
    const schema = deriveZodSchema(
      envelope([
        { name: "region", valueType: "text", required: true },
        { name: "note", valueType: "text" },
      ])
    );

    expect(schema.safeParse({ region: "cn-hangzhou", note: "x" }).success).toBe(true);
    expect(schema.safeParse({ region: "", note: "x" }).success).toBe(false);
    expect(schema.safeParse({ note: "x" }).success).toBe(false);
    expect(schema.safeParse({ region: "cn-hangzhou" }).success).toBe(true);
  });

  it("treats unknown valueType as string (graceful fallback)", () => {
    const schema = deriveZodSchema(envelope([{ name: "x", valueType: "totally-unknown", required: true }]));
    expect(schema.safeParse({ x: "ok" }).success).toBe(true);
    expect(schema.safeParse({ x: "" }).success).toBe(false);
  });

  it("validates number with min/max", () => {
    const schema = deriveZodSchema(envelope([{ name: "port", valueType: "number", min: 1, max: 65535 }]));
    expect(schema.safeParse({ port: 443 }).success).toBe(true);
    expect(schema.safeParse({ port: 0 }).success).toBe(false);
    expect(schema.safeParse({ port: 70000 }).success).toBe(false);
  });
});

describe("deriveZodSchema - discriminated union", () => {
  const cdn = (): ProviderSchemaEnvelope =>
    envelope([
      { name: "domainMatchPattern", valueType: "radio", options: [{ value: "exact" }, { value: "wildcard" }, { value: "certsan" }] },
      {
        name: "domain",
        valueType: "text",
        visibleWhen: [{ field: "domainMatchPattern", op: "notIn", values: ["certsan"] }],
        requiredWhen: [{ field: "domainMatchPattern", op: "in", values: ["exact", "wildcard"] }],
        validateWhen: [{ field: "domainMatchPattern", op: "in", values: ["exact", "wildcard"], validator: "domain", params: { allowWildcard: true } }],
      },
    ]);

  it("requires domain for exact/wildcard but not certsan", () => {
    const schema = deriveZodSchema(cdn());
    expect(schema.safeParse({ domainMatchPattern: "exact", domain: "example.com" }).success).toBe(true);
    expect(schema.safeParse({ domainMatchPattern: "exact", domain: "" }).success).toBe(false);
    expect(schema.safeParse({ domainMatchPattern: "certsan" }).success).toBe(true);
  });

  it("runs the domain validator on matching variants", () => {
    const schema = deriveZodSchema(cdn());
    expect(schema.safeParse({ domainMatchPattern: "exact", domain: "not a domain" }).success).toBe(false);
    expect(schema.safeParse({ domainMatchPattern: "wildcard", domain: "*.example.com" }).success).toBe(true);
  });

  it("rejects an unknown discriminator value", () => {
    const schema = deriveZodSchema(cdn());
    expect(schema.safeParse({ domainMatchPattern: "bogus", domain: "example.com" }).success).toBe(false);
  });
});

describe("deriveZodSchema - multi-discriminator falls back to object+superRefine", () => {
  it("enforces conditional required across two discriminators without throwing", () => {
    const schema = deriveZodSchema(
      envelope([
        { name: "useSCP", valueType: "switch" },
        { name: "fileFormat", valueType: "select", options: [{ value: "PEM" }, { value: "PFX" }] },
        {
          name: "pfxPassword",
          valueType: "text",
          requiredWhen: [{ field: "fileFormat", op: "eq", values: ["PFX"] }],
          visibleWhen: [{ field: "fileFormat", op: "eq", values: ["PFX"] }],
        },
      ])
    );

    expect(schema.safeParse({ useSCP: false, fileFormat: "PEM" }).success).toBe(true);
    expect(schema.safeParse({ useSCP: false, fileFormat: "PFX", pfxPassword: "secret" }).success).toBe(true);
    expect(schema.safeParse({ useSCP: false, fileFormat: "PFX", pfxPassword: "" }).success).toBe(false);
  });
});

describe("matchCondition", () => {
  it("evaluates eq/ne/in/notIn", () => {
    const c = (op: SchemaConditionOp, values: string[]) => ({ field: "f", op, values });
    expect(matchCondition(c("eq", ["a"]), "a")).toBe(true);
    expect(matchCondition(c("eq", ["a"]), "b")).toBe(false);
    expect(matchCondition(c("ne", ["a"]), "b")).toBe(true);
    expect(matchCondition(c("in", ["a", "b"]), "b")).toBe(true);
    expect(matchCondition(c("notIn", ["a"]), "b")).toBe(true);
    expect(matchCondition(c("notIn", ["a"]), "a")).toBe(false);
  });
});

describe("deriveZodSchema - graceful degradation (future plugin)", () => {
  it("derives without throwing for a future schemaVersion + unknown valueType", () => {
    const future = envelope([
      { name: "futureField", valueType: "monaco-editor" },
      { name: "port", valueType: "number", min: 1, max: 65535 },
    ]);
    future.schemaVersion = "form/v999";

    expect(() => deriveZodSchema(future)).not.toThrow();

    const schema = deriveZodSchema(future);
    expect(schema.safeParse({ futureField: "anything" }).success).toBe(true);
    expect(schema.safeParse({ port: 443 }).success).toBe(true);
    expect(schema.safeParse({ port: 0 }).success).toBe(false);
  });
});

describe("saved identifier types", () => {
  it.each([42, "42", "host-name"])("preserves a saved Mohua identifier %s", (hostId) => {
    const schema = deriveZodSchema(envelope([{ name: "hostId", valueType: "text", numericValue: "positiveInteger", required: true }]));
    expect(schema.parse({ hostId })).toEqual({ hostId });
  });

  it.each([0, -1, 1.5, "", undefined, null])("rejects an invalid required identifier %s", (hostId) => {
    const schema = deriveZodSchema(envelope([{ name: "hostId", valueType: "text", numericValue: "positiveInteger", required: true }]));
    expect(schema.safeParse({ hostId }).success).toBe(false);
  });

  it("preserves optional integer text and validates the selected target", () => {
    const schema = deriveZodSchema(
      envelope([
        { name: "deployTarget", valueType: "select", options: [{ value: "certificate" }] },
        {
          name: "certificateId",
          valueType: "text",
          numericValue: "integer",
          inputType: "number",
          requiredWhen: [{ field: "deployTarget", op: "eq", values: ["certificate"] }],
          validateWhen: [{ field: "deployTarget", op: "eq", values: ["certificate"], validator: "positiveInteger" }],
        },
        { name: "clientId", valueType: "text", numericValue: "integer", inputType: "number" },
      ])
    );
    for (const certificateId of [42, "42"]) {
      const value = { deployTarget: "certificate", certificateId, clientId: "" };
      expect(schema.parse(value)).toEqual(value);
    }
    for (const certificateId of ["", "bad", "0", 0, -1, 1.5, "1.5"]) {
      expect(schema.safeParse({ deployTarget: "certificate", certificateId }).success).toBe(false);
    }
  });
});

describe("panel form conditions", () => {
  const schema = deriveZodSchema(
    envelope([
      { name: "deployTarget", valueType: "select", required: true, options: [{ value: "website" }, { value: "certificate" }] },
      {
        name: "match",
        valueType: "radio",
        options: [{ value: "exact" }, { value: "certsan" }],
        visibleWhen: [{ field: "deployTarget", op: "eq", values: ["website"] }],
      },
      {
        name: "websiteId",
        valueType: "text",
        required: true,
        visibleWhen: [
          { field: "deployTarget", op: "eq", values: ["website"] },
          { field: "match", op: "eq", values: ["exact"] },
        ],
      },
      { name: "certificateId", valueType: "number", required: true, visibleWhen: [{ field: "deployTarget", op: "eq", values: ["certificate"] }] },
    ])
  );

  it("validates only fields visible for the selected target and matching mode", () => {
    expect(schema.parse({ deployTarget: "website", match: "certsan", certificateId: "bad" })).toEqual({ deployTarget: "website", match: "certsan" });
    expect(schema.safeParse({ deployTarget: "website", match: "exact" }).success).toBe(false);
    expect(schema.parse({ deployTarget: "certificate", certificateId: "42", match: "bad" })).toEqual({ deployTarget: "certificate", certificateId: 42 });
  });

  it("preserves semicolon-separated site names and rejects empty entries", () => {
    const sites = deriveZodSchema(envelope([{ name: "siteNames", valueType: "multipleSplit", required: true, separator: ";" }]));
    expect(sites.parse({ siteNames: "one; two" })).toEqual({ siteNames: "one; two" });
    for (const siteNames of ["", "one;", "one; ;two"]) expect(sites.safeParse({ siteNames }).success).toBe(false);
  });

  it("accepts device hostnames and HTTP URLs with the same validation", () => {
    const host = deriveZodSchema(
      envelope([{ name: "host", valueType: "text", required: true, validateWhen: [{ field: "host", op: "ne", values: [""], validator: "hostname_or_url" }] }])
    );
    for (const value of ["ibmc.example.com", "https://ibmc.example.com:443"]) expect(host.safeParse({ host: value }).success).toBe(true);
    for (const value of ["", "bad host", "ftp://example.com"]) expect(host.safeParse({ host: value }).success).toBe(false);
  });
});

it("accepts a saved backend-supported choice absent from the original select options", () => {
  const schema = deriveZodSchema(envelope([{ name: "siteType", valueType: "select", required: true, options: [{ value: "php" }], allowCustomValue: true }]));
  expect(schema.parse({ siteType: "aspx" })).toEqual({ siteType: "aspx" });
  expect(schema.safeParse({ siteType: "" }).success).toBe(false);
});
