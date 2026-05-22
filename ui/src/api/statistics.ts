import { type Statistics } from "@/domain/statistics";

import { get as httpGet } from "./_api";

export const get = async () => {
  type RespData = Statistics;

  return httpGet<RespData>({
    url: `/api/statistics`,
  });
};
