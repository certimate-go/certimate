import { post as httpPost } from "./_api";

export const testPushNotification = async ({ provider, accessId }: { provider: string; accessId: string }) => {
  return httpPost({
    url: `/api/notifications/test`,
    body: {
      provider,
      accessId,
    },
  });
};
