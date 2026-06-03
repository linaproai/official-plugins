import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  bindCapabilityTier,
  clearTier,
  clearTierUpdatedAt,
  createProviderEndpoint,
  createProviderModel,
  createProviderWithModel,
  deleteProvider,
  saveModelCapabilities,
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

test.describe("TC-5 智能中心多模态档位", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-5a: 能力类型 Tab、三档绑定、默认参数和错误路径", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-text-tier-model-${suffix}`,
        providerName: `E2E Tier Provider ${suffix}`,
      });
      const endpointId = await createProviderEndpoint(api, fixture.providerId, {
        baseUrl: `https://example.com/e2e-document-${suffix}/v1`,
        protocol: "openai-compatible",
      });
      const documentModelId = await createProviderModel(
        api,
        fixture.providerId,
        {
          capabilityMethod: "analyze",
          capabilityType: "document",
          endpointId,
          maxInputTokens: 8192,
          maxOutputTokens: 2048,
          modelName: `e2e-document-model-${suffix}`,
          protocol: "openai-compatible",
        },
      );
      await saveModelCapabilities(api, documentModelId, [
        {
          capabilityMethod: "analyze",
          capabilityType: "document",
          defaultParamsJson: '{"maxOutputTokens":2048}',
          enabled: 1,
          endpointId,
          inputModalities: ["document"],
          outputModalities: ["text"],
        },
      ]);
      await bindCapabilityTier(
        api,
        "basic",
        { modelId: documentModelId, providerId: fixture.providerId },
        {
          capabilityMethod: "analyze",
          capabilityType: "document",
        },
      );
      clearTierUpdatedAt("basic", "document", "analyze");

      const routePattern = "**/x/linapro-ai-core/api/v1/ai/tiers/basic/test";
      const gate = createGate();
      await adminPage.route(routePattern, async (route) => {
        const body = route.request().postDataJSON() as Record<string, unknown>;
        expect(body.capabilityType).toBe("document");
        expect(body.capabilityMethod).toBe("analyze");
        await gate.promise;
        await route.fulfill({
          body: JSON.stringify({
            code: 0,
            data: {
              errorSummary: "E2E document analyze test failure",
              latencyMs: 0,
              modelName: `e2e-document-model-${suffix}`,
              protocol: "openai-compatible",
              providerName: fixture.providerName,
              status: "failed",
              testedAt: Date.now(),
              thinkingEffort: "",
            },
          }),
          contentType: "application/json",
          status: 200,
        });
      });

      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoTiers();
        await smartCenter.assertTierCapabilityTypeTabs();
        await smartCenter.assertTierTabsVisualStyle();
        await smartCenter.selectTierCapabilityType("document");
        await smartCenter.assertTierTypePage(
          "document",
          '{"maxOutputTokens":2048}',
        );
        await smartCenter.assertTierUpdatedAtHidden(/基础|Basic/i);
        await smartCenter.captureEvidence("TC005-document-tier-tab-list");
        await smartCenter.assertTierDrawerWithoutThinkingEffort(/基础|Basic/i);

        await smartCenter.clickSavedTierTestAndAssertLoading(/基础|Basic/i);
        await smartCenter.captureEvidence("TC005-document-tier-test-loading");
        gate.release();
        await expect(
          adminPage.getByText("E2E document analyze test failure").first(),
        ).toBeVisible();

        const invalidResponse = await updateTierRaw(api, "standard", {
          capabilityMethod: "analyze",
          capabilityType: "document",
          enabled: 1,
          modelId: fixture.modelId,
          providerId: fixture.providerId,
        });
        await expect(invalidResponse.text()).resolves.toMatch(
          /AI_CORE_MODEL_NOT_FOUND|AI model does not exist|模型不存在/i,
        );
      } finally {
        await adminPage.unroute(routePattern).catch(() => {});
        await clearTier(api, "basic", "document", "analyze").catch(() => {});
        await clearTier(api, "standard", "document", "analyze").catch(() => {});
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });
});
