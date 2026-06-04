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
  listMethodDefaults,
  listTiers,
  updateMethodDefault,
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
  }) => {
    const smartCenter = new SmartCenterPage(adminPage);
    await smartCenter.gotoTiers();

    await expect(adminPage.getByText(/基础|Basic/i)).toBeVisible();
    await expect(adminPage.getByText(/标准|Standard/i)).toBeVisible();
    await expect(adminPage.getByText(/高级|Advanced/i)).toBeVisible();
    await smartCenter.assertTierThinkingEffortLabel();
  });

  test("TC-2b: 编辑档位默认使用模型默认并可在双列表单中保存方法参数", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const originalDefaults = await listMethodDefaults(api);
      const originalTextDefault = originalDefaults.find(
        (item: any) =>
          item.capabilityType === "text" &&
          item.capabilityMethod === "generate",
      );
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-default-config-model-${suffix}`,
        providerName: `E2E Default Config Provider ${suffix}`,
      });
      const updatedMaxOutputTokens = 1200 + (suffix % 100);
      const updatedParamsJson = `{
  "maxOutputTokens": ${updatedMaxOutputTokens},
  "temperature": 0.2
}`;

      try {
        await updateMethodDefault(
          api,
          "text",
          "generate",
          '{"maxOutputTokens":1024}',
        );
        await bindTier(api, "basic", fixture);

        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoTiers();
        await smartCenter.assertTierDrawerDefaultConfig(
          /基础|Basic/i,
          '"maxOutputTokens": 1024',
        );
        await smartCenter.assertTierDefaultParamsFormLayout();
        await smartCenter.fillTierDefaultParams(updatedParamsJson);
        await smartCenter.captureEvidence("TC002-tier-default-config-drawer");
        await smartCenter.saveTierDrawer();

        const [tiers, defaults] = await Promise.all([
          listTiers(api),
          listMethodDefaults(api),
        ]);
        const basicTier = tiers.find((item: any) => item.code === "basic");
        const textDefault = defaults.find(
          (item: any) =>
            item.capabilityType === "text" &&
            item.capabilityMethod === "generate",
        );
        expect(basicTier?.defaultEffort).toBe("");
        expect(textDefault?.defaultParamsJson).toBe(
          `{"maxOutputTokens":${updatedMaxOutputTokens},"temperature":0.2}`,
        );
      } finally {
        await updateMethodDefault(
          api,
          "text",
          "generate",
          originalTextDefault?.defaultParamsJson || '{"maxOutputTokens":1024}',
        ).catch(() => {});
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2c: 档位模型选择按供应商协议分组展示", async ({
    adminPage,
  }) => {
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
        capabilityMethod: "generate",
        capabilityType: "text",
        endpointId: anthropicEndpointId,
        maxOutputTokens: 512,
        modelName: anthropicModelName,
        protocol: "anthropic",
        supportedEfforts: ["low", "medium"],
        supportsThinking: 1,
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
        supportedEfforts: ["low"],
        supportsThinking: 1,
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

  test("TC-2e: 不支持的 thinking effort 给出校验提示", async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
        supportedEfforts: ["low"],
        supportsThinking: 1,
      });
      try {
        const response = await updateTierRaw(api, "basic", {
          defaultEffort: "max",
          enabled: 1,
          modelId: fixture.modelId,
          providerId: fixture.providerId,
        });

        expect(response.ok()).toBe(true);
        await expect(response.text()).resolves.toMatch(
          /不支持.*thinking effort|does not support this thinking effort|THINKING_EFFORT/i,
        );
      } finally {
        await clearTier(api, "basic").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-2f: 禁用档位保留已有供应商模型绑定", async () => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-disable-model-${suffix}`,
        providerName: `E2E Disable Provider ${suffix}`,
        supportedEfforts: ["low"],
        supportsThinking: 1,
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
