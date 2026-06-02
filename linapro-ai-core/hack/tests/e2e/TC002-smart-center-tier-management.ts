import { expect, test } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';
import { SmartCenterPage } from '../pages/SmartCenterPage';
import {
  bindTier,
  clearTier,
  createProviderWithModel,
  deleteProvider,
  deleteProviderRaw,
  updateTierRaw,
  withAdminApi,
} from '../support/ai-core-api';

test.describe('TC-2 智能中心档位管理', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(['linapro-ai-core']);
  });

  test('TC-2a: 三个文本能力档位稳定展示', async ({ adminPage }) => {
    const smartCenter = new SmartCenterPage(adminPage);
    await smartCenter.gotoTiers();

    await expect(adminPage.getByText(/基础|Basic/i)).toBeVisible();
    await expect(adminPage.getByText(/标准|Standard/i)).toBeVisible();
    await expect(adminPage.getByText(/高级|Advanced/i)).toBeVisible();
  });

  test('TC-2b: 不支持的 thinking effort 给出校验提示', async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
        supportedEfforts: ['low'],
        supportsThinking: 1,
      });
      try {
        const response = await updateTierRaw(api, 'basic', {
          defaultEffort: 'max',
          enabled: 1,
          modelId: fixture.modelId,
          providerId: fixture.providerId,
        });

        expect(response.ok()).toBe(true);
        await expect(response.text()).resolves.toMatch(
          /不支持.*thinking effort|does not support this thinking effort|THINKING_EFFORT/i,
        );
      } finally {
        await clearTier(api, 'basic').catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test('TC-2c: 禁用档位保留已有供应商模型绑定', async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-disable-model-${suffix}`,
        providerName: `E2E Disable Provider ${suffix}`,
        supportedEfforts: ['low'],
        supportsThinking: 1,
      });
      try {
        await bindTier(api, 'standard', fixture, 'low');

        const response = await updateTierRaw(api, 'standard', {
          defaultEffort: 'low',
          enabled: 0,
          modelId: 0,
          providerId: 0,
        });
        expect(response.ok()).toBe(true);

        const deleteResponse = await deleteProviderRaw(api, fixture.providerId);
        await expect(deleteResponse.text()).resolves.toMatch(
          /正在被能力档位使用|used by a capability tier|PROVIDER_IN_USE/i,
        );
      } finally {
        await clearTier(api, 'standard').catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });
});
