import { ClientResponseError } from "pocketbase";

import { type CertificateFormatType } from "@/domain/certificate";
import { getPocketBase } from "@/repository/_pocketbase";

export const archive = async (certificateId: string, format?: CertificateFormatType) => {
  const pb = getPocketBase();

  type RespData = {
    fileBytes: string;
  };
  const resp = await pb.send<BaseResponse<RespData>>(`/api/certificates/${encodeURIComponent(certificateId)}/archive`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: {
      format: format,
    },
  });

  if (resp.code != 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }

  return resp;
};

export const revoke = async (certificateId: string) => {
  const pb = getPocketBase();

  const resp = await pb.send<BaseResponse>(`/api/certificates/${encodeURIComponent(certificateId)}/revoke`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (resp.code != 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }

  return resp;
};
