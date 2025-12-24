import { ClientResponseError } from "pocketbase";

export const unwrapErrMsg = (error: unknown): string => {
  if (error instanceof ClientResponseError) {
    return Object.keys(error.response ?? {}).length ? unwrapErrMsg(error.response) : error.message;
  } else if (error instanceof Error) {
    return error.message;
  } else if (typeof error === "object" && error != null) {
    if ("message" in error) {
      return unwrapErrMsg(error.message);
    } else if ("msg" in error) {
      return unwrapErrMsg(error.msg);
    }
  } else if (typeof error === "string") {
    return error || "Unknown error";
  }

  return "Unknown error";
};
