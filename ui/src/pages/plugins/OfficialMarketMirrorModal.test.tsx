import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterAll, afterEach, beforeAll, expect, it, vi } from "vitest";

import { type MarketMirrorSettings, fetchOfficialMirror, saveOfficialMirror } from "@/api/pluginmarket";
import OfficialMarketMirrorModal from "./OfficialMarketMirrorModal";

vi.mock("@/api/pluginmarket", () => ({ fetchOfficialMirror: vi.fn(), saveOfficialMirror: vi.fn() }));
vi.mock("react-i18next", () => ({ useTranslation: () => locale }));

const locale = { t: (key: string) => key };
const defaultSettings = { mirror: { mode: "default" }, defaultBase: "https://cdn.jsdelivr.net", presets: [] } satisfies MarketMirrorSettings;

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

it("edits an existing custom mirror and saves its trimmed address", async () => {
  vi.mocked(fetchOfficialMirror).mockResolvedValue({ ...defaultSettings, mirror: { mode: "custom", baseUrl: "https://first.example" } });
  vi.mocked(saveOfficialMirror).mockResolvedValue({ ...defaultSettings, mirror: { mode: "custom", baseUrl: "https://second.example/path" } });
  const onSaved = vi.fn();
  render(<OfficialMarketMirrorModal onClose={vi.fn()} onSaved={onSaved} />);
  const field = await screen.findByLabelText("plugin.market.mirror.base");
  expect(field).toHaveProperty("value", "https://first.example");
  fireEvent.change(field, { target: { value: " https://second.example/path " } });
  fireEvent.click(screen.getByRole("button", { name: "plugin.market.mirror.save" }));
  await waitFor(() => expect(onSaved).toHaveBeenCalledTimes(1));
  expect(saveOfficialMirror).toHaveBeenCalledWith({ mode: "custom", baseUrl: "https://second.example/path" });
});

it("loads a previously saved preset as a custom HTTPS mirror", async () => {
  const baseUrl = "https://mirror.example";
  vi.mocked(fetchOfficialMirror).mockResolvedValue({
    ...defaultSettings,
    mirror: { mode: "preset", preset: "example" },
    presets: [{ id: "example", name: "Example", baseUrl }],
  });
  vi.mocked(saveOfficialMirror).mockResolvedValue({ ...defaultSettings, mirror: { mode: "custom", baseUrl } });
  const onSaved = vi.fn();
  render(<OfficialMarketMirrorModal onClose={vi.fn()} onSaved={onSaved} />);
  expect(await screen.findByLabelText("plugin.market.mirror.base")).toHaveProperty("value", baseUrl);
  expect(screen.queryByRole("radio", { name: "plugin.market.mirror.preset" })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "plugin.market.mirror.save" }));
  await waitFor(() => expect(onSaved).toHaveBeenCalledTimes(1));
  expect(saveOfficialMirror).toHaveBeenCalledWith({ mode: "custom", baseUrl });
});

it("can restore the default when the saved mirror reports an error", async () => {
  vi.mocked(fetchOfficialMirror).mockResolvedValue({
    ...defaultSettings,
    mirror: { mode: "custom", baseUrl: "https://offline.example" },
    error: "Mirror unavailable",
  });
  vi.mocked(saveOfficialMirror).mockResolvedValue(defaultSettings);
  const onSaved = vi.fn();
  render(<OfficialMarketMirrorModal onClose={vi.fn()} onSaved={onSaved} />);
  expect(await screen.findByText("Mirror unavailable")).toBeTruthy();
  fireEvent.click(screen.getByRole("radio", { name: "plugin.market.mirror.default" }));
  fireEvent.click(screen.getByRole("button", { name: "plugin.market.mirror.save" }));
  await waitFor(() => expect(onSaved).toHaveBeenCalledTimes(1));
  expect(saveOfficialMirror).toHaveBeenCalledWith({ mode: "default" });
});
