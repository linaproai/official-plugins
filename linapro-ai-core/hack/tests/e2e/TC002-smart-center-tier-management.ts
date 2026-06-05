import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  bindTier,
  clearTier,
  createProviderModel,
  createProviderWithModel,
  deleteProvider,
  deleteProviderRaw,
  listProviderEndpoints,
  listTiers,
  updateTierRaw,
  withAdminApi,
} from "../support/ai-core-api";

function createGate() {
  let release!: () => void;
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}

test.describe("TC-2 智能中心档位管理", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-2a: 三个文本能力档位稳定展示且列表不展示默认值", async ({
    adminPage,
    mainLayout,
  }) => {
    const smartCenter = new SmartCenterPage(adminPage);
    await mainLayout.switchLanguage("English");
    await smartCenter.gotoTiers();

    const mainContent = adminPage.locator("#__vben_main_content");
    await expect(mainContent.getByText("AI Tiers").first()).toBeVisible();
    await expect(mainContent.getByText("Tier Management")).toHaveCount(0);
    await expect(adminPage.getByText(/基础|Basic/i)).toBeVisible();
    await expect(adminPage.getByText(/标准|Standard/i)).toBeVisible();
    await expect(adminPage.getByText(/高级|Advanced/i)).toBeVisible();
    await smartCenter.assertTierThinkingEffortLabel();
  });

  test("TC-2b: 编辑档位默认使用模型默认且不维护默认参数 JSON", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-default-config-model-${suffix}`,
        providerName: `E2E Default Config Provider ${suffix}`,
      });
      const methodDefaultsRoute =
        "**/x/linapro-ai-core/api/v1/ai/method-defaults**";
      let methodDefaultsCalled = false;
      await adminPage.route(methodDefaultsRoute, async (route) => {
        methodDefaultsCalled = true;
        await route.abort();
      });

      try {
        await bindTier(api, "basic", fixture);

        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoTiers();
        await smartCenter.assertTierDrawerDefaultConfig(/基础|Basic/i);
        await smartCenter.captureEvidence("TC002-tier-config-drawer");
        await smartCenter.saveTierDrawer();

        const tiers = await listTiers(api);
        const basicTier = tiers.find((item: any) => item.code === "basic");
        expect(basicTier?.defaultEffort).toBe("");
        expect(methodDefaultsCalled).toBe(false);
      } finally {
        await adminPage.unroute(methodDefaultsRoute).catch(() => {});
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2c: 档位模型选择按渠道协议分组展示", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const anthropicModelName = `e2e-anthropic-tier-model-${suffix}`;
      const openAIModelName = `e2e-openai-tier-model-${suffix}`;
      const fixture = await createProviderWithModel(api, {
        anthropicEndpointUrl: "http://127.0.0.1:65535/anthropic/v1",
        modelName: openAIModelName,
        providerName: `E2E Tier Group Provider ${suffix}`,
      });
      const endpoints = await listProviderEndpoints(api, fixture.providerId, {
        protocol: "anthropic",
      });
      const anthropicEndpointId = Number(endpoints[0]?.id || 0);
      expect(anthropicEndpointId).toBeGreaterThan(0);
      await createProviderModel(api, fixture.providerId, {
        endpointId: anthropicEndpointId,
        modelName: anthropicModelName,
        protocol: "anthropic",
      });

      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoTiers();
        await smartCenter.assertTierModelOptionsGrouped({
          anthropicModelName,
          openAIModelName,
          providerName: fixture.providerName,
          tierName: /基础|Basic/i,
        });
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2d: 测试按钮请求中显示 loading 并禁止重复点击", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-loading-model-${suffix}`,
        providerName: `E2E Loading Provider ${suffix}`,
      });
      await bindTier(api, "basic", fixture, "low");

      const routePattern = "**/x/linapro-ai-core/api/v1/ai/tiers/basic/test";
      const gates = [createGate(), createGate()];
      let routeCalls = 0;
      await adminPage.route(routePattern, async (route) => {
        const current = routeCalls;
        routeCalls += 1;
        await gates[current]?.promise;
        await route.fulfill({
          body: JSON.stringify({
            code: 0,
            data: {
              errorSummary:
                current === 0
                  ? "E2E delayed saved test"
                  : "E2E delayed draft test",
              latencyMs: 0,
              modelName: fixture.modelName,
              protocol: "openai",
              providerName: fixture.providerName,
              status: "failed",
              testedAt: Date.now(),
              thinkingEffort: "low",
            },
          }),
          contentType: "application/json",
          status: 200,
        });
      });

      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoTiers();
        await smartCenter.clickSavedTierTestAndAssertLoading(/基础|Basic/i);
        gates[0].release();
        await expect(
          adminPage.getByText("E2E delayed saved test").first(),
        ).toBeVisible();

        await smartCenter.clickDraftTierTestAndAssertLoading(/基础|Basic/i);
        gates[1].release();
        await expect(
          adminPage.getByText("E2E delayed draft test").first(),
        ).toBeVisible();
        await smartCenter.assertDraftTierCurrentTestLatency("0ms");
        await smartCenter.cancelDrawer();
      } finally {
        await adminPage.unroute(routePattern).catch(() => {});
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2e: 默认 thinking effort 只做枚举校验不依赖模型声明", async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
      });
      try {
        const response = await updateTierRaw(api, "basic", {
          defaultEffort: "max",
          enabled: 1,
          modelId: fixture.modelId,
          providerId: fixture.providerId,
        });

        expect(response.ok()).toBe(true);
        const tiers = await listTiers(api);
        const basicTier = tiers.find((item: any) => item.code === "basic");
        expect(basicTier?.defaultEffort).toBe("max");
      } finally {
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2f: 禁用档位保留已有渠道模型绑定", async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-disable-model-${suffix}`,
        providerName: `E2E Disable Provider ${suffix}`,
      });
      try {
        await bindTier(api, "standard", fixture, "low");

        const response = await updateTierRaw(api, "standard", {
          defaultEffort: "low",
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
        await clearTier(api, "standard").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });
});
