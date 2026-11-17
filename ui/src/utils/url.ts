const baseUrl = new URL(import.meta.env.BASE_URL, document.baseURI);

export const getBasePath = () => {
  return baseUrl.pathname;
};

export const withBasePath = (path: string) => {
  // 完整 URL 和协议相对 URL 不属于应用内资源，保持调用方传入的地址不变。
  if (/^[a-z][a-z\d+.-]*:/i.test(path) || path.startsWith("//")) {
    return path;
  }

  const url = new URL(path.replace(/^\/+/, ""), baseUrl);
  return `${url.pathname}${url.search}${url.hash}`;
};
