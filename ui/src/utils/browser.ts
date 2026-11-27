export const isBrowserHappy = () => {
  try {
    if (typeof Promise.withResolvers !== "function") return false;
    if (typeof Promise.try !== "function") return false;
    if (typeof CSS.supports !== "function") return false;
    if (!CSS.supports("color", "oklch(0 0 0)")) return false;
  } catch (_) {
    return false;
  }

  return true;
};

export const isTouchSupported = () => {
  if (!("ontouchstart" in window)) return false;
  if (!("maxTouchPoints" in window.navigator)) return false;
  if (window.navigator.maxTouchPoints === 0) return false;

  return true;
};
