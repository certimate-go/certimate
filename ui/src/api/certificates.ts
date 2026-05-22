import { type CertificateFormatType } from "@/domain/certificate";

import { post as httpPost } from "./_api";

export const download = (certificateId: string, format?: CertificateFormatType) => {
  type RespData = {
    zipBytes: string;
  };

  return httpPost<RespData>({
    url: `/api/certificates/${encodeURIComponent(certificateId)}/download`,
    body: {
      fileFormat: format,
    },
  });
};

export const revoke = (certificateId: string) => {
  return httpPost({
    url: `/api/certificates/${encodeURIComponent(certificateId)}/revoke`,
  });
};
