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

const sourcePluginIDs = ["linapro-monitor-loginlog"] as const;
const pluginID = "linapro-monitor-loginlog";

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

test.describe("TC-3 Login log backend error localization", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-3a: login-log not-found error keeps stable code while message follows request locale", async () => {
    for (const locale of ["zh-CN", "en-US"] as const) {
      const payload = await expectBackendError(
        await adminApi.get(pluginApiPath(pluginID, "loginlog/99999999"), {
          headers: { "Accept-Language": locale },
        }),
      );

      expect(payload.errorCode).toBe("MONITOR_LOGINLOG_NOT_FOUND");
      expect(payload.messageKey).toBe("error.monitor.loginlog.not.found");
      expect(payload.message).toBe(
        locale === "en-US" ? "Login log does not exist" : "登录日志不存在",
      );
    }
  });
});
