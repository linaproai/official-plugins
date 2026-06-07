import type { APIRequestContext } from "@host-tests/support/api/job";

import {
  createAdminApiContext,
  createApiContext,
  createUser,
  deleteUser,
  expectSuccess,
} from "@host-tests/support/api/job";

const recipientPassword = "test123456";

export type MessageRecipient = {
  api: APIRequestContext;
  cleanup: () => Promise<void>;
  password: string;
  username: string;
};

function normalizeLabel(label: string) {
  return label
    .replaceAll(/[^a-z0-9]+/gi, "_")
    .replace(/^_+|_+$/g, "")
    .toLowerCase();
}

export async function createMessageRecipient(
  label: string,
): Promise<MessageRecipient> {
  const adminApi = await createAdminApiContext();
  const username = `e2e_notice_${normalizeLabel(label)}_${Date.now()}_${Math.random()
    .toString(36)
    .slice(2, 8)}`;
  let userID = 0;
  let userApi: APIRequestContext | undefined;

  try {
    userID = (
      await createUser(adminApi, {
        username,
        password: recipientPassword,
        nickname: "E2E Notice Recipient",
      })
    ).id;
    userApi = await createApiContext(username, recipientPassword);
    await clearMessages(userApi);
  } catch (error) {
    await userApi?.dispose().catch(() => {});
    if (userID > 0) {
      await deleteUser(adminApi, userID).catch(() => {});
    }
    await adminApi.dispose();
    throw error;
  }

  let cleaned = false;

  return {
    api: userApi,
    password: recipientPassword,
    username,
    cleanup: async () => {
      if (cleaned) {
        return;
      }
      cleaned = true;
      await clearMessages(userApi).catch(() => {});
      await userApi.post("auth/logout").catch(() => {});
      await userApi.dispose().catch(() => {});
      if (userID > 0) {
        await deleteUser(adminApi, userID).catch(() => {});
      }
      await adminApi.dispose();
    },
  };
}

export async function clearMessages(api: APIRequestContext) {
  await expectSuccess(await api.delete("user/message/clear"));
}

export async function unreadCount(api: APIRequestContext) {
  const result = await expectSuccess<{ count: number }>(
    await api.get("user/message/count"),
  );
  return result.count;
}
