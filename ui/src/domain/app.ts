import i18next from "i18next";

// fallback policy: .env > git tag > "v0.0.0-dev"
export const APP_VERSION: string = "v" + (__APP_VERSION__ || "0.0.0-dev").replace(/^v/, "");

export const APP_REPO_URL = "https://github.com/certimate-go/certimate";
export const APP_DOWNLOAD_URL = APP_REPO_URL + "/releases";
export let APP_DOCUMENT_URL = "https://docs.certimate.me";

i18next.on("languageChanged", (language) => {
  if (language.startsWith("zh")) {
    APP_DOCUMENT_URL = "https://docs.certimate.me";
  } else {
    APP_DOCUMENT_URL = "https://docs.certimate.me/en/";
  }
});
