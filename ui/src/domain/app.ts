export const APP_REPO_URL = "https://github.com/certimate-go/certimate";

export const APP_DOWNLOAD_URL = APP_REPO_URL + "/releases";

export const APP_DOCUMENT_URL = "https://docs.certimate.me";

// fallback policy: .env > git tag > "v0.0.0-dev"
export const APP_VERSION = import.meta.env.VITE_APP_VERSION || __APP_VERSION__ || "v0.0.0-dev";
