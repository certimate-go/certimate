import { describe, expect, it } from "vitest";

import type { SchemaColumn } from "@/api/providerschema";

import { buildValueEnum, discriminatorFields, isColumnVisible, mapValueType } from "./schemaValueTypes";

describe("mapValueType", () => {
  it("maps known valueTypes to ProComponents valueTypes", () => {
    expect(mapValueType("text")).toBe("text");
    expect(mapValueType("textarea")).toBe("textarea");
    expect(mapValueType("select")).toBe("select");
    expect(mapValueType("radio")).toBe("radio");
    expect(mapValueType("switch")).toBe("switch");
    expect(mapValueType("number")).toBe("digit");
    expect(mapValueType("secret")).toBe("password");
    expect(mapValueType("code")).toBe("code");
  });

  it("falls back unknown valueTypes to text (graceful degradation)", () => {
    expect(mapValueType("totally-unknown")).toBe("text");
    expect(mapValueType("monaco")).toBe("text");
    expect(mapValueType("")).toBe("text");
  });
});

describe("buildValueEnum", () => {
  const t = (k: string) => `T(${k})`;

  it("builds a valueEnum for select/radio from options", () => {
    const col: SchemaColumn = {
      name: "fmt",
      valueType: "select",
      options: [{ value: "PEM", labelKey: "k.pem" }, { value: "PFX" }],
    };
    expect(buildValueEnum(col, t)).toEqual({
      PEM: { text: "T(k.pem)" },
      PFX: { text: "PFX" },
    });
  });

  it("returns undefined for non-select fields or fields without options", () => {
    expect(buildValueEnum({ name: "x", valueType: "text" } as SchemaColumn, t)).toBeUndefined();
    expect(buildValueEnum({ name: "x", valueType: "select" } as SchemaColumn, t)).toBeUndefined();
  });
});

describe("isColumnVisible", () => {
  const col = (visibleWhen: SchemaColumn["visibleWhen"]): SchemaColumn => ({ name: "domain", valueType: "text", visibleWhen });

  it("is visible when it has no visibleWhen", () => {
    expect(isColumnVisible(col(undefined), () => undefined)).toBe(true);
  });

  it("hides when a visibleWhen condition does not hold", () => {
    const c = col([{ field: "domainMatchPattern", op: "notIn", values: ["certsan"] }]);
    expect(isColumnVisible(c, () => "certsan")).toBe(false);
    expect(isColumnVisible(c, () => "exact")).toBe(true);
  });
});

describe("discriminatorFields", () => {
  it("collects the union of dependency fields", () => {
    const cols: SchemaColumn[] = [
      { name: "a", valueType: "text", dependencies: ["fmt", "useSCP"] },
      { name: "b", valueType: "text", dependencies: ["fmt"] },
    ];
    expect(discriminatorFields(cols).sort()).toEqual(["fmt", "useSCP"]);
  });
});
