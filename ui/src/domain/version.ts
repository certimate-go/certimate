// .env > git tag > v0.0.0-beta
// https://vite.dev/guide/env-and-mode.html#env-variables
export const version = import.meta.env.VITE_APP_VERSION || __APP_VERSION__ || "v0.0.0-dev";
