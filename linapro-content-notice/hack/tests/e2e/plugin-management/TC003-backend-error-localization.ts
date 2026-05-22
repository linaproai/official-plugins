import type { APIRequestContext, APIResponse } from "@host-tests/support/playwright";

import { test, expect } from '@host-tests/fixtures/auth';
import { pluginApiPath } from '@host-tests/fixtures/config';
import {
  createAdminApiContext,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
} from '@host-tests/support/api/job';

type ErrorEnvelope = {
  code: number;
  errorCode?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
};

const sourcePluginIDs = [
  "linapro-content-notice",
] as const;
const pluginID = "linapro-content-notice";

const errorCases = [
  {
    errorCode: "CONTENT_NOTICE_NOT_FOUND",
    messageKey: "error.content.notice.not.found",
    messages: {
      "en-US": "Notice does not exist",
      "zh-CN": "通知公告不存在",
    },
    path: "notice/99999999",
  },
] as const;

async function ensureSourcePluginsEnabled(
  api: APIRequestContext,
  pluginIDs: readonly string[],
) {
  await syncPlugins(api);
  for (const pluginID of pluginIDs) {
    let plugin = await getPlugin(api, pluginID);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginID);
      plugin = await getPlugin(api, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(api, pluginID);
    }
  }
}

async function expectBackendError(
  response: APIResponse,
): Promise<ErrorEnvelope> {
  const payload = (await response.json()) as ErrorEnvelope;
  expect(payload.code).not.toBe(0);
  return payload;
}

test.describe("TC-3 Backend error localization", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-3a: business errors keep stable codes while messages follow request locale", async () => {
    for (const errorCase of errorCases) {
      for (const locale of ["zh-CN", "en-US"] as const) {
        const payload = await expectBackendError(
          await adminApi.get(pluginApiPath(pluginID, errorCase.path), {
            headers: { "Accept-Language": locale },
          }),
        );

        expect(payload.errorCode).toBe(errorCase.errorCode);
        expect(payload.messageKey).toBe(errorCase.messageKey);
        expect(payload.message).toBe(errorCase.messages[locale]);
      }
    }
  });
});
