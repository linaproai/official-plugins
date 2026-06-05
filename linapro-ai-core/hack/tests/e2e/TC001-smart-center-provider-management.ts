import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  bindTier,
  clearTier,
  createProviderModel,
  createProviderWithModel,
  deleteProvider,
  insertProviderModelIdentityOnly,
  listProviderModels,
  withAdminApi,
} from "../support/ai-core-api";

const legacyChineseProviderTerm = "\u4f9b\u5e94\u5546";
const legacyChineseProviderMenuTerm = `${legacyChineseProviderTerm}\u7ba1\u7406`;
const legacyChineseProviderListTerm = `${legacyChineseProviderTerm}\u5217\u8868`;

test.describe("TC-1 智能中心渠道管理", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-1a: 渠道列表可查看模型维护结果", async ({ adminPage }) => {
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
          endpointId: fixture.endpointId,
          modelName: managedModelName,
          protocol: "openai",
        });
        for (const modelName of overflowModelNames) {
          await createProviderModel(api, fixture.providerId, {
            endpointId: fixture.endpointId,
            modelName,
            protocol: "openai",
          });
        }
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.assertProviderPageWithoutTabs();
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
        await smartCenter.deleteModelFromProviderRow(
          fixture.providerName,
          multiProtocolModelName,
        );
        const modelsAfterAggregateDelete = await listProviderModels(
          api,
          fixture.providerId,
        );
        expect(
          modelsAfterAggregateDelete.filter(
            (item: { modelName?: string } | undefined) =>
              item?.modelName === multiProtocolModelName,
          ),
        ).toHaveLength(0);
        await smartCenter.assertModelManagementProjection({
          endpointUrl: fixture.openaiEndpointUrl,
          modelName: managedModelName,
          protocolLabel: /OpenAI/i,
          providerName: fixture.providerName,
        });
        await smartCenter.assertModelManagementHidesCapabilityControls(
          managedModelName,
        );
        const listedModels = await listProviderModels(api, fixture.providerId);
        expect(
          listedModels.filter(
            (item: { modelName?: string } | undefined) =>
              item?.modelName === managedModelName,
          ),
        ).toHaveLength(1);
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

  test("TC-1b: 编辑渠道名称和接入配置", async ({ adminPage }) => {
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

  test("TC-1c: 被档位引用的渠道不能删除", async ({ adminPage }) => {
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

  test("TC-1d: 渠道和模型菜单拆分且搜索区使用默认样式", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("简体中文");
    await mainLayout.expandSidebarGroup("智能中心");
    const providerMenu = mainLayout.sidebarMenuItem("渠道管理");
    const modelMenu = mainLayout.sidebarMenuItem("模型管理");
    const tierMenu = mainLayout.sidebarMenuItem("档位管理");
    await expect(providerMenu).toBeVisible();
    await expect(modelMenu).toBeVisible();
    await expect(tierMenu).toBeVisible();
    await expect(
      mainLayout.sidebarMenuItem(legacyChineseProviderTerm),
    ).toHaveCount(0);
    await expect(
      mainLayout.sidebarMenuItem(legacyChineseProviderMenuTerm),
    ).toHaveCount(0);
    const [providerBox, modelBox, tierBox] = await Promise.all([
      providerMenu.boundingBox(),
      modelMenu.boundingBox(),
      tierMenu.boundingBox(),
    ]);
    expect(modelBox?.y ?? 0).toBeGreaterThan(providerBox?.y ?? 0);
    expect(tierBox?.y ?? 0).toBeGreaterThan(modelBox?.y ?? 0);

    const smartCenter = new SmartCenterPage(adminPage);
    await smartCenter.gotoProviders();
    await smartCenter.assertProviderPageWithoutTabs();
    await smartCenter.assertProviderSearchFormDefaultSpacing();
    await expect(
      adminPage.getByText("渠道列表", { exact: true }),
    ).toBeVisible();
    await expect(
      adminPage.getByText(legacyChineseProviderListTerm, { exact: true }),
    ).toHaveCount(0);
    await smartCenter.captureEvidence(
      "TC001-provider-management-search-default",
    );
    await smartCenter.gotoModels();
    await expect(
      adminPage.getByTestId("ai-model-management-page"),
    ).toBeVisible();
    await smartCenter.captureEvidence("TC001-provider-model-split-menus");
  });

  test("TC-1e: 同步态模型身份显示可维护状态", async ({ adminPage }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const fixture = await createProviderWithModel(api, {
        modelName: `e2e-model-${suffix}`,
        providerName: `E2E Provider ${suffix}`,
      });
      const syncedModelName = `e2e-synced-identity-${suffix}`;
      try {
        insertProviderModelIdentityOnly({
          endpointId: fixture.endpointId,
          modelName: syncedModelName,
          protocol: "openai",
          providerId: fixture.providerId,
        });
        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.gotoProviders();
        await smartCenter.searchProvider(fixture.providerName);
        await smartCenter.assertProviderIdentityModel(
          fixture.providerName,
          syncedModelName,
        );
        await smartCenter.captureEvidence("TC001-provider-identity-model");
      } finally {
        await deleteProvider(api, fixture.providerId).catch(() => {});
      }
    });
  });

  test("TC-1f: 模型管理支持渠道筛选且不显示能力方法筛选", async ({
    adminPage,
  }) => {
    await withAdminApi(async (api) => {
      const suffix = Date.now();
      const primary = await createProviderWithModel(api, {
        modelName: `e2e-filter-text-${suffix}`,
        providerName: `E2E Filter Provider ${suffix}`,
      });
      const secondary = await createProviderWithModel(api, {
        modelName: `e2e-filter-other-text-${suffix}`,
        providerName: `E2E Filter Other Provider ${suffix}`,
      });
      const primaryExtraModelName = `e2e-filter-extra-${suffix}`;
      const secondaryExtraModelName = `e2e-filter-other-extra-${suffix}`;
      try {
        await createProviderModel(api, primary.providerId, {
          endpointId: primary.endpointId,
          modelName: primaryExtraModelName,
          protocol: "openai",
        });
        await createProviderModel(api, secondary.providerId, {
          endpointId: secondary.endpointId,
          modelName: secondaryExtraModelName,
          protocol: "openai",
        });

        const smartCenter = new SmartCenterPage(adminPage);
        await smartCenter.filterModelsByProviderOnly({
          expectedModelNames: [primary.modelName, primaryExtraModelName],
          hiddenModelNames: [secondary.modelName, secondaryExtraModelName],
          providerId: primary.providerId,
          providerName: primary.providerName,
        });
      } finally {
        await deleteProvider(api, primary.providerId).catch(() => {});
        await deleteProvider(api, secondary.providerId).catch(() => {});
      }
    });
  });
});
