import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { App } from "antd";
import { ClientResponseError } from "pocketbase";
import { afterAll, afterEach, beforeAll, expect, it, vi } from "vitest";

import * as MarketAPI from "@/api/pluginmarket";
import PluginList from "./PluginList";

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
