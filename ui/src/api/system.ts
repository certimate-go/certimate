import { ClientResponseError } from "pocketbase";

import { type SystemEnvironment } from "@/domain/system";
import { getPocketBase } from "@/repository/_pocketbase";

export const getEnvironment = async () => {
  const pb = getPocketBase();

  const resp = await pb.send<BaseResponse<SystemEnvironment>>("/api/system/environment", {
    method: "GET",
  });

  if (resp.code != 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }

  return resp;
};

