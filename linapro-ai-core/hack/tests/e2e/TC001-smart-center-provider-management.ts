import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  bindTier,
  clearTier,
  createProviderWithModel,
  deleteProvider,
  withAdminApi,
} from "../support/ai-core-api";

test.describe("TC-1 智能中心供应商管理", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-1a: 供应商列表可查看模型维护结果", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        anthropicEndpointUrl: "http://127.0.0.1:65535/anthropic",
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
        secretRef: "sk-1234567890",
        websiteUrl: `https://example.com/e2e-provider-${suffix}`,
      });
      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.assertCreateProviderDrawerChineseTranslations();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.assertProviderVisible(fixture.providerName);
        await smartCenter.assertProviderListProjection(fixture);
        await smartCenter.assertProviderSyncActions({
          providerName: fixture.providerName,
          syncAnthropic: true,
          syncOpenAI: true,
        });
        await smartCenter.assertCreateModelDrawerChineseTranslations(
          fixture.providerName,
        );
        await smartCenter.deleteModelFromProviderRow(
          fixture.providerName,
          fixture.modelName,
        );
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-1b: 编辑供应商名称", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
      });
      const renamedProviderName = `E2E Provider Renamed ${suffix}`;
      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.openProvider(fixture.providerName);
        await smartCenter.assertEditProviderMetadataForm();
        await smartCenter.fillProvider({
          name: renamedProviderName,
        });
        await smartCenter.confirmDrawer();

        await smartCenter.searchProvider(renamedProviderName);
        await smartCenter.assertProviderVisible(renamedProviderName);
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-1c: 被档位引用的供应商不能删除", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
      });
      await bindTier(api, "basic", fixture);
      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.deleteProvider(fixture.providerName);
        await expect(
          adminPage.getByText(/正在被能力档位使用|used by a capability tier/i),
        ).toBeVisible();
      } finally {
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });
});
