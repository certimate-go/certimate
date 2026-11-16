import "reflect-metadata";
import type * as ReactI18Next from "react-i18next";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { Form, type FormInstance } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

import { type PluginCatalogEntry } from "@/api/plugincatalog";
import * as ProviderSchemaAPI from "@/api/providerschema";
import { applyPluginCatalog } from "@/domain/pluginCatalog";
import { PURGE_PROVIDERS } from "@/domain/provider";

import BizPurgeNodeConfigForm from "./BizPurgeNodeConfigForm";

vi.mock("@/api/providerschema", async (importOriginal) => ({
  ...(await importOriginal<typeof ProviderSchemaAPI>()),
  getProviderSchema: vi.fn(),
}));
const { accessesFixture, fetchAccessesMock } = vi.hoisted(() => ({
  accessesFixture: [{ id: "access-1", provider: "baiducloud", reserve: false }],
  fetchAccessesMock: vi.fn(),
}));
vi.mock("@/stores/access", () => ({
  useAccessesStore: (selector: (state: unknown) => unknown) => selector({ accesses: accessesFixture, fetchAccesses: fetchAccessesMock }),
}));
vi.mock("@/stores/pluginCatalog", () => ({
  usePluginCatalogStore: (selector: (state: { version: number }) => unknown) => selector({ version: catalogVersionRef.value }),
}));
vi.mock("@/components/access/AccessEditDrawer", () => ({ default: () => null }));
vi.mock("@/components/access/AccessSelect", () => ({
  default: (props: { value?: string; onChange?: (v: string) => void }) => (
    <input aria-label="providerAccessId" value={props.value ?? ""} onChange={(e) => props.onChange?.(e.target.value)} />
  ),
}));
vi.mock("react-i18next", async (importOriginal) => ({
  ...(await importOriginal<typeof ReactI18Next>()),
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: "zh" } }),
  getI18n: () => ({ t: (key: string) => key, language: "zh" }),
}));

const catalogVersionRef = { value: 0 };
const legacyProvider = PURGE_PROVIDERS.BAIDUCLOUD_CERT;
const canonicalProvider = `${legacyProvider}-purger`;
const savedConfig = { provider: legacyProvider, providerAccessId: "access-1", providerConfig: { region: "bj" }, expiredDays: 7 };

const purgePluginEntry = (): PluginCatalogEntry => ({
  source: "plugin-dir",
  capability: "purge",
  providerType: canonicalProvider,
  accessProviderType: "baiducloud",
  ownsAccess: false,
  category: "",
  displayNameKey: "plugin.baiducloud-cert-purger.name",
  accessDisplayNameKey: "",
});

const purgeSchemaEnvelope = (): ProviderSchemaAPI.ProviderSchemaEnvelope => ({
  schemaVersion: "form/v1",
  provider: canonicalProvider,
  category: "purge",
  schema: { columns: [{ name: "region", valueType: "text", required: true, labelKey: "plugin.purge.region" }] },
});

function Harness({ config, onForm }: { config: Record<string, unknown>; onForm: (form: FormInstance) => void }) {
  const [form] = Form.useForm();
  onForm(form);
  const node = {
    flowNodeType: "bizPurge",
    form: { getValueIn: (key: string) => (key === "config" ? config : void 0) },
  } as never;

  return <BizPurgeNodeConfigForm form={form} node={node} />;
}

const renderForm = async (config: Record<string, unknown>) => {
  let form: FormInstance | undefined;
  render(<Harness config={config} onForm={(f) => (form = f)} />);
  await waitFor(() => expect(form).toBeDefined());
  return form!;
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
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
  );
});
afterEach(() => {
  cleanup();
  applyPluginCatalog([]);
  catalogVersionRef.value = 0;
  vi.mocked(ProviderSchemaAPI.getProviderSchema).mockReset().mockResolvedValue(null);
});
afterAll(() => vi.unstubAllGlobals());

describe("BizPurgeNodeConfigForm.getSchema", () => {
  it("accepts a legacy node config with access and expiry once the plugin is installed", () => {
    applyPluginCatalog([purgePluginEntry()]);
    const res = BizPurgeNodeConfigForm.getSchema({}).safeParse(savedConfig);
    expect(res.success).toBe(true);
  });

  it("rejects the legacy provider id while the plugin is missing", () => {
    const res = BizPurgeNodeConfigForm.getSchema({}).safeParse(savedConfig);
    expect(res.success).toBe(false);
    if (!res.success) {
      expect(res.error.issues.some((issue) => issue.path[0] === "provider")).toBe(true);
    }
  });

  it("rejects a plugin provider without access config", () => {
    applyPluginCatalog([purgePluginEntry()]);
    const res = BizPurgeNodeConfigForm.getSchema({}).safeParse({ provider: canonicalProvider, expiredDays: 7 });
    expect(res.success).toBe(false);
    if (!res.success) {
      expect(res.error.issues.some((issue) => issue.path[0] === "providerAccessId")).toBe(true);
    }
  });
});

describe("BizPurgeNodeConfigForm.purgeSchemaUnavailableRule", () => {
  const rule = BizPurgeNodeConfigForm.purgeSchemaUnavailableRule((key) => key)("workflow_node.purge.form.provider.unavailable");

  it("blocks saving while the rule is unavailable", async () => {
    await expect(rule(true).validator()).rejects.toThrow("workflow_node.purge.form.provider.unavailable");
  });

  it("passes once the schema is available", async () => {
    await expect(rule(false).validator()).resolves.toBeUndefined();
  });
});

describe("BizPurgeNodeConfigForm rendering", () => {
  it("keeps the legacy provider id and saved config while the plugin is missing", async () => {
    const form = await renderForm(savedConfig);

    expect(ProviderSchemaAPI.getProviderSchema).not.toHaveBeenCalled();
    expect(form.getFieldValue("provider")).toBe(legacyProvider);
    expect(form.getFieldValue("providerAccessId")).toBe("access-1");
    expect(form.getFieldValue("providerConfig")).toEqual({ region: "bj" });
  });

  it("normalizes the legacy id to the canonical plugin id once installed", async () => {
    applyPluginCatalog([purgePluginEntry()]);
    vi.mocked(ProviderSchemaAPI.getProviderSchema).mockResolvedValue(purgeSchemaEnvelope());

    const form = await renderForm(savedConfig);

    await waitFor(() => {
      expect(form.getFieldValue("provider")).toBe(canonicalProvider);
    });
    const values = await waitFor(() => form.validateFields());
    expect(values.provider).toBe(canonicalProvider);
    expect(form.getFieldValue("providerAccessId")).toBe("access-1");
    expect(form.getFieldValue("providerConfig")).toEqual({ region: "bj" });
  });

  it("keeps saved config when the plugin schema cannot be loaded", async () => {
    applyPluginCatalog([purgePluginEntry()]);

    const form = await renderForm(savedConfig);

    await waitFor(() => {
      expect(ProviderSchemaAPI.getProviderSchema).toHaveBeenCalledWith(canonicalProvider, "purge");
    });
    expect(form.getFieldValue("provider")).toBe(canonicalProvider);
    expect(form.getFieldValue("providerAccessId")).toBe("access-1");
    expect(form.getFieldValue("providerConfig")).toEqual({ region: "bj" });
  });

  it("renders nested schema fields from the plugin envelope", async () => {
    applyPluginCatalog([purgePluginEntry()]);
    vi.mocked(ProviderSchemaAPI.getProviderSchema).mockResolvedValue(purgeSchemaEnvelope());

    const form = await renderForm(savedConfig);

    await waitFor(() => {
      expect(form.getFieldValue("provider")).toBe(canonicalProvider);
    });
    await waitFor(() => {
      expect(screen.getByText("plugin.purge.region")).toBeInTheDocument();
    });
    const values = await form.validateFields();
    expect(values.provider).toBe(canonicalProvider);
  });
});

describe("BizPurgeNodeConfigForm zod rule", () => {
  it("attaches the provider issue to the provider field", async () => {
    let form: FormInstance | undefined;
    const onForm = (inst: FormInstance) => {
      form = inst;
    };
    const Mini = () => {
      const [f] = Form.useForm();
      onForm(f);
      const rule = createSchemaFieldRule(BizPurgeNodeConfigForm.getSchema({}));
      return (
        <Form form={f} initialValues={{ provider: legacyProvider, providerAccessId: "access-1", expiredDays: 7 }}>
          <Form.Item name="provider" rules={[rule]}>
            <input />
          </Form.Item>
          <Form.Item name="providerAccessId" rules={[rule]}>
            <input />
          </Form.Item>
        </Form>
      );
    };
    render(<Mini />);
    await expect(form!.validateFields()).rejects.toMatchObject({
      errorFields: [expect.objectContaining({ name: ["provider"] })],
    });
  });
});
