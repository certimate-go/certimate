import { get as httpGet, post as httpPost } from "./_api";

export const getStatus = () => {
  type RespData = {
    installed: boolean;
  };

  return httpGet<RespData>({
    url: `/api/installer/status`,
  });
};

export const initSuperuser = (params: { token: string; email: string; password: string }) => {
  return httpPost({
    url: `/api/installer/superuser`,
    body: {
      ...params,
    },
  });
};
