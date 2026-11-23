// import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import dayjs from "dayjs";
import dayjsUtc from "dayjs/plugin/utc";

import App from "./App";
import "./i18n";
import "./index.css";
import "./global.css";
import { usePluginCatalogStore } from "./stores/pluginCatalog";

dayjs.extend(dayjsUtc);

const renderApp = () => {
  createRoot(document.getElementById("root")!).render(
    // TODO: React v19 严格模式与 flowgram.ai 有兼容问题，暂时注释掉
    // <StrictMode>
    <App />
    // </StrictMode>
  );
};

const pluginCatalogReady = usePluginCatalogStore
  .getState()
  .init()
  .catch(() => undefined);
const renderDeadline = new Promise<void>((resolve) => setTimeout(resolve, 2000));
void Promise.race([pluginCatalogReady, renderDeadline]).finally(renderApp);
