import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Form } from "antd";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

import type { ProviderSchemaEnvelope } from "@/api/providerschema";

import { FormNestedFieldsContextProvider } from "./_context";
import SchemaConfigFields from "./SchemaConfigFields";

vi.mock("@/components/CodeTextInput", () => ({ default: () => null }));

const envelope: ProviderSchemaEnvelope = {
  schemaVersion: "form/v1",
  provider: "numeric-fixture",
  category: "deploy",
  schema: {
    columns: [{ name: "certificateId", valueType: "number", required: true, min: 1, max: Number.MAX_SAFE_INTEGER }],
  },
};

const flexEnvelope: ProviderSchemaEnvelope = {
  schemaVersion: "form/v1",
  provider: "flexcdn",
  category: "deploy",
  schema: {
    columns: [
      { name: "deployTarget", valueType: "select", required: true, default: "certificate", options: [{ value: "certificate" }] },
      {
        name: "certificateId",
        valueType: "number",
        required: true,
        min: 1,
        max: Number.MAX_SAFE_INTEGER,
        dependencies: ["deployTarget"],
        visibleWhen: [{ field: "deployTarget", op: "eq", values: ["certificate"] }],
        validateWhen: [{ field: "deployTarget", op: "eq", values: ["certificate"], validator: "regex", params: { pattern: "^[1-9][0-9]*$" } }],
      },
    ],
  },
};

beforeAll(() => {
  vi.stubGlobal("matchMedia", (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  }));
});
afterEach(cleanup);
afterAll(() => vi.unstubAllGlobals());

describe.each([
  { name: "plain numeric schema", schema: envelope, invalidValues: ["-1", "9007199254740993"], extra: {} },
  { name: "conditional FlexCDN schema", schema: flexEnvelope, invalidValues: ["-1", "9007199254740993", "1.5", ""], extra: { deployTarget: "certificate" } },
])("$name", ({ schema, invalidValues, extra }) => {
  const scenarios: Array<[number | string, string]> = [42, "42"].flatMap((initialId) =>
    invalidValues.map((inputValue): [number | string, string] => [initialId, inputValue])
  );
  it.each(scenarios)("accepts saved %s and rejects %s without clamping the identifier", async (initialId, inputValue) => {
    const finish = vi.fn();
    const failed = vi.fn();
    render(
      <Form layout="vertical" initialValues={{ providerConfig: { ...extra, certificateId: initialId } }} onFinish={finish} onFinishFailed={failed}>
        <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
          <SchemaConfigFields envelope={schema} />
        </FormNestedFieldsContextProvider>
        <button type="submit">Save</button>
      </Form>
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith({ providerConfig: { ...extra, certificateId: initialId } }));
    finish.mockClear();
    const input = screen.getByRole("spinbutton");
    if (!(input instanceof HTMLInputElement)) throw new Error("numeric field must render an input");
    fireEvent.change(input, { target: { value: inputValue } });
    fireEvent.blur(input);
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(failed).toHaveBeenCalledOnce());
    expect(finish).not.toHaveBeenCalled();
    expect(Number(input.value)).toBe(Number(inputValue));
    expect(input.value).not.toBe(String(Number.MAX_SAFE_INTEGER));
    expect(input.value).not.toBe("42");
    fireEvent.change(input, { target: { value: "42" } });
    fireEvent.blur(input);
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith({ providerConfig: { ...extra, certificateId: 42 } }));
  });
});

describe("saved text identifiers", () => {
  it.each([42, "42"])("preserves %s until the user edits it", async (initialId) => {
    const finish = vi.fn();
    const schema: ProviderSchemaEnvelope = {
      schemaVersion: "form/v1",
      provider: "saved-id",
      category: "deploy",
      schema: { columns: [{ name: "hostId", valueType: "text", numericValue: "positiveInteger", required: true }] },
    };
    render(
      <Form initialValues={{ providerConfig: { hostId: initialId } }} onFinish={finish}>
        <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
          <SchemaConfigFields envelope={schema} />
        </FormNestedFieldsContextProvider>
        <button type="submit">Save</button>
      </Form>
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith({ providerConfig: { hostId: initialId } }));
    fireEvent.change(screen.getByRole("textbox"), { target: { value: "43" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith({ providerConfig: { hostId: "43" } }));
  });
});
