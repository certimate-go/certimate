import { type SpawnSyncReturns, execFileSync } from "node:child_process";
import path from "node:path";

import tailwindcssPlugin from "@tailwindcss/vite";
import legacyPlugin from "@vitejs/plugin-legacy";
import reactPlugin from "@vitejs/plugin-react";
import fs from "fs-extra";
import { type Plugin, defineConfig, loadEnv } from "vite";

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

export default defineConfig(({ mode }) => {
  const envs = loadEnv(mode, process.cwd());
  let appVersion = undefined;
  if (!envs?.VITE_APP_VERSION) {
    try {
      appVersion = execFileSync("git", ["describe", "--match", "v[0-9]*", "--tags", "--abbrev=8"], {
        stdio: [],
      })?.toString();
    } catch (error) {
      const err = error as SpawnSyncReturns<Buffer>;
      console.warn("[Warn] failed to get version number through git", err?.stderr?.toString());
    }
  }

  return {
    define: {
      __APP_VERSION__: JSON.stringify(appVersion),
    },
    plugins: [
      reactPlugin({}),
      legacyPlugin({
        targets: ["defaults", "not IE 11"],
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
