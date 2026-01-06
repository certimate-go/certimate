import path from "node:path";

import tailwindcssPlugin from "@tailwindcss/vite";
import legacyPlugin from "@vitejs/plugin-legacy";
import reactPlugin from "@vitejs/plugin-react";
import fs from "fs-extra";
import { type Plugin, defineConfig } from "vite";

const preserveFilesPlugin = (filesToPreserve: string[]): Plugin => {
  return {
    name: "preserve-files",
    apply: "build",
    buildStart() {
      // 在构建开始时将要保留的文件或目录移动到临时位置
      filesToPreserve.forEach((file) => {
        const srcPath = path.resolve(__dirname, file);
        const tempPath = path.resolve(__dirname, `node_modules/.tmp/build/${file}`);
        if (fs.existsSync(srcPath)) {
          fs.moveSync(srcPath, tempPath, { overwrite: true });
        }
      });
    },
    closeBundle() {
      // 在构建完成后将临时位置的文件或目录移回原来的位置
      filesToPreserve.forEach((file) => {
        const srcPath = path.resolve(__dirname, file);
        const tempPath = path.resolve(__dirname, `node_modules/.tmp/build/${file}`);
        if (fs.existsSync(tempPath)) {
          fs.moveSync(tempPath, srcPath, { overwrite: true });
        }
      });
    },
  };
};

export default defineConfig(() => {
  let appVersion = undefined;
  try {
    const content = fs.readFileSync(path.resolve(__dirname, "../internal/app/app.go"), "utf-8");
    const matches = content.match(/AppVersion\s+=\s+"(.+?)"/);
    if (matches) {
      appVersion = matches[1];
      console.info("[certimate] AppVersion is " + appVersion);
    } else {
      throw new Error("AppVersion not found in '/internal/app/app.go'");
    }
  } catch (err) {
    throw new Error("Could not read app version: " + (err as Error).message);
  }

  return {
    define: {
      __APP_VERSION__: JSON.stringify(appVersion),
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes("/src/i18n/")) {
              return "locales";
            }
          },
        },
      },
    },
    plugins: [
      reactPlugin({}),
      legacyPlugin({
        targets: ["defaults", "not IE 11"],
        modernTargets: "chrome>=111, firefox>=113, safari>=15.4",
        polyfills: true,
        modernPolyfills: true,
        renderLegacyChunks: false,
        renderModernChunks: true,
      }),
      tailwindcssPlugin(),
      preserveFilesPlugin(["dist/.gitkeep"]),
    ],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      proxy: {
        "/api": "http://127.0.0.1:8090",
      },
    },
  };
});
