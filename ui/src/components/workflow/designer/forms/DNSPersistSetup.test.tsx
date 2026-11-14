import { useEffect } from "react";
import type * as ReactI18Next from "react-i18next";
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Button, Form, Input } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import "reflect-metadata";
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { checkDNSPersist, prepareDNSPersist } from "@/api/acmeAccounts";

import BizApplyNodeConfigForm from "./BizApplyNodeConfigForm";
import DNSPersistSetup from "./DNSPersistSetup";

vi.mock("@/api/acmeAccounts", () => ({ prepareDNSPersist: vi.fn(), checkDNSPersist: vi.fn() }));
vi.mock("react-i18next", async (importOriginal) => ({
  ...(await importOriginal<typeof ReactI18Next>()),
  useTranslation: () => ({ t: (key: string) => key }),
  getI18n: () => ({ t: (key: string) => key }),
}));

const prefix = "workflow_node.apply.form.dns_persist.";
const preparation = {
  accountId: "account-1",
  accountUrl: "https://ca.example/acct/1",
  records: [
    { domain: "example.com", name: "_validation-persist.example.com.", type: "TXT" as const, value: "ca.example; accounturi=https://ca.example/acct/1" },
  ],
};
const response = { code: 0, message: "", data: preparation };
const initialValues = {
  identifier: "domain",
  challengeType: "dns-persist-01",
  domains: "example.com;*.example.com",
  contactEmail: "certs@example.com",
  keySource: "auto",
  keyAlgorithm: "RSA2048",
  skipBeforeExpiryDays: 30,
};
const finish = vi.fn();
const failed = vi.fn();

function Harness({ saved = false, loadLater = false }: { saved?: boolean; loadLater?: boolean }) {
  const [form] = Form.useForm();
  const values = { ...initialValues, dnsPersistAccountId: saved ? preparation.accountId : undefined };
  const rule = createSchemaFieldRule(BizApplyNodeConfigForm.getSchema({}));
  useEffect(() => {
    if (loadLater) Promise.resolve().then(() => form.setFieldsValue({ ...initialValues, dnsPersistAccountId: preparation.accountId }));
  }, [form, loadLater]);
  return (
    <Form form={form} initialValues={loadLater ? undefined : values} onFinish={finish} onFinishFailed={failed}>
      {[
        "identifier",
        "challengeType",
        "domains",
        "contactEmail",
        "caProvider",
        "caProviderAccessId",
        "keySource",
        "keyAlgorithm",
        "acmeProfile",
        "nameservers",
        "skipBeforeExpiryDays",
      ].map((name) => (
        <Form.Item key={name} name={name} rules={[rule]}>
          <Input aria-label={name} />
        </Form.Item>
      ))}
      <Form.Item name="dnsPersistAccountId" rules={[rule]}>
        <DNSPersistSetup enabled form={form} />
      </Form.Item>
      <Button htmlType="submit">Save</Button>
    </Form>
  );
}

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
beforeEach(() => {
  vi.mocked(prepareDNSPersist).mockResolvedValue(response);
  vi.mocked(checkDNSPersist).mockResolvedValue({
    code: 0,
    message: "",
    data: { ready: true, records: preparation.records.map((record) => ({ ...record, status: "ready" })) },
  });
});
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});
afterAll(() => vi.unstubAllGlobals());

describe("DNS-PERSIST setup in the request form", () => {
  it("preserves an asynchronously loaded binding and requires regeneration after editing domains", async () => {
    render(<Harness saved loadLater />);
    await screen.findByText(prefix + "saved_binding");
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith(expect.objectContaining({ dnsPersistAccountId: preparation.accountId })));
    expect(prepareDNSPersist).not.toHaveBeenCalled();
    fireEvent.change(screen.getByLabelText("domains"), { target: { value: "changed.example.com" } });
    await waitFor(() => expect(screen.queryByText(prefix + "saved_binding")).not.toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(failed).toHaveBeenCalledOnce());
    expect(screen.getByText(prefix + "binding_required")).toBeInTheDocument();
  });

  it("generates without DNS credentials or a prior binding and retains the binding when lookup settings change", async () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: prefix + "generate" }));
    await screen.findByText(preparation.accountUrl);
    expect(prepareDNSPersist).toHaveBeenCalledWith(
      expect.objectContaining({ domains: ["example.com", "*.example.com"], contactEmail: "certs@example.com", keyAlgorithm: "RSA2048" })
    );
    expect(screen.getByText(preparation.records[0].name)).toBeInTheDocument();
    expect(screen.getByText(preparation.records[0].value)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: prefix + "check" }));
    await screen.findByText(prefix + "ready");
    fireEvent.change(screen.getByLabelText("nameservers"), { target: { value: "1.1.1.1" } });
    await waitFor(() => expect(screen.queryByText(prefix + "ready")).not.toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(finish).toHaveBeenCalledWith(expect.objectContaining({ dnsPersistAccountId: preparation.accountId })));
  });

  it("ignores a preparation response after the contact email changes", async () => {
    let resolve: ((value: typeof response) => void) | undefined;
    vi.mocked(prepareDNSPersist).mockImplementationOnce(
      () =>
        new Promise((done) => {
          resolve = done;
        })
    );
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: prefix + "generate" }));
    await waitFor(() => expect(prepareDNSPersist).toHaveBeenCalledOnce());
    fireEvent.change(screen.getByLabelText("contactEmail"), { target: { value: "changed@example.com" } });
    await act(async () => resolve?.(response));
    expect(screen.queryByText(preparation.accountUrl)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(failed).toHaveBeenCalledOnce());
    fireEvent.click(screen.getByRole("button", { name: prefix + "generate" }));
    await screen.findByText(preparation.accountUrl);
    expect(prepareDNSPersist).toHaveBeenLastCalledWith(expect.objectContaining({ contactEmail: "changed@example.com" }));
  });

  it("ignores a check response after the CA changes and recovers after a lookup failure", async () => {
    let resolve: ((value: Awaited<ReturnType<typeof checkDNSPersist>>) => void) | undefined;
    vi.mocked(checkDNSPersist).mockImplementationOnce(
      () =>
        new Promise((done) => {
          resolve = done;
        })
    );
    render(<Harness />);
    fireEvent.click(screen.getByRole("button", { name: prefix + "generate" }));
    await screen.findByText(preparation.accountUrl);
    fireEvent.click(screen.getByRole("button", { name: prefix + "check" }));
    await waitFor(() => expect(checkDNSPersist).toHaveBeenCalledOnce());
    fireEvent.change(screen.getByLabelText("caProvider"), { target: { value: "letsencrypt" } });
    await act(async () => resolve?.({ code: 0, message: "", data: { ready: true, records: [] } }));
    expect(screen.queryByText(prefix + "ready")).not.toBeInTheDocument();
    expect(screen.queryByText(preparation.accountUrl)).not.toBeInTheDocument();
    vi.mocked(checkDNSPersist).mockRejectedValueOnce(new Error("Resolver unavailable"));
    fireEvent.click(screen.getByRole("button", { name: prefix + "generate" }));
    await screen.findByText(preparation.accountUrl);
    fireEvent.click(screen.getByRole("button", { name: prefix + "check" }));
    await screen.findByText("Resolver unavailable");
    expect(screen.getByText(preparation.records[0].value)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: prefix + "check" }));
    await screen.findByText(prefix + "ready");
  });

  it("rejects IP identifiers and retains required provider validation for DNS-01", () => {
    const schema = BizApplyNodeConfigForm.getSchema({});
    expect(schema.safeParse({ ...initialValues, dnsPersistAccountId: preparation.accountId }).success).toBe(true);
    const ipResult = schema.safeParse({
      ...initialValues,
      identifier: "ip",
      domains: undefined,
      ipaddrs: "192.0.2.1",
      dnsPersistAccountId: preparation.accountId,
    });
    expect(ipResult.success).toBe(false);
    if (!ipResult.success) expect(ipResult.error.issues).toEqual(expect.arrayContaining([expect.objectContaining({ path: ["challengeType"] })]));
    const dnsResult = schema.safeParse({ ...initialValues, challengeType: "dns-01" });
    expect(dnsResult.success).toBe(false);
    if (!dnsResult.success) expect(dnsResult.error.issues).toEqual(expect.arrayContaining([expect.objectContaining({ path: ["provider"] })]));
  });
});
