import { withBasePath } from "@/utils/url";

export class SetupError extends Error {
  code: number;

  constructor(code: number, message: string) {
    super(message);
    this.code = code;
  }
}

export type SetupStatus = {
  initialized: boolean;
};

export type SubmitSetupRequest = {
  token: string;
  email: string;
  password: string;
  passwordConfirm: string;
};

const requestSetupApi = async <T>(path: string, init?: RequestInit): Promise<T> => {
  let res: Response;
  try {
    res = await fetch(withBasePath(path), {
      ...init,
      cache: "no-store",
      headers: {
        "Content-Type": "application/json",
        ...init?.headers,
      },
    });
  } catch (err) {
    throw new SetupError(-1, `Network error: ${(err as Error).message}`);
  }

  let result: { code?: number; msg?: string; data?: T };
  try {
    result = await res.json();
  } catch {
    throw new SetupError(res.status, `Unexpected response (HTTP ${res.status})`);
  }

  if (!res.ok || result.code !== 0) {
    throw new SetupError(result.code ?? res.status, result.msg || `Request failed (HTTP ${res.status})`);
  }

  return result.data as T;
};

export const getSetupStatus = () => {
  return requestSetupApi<SetupStatus>("/api/setup/status");
};

export const submitSetup = (request: SubmitSetupRequest) => {
  return requestSetupApi<null>("/api/setup", {
    method: "POST",
    body: JSON.stringify(request),
  });
};
