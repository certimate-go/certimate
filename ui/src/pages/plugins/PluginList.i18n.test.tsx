import { I18nextProvider } from "react-i18next";
import { act, cleanup, fireEvent, render, screen, within } from "@testing-library/react";
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

const renderMarket = async (resources = {}, entries?: MarketAPI.MarketEntry[]) => {
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe = vi.fn();
      unobserve = vi.fn();
      disconnect = vi.fn();
    }
  );
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
    items: entries ?? [
      {
        provider_type: "sample",
        display_name_key: "plugin.sample.name",
        category: "",
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
    totalItems: entries?.length ?? 1,
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

it("keeps source-specific names in cards, suggestions, and removal confirmation", async () => {
  const localized = {
    provider_type: "sample",
    display_name_key: "plugin.sample.name",
    category: "website",
    capability: "deploy",
    version: "1.0.0",
    description: "部署网站证书。",
    status: "installed",
    source: { id: "custom", repository: "fixture/plugins" },
    official: false,
    sourceAvailable: true,
    listed: true,
    canInstall: false,
    canUpdate: false,
    canDelete: true,
    display_name: "甲市场插件",
  } satisfies MarketAPI.MarketEntry;
  await renderMarket(
    {
      zh: {
        translation: {
          "plugin.sample.name": "已安装插件的全局名称",
          "provider.category.website": "网站托管",
          "plugin.market.capability.deploy": "部署",
          "plugin.market.action.more": "{{name}} 的操作",
          "plugin.market.action.delete": "删除",
          "plugin.market.action.delete_modal_content": "将移除 {{name}}。",
        },
      },
    },
    [localized, { ...localized, display_name: "乙市场插件", source: { id: "second", repository: "second/plugins" }, category: "custom-category" }]
  );
  expect(screen.getByRole("heading", { name: "甲市场插件" })).toBeInTheDocument();
  expect(screen.getByRole("heading", { name: "乙市场插件" })).toBeInTheDocument();
  expect(screen.getByText("部署 · 网站托管")).toBeInTheDocument();
  expect(screen.getByText("部署 · custom-category")).toBeInTheDocument();
  expect(screen.getAllByText("部署网站证书。")).toHaveLength(2);
  const search = screen.getByRole("combobox", { name: "plugin.market.action.search_placeholder" });
  fireEvent.focus(search);
  fireEvent.change(search, { target: { value: "插件" } });
  const suggestion = await screen.findByText("sample · second/plugins");
  expect(suggestion.parentElement).toHaveTextContent("乙市场插件");
  fireEvent.blur(search);
  fireEvent.click(screen.getByRole("button", { name: "甲市场插件 的操作" }));
  fireEvent.click(await screen.findByRole("menuitem", { name: "删除" }));
  expect(await within(screen.getByRole("dialog")).findByText("将移除 甲市场插件。")).toBeInTheDocument();
});
