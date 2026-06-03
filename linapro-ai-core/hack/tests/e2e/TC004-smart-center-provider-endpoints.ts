import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  createProviderEndpoint,
  createProviderModel,
  createProviderWithModel,
  deleteProvider,
  deleteProviderEndpointRaw,
  listModelCapabilities,
  listProviderEndpoints,
  saveModelCapabilities,
  withAdminApi,
} from "../support/ai-core-api";

test.describe("TC-4 智能中心供应商端点", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-4a: 供应商端点和模型能力按多模态方法维护", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-text-model-${suffix}`,
        providerName: `E2E Endpoint Provider ${suffix}`,
      });
      const endpointBaseUrl = `https://example.com/e2e-anthropic-${suffix}/v1`;
      const imageModelName = `e2e-image-model-${suffix}`;
      try {
        const endpointId = await createProviderEndpoint(
          api,
          fixture.providerId,
          {
            baseUrl: endpointBaseUrl,
            metadataJson: '{"region":"e2e"}',
            protocol: "anthropic",
            secretRef: "sk-endpoint-1234567890",
          },
        );
        const modelId = await createProviderModel(api, fixture.providerId, {
          capabilityMethod: "generate",
          capabilityType: "image",
          endpointId,
          maxInputTokens: 2048,
          maxOutputTokens: 4096,
          modelName: imageModelName,
          protocol: "anthropic",
        });
        await saveModelCapabilities(api, modelId, [
          {
            capabilityMethod: "generate",
            capabilityType: "image",
            defaultParamsJson: '{"size":"1024x1024"}',
            enabled: 1,
            endpointId,
            inputModalities: ["text"],
            maxOutputAssets: 1,
            outputModalities: ["image"],
            supportsOperation: 1,
          },
        ]);

        const endpoints = await listProviderEndpoints(api, fixture.providerId, {
          protocol: "anthropic",
        });
        expect(endpoints).toHaveLength(1);
        expect(endpoints[0].secretRef).toMatch(/\*\*\*\*\*\*\*\*\*\*/);

        const capabilities = await listModelCapabilities(api, modelId);
        expect(capabilities).toMatchObject([
          {
            capabilityMethod: "generate",
            capabilityType: "image",
            endpointId,
            outputModalities: ["image"],
          },
        ]);

        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.assertProviderVisible(fixture.providerName);
        await smartCenter.assertProviderRowEndpoint(
          fixture.providerName,
          endpointBaseUrl,
          "Anthropic",
        );
        await smartCenter.captureEvidence("TC004-provider-endpoint-list");

        await smartCenter.openProviderEndpoints(fixture.providerName);
        await smartCenter.assertEndpointDrawerChineseTranslations(
          fixture.providerName,
        );
        await smartCenter.assertEndpointVisible({
          baseUrl: endpointBaseUrl,
          protocolLabel: "Anthropic",
          secretText: "sk-**********90",
        });
        await smartCenter.captureEvidence("TC004-provider-endpoint-drawer");
        await smartCenter.cancelDrawer();

        const deleteResponse = await deleteProviderEndpointRaw(
          api,
          fixture.providerId,
          endpointId,
        );
        await expect(deleteResponse.text()).resolves.toMatch(
          /AI_CORE_PROVIDER_ENDPOINT_IN_USE|供应商端点正在被模型能力使用|provider endpoint is used/i,
        );
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });
});
