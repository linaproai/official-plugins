import { expect, type Page } from "@host-tests/support/playwright";

import { mkdirSync } from "node:fs";
import path from "node:path";

import { workspacePath } from "@host-tests/fixtures/config";
import {
  closeDialogWithEscape,
  waitForBusyIndicatorsToClear,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from "@host-tests/support/ui";

const repoRoot = path.resolve(process.cwd(), "../..");
const capabilityMethodOptionOrder = [
  "text.generate",
  "image.generate",
  "image.edit",
  "embedding.create",
  "audio.transcribe",
  "audio.synthesize",
  "vision.analyze",
  "document.analyze",
  "document.cite",
  "safety.moderate",
  "video.generate",
  "video.edit",
  "video.extend",
  "video.operation.get",
  "video.operation.cancel",
];
const tierCapabilityTypeLabels: Record<string, { en: string; zh: string }> = {
  audio: { en: "Audio", zh: "音频" },
  document: { en: "Document", zh: "文档理解" },
  embedding: { en: "Embedding", zh: "向量嵌入" },
  image: { en: "Image", zh: "图像" },
  safety: { en: "Safety", zh: "安全审核" },
  text: { en: "Text", zh: "文本" },
  video: { en: "Video", zh: "视频" },
  vision: { en: "Vision", zh: "视觉理解" },
};

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function cssAttributeValue(value: string) {
  return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

function screenshotName(name: string) {
  const timestamp = new Date().toISOString().replace(/\D/g, "").slice(0, 14);
  const safeName = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return `${timestamp}-${safeName || "screenshot"}.png`;
}

export class SmartCenterPage {
  constructor(private page: Page) {}

  private get dialog() {
    return this.page.locator('[role="dialog"]').last();
  }

  private providerNameInput() {
    return this.dialog.getByRole("textbox", { name: /名称|Name/i });
  }

  private providerApiKeyInput() {
    return this.dialog.getByLabel(/API 密钥|API Key/i).first();
  }

  private providerOpenAIBaseUrlInput() {
    return this.dialog
      .getByRole("textbox", {
        name: /OpenAI\s*(接入地址|基础地址|Access URL|Base URL)/i,
      })
      .first();
  }

  private providerAnthropicBaseUrlInput() {
    return this.dialog
      .getByRole("textbox", {
        name: /Anthropic\s*(接入地址|基础地址|Access URL|Base URL)/i,
      })
      .first();
  }

  async gotoProviders() {
    await this.page.goto(workspacePath("/ai/providers"));
    await waitForTableReady(this.page);
  }

  async gotoTiers() {
    await this.page.goto(workspacePath("/ai/tiers"));
    await waitForTableReady(this.page);
  }

  async assertTierThinkingEffortLabel() {
    await expect(
      this.page.getByText("Thinking Effort", { exact: true }),
    ).toHaveCount(0);
    await expect(this.page.getByText("默认 Thinking Effort")).toHaveCount(0);
    await expect(this.page.getByText("Default Thinking Effort")).toHaveCount(0);
  }

  async gotoInvocations() {
    await this.page.goto(workspacePath("/ai/invocations"));
    await waitForTableReady(this.page);
  }

  async openCreateProvider() {
    await this.page
      .getByRole("button", { name: /新\s*增\s*供\s*应\s*商|Add Provider/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async openCreateModel() {
    await this.page
      .getByRole("button", { name: /新\s*增\s*模\s*型|Add Model/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async assertCreateProviderDrawerChineseTranslations() {
    await this.openCreateProvider();
    await expect(this.dialog.getByText("新增供应商")).toBeVisible();
    await expect(this.providerNameInput()).toBeVisible();
    await expect(this.dialog.getByText("端点配置")).toHaveCount(0);
    await expect(this.dialog.getByText("供应商名称")).toHaveCount(0);
    await expect(this.dialog.getByText("API 密钥")).toBeVisible();
    await expect(this.providerApiKeyInput()).toBeVisible();
    await expect(this.providerApiKeyInput()).toHaveAttribute(
      "placeholder",
      /输入 API 密钥|Enter an API key/i,
    );
    await expect(
      this.dialog.getByText(/OpenAI\s*(接入地址|基础地址)/),
    ).toBeVisible();
    await expect(this.providerOpenAIBaseUrlInput()).toBeVisible();
    await expect(
      this.dialog.getByText(/Anthropic\s*(接入地址|基础地址)/),
    ).toBeVisible();
    await expect(this.providerAnthropicBaseUrlInput()).toHaveValue(
      "https://api.anthropic.com/v1",
    );
    await expect(
      this.dialog.getByPlaceholder("https://api.openai.com/v1"),
    ).toBeVisible();
    await expect(
      this.dialog.getByPlaceholder("https://api.anthropic.com/v1"),
    ).toBeVisible();
    await expect(
      this.dialog.getByText("启用", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      this.dialog.getByText("停用", { exact: true }).first(),
    ).toBeVisible();
    await expect(this.dialog.getByText("新增模型")).toHaveCount(0);
    await expect(this.dialog.getByText("模型名称")).toHaveCount(0);
    await expect(
      this.dialog.getByText("plugin.linapro-ai-core.common.enabled"),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText("plugin.linapro-ai-core.common.disabled"),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText("plugin.linapro-ai-core.effort.empty"),
    ).toHaveCount(0);
    await expect(this.dialog.getByText(/plugin\.linapro-ai-core/)).toHaveCount(
      0,
    );
    await this.cancelDrawer();
  }

  async assertCreateModelDrawerChineseTranslations(providerName: string) {
    await this.openCreateModel();
    await expect(this.dialog.getByText("新增模型")).toBeVisible();
    await expect(this.dialog.getByText("供应商")).toBeVisible();
    await expect(this.dialog.getByText("模型名称")).toBeVisible();
    await this.dialog.getByLabel(/供应商|Provider/i).click();
    await this.page.getByTitle(providerName).click();
    await expect(this.dialog.getByTitle(providerName)).toBeVisible();
    await this.cancelDrawer();
  }

  async assertProviderListProjection(input: {
    anthropicEndpointUrl?: string;
    maskedApiKey: string;
    modelName: string;
    openaiEndpointUrl: string;
    providerName: string;
    websiteUrl: string;
  }) {
    await expect(
      this.page.getByRole("button", { name: /新\s*增\s*模\s*型|Add Model/i }),
    ).toBeVisible();
    await expect(
      this.page.getByRole("button", {
        name: /新\s*增\s*供\s*应\s*商|Add Provider/i,
      }),
    ).toBeVisible();
    await expect(
      this.page.getByText("模型", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      this.page.getByText("端点", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      this.page.getByText("密钥", { exact: true }).first(),
    ).toBeVisible();
    const modelHeaderIndex = await this.providerHeaderIndex(/模型|Models/i);
    const endpointHeaderIndex =
      await this.providerHeaderIndex(/端点|Endpoint/i);
    expect(endpointHeaderIndex).toBe(modelHeaderIndex + 1);
    await expect(this.page.getByText("模型数", { exact: true })).toHaveCount(0);
    await expect(
      this.page.getByText("启用模型数", { exact: true }),
    ).toHaveCount(0);

    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: input.providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
    const websiteLink = row
      .first()
      .getByRole("link", { name: input.websiteUrl });
    await expect(websiteLink).toBeVisible();
    await expect(websiteLink).toHaveAttribute("href", input.websiteUrl);
    await expect(websiteLink).toHaveAttribute("target", "_blank");
    const popupPromise = this.page.waitForEvent("popup");
    await websiteLink.click();
    const popup = await popupPromise;
    await expect.poll(() => popup.url()).toContain(input.websiteUrl);
    await popup.close();
    await expect(row.first()).toContainText(input.modelName);
    await expect(row.first()).toContainText("OpenAI");
    const modelText = row.first().getByText(input.modelName).first();
    await expect
      .poll(async () => {
        const weight = await modelText.evaluate((node) =>
          Number.parseInt(window.getComputedStyle(node).fontWeight, 10),
        );
        return Number.isNaN(weight) ? 400 : weight;
      })
      .toBeLessThan(600);
    const deleteModelButton = row.first().getByRole("button", {
      name: new RegExp(
        `删\\s*除.*${escapeRegExp(input.modelName)}|Delete.*${escapeRegExp(input.modelName)}`,
        "i",
      ),
    });
    await expect(deleteModelButton).toBeVisible();
    const deleteIcon = deleteModelButton.locator(".ai-model-delete-icon");
    await expect(deleteIcon).toHaveCount(1);
    await expect
      .poll(async () =>
        deleteIcon.evaluate((node) => {
          const style = window.getComputedStyle(node);
          return (
            style.getPropertyValue("mask-image") ||
            style.getPropertyValue("-webkit-mask-image") ||
            ""
          );
        }),
      )
      .not.toBe("none");
    await expect
      .poll(async () => (await deleteModelButton.textContent())?.trim() || "")
      .toBe("");
    await expect(row.first()).toContainText(input.openaiEndpointUrl);
    const openaiEndpointItem = row
      .first()
      .locator(".ai-provider-endpoint-item", {
        hasText: input.openaiEndpointUrl,
      });
    const openaiEndpointTag = openaiEndpointItem.locator(
      ".ai-provider-endpoint-badge",
      {
        hasText: "OpenAI",
      },
    );
    await expect(openaiEndpointTag).toBeVisible();
    if (input.anthropicEndpointUrl) {
      await expect(row.first()).toContainText("Anthropic");
      await expect(row.first()).toContainText(input.anthropicEndpointUrl);
      const anthropicEndpointItem = row
        .first()
        .locator(".ai-provider-endpoint-item", {
          hasText: input.anthropicEndpointUrl,
        });
      const anthropicEndpointTag = anthropicEndpointItem.locator(
        ".ai-provider-endpoint-badge",
        {
          hasText: "Anthropic",
        },
      );
      await expect(anthropicEndpointTag).toBeVisible();
      const [
        openaiUrlLeft,
        anthropicUrlLeft,
        openaiTextAlign,
        anthropicTextAlign,
      ] = await Promise.all([
        openaiEndpointItem
          .locator(".ai-provider-endpoint-url")
          .evaluate((node) => node.getBoundingClientRect().left),
        anthropicEndpointItem
          .locator(".ai-provider-endpoint-url")
          .evaluate((node) => node.getBoundingClientRect().left),
        openaiEndpointItem
          .locator(".ai-provider-endpoint-url")
          .evaluate((node) => window.getComputedStyle(node).textAlign),
        anthropicEndpointItem
          .locator(".ai-provider-endpoint-url")
          .evaluate((node) => window.getComputedStyle(node).textAlign),
      ]);
      expect(Math.abs(openaiUrlLeft - anthropicUrlLeft)).toBeLessThan(1);
      expect(openaiTextAlign).toBe("left");
      expect(anthropicTextAlign).toBe("left");
      await expect(
        openaiEndpointTag.locator(".ai-provider-endpoint-icon-mark"),
      ).toHaveAttribute("data-provider-icon", "openai");
      await expect(
        anthropicEndpointTag.locator(".ai-provider-endpoint-icon-mark"),
      ).toHaveAttribute("data-provider-icon", "anthropic");
      const [openaiBadgeStyle, anthropicBadgeStyle] = await Promise.all([
        openaiEndpointTag.evaluate((node) => {
          const style = window.getComputedStyle(node);
          return {
            backgroundColor: style.backgroundColor,
            borderColor: style.borderColor,
            color: style.color,
          };
        }),
        anthropicEndpointTag.evaluate((node) => {
          const style = window.getComputedStyle(node);
          return {
            backgroundColor: style.backgroundColor,
            borderColor: style.borderColor,
            color: style.color,
          };
        }),
      ]);
      expect(openaiBadgeStyle).not.toEqual(anthropicBadgeStyle);
      for (const [item, tag] of [
        [openaiEndpointItem, openaiEndpointTag],
        [anthropicEndpointItem, anthropicEndpointTag],
      ] as const) {
        const [itemBox, tagBox, position] = await Promise.all([
          item.evaluate((node) => {
            const box = node.getBoundingClientRect();
            return { right: box.right, top: box.top };
          }),
          tag.evaluate((node) => {
            const box = node.getBoundingClientRect();
            return { right: box.right, top: box.top };
          }),
          tag.evaluate((node) => window.getComputedStyle(node).position),
        ]);
        expect(position).toBe("absolute");
        expect(tagBox.right).toBeLessThanOrEqual(itemBox.right + 1);
        expect(tagBox.top).toBeGreaterThanOrEqual(itemBox.top - 1);
        expect(tagBox.top - itemBox.top).toBeLessThan(12);
      }
    }
    await expect(row.first()).toContainText(input.maskedApiKey);
  }

  async captureEvidence(name: string) {
    const dir = path.join(repoRoot, "temp");
    mkdirSync(dir, { recursive: true });
    const hasOpenDrawer = await this.dialog.isVisible().catch(() => false);
    const pathName = path.join(dir, screenshotName(name));
    if (hasOpenDrawer) {
      await this.dialog.screenshot({ path: pathName });
      return;
    }
    await this.resetHorizontalScroll();
    await this.page.screenshot({
      fullPage: true,
      path: pathName,
    });
    await this.resetHorizontalScroll();
  }

  async assertProviderRowEndpoint(
    providerName: string,
    baseUrl: string,
    protocolLabel: string,
  ) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
    await expect(row.first()).toContainText(protocolLabel);
    await expect(row.first()).toContainText(baseUrl);
  }

  async assertProviderRowSecret(providerName: string, secretText: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
    await expect(row.first()).toContainText(secretText);
  }

  async assertProviderVisible(providerName: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await expect(row.first()).toBeVisible();
  }

  async assertProviderSyncActions(input: {
    providerName: string;
    syncAnthropic?: boolean;
    syncOpenAI?: boolean;
  }) {
    const actionRow = await this.providerActionRow(input.providerName);
    await expect(
      actionRow.getByRole("button", { name: /端\s*点|Endpoints/i }),
    ).toHaveCount(0);
    await expect(
      actionRow.getByRole("button", { name: /编\s*辑|Edit/i }),
    ).toBeVisible();
    await expect(
      actionRow.getByRole("button", { name: /删\s*除|Delete/i }),
    ).toBeVisible();
    const openaiSync = actionRow.getByRole("button", {
      name: /同步 openai 模型|Sync openai Models/i,
    });
    const anthropicSync = actionRow.getByRole("button", {
      name: /同步 anthropic 模型|Sync anthropic Models/i,
    });
    if (input.syncOpenAI) {
      await expect(openaiSync).toBeVisible();
    }
    if (input.syncAnthropic) {
      await expect(anthropicSync).toBeVisible();
    }
    const actionList = actionRow.locator(".ai-provider-action-list");
    await expect(actionList).toBeVisible();
    await expect
      .poll(async () =>
        actionList.evaluate((node) => {
          const listBox = node.getBoundingClientRect();
          const cellBox = node.closest(".vxe-cell")?.getBoundingClientRect();
          return Boolean(cellBox && listBox.height <= cellBox.height + 1);
        }),
      )
      .toBe(true);
  }

  async cancelDrawer() {
    await closeDialogWithEscape(this.page, this.dialog, 2_000);
    if (await this.dialog.isHidden().catch(() => false)) {
      return;
    }
    await this.dialog
      .locator(".ant-drawer-close, .ant-modal-close")
      .first()
      .click({ force: true, timeout: 5_000 });
    await expect(this.dialog).toBeHidden({ timeout: 10_000 });
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openProvider(name: string) {
    const actionRow = await this.providerActionRow(name);
    await this.clickFixedActionButton(
      actionRow.getByRole("button", { name: /编\s*辑|Edit/i }),
    );
    await waitForDialogReady(this.dialog);
  }

  async fillProvider(data: {
    anthropicBaseUrl?: string;
    name?: string;
    openaiBaseUrl?: string;
    remark?: string;
    secretRef?: string;
    websiteUrl?: string;
  }) {
    if (data.name !== undefined) {
      await this.providerNameInput().fill(data.name);
    }
    if (data.websiteUrl !== undefined) {
      await this.dialog
        .getByRole("textbox", { name: /官网地址|Website/i })
        .fill(data.websiteUrl);
    }
    if (data.secretRef !== undefined) {
      await this.providerApiKeyInput().fill(data.secretRef);
    }
    if (data.openaiBaseUrl !== undefined) {
      await this.providerOpenAIBaseUrlInput().fill(data.openaiBaseUrl);
    }
    if (data.anthropicBaseUrl !== undefined) {
      await this.providerAnthropicBaseUrlInput().fill(data.anthropicBaseUrl);
    }
    if (data.remark !== undefined) {
      await this.dialog.getByLabel(/备注|Remark/i).fill(data.remark);
    }
  }

  async assertEditProviderMetadataForm(input?: {
    anthropicEndpointUrl?: string;
    openaiEndpointUrl?: string;
  }) {
    await expect(this.providerNameInput()).toBeVisible();
    await expect(this.dialog.getByText("端点配置")).toHaveCount(0);
    await expect(this.dialog.getByText("供应商名称")).toHaveCount(0);
    await expect(this.providerApiKeyInput()).toBeVisible();
    await expect(this.providerApiKeyInput()).toHaveValue("");
    await expect(this.providerApiKeyInput()).toHaveAttribute(
      "placeholder",
      /留空则保持原密钥|Leave blank to keep the existing secret/i,
    );
    expect(
      await this.providerApiKeyInput().getAttribute("placeholder"),
    ).not.toMatch(/端点|endpoint/i);
    await expect(this.providerOpenAIBaseUrlInput()).toBeVisible();
    await expect(this.providerAnthropicBaseUrlInput()).toBeVisible();
    await expect(this.dialog.getByText(/plugin\.linapro-ai-core/)).toHaveCount(
      0,
    );
    if (input?.openaiEndpointUrl) {
      await expect(this.providerOpenAIBaseUrlInput()).toHaveValue(
        input.openaiEndpointUrl,
      );
    }
    if (input?.anthropicEndpointUrl) {
      await expect(this.providerAnthropicBaseUrlInput()).toHaveValue(
        input.anthropicEndpointUrl,
      );
    }
  }

  async deleteModelFromProviderRow(providerName: string, modelName: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
    const deleteResponse = this.page.waitForResponse(
      (response) =>
        response.request().method() === "DELETE" &&
        /\/x\/linapro-ai-core\/api\/v1\/ai\/models\/\d+$/.test(response.url()),
      { timeout: 20_000 },
    );
    await row
      .first()
      .getByRole("button", {
        name: new RegExp(
          `删\\s*除.*${escapeRegExp(modelName)}|Delete.*${escapeRegExp(modelName)}`,
          "i",
        ),
      })
      .click();
    await this.page
      .locator(".ant-popover")
      .getByRole("button", { name: /确\s*定|OK/i })
      .click();
    const response = await deleteResponse;
    expect(response.ok()).toBeTruthy();
    await waitForBusyIndicatorsToClear(this.page);
    await this.searchProvider(providerName);
    await expect(row.first()).not.toContainText(modelName, { timeout: 10_000 });
  }

  async fillModel(data: { modelName: string; maxOutputTokens?: string }) {
    await this.dialog
      .getByRole("textbox", { name: /模型名称|Model Name/i })
      .fill(data.modelName);
    const maxOutput = this.dialog
      .getByLabel(/最大输出 Tokens|Max Output Tokens/i)
      .first();
    await maxOutput.fill(data.maxOutputTokens || "256");
  }

  async saveModel() {
    await this.dialog
      .getByRole("button", { name: /保\s*存|Save/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async confirmDrawer() {
    await this.dialog
      .getByRole("button", { name: /确\s*认|Confirm/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async searchProvider(name: string) {
    await this.page
      .getByLabel(/供应商|Provider/i)
      .first()
      .fill(name);
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async deleteProvider(name: string) {
    const actionRow = await this.providerActionRow(name);
    await actionRow.getByRole("button", { name: /删\s*除|Delete/i }).click();
    await this.page
      .locator(".ant-popover")
      .getByRole("button", { name: /确\s*定|OK/i })
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async configureTier(
    tierName: RegExp,
    providerName: string,
    modelName: string,
  ) {
    const rowIndex = await this.tierRowIndex(tierName);
    await this.page
      .getByRole("button", { name: /编\s*辑|Edit/i })
      .nth(rowIndex)
      .click();
    await waitForDialogReady(this.dialog);

    await this.dialog.getByLabel(/供应商|Provider/i).click();
    await this.page.getByTitle(providerName).click();
    await this.dialog.getByLabel(/模型|Model/i).click();
    await this.page.getByTitle(new RegExp(modelName)).click();
    await this.confirmDrawer();
  }

  async editTier(tierName: RegExp) {
    const rowIndex = await this.tierRowIndex(tierName);
    await this.page
      .getByRole("button", { name: /编\s*辑|Edit/i })
      .nth(rowIndex)
      .click();
    await waitForDialogReady(this.dialog);
    return this.dialog;
  }

  async clickSavedTierTestAndAssertLoading(tierName: RegExp) {
    const rowIndex = await this.tierRowIndex(tierName);
    const button = this.page
      .getByRole("button", { name: /测\s*试|Test/i })
      .nth(rowIndex);
    await button.click();
    await expect(button).toBeDisabled();
    await expect(
      button.locator(".ant-btn-loading-icon, .anticon-loading").first(),
    ).toBeVisible();
  }

  async clickDraftTierTestAndAssertLoading(tierName: RegExp) {
    const dialog = await this.editTier(tierName);
    const button = dialog.getByRole("button", {
      name: /测\s*试\s*当\s*前\s*配\s*置|Test Current Config/i,
    });
    await button.click();
    await expect(button).toBeDisabled();
    await expect(
      button.locator(".ant-btn-loading-icon, .anticon-loading").first(),
    ).toBeVisible();
  }

  async selectTierCapabilityType(capabilityType: string) {
    const tab = this.tierCapabilityTypeTab(capabilityType);
    await tab.click();
    await expect(tab).toHaveAttribute("aria-selected", "true");
    await waitForTableReady(this.page);
  }

  async assertTierCapabilityTypeTabs() {
    for (const [capabilityType, label] of Object.entries(
      tierCapabilityTypeLabels,
    )) {
      await expect
        .poll(async () => this.tierCapabilityTypeTabByLabel(label).count())
        .toBeGreaterThan(0);
      await expect(
        this.page.getByTestId(`ai-tier-capability-tab-icon-${capabilityType}`),
      ).toBeVisible();
    }
    await expect(
      this.page.getByLabel(/能力方法|Capability Method/i),
    ).toHaveCount(0);
    await expect(
      this.page.getByText(/plugin\.linapro-ai-core\.capability\.types/),
    ).toHaveCount(0);
    await expect(this.page.getByText("document.analyze")).toHaveCount(0);
  }

  async assertTierTabsVisualStyle() {
    const tabs = this.page.getByTestId("ai-tier-capability-tabs");
    await expect(tabs).not.toHaveClass(/ant-tabs-card/);
    const nav = tabs.locator(".ant-tabs-nav").first();
    const contentHolder = tabs.locator(".ant-tabs-content-holder").first();
    const content = this.page.getByTestId("ai-tier-capability-content");
    const firstTab = tabs.locator('[role="tab"]').first();
    const activeTab = tabs.locator(".ant-tabs-tab-active").first();
    const activeButton = activeTab.locator(".ant-tabs-tab-btn").first();
    const activeIcon = activeTab.locator(".tier-capability-tab-icon").first();
    const inactiveTab = tabs
      .locator(".ant-tabs-tab:not(.ant-tabs-tab-active)")
      .first();
    const inactiveButton = inactiveTab.locator(".ant-tabs-tab-btn").first();
    const inkBar = tabs.locator(".ant-tabs-ink-bar").first();
    await expect(nav).toBeVisible();
    await expect(contentHolder).toBeVisible();
    await expect(content).toBeVisible();
    await expect(activeTab).toBeVisible();
    await expect(inactiveTab).toBeVisible();
    await expect(inkBar).toBeVisible();

    const [
      activeBg,
      inactiveBg,
      activeColor,
      inactiveColor,
      activeIconColor,
      contentBg,
      contentBorderWidth,
      inkBg,
      inkBox,
      navDividerWidth,
      tabsBox,
      firstTabBox,
      navBox,
      contentHolderBox,
    ] = await Promise.all([
      activeTab.evaluate(
        (node) => window.getComputedStyle(node).backgroundColor,
      ),
      inactiveTab.evaluate(
        (node) => window.getComputedStyle(node).backgroundColor,
      ),
      activeButton.evaluate((node) => window.getComputedStyle(node).color),
      inactiveButton.evaluate((node) => window.getComputedStyle(node).color),
      activeIcon.evaluate((node) => window.getComputedStyle(node).color),
      contentHolder.evaluate(
        (node) => window.getComputedStyle(node).backgroundColor,
      ),
      contentHolder.evaluate((node) =>
        Number.parseFloat(window.getComputedStyle(node).borderTopWidth),
      ),
      inkBar.evaluate((node) => window.getComputedStyle(node).backgroundColor),
      inkBar.boundingBox(),
      nav.evaluate((node) =>
        Number.parseFloat(
          window.getComputedStyle(node, "::before").borderBottomWidth,
        ),
      ),
      tabs.boundingBox(),
      firstTab.boundingBox(),
      nav.boundingBox(),
      contentHolder.boundingBox(),
    ]);
    expect(activeBg).toBe("rgba(0, 0, 0, 0)");
    expect(inactiveBg).toBe("rgba(0, 0, 0, 0)");
    expect(activeColor).not.toBe(inactiveColor);
    expect(activeIconColor).toBe(activeColor);
    expect(contentBg).not.toBe("rgba(0, 0, 0, 0)");
    expect(contentBorderWidth).toBe(0);
    expect(inkBg).toBe(activeColor);
    expect(inkBox).not.toBeNull();
    expect(inkBox!.height).toBeGreaterThanOrEqual(2);
    expect(inkBox!.width).toBeGreaterThan(0);
    expect(navDividerWidth).toBeGreaterThanOrEqual(1);
    expect(tabsBox).not.toBeNull();
    expect(firstTabBox).not.toBeNull();
    expect(firstTabBox!.x - tabsBox!.x).toBeGreaterThanOrEqual(16);
    expect(navBox).not.toBeNull();
    expect(contentHolderBox).not.toBeNull();
    expect(
      Math.round(contentHolderBox!.y - (navBox!.y + navBox!.height)),
    ).toBeLessThanOrEqual(1);
  }

  async assertTierTypePage(capabilityType: string, defaultParams: string) {
    await expect(this.tierCapabilityTypeTab(capabilityType)).toHaveAttribute(
      "aria-selected",
      "true",
    );
    await expect(this.page.getByText(defaultParams)).toHaveCount(0);
    await expect(this.page.getByText(/基础|Basic/i)).toBeVisible();
    await expect(this.page.getByText(/标准|Standard/i)).toBeVisible();
    await expect(this.page.getByText(/高级|Advanced/i)).toBeVisible();
  }

  async assertTierUpdatedAtHidden(tierName: RegExp) {
    const headerIndex = await this.tierUpdatedAtColumnIndex();
    const row = this.page
      .locator(".vxe-table--main-wrapper .vxe-body--row:visible", {
        hasText: tierName,
      })
      .first();
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const updatedAtCell = row
      .locator(".vxe-body--column:visible")
      .nth(headerIndex);
    await expect
      .poll(async () => (await updatedAtCell.innerText()).trim())
      .toBe("");
  }

  async assertTierDrawerWithoutThinkingEffort(tierName: RegExp) {
    await this.editTier(tierName);
    await expect(
      this.dialog.getByText("供应商", { exact: true }),
    ).toBeVisible();
    await expect(this.dialog.getByText("模型", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("Thinking Effort")).toHaveCount(0);
    await this.cancelDrawer();
  }

  async assertTierDrawerDefaultConfig(
    tierName: RegExp,
    expectedParamsFragment: string,
  ) {
    await this.editTier(tierName);
    await expect(
      this.dialog.getByText("Thinking Effort", { exact: true }),
    ).toBeVisible();
    await expect(
      this.dialog.getByText(/模型默认|Model default/i).first(),
    ).toBeVisible();
    await expect(
      this.dialog.getByText(/默认参数 JSON|Default Params JSON/i),
    ).toBeVisible();
    const editorInput = this.dialog.getByTestId(
      "ai-tier-default-params-editor-input",
    );
    await expect(editorInput).toBeVisible();
    await expect(editorInput).toHaveValue(
      new RegExp(escapeRegExp(expectedParamsFragment)),
    );
    await expect(
      this.dialog
        .getByTestId("ai-tier-default-params-editor")
        .locator(".json-token-key")
        .first(),
    ).toBeVisible();
  }

  async assertTierDefaultParamsFormLayout() {
    const defaultParamsLabel = this.dialog
      .locator("label", {
        hasText: /默认参数 JSON|Default Params JSON/i,
      })
      .first();
    const providerLabel = this.dialog
      .locator("label", { hasText: /供应商|Provider/i })
      .first();
    const editor = this.dialog.getByTestId("ai-tier-default-params-editor");

    await expect(defaultParamsLabel).toBeVisible();
    await expect(providerLabel).toBeVisible();
    await expect(editor).toBeVisible();

    const [defaultParamsMetrics, providerMetrics, editorMetrics] =
      await Promise.all([
        defaultParamsLabel.evaluate((node) => {
          const formItem = node.closest(".relative.flex");
          const control = formItem?.querySelector(".flex-auto");
          const labelBox = node.getBoundingClientRect();
          const controlBox = control?.getBoundingClientRect();
          if (!controlBox) {
            throw new Error("default params form control was not found");
          }
          return {
            controlLeft: controlBox.left,
            controlRight: controlBox.right,
            labelLeft: labelBox.left,
          };
        }),
        providerLabel.evaluate((node) => {
          const formItem = node.closest(".relative.flex");
          const control = formItem?.querySelector(".flex-auto");
          const labelBox = node.getBoundingClientRect();
          const controlBox = control?.getBoundingClientRect();
          if (!controlBox) {
            throw new Error("provider form control was not found");
          }
          return {
            controlLeft: controlBox.left,
            controlRight: controlBox.right,
            labelLeft: labelBox.left,
          };
        }),
        editor.evaluate((node) => {
          const box = node.getBoundingClientRect();
          return {
            left: box.left,
            right: box.right,
          };
        }),
      ]);

    expect(
      Math.abs(defaultParamsMetrics.labelLeft - providerMetrics.labelLeft),
    ).toBeLessThan(1);
    expect(
      Math.abs(defaultParamsMetrics.controlLeft - providerMetrics.controlLeft),
    ).toBeLessThan(2);
    expect(
      Math.abs(
        defaultParamsMetrics.controlRight - providerMetrics.controlRight,
      ),
    ).toBeLessThan(2);
    expect(
      Math.abs(editorMetrics.left - defaultParamsMetrics.controlLeft),
    ).toBeLessThan(2);
    expect(
      Math.abs(editorMetrics.right - defaultParamsMetrics.controlRight),
    ).toBeLessThan(2);
  }

  async fillTierDefaultParams(defaultParamsJson: string) {
    const editorInput = this.dialog.getByTestId(
      "ai-tier-default-params-editor-input",
    );
    await editorInput.fill(defaultParamsJson);
    await expect(
      this.dialog
        .getByTestId("ai-tier-default-params-editor")
        .locator(".json-token-key")
        .first(),
    ).toBeVisible();
  }

  async saveTierDrawer() {
    await this.confirmDrawer();
  }

  async openInvocationDetail() {
    await this.page
      .getByRole("button", { name: /详\s*情|Detail/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async filterInvocationsByPurpose(purpose: string) {
    await this.page.getByLabel(/用途|Purpose/i).fill(purpose);
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForTableReady(this.page);
  }

  async filterInvocationsByCapabilityAndPurpose(
    capabilityKey: string,
    purpose: string,
  ) {
    const currentCapabilityKey = await this.openCapabilityMethodSelect();
    await this.selectCapabilityDropdownOption(
      capabilityKey,
      currentCapabilityKey,
    );
    await this.page.getByLabel(/用途|Purpose/i).fill(purpose);
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForTableReady(this.page);
  }

  private async openCapabilityMethodSelect() {
    const combobox = this.page
      .getByLabel(/能力方法|Capability Method/i)
      .first();
    const select = combobox.locator(
      'xpath=ancestor::*[contains(concat(" ", normalize-space(@class), " "), " ant-select ")][1]',
    );
    const currentCapabilityKey =
      (await select
        .locator(".ant-select-selection-item")
        .first()
        .textContent()) || "text.generate";
    await select.locator(".ant-select-selector").click();
    await this.visibleCapabilityDropdown().waitFor({
      state: "visible",
      timeout: 5_000,
    });
    return currentCapabilityKey.trim() || "text.generate";
  }

  private async selectCapabilityDropdownOption(
    label: string,
    currentLabel: string,
  ) {
    const targetIndex = capabilityMethodOptionOrder.indexOf(label);
    if (targetIndex < 0) {
      throw new Error(`未知能力方法选项: ${label}`);
    }

    const dropdown = this.visibleCapabilityDropdown();
    const visibleOptions = dropdown.locator(".ant-select-item-option:visible");
    await expect
      .poll(async () => visibleOptions.count(), { timeout: 5_000 })
      .toBeGreaterThan(0);

    if (await this.clickVisibleCapabilityOption(label)) {
      await this.expectCapabilityMethodSelected(label);
      return;
    }

    await this.ensureCapabilityDropdownOpen();
    const currentIndex = capabilityMethodOptionOrder.indexOf(currentLabel);
    const startIndex = currentIndex >= 0 ? currentIndex : 0;
    const key = targetIndex >= startIndex ? "ArrowDown" : "ArrowUp";
    const steps = Math.abs(targetIndex - startIndex);
    for (let index = 0; index < steps; index += 1) {
      await this.page.keyboard.press(key);
    }
    await this.page.keyboard.press("Enter");
    await this.expectCapabilityMethodSelected(label);
  }

  private visibleCapabilityDropdown() {
    return this.page
      .locator(".ant-select-dropdown:not(.ant-select-dropdown-hidden):visible")
      .last();
  }

  private async ensureCapabilityDropdownOpen() {
    if ((await this.visibleCapabilityDropdown().count()) > 0) {
      return;
    }
    const combobox = this.page
      .getByLabel(/能力方法|Capability Method/i)
      .first();
    const select = combobox.locator(
      'xpath=ancestor::*[contains(concat(" ", normalize-space(@class), " "), " ant-select ")][1]',
    );
    await select.locator(".ant-select-selector").click();
    await this.visibleCapabilityDropdown().waitFor({
      state: "visible",
      timeout: 5_000,
    });
  }

  private async clickVisibleCapabilityOption(label: string) {
    for (let attempt = 0; attempt < 3; attempt += 1) {
      const option = this.visibleCapabilityDropdown()
        .locator(".ant-select-item-option:visible")
        .filter({ hasText: label })
        .last();
      if ((await option.count()) === 0) {
        return false;
      }
      try {
        await expect(option).toBeVisible({ timeout: 1_000 });
        await option.click({ timeout: 2_000 });
        return true;
      } catch {
        if (attempt === 2) {
          return false;
        }
        await this.page.waitForTimeout(100);
      }
    }
    return false;
  }

  private async expectCapabilityMethodSelected(label: string) {
    await expect
      .poll(async () => {
        const combobox = this.page
          .getByLabel(/能力方法|Capability Method/i)
          .first();
        const select = combobox.locator(
          'xpath=ancestor::*[contains(concat(" ", normalize-space(@class), " "), " ant-select ")][1]',
        );
        return (
          (await select
            .locator(".ant-select-selection-item")
            .first()
            .textContent()) || ""
        ).trim();
      })
      .toBe(label);
  }

  private tierCapabilityTypeLabel(capabilityType: string) {
    const labels = tierCapabilityTypeLabels[capabilityType];
    if (!labels) {
      throw new Error(`未知能力类型 Tab: ${capabilityType}`);
    }
    return labels;
  }

  private tierCapabilityTypeTab(capabilityType: string) {
    return this.tierCapabilityTypeTabByLabel(
      this.tierCapabilityTypeLabel(capabilityType),
    ).first();
  }

  private tierCapabilityTypeTabByLabel(label: { en: string; zh: string }) {
    return this.page.getByRole("tab", {
      name: new RegExp(
        `${escapeRegExp(label.zh)}|${escapeRegExp(label.en)}`,
        "i",
      ),
    });
  }

  private async tierUpdatedAtColumnIndex() {
    const headers = this.page.locator(
      ".vxe-table--main-wrapper .vxe-header--column:visible",
    );
    const count = await headers.count();
    for (let index = 0; index < count; index += 1) {
      const text = (await headers.nth(index).innerText()).trim();
      if (/更新时间|Updated At/i.test(text)) {
        return index;
      }
    }
    throw new Error("未找到档位表更新时间列");
  }

  private async providerHeaderIndex(header: RegExp) {
    const headers = this.page.locator(
      ".vxe-table--main-wrapper .vxe-header--column:visible",
    );
    const count = await headers.count();
    for (let index = 0; index < count; index += 1) {
      const text = (await headers.nth(index).innerText()).trim();
      if (header.test(text)) {
        return index;
      }
    }
    throw new Error(`未找到供应商表列: ${header}`);
  }

  private async tierRowIndex(tierName: RegExp) {
    const rows = this.page.locator(".vxe-table--body .vxe-body--row:visible");
    const count = await rows.count();
    for (let index = 0; index < count; index += 1) {
      const text = await rows.nth(index).textContent();
      if (text && tierName.test(text)) {
        return index;
      }
    }
    throw new Error(`未找到档位行: ${tierName}`);
  }

  private async resetHorizontalScroll() {
    await this.page.evaluate(() => {
      window.scrollTo({ left: 0, top: window.scrollY });
      document.documentElement.scrollLeft = 0;
      document.body.scrollLeft = 0;
    });
  }

  private providerMainRow(providerName: string) {
    return this.page
      .locator(".vxe-table--main-wrapper .vxe-body--row:visible", {
        hasText: providerName,
      })
      .first();
  }

  private async providerActionRow(providerName: string) {
    const row = this.providerMainRow(providerName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const rowID = await row.getAttribute("rowid");
    expect(rowID, `未找到供应商行 rowid: ${providerName}`).toBeTruthy();
    const actionRow = this.page
      .locator(
        `.vxe-table--fixed-right-wrapper .vxe-body--row[rowid="${cssAttributeValue(rowID || "")}"]:visible`,
      )
      .first();
    await actionRow.waitFor({ state: "visible", timeout: 10_000 });
    return actionRow;
  }

  private async clickFixedActionButton(button: ReturnType<Page["locator"]>) {
    try {
      await button.click({ timeout: 2_000 });
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      if (!message.includes("intercepts pointer events")) {
        throw error;
      }
      await button.evaluate((node) => {
        if (!(node instanceof HTMLButtonElement)) {
          throw new Error("fixed action target is not a button");
        }
        node.click();
      });
    }
  }
}
