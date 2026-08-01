import path from "node:path";

import fs from "fs-extra";
import { type Plugin } from "vite";

interface Options {
  files: string[];
}

export default function (options: Options): Plugin {
  const tmpdir = path.resolve("./", "node_modules/.vite/certimate-temp");

  return {
    name: "preserve-files",
    apply: "build",
    buildStart() {
      // 在构建开始时将要保留的文件或目录移动到临时位置
      options.files.forEach((file) => {
        const src = path.resolve("./", file);
        const dist = path.resolve(tmpdir, file);
        if (fs.existsSync(src)) {
          fs.moveSync(src, dist, { overwrite: true });
        }
      });
    },
    closeBundle() {
      // 在构建完成后将临时位置的文件或目录移回原来的位置
      options.files.forEach((file) => {
        const src = path.resolve("./", file);
        const dist = path.resolve(tmpdir, file);
        if (fs.existsSync(dist)) {
          fs.moveSync(dist, src, { overwrite: true });
        }
      });

      // 清理临时目录
      fs.emptyDirSync(tmpdir);
    },
  };
}
