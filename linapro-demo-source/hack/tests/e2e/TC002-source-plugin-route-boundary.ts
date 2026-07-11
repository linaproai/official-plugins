import { test, expect } from '@host-tests/fixtures/auth';
import { config } from '@host-tests/fixtures/config';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';
import { request as playwrightRequest } from '@host-tests/support/playwright';

const pluginID = 'linapro-demo-source';

test.describe('TC-2 源码插件路由边界', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([pluginID]);
  });

  test('TC-2a: 源码插件公开路由不被管理工作台 fallback 吞掉', async () => {
    const publicApi = await playwrightRequest.newContext({
      baseURL: config.publicBaseURL,
    });
    try {
      const response = await publicApi.get('/portal/linapro-demo-source/ping');
      expect(response.status()).toBe(200);
      await expect(response.text()).resolves.toBe('linapro-demo-source-public-pong');
    } finally {
      await publicApi.dispose();
    }
  });

  test('TC-2b: 源码插件 API 使用统一插件 API 前缀', async () => {
    const publicApi = await playwrightRequest.newContext({
      baseURL: config.publicBaseURL,
    });
    try {
      const response = await publicApi.get(
        '/x/linapro-demo-source/api/v1/ping',
      );
      expect(response.status()).toBe(200);
      const payload = await response.json();
      expect(payload?.code).toBe(0);
      expect(payload?.data?.message).toBe('pong');
    } finally {
      await publicApi.dispose();
    }
  });
});
