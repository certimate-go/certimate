import { useEffect, useState } from "react";
import { useRequest } from "ahooks";

import { APP_VERSION } from "@/domain/app";

export type UseVersionCheckerReturns = {
  hasUpdate: boolean;
  checkUpdate: () => Promise<boolean>;
};

const extractSemver = (vers: string) => {
  let semver = String(vers ?? "");
  semver = semver.replace(/^v/i, "");
  semver = semver.split("-")[0];
  return semver;
};

const compareVersions = (a: string, b: string) => {
  const aSemver = extractSemver(a);
  const bSemver = extractSemver(b);
  const aSemverParts = aSemver.split(".");
  const bSemverParts = bSemver.split(".");

  const len = Math.max(aSemverParts.length, bSemverParts.length);
  for (let i = 0; i < len; i++) {
    const aPart = parseInt(aSemverParts[i] ?? "0");
    const bPart = parseInt(bSemverParts[i] ?? "0");
    if (aPart > bPart) return 1;
    if (bPart > aPart) return -1;
  }

  return 0;
};

const LOCAL_STORAGE_KEY = "certimate-ui-newver";

/**
 * 获取版本检查器。
 * @returns {UseVersionCheckerReturns}
 */
const useVersionChecker = () => {
  const [hasUpdate, setHasUpdate] = useState(() => {
    const newver = localStorage.getItem(LOCAL_STORAGE_KEY)!;
    if (newver) {
      return compareVersions(newver, APP_VERSION) === 1;
    }

    return false;
  });

  const { refresh, cancel } = useRequest(
    async () => {
      type ReleaseInfo = {
        id: number;
        name: string;
        body: string;
        prerelease: boolean;
      };

      let releases: ReleaseInfo[] = [];
      try {
        // try to fetch from GitHub
        releases = await fetch("https://api.github.com/repos/certimate-go/certimate/releases").then((res) => {
          if (res.ok) {
            return res.json().then((res) => Array.from(res) as ReleaseInfo[]);
          } else {
            throw new Error("Failed to check update from GitHub");
          }
        });
      } catch {
        // fallback to fetch from Gitee
        releases = await fetch("https://gitee.com/api/v5/repos/certimate-go/certimate/releases").then((res) => {
          if (res.ok) {
            return res.json().then((res) => Array.from(res) as ReleaseInfo[]);
          } else {
            throw new Error("Failed to check update from GitHub");
          }
        });
      }

      const cIdx = releases.findIndex((e) => e.name === APP_VERSION);
      if (cIdx === 0) {
        return false;
      }

      const nIdx = releases.findIndex((e) => compareVersions(e.name, APP_VERSION) !== -1);
      if (cIdx !== -1 && cIdx <= nIdx) {
        return false;
      }

      if (releases[nIdx]) {
        localStorage.setItem(LOCAL_STORAGE_KEY, releases[nIdx].name);
      } else {
        localStorage.removeItem(LOCAL_STORAGE_KEY);
      }

      return !!releases[nIdx];
    },
    {
      manual: true,
      focusTimespan: 15 * 60 * 1000,
      pollingInterval: 6 * 60 * 60 * 1000,
      throttleWait: 60 * 1000,
      onSuccess: (res) => {
        setHasUpdate(res);
      },
    }
  );

  useEffect(() => {
    refresh();

    return () => cancel();
  }, []);

  return {
    hasUpdate: hasUpdate,
    checkUpdate: refresh,
  };
};

export default useVersionChecker;
