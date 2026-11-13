import { I18nextProvider } from "react-i18next";
import { act, cleanup, render, screen } from "@testing-library/react";
import { App } from "antd";
import { createInstance } from "i18next";
import { afterEach, expect, it, vi } from "vitest";

import * as MarketAPI from "@/api/pluginmarket";
import PluginList from "./PluginList";

vi.mock("@/api/pluginmarket", async (importOriginal) => ({
  ...(await importOriginal<typeof MarketAPI>()),
  fetchMarketListing: vi.fn(),
}));
vi.mock("@/stores/pluginCatalog", () => ({ usePluginCatalogStore: () => vi.fn() }));

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

const renderMarket = async (resources = {}) => {
  vi.stubGlobal("matchMedia", (media: string) => ({
    matches: false,
    media,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }));
  const i18n = createInstance();
  const logs = vi.spyOn(console, "log").mockImplementation(() => {});
  await i18n.init({ lng: "zh", fallbackLng: "en", debug: true, resources, react: { bindI18nStore: "added" } });
  vi.mocked(MarketAPI.fetchMarketListing).mockResolvedValue({
    items: [
      {
        provider_type: "sample",
        display_name_key: "plugin.sample.name",
        deploy_category: "",
        version: "1.0.0",
        description: "",
        status: "not_installed",
        source: { id: "custom", repository: "fixture/plugins" },
        official: false,
        sourceAvailable: true,
        listed: true,
        canInstall: true,
        canUpdate: false,
        canDelete: false,
      },
    ],
    page: 1,
    perPage: 30,
    totalItems: 1,
    totalPages: 1,
    refreshFailed: false,
    refreshing: false,
    activeJobs: [],
    sources: [],
  });
  render(
    <I18nextProvider i18n={i18n}>
      <App>
        <PluginList />
      </App>
    </I18nextProvider>
  );
  await screen.findByText("fixture/plugins");
  return { i18n, logs };
};

it("displays the fallback without missing-key logs and updates when plugin translations arrive", async () => {
  const { i18n, logs } = await renderMarket();
  expect(screen.getByRole("heading", { name: "sample" })).toBeInTheDocument();
  expect(logs.mock.calls.filter((args) => args.includes("plugin.sample.name"))).toEqual([]);
  await act(async () => {
    i18n.addResourceBundle("zh", "translation", { "plugin.sample.name": "示例插件" }, true, true);
  });
  expect(await screen.findByRole("heading", { name: "示例插件" })).toBeInTheDocument();
  expect(logs.mock.calls.filter((args) => args.includes("plugin.sample.name"))).toEqual([]);
});

it("preserves translated names from the fallback language", async () => {
  const { logs } = await renderMarket({ en: { translation: { "plugin.sample.name": "Sample plugin" } } });
  expect(screen.getByRole("heading", { name: "Sample plugin" })).toBeInTheDocument();
  expect(logs.mock.calls.filter((args) => args.includes("plugin.sample.name"))).toEqual([]);
});
