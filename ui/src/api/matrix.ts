import { ClientResponseError } from "pocketbase";

import { getPocketBase } from "@/repository/_pocketbase";

export type MatrixVerifyStep = {
  name: string;
  ok: boolean;
  message: string;
  detail?: string;
  code?: string;
  retryAfterSec?: number;
};

export type MatrixVerifyConnectionResult = {
  ok: boolean;
  userId?: string;
  sessionAccessToken?: string;
  sessionDeviceId?: string;
  sessionSaved?: boolean;
  steps: MatrixVerifyStep[];
};

export type MatrixTestSendResult = {
  ok: boolean;
  userId?: string;
  sessionAccessToken?: string;
  sessionDeviceId?: string;
  sessionSaved?: boolean;
};

export const verifyMatrixConnection = async ({ config, accessId }: { config: Record<string, unknown>; accessId?: string }) => {
  const pb = getPocketBase();

  const resp = await pb.send<{ code: number; msg: string; data: MatrixVerifyConnectionResult }>("/api/notifications/matrix/verify", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: {
      config,
      accessId,
    },
  });

  if (resp.code != 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }

  return resp.data;
};

export const sendMatrixTestMessage = async ({
  config,
  accessId,
  subject,
  message,
}: {
  config: Record<string, unknown>;
  accessId?: string;
  subject?: string;
  message?: string;
}) => {
  const pb = getPocketBase();

  const resp = await pb.send<{ code: number; msg: string; data: MatrixTestSendResult }>("/api/notifications/matrix/test-send", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: {
      config,
      accessId,
      subject,
      message,
    },
  });

  if (resp.code != 0) {
    throw new ClientResponseError({ status: resp.code, response: resp, data: {} });
  }

  return resp.data;
};
