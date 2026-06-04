import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  bindTier,
  clearTier,
  createProviderModel,
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
        anthropicEndpointUrl: `https://api.anthropic.example.com/v1/workspaces/${suffix}/long-provider-endpoint-url-rendering-check`,
        modelName: "gpt-4o",
        openaiEndpointUrl: `https://api.openai.example.com/v1/organizations/${suffix}/projects/long-provider-endpoint-url-rendering-check`,
        providerName: `E2E Provider ${suffix}`,
        secretRef: "sk-1234567890",
        websiteUrl: `https://example.com/e2e-provider-${suffix}`,
      });
      const multiProtocolModelName = "mimo";
      const managedModelName = `e2e-managed-model-${suffix}`;
      const renamedManagedModelName = `e2e-managed-model-renamed-${suffix}`;
      const overflowModelNames = [
        "claude-3-5",
        "o4-mini",
        "qwen3",
        "deepseek-v3",
        "gpt-4.1",
        "gemini-2.5",
        "mistral",
        "llama-4",
      ];
      try {
        await createProviderModel(api, fixture.providerId, {
          capabilityMethod: "generate",
          capabilityType: "text",
          endpointId: fixture.endpointId,
          maxOutputTokens: 384,
          modelName: managedModelName,
          protocol: "openai",
        });
        for (const modelName of overflowModelNames) {
          await createProviderModel(api, fixture.providerId, {
            capabilityMethod: "generate",
            capabilityType: "text",
            endpointId: fixture.endpointId,
            maxOutputTokens: 384,
            modelName,
            protocol: "openai",
          });
        }
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.assertProviderTabs();
        await smartCenter.assertCreateProviderDrawerChineseTranslations();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.assertProviderVisible(fixture.providerName);
        await smartCenter.assertProviderListProjection(fixture);
        await smartCenter.assertProviderSyncActions({
          providerName: fixture.providerName,
        });
        await smartCenter.captureEvidence("TC001-provider-list-layout");
        await smartCenter.assertProviderRowAddModelDefaults(
          fixture.providerName,
        );
        await smartCenter.assertCreateModelDrawerChineseTranslations(
          fixture.providerName,
        );
        await smartCenter.createModelForProviderProtocols({
          modelName: multiProtocolModelName,
          providerName: fixture.providerName,
          protocolLabels: [/OpenAI/i, /Anthropic/i],
        });
        await smartCenter.assertModelManagementProjection({
          endpointUrl: fixture.openaiEndpointUrl,
          modelName: managedModelName,
          protocolLabel: /OpenAI/i,
          providerName: fixture.providerName,
        });
        await smartCenter.renameModelFromModelManagement({
          modelName: managedModelName,
          nextModelName: renamedManagedModelName,
        });
        await smartCenter.deleteModelFromModelManagement(
          renamedManagedModelName,
        );
        await smartCenter.openProviderManagementTab();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.deleteModelFromProviderRow(
          fixture.providerName,
          fixture.modelName,
        );
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-1b: 编辑供应商名称和接入配置", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        anthropicEndpointUrl: `http://127.0.0.1:65535/anthropic-${suffix}`,
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
      });
      const renamedProviderName = `E2E Provider Renamed ${suffix}`;
      const updatedOpenaiUrl = `https://example.com/e2e-openai-${suffix}/v1`;
      const updatedAnthropicUrl = `https://example.com/e2e-anthropic-${suffix}/v1`;
      const updatedSecret = "sk-updated-1234567890";
      try {
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.openProvider(fixture.providerName);
        await smartCenter.assertEditProviderMetadataForm({
          anthropicEndpointUrl: fixture.anthropicEndpointUrl,
          openaiEndpointUrl: fixture.openaiEndpointUrl,
        });
        await smartCenter.captureEvidence(
          "TC001-provider-edit-agent-box-fields",
        );
        await smartCenter.fillProvider({
          anthropicBaseUrl: updatedAnthropicUrl,
          name: renamedProviderName,
          openaiBaseUrl: updatedOpenaiUrl,
          secretRef: updatedSecret,
        });
        await smartCenter.confirmDrawer();

        await smartCenter.searchProvider(renamedProviderName);
        await smartCenter.assertProviderVisible(renamedProviderName);
        await smartCenter.assertProviderRowEndpoint(
          renamedProviderName,
          updatedOpenaiUrl,
          "OpenAI",
        );
        await smartCenter.assertProviderRowEndpoint(
          renamedProviderName,
          updatedAnthropicUrl,
          "Anthropic",
        );
        await smartCenter.assertProviderRowSecret(
          renamedProviderName,
          "sk-**********90",
        );
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
