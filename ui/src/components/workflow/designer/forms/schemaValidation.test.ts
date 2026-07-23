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
