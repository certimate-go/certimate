import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import dayjs from "dayjs";
import dayjsUtc from "dayjs/plugin/utc";

import App from "./App";
import "./i18n";
import "./index.css";
import "./global.css";

dayjs.extend(dayjsUtc);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
