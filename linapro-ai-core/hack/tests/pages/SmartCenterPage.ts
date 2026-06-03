import { expect, type Page } from "@host-tests/support/playwright";

import { mkdirSync } from "node:fs";
import path from "node:path";

import { workspacePath } from "@host-tests/fixtures/config";
import {
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

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function cssAttributeValue(value: string) {
  return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

function screenshotName(name: string) {
  const timestamp = new Date().toISOString().replace(/\D/g, "").slice(0, 14);
  const safeName = name.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
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
      this.page.getByText("Thinking Effort", { exact: true }).first(),
    ).toBeVisible();
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
    await expect(this.dialog.getByText("供应商名称")).toHaveCount(0);
    await expect(this.dialog.getByText("OpenAI 接入地址")).toHaveCount(0);
    await expect(this.dialog.getByText("Anthropic 接入地址")).toHaveCount(0);
    await expect(this.dialog.getByText("OpenAI Base URL")).toHaveCount(0);
    await expect(this.dialog.getByText("Anthropic Base URL")).toHaveCount(0);
    await expect(this.dialog.getByText("API 密钥")).toHaveCount(0);
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
    const openaiEndpointTag = row.first().locator(".ant-tag", {
      hasText: "OpenAI",
    });
    await expect(openaiEndpointTag).toBeVisible();
    if (input.anthropicEndpointUrl) {
      await expect(row.first()).toContainText("Anthropic");
      await expect(row.first()).toContainText(input.anthropicEndpointUrl);
      const anthropicEndpointTag = row.first().locator(".ant-tag", {
        hasText: "Anthropic",
      });
      await expect(anthropicEndpointTag).toBeVisible();
      const [openaiWidth, anthropicWidth] = await Promise.all([
        openaiEndpointTag.evaluate(
          (node) => node.getBoundingClientRect().width,
        ),
        anthropicEndpointTag.evaluate(
          (node) => node.getBoundingClientRect().width,
        ),
      ]);
      expect(Math.abs(openaiWidth - anthropicWidth)).toBeLessThan(1);
      await expect
        .poll(() =>
          openaiEndpointTag.evaluate(
            (node) => window.getComputedStyle(node).justifyContent,
          ),
        )
        .toBe("flex-end");
      await expect
        .poll(() =>
          anthropicEndpointTag.evaluate(
            (node) => window.getComputedStyle(node).justifyContent,
          ),
        )
        .toBe("flex-end");
    }
    await expect(row.first()).toContainText(input.maskedApiKey);
  }

  async captureEvidence(name: string) {
    const dir = path.join(repoRoot, "temp");
    mkdirSync(dir, { recursive: true });
    await this.page.screenshot({
      fullPage: true,
      path: path.join(dir, screenshotName(name)),
    });
  }

  async openProviderEndpoints(providerName: string) {
    const actionRow = await this.providerActionRow(providerName);
    const endpointButton = actionRow.getByRole("button", {
      name: /端\s*点|Endpoints/i,
    });
    await expect(endpointButton).toBeVisible();
    await this.clickFixedActionButton(endpointButton);
    await waitForDialogReady(this.dialog);
  }

  async assertEndpointDrawerChineseTranslations(providerName: string) {
    await expect(
      this.dialog.getByText(`供应商端点 - ${providerName}`),
    ).toBeVisible();
    await expect(this.dialog.getByText("协议")).toBeVisible();
    await expect(this.dialog.getByText("基础地址")).toBeVisible();
    await expect(this.dialog.getByText("密钥", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("元数据 JSON")).toBeVisible();
    await expect(
      this.dialog.getByRole("button", { name: /新增端点|Add Endpoint/i }),
    ).toBeVisible();
    await expect(this.dialog.getByText(/plugin\.linapro-ai-core/)).toHaveCount(0);
  }

  async assertEndpointVisible(input: {
    baseUrl: string;
    protocolLabel: string;
    secretText?: string;
  }) {
    await expect(
      this.dialog.locator(".ant-tag", { hasText: input.protocolLabel }).first(),
    ).toBeVisible();
    await expect(this.dialog.getByText(input.baseUrl)).toBeVisible();
    if (input.secretText) {
      await expect(this.dialog.getByText(input.secretText)).toBeVisible();
    }
  }

  async assertProviderRowEndpoint(providerName: string, baseUrl: string, protocolLabel: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
    await expect(row.first()).toContainText(protocolLabel);
    await expect(row.first()).toContainText(baseUrl);
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
      this.page.getByRole("button", { name: /更多|More/i }),
    ).toHaveCount(0);
    const openaiSync = actionRow.getByRole("button", {
      name: /同步 openai 模型|Sync OpenAI Models/i,
    });
    const anthropicSync = actionRow.getByRole("button", {
      name: /同步 anthropic 模型|Sync Anthropic Models/i,
    });
    if (input.syncOpenAI) {
      await expect(openaiSync).toBeVisible();
    }
    if (input.syncAnthropic) {
      await expect(anthropicSync).toBeVisible();
    }
  }

  async cancelDrawer() {
    await this.dialog
      .getByRole("button", { name: /取\s*消|Cancel/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openProvider(name: string) {
    const actionRow = await this.providerActionRow(name);
    await actionRow
      .getByRole("button", { name: /编\s*辑|Edit/i })
      .click();
    await waitForDialogReady(this.dialog);
  }

  async fillProvider(data: { name: string }) {
    await this.providerNameInput().fill(data.name);
  }

  async assertEditProviderMetadataForm() {
    await expect(this.providerNameInput()).toBeVisible();
    await expect(this.dialog.getByText("供应商名称")).toHaveCount(0);
    await expect(this.dialog.getByLabel(/API 密钥|API Key/i)).toHaveCount(0);
    await expect(
      this.dialog.getByRole("textbox", {
        name: /OpenAI 接入地址|OpenAI Access URL/i,
      }),
    ).toHaveCount(0);
  }

  async deleteModelFromProviderRow(providerName: string, modelName: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await row.first().waitFor({ state: "visible", timeout: 10_000 });
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
    await waitForBusyIndicatorsToClear(this.page);
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
    await actionRow
      .getByRole("button", { name: /删\s*除|Delete/i })
      .click();
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

  async selectTierCapabilityMethod(capabilityKey: string) {
    const currentCapabilityKey = await this.openCapabilityMethodSelect();
    await this.selectCapabilityDropdownOption(
      capabilityKey,
      currentCapabilityKey,
    );
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForTableReady(this.page);
  }

  async assertTierMethodPage(capabilityKey: string, defaultParams: string) {
    await expect(this.page.getByText(`${capabilityKey} 默认参数`)).toBeVisible();
    await expect(this.page.getByText(defaultParams)).toBeVisible();
    await expect(this.page.getByText(/基础|Basic/i)).toBeVisible();
    await expect(this.page.getByText(/标准|Standard/i)).toBeVisible();
    await expect(this.page.getByText(/高级|Advanced/i)).toBeVisible();
  }

  async assertTierDrawerWithoutThinkingEffort(tierName: RegExp) {
    await this.editTier(tierName);
    await expect(this.dialog.getByText("供应商", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("模型", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("Thinking Effort")).toHaveCount(0);
    await this.cancelDrawer();
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

  async filterInvocationsByCapabilityAndPurpose(capabilityKey: string, purpose: string) {
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
      (await select.locator(".ant-select-selection-item").first().textContent()) ||
      "text.generate";
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
          (await select.locator(".ant-select-selection-item").first().textContent()) ||
          ""
        ).trim();
      })
      .toBe(label);
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
