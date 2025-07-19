import { useRequest } from "ahooks";

import { APP_VERSION } from "@/domain/app";

export type UseVersionCheckerReturns = {
  hasNewVersion: boolean;
  checkNewVersion: () => void;
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

/**
 * 获取版本检查器。
 * @returns {UseVersionCheckerReturns}
 */
const useVersionChecker = () => {
  const { data, refresh } = useRequest(
    async () => {
      const releases = await fetch("https://api.github.com/repos/certimate-go/certimate/releases")
        .then((res) => res.json())
        .then((res) => Array.from(res));

      const cIdx = releases.findIndex((e: any) => e.name === APP_VERSION);
      if (cIdx === 0) {
        return false;
      }

      const nIdx = releases.findIndex((e: any) => compareVersions(e.name, APP_VERSION) !== -1);
      if (cIdx !== -1 && cIdx <= nIdx) {
        return false;
      }

      return !!releases[nIdx];
    },
    {
      pollingInterval: 6 * 60 * 60 * 1000,
      refreshOnWindowFocus: true,
      focusTimespan: 15 * 60 * 1000,
    }
  );

  return {
    hasNewVersion: !!data,
    checkNewVersion: refresh,
  };
};

export default useVersionChecker;
