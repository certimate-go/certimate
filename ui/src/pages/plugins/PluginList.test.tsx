import { useState } from "react";
import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { App } from "antd";
import { ClientResponseError } from "pocketbase";
import { afterAll, afterEach, beforeAll, expect, it, vi } from "vitest";

import * as MarketAPI from "@/api/pluginmarket";
import PluginList from "./PluginList";
import PluginMarketSearch from "./PluginMarketSearch";

vi.mock("@/api/pluginmarket", async (importOriginal) => ({
  ...(await importOriginal<typeof MarketAPI>()),
  fetchMarketListing: vi.fn(),
  installPlugin: vi.fn(),
  getInstallStatus: vi.fn(),
}));
vi.mock("@/stores/pluginCatalog", () => ({ usePluginCatalogStore: () => vi.fn() }));
vi.mock("react-i18next", () => ({
  useTranslation: () => translations,
}));

const translations = { t: (key: string, fallback?: string) => (typeof fallback === "string" ? fallback : key), i18n: { language: "en" } };
const official = { id: "official", repository: "certimate-go/plugins" };
const custom = { id: "custom", repository: "fixture/plugins" };
const base: MarketAPI.MarketEntry = {
  provider_type: "sample",
  display_name_key: "",
  deploy_category: "",
  version: "1.0.0",
  description: "",
  status: "not_installed",
  source: official,
  official: true,
  sourceAvailable: true,
  listed: true,
  canInstall: true,
  canUpdate: false,
  canDelete: false,
};
const listing: MarketAPI.MarketPage = {
  items: [base, { ...base, source: custom, official: false }],
  page: 1,
  perPage: 30,
  totalItems: 2,
  totalPages: 1,
  refreshFailed: false,
  refreshing: false,
  activeJobs: [],
  sources: [
    { ...official, official: true, stale: false, refreshing: false },
    { ...custom, official: false, stale: false, refreshing: false },
  ],
};

beforeAll(() => {
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe = vi.fn();
      unobserve = vi.fn();
      disconnect = vi.fn();
    }
  );
  vi.stubGlobal("matchMedia", (query: string) => ({
    matches: false,
    media: query,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }));
});
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});
afterAll(() => vi.unstubAllGlobals());

it("badges the official listing and sends the selected custom source without polling on an ownership conflict", async () => {
  vi.mocked(MarketAPI.fetchMarketListing).mockResolvedValue(listing);
  vi.mocked(MarketAPI.installPlugin).mockRejectedValue(
    new ClientResponseError({ status: 409, response: { message: "Owned by another source", data: { reason: "ownership_conflict" } } })
  );
  const { container } = render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  const cards = Array.from(container.querySelectorAll<HTMLElement>(".ant-card"));
  expect(cards).toHaveLength(2);
  const customCard = cards.find((card) => card.textContent?.includes(custom.repository));
  expect(customCard).toBeDefined();
  expect(screen.getAllByRole("img", { name: "plugin.market.sources.official" })).toHaveLength(1);
  if (!customCard) throw new Error("missing custom card");
  expect(within(customCard).queryByRole("img", { name: "plugin.market.sources.official" })).toBeNull();
  fireEvent.click(within(customCard).getByRole("button", { name: "plugin.market.action.install" }));
  await waitFor(() => expect(MarketAPI.installPlugin).toHaveBeenCalledWith({ providerType: "sample", source: custom.id }));
  await screen.findByText("Owned by another source");
  expect(MarketAPI.getInstallStatus).not.toHaveBeenCalled();
});

it("renders local removal only on the owning listing", async () => {
  vi.mocked(MarketAPI.fetchMarketListing).mockResolvedValue({
    ...listing,
    items: [
      { ...base, status: "ownership_conflict", canInstall: false, installedOrigin: custom },
      { ...base, source: custom, official: false, status: "installed", canInstall: false, canDelete: true },
    ],
  });
  render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  expect(screen.getAllByRole("button", { name: "plugin.market.action.more" })).toHaveLength(1);
  expect(screen.queryByRole("button", { name: "plugin.market.action.install" })).toBeNull();
});

const mockSearchListing = () => {
  vi.mocked(MarketAPI.fetchMarketListing).mockImplementation(async (query) => {
    const items = query.providerType
      ? listing.items.filter((entry) => entry.provider_type === query.providerType && entry.source.id === query.source)
      : listing.items;
    return { ...listing, items, totalItems: items.length };
  });
};

const searchInput = () => screen.getByRole("combobox", { name: "plugin.market.action.search_placeholder" });
const pressKey = (key: string, keyCode: number) => {
  fireEvent.keyDown(searchInput(), { key, code: key, keyCode, which: keyCode });
  fireEvent.keyUp(searchInput(), { key, code: key, keyCode, which: keyCode });
};

it("suggests matches without changing the grid and selects a single source, then allows a broad search", async () => {
  mockSearchListing();
  const { container } = render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  fireEvent.focus(searchInput());
  fireEvent.change(searchInput(), { target: { value: "sam" } });
  await screen.findByText("sample · fixture/plugins");
  expect(container.querySelectorAll(".ant-card")).toHaveLength(2);
  fireEvent.click(screen.getByText("sample · fixture/plugins"));
  await waitFor(() =>
    expect(MarketAPI.fetchMarketListing).toHaveBeenLastCalledWith(
      expect.objectContaining({
        keyword: "",
        providerType: "sample",
        source: custom.id,
        page: 1,
      })
    )
  );
  await waitFor(() => expect(container.querySelectorAll(".ant-card")).toHaveLength(1));
  expect(searchInput()).toHaveValue("sample");
  fireEvent.click(screen.getByRole("button", { name: "search" }));
  await waitFor(() =>
    expect(MarketAPI.fetchMarketListing).toHaveBeenLastCalledWith(
      expect.objectContaining({
        keyword: "sample",
        providerType: undefined,
        source: "",
      })
    )
  );
  await waitFor(() => expect(container.querySelectorAll(".ant-card")).toHaveLength(2));
});

it("submits a keyword on bare Enter even while suggestions are open", async () => {
  mockSearchListing();
  render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  fireEvent.focus(searchInput());
  fireEvent.change(searchInput(), { target: { value: "sam" } });
  await screen.findByText("sample · fixture/plugins");
  pressKey("Enter", 13);
  await waitFor(() =>
    expect(MarketAPI.fetchMarketListing).toHaveBeenLastCalledWith(
      expect.objectContaining({
        keyword: "sam",
        providerType: undefined,
        perPage: 30,
      })
    )
  );
});

it("selects the highlighted suggestion on Enter without overwriting it with keyword search", async () => {
  mockSearchListing();
  const { container } = render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  fireEvent.focus(searchInput());
  fireEvent.change(searchInput(), { target: { value: "sam" } });
  await screen.findByText("sample · fixture/plugins");
  pressKey("ArrowDown", 40);
  pressKey("Enter", 13);
  await waitFor(() =>
    expect(MarketAPI.fetchMarketListing).toHaveBeenLastCalledWith(
      expect.objectContaining({
        keyword: "",
        providerType: "sample",
        source: official.id,
        perPage: 30,
      })
    )
  );
  await waitFor(() => expect(container.querySelectorAll(".ant-card")).toHaveLength(1));
  expect(searchInput()).toHaveValue("sample");
});

it("ignores delayed suggestions after a keyword submit and keeps search usable after suggestion failure", async () => {
  let resolveSuggestion: (value: MarketAPI.MarketPage) => void = () => {};
  vi.mocked(MarketAPI.fetchMarketListing).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        resolveSuggestion = resolve;
      })
  );
  const onSearch = vi.fn();
  const SearchFixture = () => {
    const [draft, setDraft] = useState("");
    return (
      <PluginMarketSearch
        draft={draft}
        filters={{ status: "all", capability: "all", source: "", locale: "en" }}
        onDraftChange={setDraft}
        onSearch={onSearch}
        onSelect={vi.fn()}
      />
    );
  };
  render(<SearchFixture />);
  fireEvent.focus(searchInput());
  fireEvent.change(searchInput(), { target: { value: "sam" } });
  await waitFor(() => expect(MarketAPI.fetchMarketListing).toHaveBeenCalledTimes(1));
  pressKey("Enter", 13);
  expect(onSearch).toHaveBeenLastCalledWith("sam");
  await act(async () => resolveSuggestion(listing));
  expect(screen.queryByText("sample · fixture/plugins")).toBeNull();

  vi.mocked(MarketAPI.fetchMarketListing).mockRejectedValueOnce(new Error("unavailable"));
  fireEvent.change(searchInput(), { target: { value: "other" } });
  await waitFor(() => expect(MarketAPI.fetchMarketListing).toHaveBeenCalledTimes(2));
  pressKey("Enter", 13);
  expect(onSearch).toHaveBeenLastCalledWith("other");
});

it("keeps showing suggestions when whitespace around the keyword changes", async () => {
  mockSearchListing();
  render(
    <App>
      <PluginList />
    </App>
  );
  await screen.findByText(custom.repository);
  fireEvent.focus(searchInput());
  for (const value of ["sam", "sam ", " sam ", "sam"]) {
    fireEvent.change(searchInput(), { target: { value } });
    await screen.findByText("sample · fixture/plugins");
  }
});
