import {
  expect,
  type Locator,
  type Page,
  type Response,
} from "@host-tests/support/playwright";

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
const legacyChineseProviderPattern = new RegExp("\u4f9b\u5e94\u5546");
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

function screenshotName(name: string, timestamp: string) {
  const safeName = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  return `${timestamp}-${safeName || "screenshot"}.png`;
}

function screenshotTimestamp() {
  return new Date().toISOString().replace(/\D/g, "").slice(0, 14);
}

function formatPickerDate(value: Date) {
  const year = value.getFullYear();
  const month = `${value.getMonth() + 1}`.padStart(2, "0");
  const day = `${value.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function todayToTomorrowPickerRange() {
  const start = new Date();
  const end = new Date(start);
  end.setDate(start.getDate() + 1);
  return {
    end: formatPickerDate(end),
    start: formatPickerDate(start),
  };
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

  async gotoModels() {
    await this.page.goto(workspacePath("/ai/models"));
    await waitForTableReady(this.page);
  }

  async assertProviderPageWithoutTabs() {
    const page = this.providerPage();
    await expect(page).toBeVisible();
    await expect(
      this.page.getByTestId("ai-provider-management-tabs"),
    ).toHaveCount(0);
    await expect(page.getByRole("tab")).toHaveCount(0);
    await expect(page.getByText("模型管理", { exact: true })).toHaveCount(0);
    await expect(
      page.getByText("Model Management", { exact: true }),
    ).toHaveCount(0);
    await expect(
      page.getByText(/plugin\.linapro-ai-core\.provider\.tabs/),
    ).toHaveCount(0);
    await this.assertProviderPageHeightStable();
  }

  async openProviderManagementTab() {
    await this.gotoProviders();
  }

  async openModelManagementTab() {
    await this.gotoModels();
    await expect(
      this.page.getByTestId("ai-model-management-page"),
    ).toBeVisible();
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

  async assertInvocationMethodLabelSingleLine() {
    const label = this.page
      .locator("label", { hasText: /Invocation Method/i })
      .first();
    await expect(label).toBeVisible();
    const metrics = await label.evaluate((node) => {
      const styles = window.getComputedStyle(node);
      const lineHeight = Number.parseFloat(styles.lineHeight);
      const box = node.getBoundingClientRect();
      return {
        height: box.height,
        lineHeight,
      };
    });
    expect(metrics.lineHeight).toBeGreaterThan(0);
    expect(metrics.height).toBeLessThanOrEqual(metrics.lineHeight * 1.35);
  }

  async openCreateProvider() {
    await this.page
      .getByRole("button", { name: /新\s*增\s*渠\s*道|Add Provider/i })
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
    await expect(this.dialog.getByText("新增渠道")).toBeVisible();
    await expect(this.providerNameInput()).toBeVisible();
    await expect(this.dialog.getByText("端点配置")).toHaveCount(0);
    await expect(this.dialog.getByText("渠道名称")).toHaveCount(0);
    await expect(
      this.dialog.getByText(legacyChineseProviderPattern),
    ).toHaveCount(0);
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
    await expect(this.providerAnthropicBaseUrlInput()).toHaveValue("");
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
    await expect(this.dialog.getByText("渠道")).toBeVisible();
    await expect(this.dialog.getByText("模型名称")).toBeVisible();
    await this.assertModelDrawerLabelsSingleLine(["渠道", "协议", "模型名称"]);
    await expect(
      this.dialog.getByText(/能力方法|Capability Method/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/支持 Thinking|Supports Thinking/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(
        /支持的 Thinking Effort|Supported Thinking Efforts/i,
      ),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/最大输入 Tokens|Max Input Tokens/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/最大输出 Tokens|Max Output Tokens/i),
    ).toHaveCount(0);
    await expect(this.dialog.getByText(/端点\s*\/\s*协议/)).toHaveCount(0);
    await expect(
      this.dialog.getByText(/Endpoint\s*\/\s*Protocol/i),
    ).toHaveCount(0);
    await this.captureEvidence("TC001-create-model-drawer-default");
    await this.dialog.getByLabel(/渠道|Provider/i).click();
    await this.page.getByTitle(providerName).click();
    await expect(this.dialog.getByTitle(providerName)).toBeVisible();
    await this.assertModelProtocolOptions();
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
      this.page
        .getByRole("button", { name: /新\s*增\s*模\s*型|Add Model/i })
        .first(),
    ).toBeVisible();
    await expect(
      this.page.getByRole("button", {
        name: /新\s*增\s*渠\s*道|Add Provider/i,
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
    const modelCell = row
      .first()
      .locator(".vxe-body--column:visible")
      .nth(modelHeaderIndex);
    const endpointCell = row
      .first()
      .locator(".vxe-body--column:visible")
      .nth(endpointHeaderIndex);
    const modelRows = modelCell.locator(".ai-provider-model-row");
    await expect(modelRows).toHaveCount(1);
    const modelRow = modelRows.first();
    await expect(modelRow).toBeVisible();
    const modelText = modelCell.locator(".ai-provider-model-name").first();
    const modelTag = modelCell.locator(".ai-provider-model-tag").first();
    await expect(modelText).toHaveText(input.modelName);
    await expect(modelCell).not.toContainText("OpenAI");
    await expect(modelCell).not.toContainText("Anthropic");
    await expect
      .poll(async () => {
        const weight = await modelText.evaluate((node) =>
          Number.parseInt(window.getComputedStyle(node).fontWeight, 10),
        );
        return Number.isNaN(weight) ? 400 : weight;
      })
      .toBeLessThan(600);
    const [
      modelCellBox,
      endpointCellBox,
      modelListLayout,
      modelRowLayout,
      modelTagStyle,
      modelTextStyle,
    ] = await Promise.all([
      modelCell.evaluate((node) => {
        const box = node.getBoundingClientRect();
        return { right: box.right, width: box.width };
      }),
      endpointCell.evaluate((node) => {
        const box = node.getBoundingClientRect();
        return { left: box.left, right: box.right, width: box.width };
      }),
      modelCell.locator(".ai-provider-model-list").evaluate((node) => {
        const style = window.getComputedStyle(node);
        return {
          alignItems: style.alignItems,
          flexDirection: style.flexDirection,
        };
      }),
      modelRow.evaluate((node) => {
        const box = node.getBoundingClientRect();
        const style = window.getComputedStyle(node);
        return {
          columnGap: Number.parseFloat(style.columnGap),
          flexWrap: style.flexWrap,
          overflowX: style.overflowX,
          right: box.right,
          rowGap: Number.parseFloat(style.rowGap),
        };
      }),
      modelTag.evaluate((node) => {
        const style = window.getComputedStyle(node);
        return {
          display: style.display,
          fontSize: Number.parseFloat(style.fontSize),
          paddingLeft: Number.parseFloat(style.paddingLeft),
        };
      }),
      modelText.evaluate((node) => {
        const style = window.getComputedStyle(node);
        return {
          fontSize: Number.parseFloat(style.fontSize),
          overflow: style.overflow,
          textOverflow: style.textOverflow,
          whiteSpace: style.whiteSpace,
          wordBreak: style.wordBreak,
        };
      }),
    ]);
    expect(modelListLayout.flexDirection).toBe("column");
    expect(modelListLayout.alignItems).toBe("flex-start");
    expect(modelRowLayout.flexWrap).toBe("wrap");
    expect(modelRowLayout.columnGap).toBeLessThanOrEqual(6);
    expect(modelRowLayout.rowGap).toBeLessThanOrEqual(6);
    expect(["auto", "scroll"]).not.toContain(modelRowLayout.overflowX);
    expect(["flex", "inline-flex"]).toContain(modelTagStyle.display);
    expect(modelTagStyle.fontSize).toBeLessThanOrEqual(12);
    expect(modelTagStyle.paddingLeft).toBeLessThanOrEqual(8);
    expect(modelTextStyle.fontSize).toBeLessThanOrEqual(12);
    expect(modelTextStyle.whiteSpace).toBe("normal");
    expect(modelTextStyle.overflow).not.toBe("hidden");
    expect(modelTextStyle.textOverflow).not.toBe("ellipsis");
    expect(modelTextStyle.wordBreak).toBe("break-all");
    expect(modelRowLayout.right).toBeLessThanOrEqual(modelCellBox.right + 1);
    expect(modelCellBox.right).toBeLessThanOrEqual(endpointCellBox.left + 1);
    expect(modelCellBox.width).toBeGreaterThanOrEqual(400);
    expect(endpointCellBox.width).toBeGreaterThanOrEqual(400);
    await this.assertProviderModelListFullyVisible(modelCell);
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
    await this.assertEndpointBadgeLayout(
      openaiEndpointItem,
      openaiEndpointTag,
      input.openaiEndpointUrl,
    );
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
      await this.assertEndpointBadgeLayout(
        anthropicEndpointItem,
        anthropicEndpointTag,
        input.anthropicEndpointUrl,
      );
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

  private async assertProviderModelListFullyVisible(
    modelCell: ReturnType<Page["locator"]>,
  ) {
    await expect(modelCell.locator(".ai-provider-model-toggle")).toHaveCount(0);
    await expect(
      modelCell.locator(".ai-provider-model-list-content"),
    ).toHaveCount(0);
    const modelList = modelCell.locator(".ai-provider-model-list");
    const tags = modelCell.locator(".ai-provider-model-tag");
    await expect.poll(async () => tags.count()).toBeGreaterThanOrEqual(10);
    await expect(tags.first()).toBeVisible();
    const listMetrics = await modelList.evaluate((node) => {
      const box = node.getBoundingClientRect();
      const style = window.getComputedStyle(node);
      return {
        height: box.height,
        maxHeight: style.maxHeight,
        overflow: style.overflow,
      };
    });
    expect(listMetrics.height).toBeGreaterThan(104);
    expect(listMetrics.maxHeight).toBe("none");
    expect(listMetrics.overflow).toBe("visible");
    const hiddenTagCount = await tags.evaluateAll(
      (nodes) =>
        nodes.filter((node) => {
          const box = node.getBoundingClientRect();
          return box.width === 0 || box.height === 0;
        }).length,
    );
    expect(hiddenTagCount).toBe(0);
  }

  async captureEvidence(name: string) {
    const timestamp = screenshotTimestamp();
    const dir = path.join(repoRoot, "temp", timestamp.slice(0, 8));
    mkdirSync(dir, { recursive: true });
    await waitForBusyIndicatorsToClear(this.page);
    await this.page
      .locator(".ant-message-notice:visible")
      .first()
      .waitFor({ state: "hidden", timeout: 3000 })
      .catch(() => {});
    const hasOpenDrawer = await this.dialog.isVisible().catch(() => false);
    const pathName = path.join(dir, screenshotName(name, timestamp));
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

  async createModelForProviderProtocols(input: {
    modelName: string;
    providerName: string;
    protocolLabels: RegExp[];
  }) {
    await this.openCreateModel();
    await this.dialog.getByLabel(/渠道|Provider/i).click();
    await this.page.getByTitle(input.providerName).click();
    await expect(this.dialog.getByTitle(input.providerName)).toBeVisible();
    await this.fillModel({ modelName: input.modelName });
    await this.selectModelProtocols(input.protocolLabels);

    const responses: Response[] = [];
    const onResponse = (response: Response) => {
      const request = response.request();
      if (
        request.method() === "POST" &&
        /\/x\/linapro-ai-core\/api\/v1\/ai\/providers\/\d+\/models$/.test(
          response.url(),
        )
      ) {
        responses.push(response);
      }
    };
    this.page.on("response", onResponse);
    try {
      await this.saveModel();
      await expect
        .poll(() => responses.length, {
          message: "等待多协议模型创建请求完成",
          timeout: 20_000,
        })
        .toBe(input.protocolLabels.length);
      await expect(this.dialog).toBeHidden({ timeout: 20_000 });
      await waitForBusyIndicatorsToClear(this.page);
    } finally {
      this.page.off("response", onResponse);
    }
    for (const response of responses) {
      expect(response.ok(), `创建模型响应状态: ${response.status()}`).toBe(
        true,
      );
    }

    await this.searchProvider(input.providerName);
    await this.assertProviderModelNameAggregated(input);
    await this.captureEvidence("TC001-provider-list-multi-protocol-model");
  }

  async assertProviderModelNameAggregated(input: {
    modelName: string;
    providerName: string;
    protocolLabels: RegExp[];
  }) {
    const row = this.providerMainRow(input.providerName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const modelHeaderIndex = await this.providerHeaderIndex(/模型|Models/i);
    const modelCell = row
      .locator(".vxe-body--column:visible")
      .nth(modelHeaderIndex);
    const modelTags = modelCell.locator(".ai-provider-model-tag", {
      hasText: input.modelName,
    });
    await expect(modelCell.locator(".ai-provider-model-toggle")).toHaveCount(0);
    await expect(modelTags).toHaveCount(1);
    await expect(
      modelTags.first().locator(".ai-provider-model-name"),
    ).toHaveText(input.modelName);
    for (const label of input.protocolLabels) {
      await expect(modelCell.getByText(label)).toHaveCount(0);
    }
    const maxTagsOnOneLine = await modelCell
      .locator(".ai-provider-model-tag")
      .evaluateAll((nodes) => {
        const groups = new Map<number, number>();
        for (const node of nodes) {
          const top = Math.round(node.getBoundingClientRect().top);
          groups.set(top, (groups.get(top) || 0) + 1);
        }
        return Math.max(...groups.values());
      });
    expect(maxTagsOnOneLine).toBeGreaterThanOrEqual(2);
  }

  async assertProviderSyncActions(input: { providerName: string }) {
    const actionRow = await this.providerActionRow(input.providerName);
    const primaryActions = actionRow.locator(".ai-provider-action-primary");
    const actionRows = actionRow.locator(".ai-provider-action-row");
    await expect(primaryActions).toBeVisible();
    await expect(
      actionRow.getByRole("button", { name: /端\s*点|Endpoints/i }),
    ).toHaveCount(0);
    await expect(
      actionRow.getByRole("button", {
        name: /同步 OpenAI 模型|Sync OpenAI Models/i,
      }),
    ).toHaveCount(0);
    await expect(
      actionRow.getByRole("button", {
        name: /同步 Anthropic 模型|Sync Anthropic Models/i,
      }),
    ).toHaveCount(0);
    const editButton = primaryActions.getByRole("button", {
      name: /编\s*辑|Edit/i,
    });
    const deleteButton = primaryActions.getByRole("button", {
      name: /删\s*除|Delete/i,
    });
    await expect(editButton).toBeVisible();
    await expect(deleteButton).toBeVisible();
    await expect(actionRows).toHaveCount(2);
    const addModelButton = actionRows
      .nth(0)
      .getByRole("button", { name: /新\s*增\s*模\s*型|Add Model/i });
    const syncButton = actionRows
      .nth(1)
      .getByRole("button", { name: /同步模型|Sync Models/i });
    await expect(addModelButton).toBeVisible();
    await expect(syncButton).toBeVisible();
    const actionList = actionRow.locator(".ai-provider-action-list");
    await expect(actionList).toBeVisible();
    const [editBox, deleteBox, primaryBox, addBox, syncBox, actionListLayout] =
      await Promise.all([
        editButton.evaluate((node) => {
          const box = node.getBoundingClientRect();
          return { left: box.left, top: box.top };
        }),
        deleteButton.evaluate((node) => {
          const box = node.getBoundingClientRect();
          return { left: box.left, top: box.top };
        }),
        primaryActions.evaluate((node) => {
          const box = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            bottom: box.bottom,
            centerX: box.left + box.width / 2,
            justifyContent: style.justifyContent,
          };
        }),
        addModelButton.evaluate((node) => {
          const box = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            backgroundColor: style.backgroundColor,
            borderColor: style.borderColor,
            centerX: box.left + box.width / 2,
            color: style.color,
            top: box.top,
          };
        }),
        syncButton.evaluate((node) => {
          const box = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            backgroundColor: style.backgroundColor,
            borderColor: style.borderColor,
            centerX: box.left + box.width / 2,
            color: style.color,
            top: box.top,
          };
        }),
        actionList.evaluate((node) => {
          const listBox = node.getBoundingClientRect();
          const cellBox = node.closest(".vxe-cell")?.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            alignItems: style.alignItems,
            centerDelta: cellBox
              ? Math.abs(
                  listBox.left +
                    listBox.width / 2 -
                    (cellBox.left + cellBox.width / 2),
                )
              : Number.POSITIVE_INFINITY,
            rowGap: Number.parseFloat(style.rowGap),
          };
        }),
      ]);
    expect(deleteBox.left).toBeGreaterThan(editBox.left);
    expect(Math.abs(deleteBox.top - editBox.top)).toBeLessThan(2);
    expect(actionListLayout.alignItems).toBe("center");
    expect(actionListLayout.centerDelta).toBeLessThanOrEqual(2);
    expect(actionListLayout.rowGap).toBeGreaterThanOrEqual(8);
    expect(primaryBox.justifyContent).toBe("center");
    expect(addBox.top).toBeGreaterThanOrEqual(primaryBox.bottom - 1);
    expect(syncBox.top).toBeGreaterThan(addBox.top);
    expect(Math.abs(addBox.centerX - syncBox.centerX)).toBeLessThan(2);
    expect(Math.abs(primaryBox.centerX - addBox.centerX)).toBeLessThan(2);
    expect(syncBox.color).toBe(addBox.color);
    expect(syncBox.borderColor).toBe(addBox.borderColor);
    expect(syncBox.backgroundColor).toBe(addBox.backgroundColor);
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

  async assertProviderRowAddModelDefaults(providerName: string) {
    const actionRow = await this.providerActionRow(providerName);
    await actionRow
      .getByRole("button", { name: /新\s*增\s*模\s*型|Add Model/i })
      .click();
    await waitForDialogReady(this.dialog);
    await expect(
      this.dialog.getByText(/新\s*增\s*模\s*型|Add Model/i).first(),
    ).toBeVisible();
    const providerField = this.dialog
      .locator(".relative.flex", { hasText: /渠道|Provider/i })
      .first();
    await expect(
      providerField.locator(".ant-select-selection-item").first(),
    ).toHaveText(providerName);
    await this.captureEvidence("TC001-provider-row-add-model-default");
    await this.cancelDrawer();
  }

  async assertProviderPageHeightStable() {
    const samples: Array<{ body: number; tabs: number }> = [];
    for (let index = 0; index < 6; index += 1) {
      await this.page.waitForTimeout(200);
      samples.push(
        await this.page.evaluate(() => {
          const page = document.querySelector(
            '[data-testid="ai-provider-management-page"]',
          );
          return {
            body: document.documentElement.scrollHeight,
            tabs: Math.round(page?.getBoundingClientRect().height || 0),
          };
        }),
      );
    }
    const bodyHeights = samples.map((item) => item.body);
    const tabHeights = samples.map((item) => item.tabs);
    const bodyGrowth = bodyHeights.at(-1)! - bodyHeights[0]!;
    const tabGrowth = tabHeights.at(-1)! - tabHeights[0]!;
    expect(bodyGrowth).toBeLessThanOrEqual(8);
    expect(tabGrowth).toBeLessThanOrEqual(8);
    expect(
      Math.max(...bodyHeights) - Math.min(...bodyHeights),
    ).toBeLessThanOrEqual(16);
    expect(
      Math.max(...tabHeights) - Math.min(...tabHeights),
    ).toBeLessThanOrEqual(16);
  }

  async assertProviderSearchFormDefaultSpacing() {
    const metrics = await this.page.evaluate(() => {
      const content = document.querySelector(
        '[data-testid="ai-provider-management-page"]',
      );
      const header = content?.querySelector(".vxe-grid--layout-header-wrapper");
      const toolbar = content?.querySelector(".vxe-toolbar");
      if (!header || !toolbar) {
        return null;
      }
      const headerBox = header.getBoundingClientRect();
      const separator = Array.from(header.querySelectorAll("div")).find(
        (node) => {
          const box = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return (
            style.position === "absolute" &&
            box.height > 0 &&
            box.width >= headerBox.width - 4 &&
            Math.abs(box.bottom - headerBox.bottom) <= 12
          );
        },
      );
      if (!separator) {
        return null;
      }
      const searchControls = Array.from(
        header.querySelectorAll<HTMLElement>(
          ".ant-input, .ant-select-selector, .ant-picker, input",
        ),
      ).filter((node) => {
        const box = node.getBoundingClientRect();
        return (
          box.width > 0 &&
          box.height > 0 &&
          box.top >= headerBox.top - 1 &&
          box.bottom <= headerBox.bottom + 1
        );
      });
      if (searchControls.length === 0) {
        return null;
      }
      const separatorBox = separator.getBoundingClientRect();
      const toolbarBox = toolbar.getBoundingClientRect();
      const inputTop = Math.min(
        ...searchControls.map((node) => node.getBoundingClientRect().top),
      );
      return {
        headerHeight: headerBox.height,
        inputTopGap: inputTop - headerBox.top,
        separatorBottomDelta: Math.abs(separatorBox.bottom - headerBox.bottom),
        separatorHeight: separatorBox.height,
        toolbarTopGap: toolbarBox.top - separatorBox.bottom,
      };
    });

    expect(metrics, "渠道搜索表单应使用共享 VXE 默认分隔样式").not.toBeNull();
    if (!metrics) {
      return;
    }
    expect(metrics.headerHeight).toBeGreaterThanOrEqual(68);
    expect(metrics.inputTopGap).toBeGreaterThanOrEqual(10);
    expect(metrics.separatorHeight).toBeGreaterThanOrEqual(8);
    expect(metrics.separatorBottomDelta).toBeGreaterThanOrEqual(4);
    expect(metrics.toolbarTopGap).toBeGreaterThanOrEqual(6);
  }

  async assertProviderIdentityModel(providerName: string, modelName: string) {
    const row = this.page.locator(".vxe-body--row:visible", {
      hasText: providerName,
    });
    await expect(row.first()).toBeVisible();
    await expect(row.first()).toContainText(modelName);
    await expect(row.first()).not.toContainText(/未声明能力|Unclassified/i);
    await expect(
      row.first().getByText(/plugin\.linapro-ai-core\.capability/),
    ).toHaveCount(0);
  }

  async assertModelManagementProjection(input: {
    endpointUrl: string;
    modelName: string;
    protocolLabel: RegExp;
    providerName: string;
  }) {
    await this.openModelManagementTab();
    await expect(
      this.modelPage().getByRole("button", {
        name: /新\s*增\s*模\s*型|Add Model/i,
      }),
    ).toBeVisible();
    await this.searchModel(input.modelName);
    const row = this.modelMainRow(input.modelName);
    await expect(row).toBeVisible();
    await expect(row).toContainText(input.providerName);
    await expect(row).toContainText(input.protocolLabel);
    await expect(row).toContainText(input.endpointUrl);
    const endpointHeaderIndex = await this.modelHeaderIndex(/端点|Endpoint/i);
    const endpointCell = row
      .locator(".vxe-body--column:visible")
      .nth(endpointHeaderIndex);
    await expect
      .poll(async () =>
        endpointCell.evaluate((node) => node.getBoundingClientRect().width),
      )
      .toBeGreaterThanOrEqual(460);
    await expect(this.page.getByText(/plugin\.linapro-ai-core/)).toHaveCount(0);
  }

  async assertModelManagementHidesCapabilityControls(modelName: string) {
    await this.openModelManagementTab();
    await this.searchModel(modelName);
    const row = this.modelMainRow(modelName);
    await expect(row).toBeVisible();
    await expect(row).not.toContainText(/text\.generate|document\.analyze/i);
    await expect(row).not.toContainText(/未声明能力|Unclassified/i);
    const actionRow = await this.modelActionRow(modelName);
    await actionRow.getByRole("button", { name: /编\s*辑|Edit/i }).click();
    await waitForDialogReady(this.dialog);
    await expect(
      this.dialog.getByText(/能力方法|Capability Method/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/支持 Thinking|Supports Thinking/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(
        /支持的 Thinking Effort|Supported Thinking Efforts/i,
      ),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/最大输入 Tokens|Max Input Tokens/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByText(/最大输出 Tokens|Max Output Tokens/i),
    ).toHaveCount(0);
    await this.cancelDrawer();
  }

  async assertModelSearchFormFiltersInSingleRow() {
    await this.openModelManagementTab();
    const page = this.modelPage();
    await expect(
      page.locator("label", { hasText: /模型名称|Model Name/i }).first(),
    ).toBeVisible();
    await expect(
      page.locator("label", { hasText: /渠道|Provider/i }).first(),
    ).toBeVisible();
    await expect(
      page.locator("label", { hasText: /能力方法|Capability Method/i }).first(),
    ).toHaveCount(0);
    await expect(
      page.locator("label", { hasText: /状态|Status/i }).first(),
    ).toBeVisible();
    await expect(page.getByText(/plugin\.linapro-ai-core/)).toHaveCount(0);

    const metrics = await page.evaluate((root) => {
      function fieldCenter(pattern: RegExp) {
        const label = Array.from(root.querySelectorAll("label")).find((node) =>
          pattern.test(node.textContent || ""),
        );
        if (!label) {
          return null;
        }
        const field =
          label.closest(".relative.flex") ||
          label.closest(".ant-form-item") ||
          label;
        const box = field.getBoundingClientRect();
        return box.top + box.height / 2;
      }

      const centers = [
        fieldCenter(/模型名称|Model Name/i),
        fieldCenter(/渠道|Provider/i),
        fieldCenter(/状态|Status/i),
      ].filter((value): value is number => typeof value === "number");

      return {
        count: centers.length,
        maxCenter: Math.max(...centers),
        minCenter: Math.min(...centers),
      };
    });

    expect(metrics.count).toBe(3);
    expect(metrics.maxCenter - metrics.minCenter).toBeLessThanOrEqual(12);
  }

  async filterModelsByProviderOnly(input: {
    expectedModelNames: string[];
    hiddenModelNames: string[];
    providerId: number;
    providerName: string;
  }) {
    await this.openModelManagementTab();
    await this.assertModelSearchFormFiltersInSingleRow();
    await this.modelPage().getByRole("textbox").first().fill("e2e-filter");
    await this.selectModelProviderFilter(input.providerName);
    const responsePromise = this.page.waitForResponse((response) => {
      if (
        response.request().method() !== "GET" ||
        !/\/x\/linapro-ai-core\/api\/v1\/ai\/models/.test(response.url())
      ) {
        return false;
      }
      const url = new URL(response.url());
      return (
        url.searchParams.get("providerId") === String(input.providerId) &&
        !url.searchParams.has("capabilityType") &&
        !url.searchParams.has("capabilityMethod")
      );
    });
    await this.modelPage()
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    const response = await responsePromise;
    expect(response.ok(), `模型筛选响应状态: ${response.status()}`).toBe(true);
    await waitForBusyIndicatorsToClear(this.page);
    for (const modelName of input.expectedModelNames) {
      await expect(this.modelMainRow(modelName)).toBeVisible();
    }
    for (const modelName of input.hiddenModelNames) {
      await expect(this.modelMainRow(modelName)).toBeHidden({
        timeout: 10_000,
      });
    }
    await this.captureEvidence("TC001-model-management-filter-row");
  }

  async renameModelFromModelManagement(input: {
    modelName: string;
    nextModelName: string;
  }) {
    await this.openModelManagementTab();
    await this.searchModel(input.modelName);
    const actionRow = await this.modelActionRow(input.modelName);
    await actionRow.getByRole("button", { name: /编\s*辑|Edit/i }).click();
    await waitForDialogReady(this.dialog);
    await expect(
      this.dialog.getByText(/编\s*辑\s*模\s*型|Edit Model/i).first(),
    ).toBeVisible();
    await this.fillModel({ modelName: input.nextModelName });
    const updateResponse = this.page.waitForResponse(
      (response) =>
        response.request().method() === "PUT" &&
        /\/x\/linapro-ai-core\/api\/v1\/ai\/models\/\d+$/.test(response.url()),
      { timeout: 20_000 },
    );
    await this.saveModel();
    const response = await updateResponse;
    expect(response.ok(), `更新模型响应状态: ${response.status()}`).toBe(true);
    await expect(this.dialog).toBeHidden({ timeout: 20_000 });
    await waitForBusyIndicatorsToClear(this.page);
    await this.searchModel(input.nextModelName);
    await expect(this.modelMainRow(input.nextModelName)).toBeVisible();
  }

  async deleteModelFromModelManagement(modelName: string) {
    await this.openModelManagementTab();
    await this.searchModel(modelName);
    const actionRow = await this.modelActionRow(modelName);
    const deleteResponse = this.page.waitForResponse(
      (response) =>
        response.request().method() === "DELETE" &&
        /\/x\/linapro-ai-core\/api\/v1\/ai\/models\/\d+$/.test(response.url()),
      { timeout: 20_000 },
    );
    await actionRow.getByRole("button", { name: /删\s*除|Delete/i }).click();
    await this.page
      .locator(".ant-popover")
      .getByRole("button", { name: /确\s*定|OK/i })
      .click();
    const response = await deleteResponse;
    expect(response.ok(), `删除模型响应状态: ${response.status()}`).toBe(true);
    await waitForBusyIndicatorsToClear(this.page);
    await this.searchModel(modelName);
    await expect(this.modelMainRow(modelName)).toBeHidden({ timeout: 10_000 });
  }

  private async assertEndpointBadgeLayout(
    endpointItem: ReturnType<Page["locator"]>,
    endpointTag: ReturnType<Page["locator"]>,
    expectedUrl: string,
  ) {
    const endpointUrl = endpointItem.locator(".ai-provider-endpoint-url");
    const iconMark = endpointTag.locator(".ai-provider-endpoint-icon-mark");
    await expect(endpointUrl).toBeVisible();
    await expect(endpointUrl).toHaveText(expectedUrl);
    await expect(iconMark).toBeVisible();
    await expect(iconMark.locator("svg").first()).toBeVisible();
    await expect
      .poll(async () => (await iconMark.textContent())?.trim() || "")
      .toBe("");
    const [urlStyle, tagStyle] = await Promise.all([
      endpointUrl.evaluate((node) => {
        const style = window.getComputedStyle(node);
        return {
          fontSize: Number.parseFloat(style.fontSize),
          overflow: style.overflow,
          overflowX: style.overflowX,
          textOverflow: style.textOverflow,
          whiteSpace: style.whiteSpace,
          wordBreak: style.wordBreak,
        };
      }),
      endpointTag.evaluate((node) => {
        const style = window.getComputedStyle(node);
        return {
          fontSize: Number.parseFloat(style.fontSize),
        };
      }),
    ]);
    expect(urlStyle.whiteSpace).toBe("normal");
    expect(urlStyle.overflow).not.toBe("hidden");
    expect(urlStyle.overflowX).not.toBe("hidden");
    expect(urlStyle.textOverflow).not.toBe("ellipsis");
    expect(urlStyle.wordBreak).toBe("break-all");
    expect(tagStyle.fontSize).toBeLessThan(urlStyle.fontSize);
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
    await expect(this.dialog.getByText("渠道名称")).toHaveCount(0);
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

  async fillModel(data: { modelName: string }) {
    await this.dialog
      .getByRole("textbox", { name: /模型名称|Model Name/i })
      .fill(data.modelName);
  }

  private async assertModelProtocolOptions() {
    await this.modelProtocolField().locator(".ant-select-selector").click();
    const dropdown = this.page.locator(".ant-select-dropdown:visible").last();
    await expect(dropdown.getByTitle(/OpenAI/i).first()).toBeVisible();
    await expect(dropdown.getByTitle(/Anthropic/i).first()).toBeVisible();
    await expect(dropdown.getByTitle("OpenAI Compatible")).toHaveCount(0);
    await expect(dropdown.getByTitle("Anthropic Compatible")).toHaveCount(0);
    await expect(dropdown.getByTitle("Voyage")).toHaveCount(0);
    await expect(dropdown.getByText(/https?:\/\//i)).toHaveCount(0);
    await expect(dropdown.getByText(/\/v1/i)).toHaveCount(0);
    await dropdown
      .getByTitle(/OpenAI/i)
      .first()
      .evaluate((node) => (node as HTMLElement).click());
    await dropdown
      .getByTitle(/Anthropic/i)
      .first()
      .evaluate((node) => (node as HTMLElement).click());
    await this.assertModelProtocolSelectionCount(2);
    await this.page.keyboard.press("Escape");
    await expect(dropdown).toBeHidden();
  }

  private modelProtocolField() {
    return this.dialog
      .locator(".relative.flex", {
        hasText: /协议|Protocol/i,
      })
      .first();
  }

  private async selectModelProtocols(protocolLabels: RegExp[]) {
    await this.modelProtocolField().locator(".ant-select-selector").click();
    const dropdown = this.page.locator(".ant-select-dropdown:visible").last();
    for (const label of protocolLabels) {
      await dropdown
        .locator(".ant-select-item-option", { hasText: label })
        .first()
        .evaluate((node) => (node as HTMLElement).click());
    }
    await this.assertModelProtocolSelectionCount(protocolLabels.length);
    await this.page.keyboard.press("Escape");
  }

  private async assertModelProtocolSelectionCount(expectedCount: number) {
    await expect
      .poll(async () =>
        this.modelProtocolField().evaluate((node) => {
          const visibleItems = node.querySelectorAll(
            ".ant-select-selection-overflow-item:not(.ant-select-selection-overflow-item-rest) .ant-select-selection-item",
          ).length;
          const restText =
            node.querySelector(".ant-select-selection-overflow-item-rest")
              ?.textContent || "";
          const restCount = Number(restText.match(/\+(\d+)/)?.[1] || 0);
          return visibleItems + restCount;
        }),
      )
      .toBe(expectedCount);
  }

  private async assertModelDrawerLabelsSingleLine(labels: string[]) {
    for (const labelText of labels) {
      const label = this.dialog
        .locator("label", { hasText: labelText })
        .first();
      await expect(label).toBeVisible();
      const metrics = await label.evaluate((node) => {
        const style = window.getComputedStyle(node);
        const lineHeight = Number.parseFloat(style.lineHeight);
        const box = node.getBoundingClientRect();
        return {
          height: box.height,
          lineHeight: Number.isNaN(lineHeight) ? 24 : lineHeight,
          whiteSpace: style.whiteSpace,
        };
      });
      expect(metrics.whiteSpace).toBe("nowrap");
      expect(metrics.height).toBeLessThanOrEqual(metrics.lineHeight + 4);
    }
  }

  async saveModel() {
    const confirmButton = this.dialog
      .getByRole("button", { name: /保\s*存|确\s*认|Save|Confirm/i })
      .last();
    await expect(confirmButton).toBeVisible({ timeout: 10_000 });
    await confirmButton.evaluate((node) => (node as HTMLElement).click());
  }

  async confirmDrawer() {
    await this.dialog
      .getByRole("button", { name: /确\s*认|Confirm/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async searchProvider(name: string) {
    const page = this.providerPage();
    await page.getByRole("textbox").first().fill(name);
    await page
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

    await this.dialog.getByLabel(/渠道|Provider/i).click();
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

  async assertDraftTierCurrentTestLatency(expectedLatency: string) {
    const result = this.dialog.getByTestId("ai-tier-current-test-result");
    await expect(result).toBeVisible();
    await expect(result).toContainText(/耗时|Latency/i);
    await expect(result).toContainText(expectedLatency);
  }

  async assertTierModelOptionsGrouped(input: {
    anthropicModelName: string;
    openAIModelName: string;
    providerName: string;
    tierName: RegExp;
  }) {
    await this.editTier(input.tierName);
    await this.dialog.getByLabel(/渠道|Provider/i).click();
    const modelResponse = this.page.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        /\/x\/linapro-ai-core\/api\/v1\/ai\/providers\/\d+\/models/.test(
          response.url(),
        ),
      { timeout: 20_000 },
    );
    await this.page.getByTitle(input.providerName).click();
    await modelResponse;
    await this.dialog.getByLabel(/模型|Model/i).click();
    const dropdown = this.page
      .locator(".ant-select-dropdown:not(.ant-select-dropdown-hidden):visible")
      .last();
    await expect(dropdown.locator(".ant-select-item-group")).toContainText([
      /OpenAI/i,
      /Anthropic/i,
    ]);
    const openAIOption = dropdown.getByTitle(
      new RegExp(escapeRegExp(input.openAIModelName)),
    );
    const anthropicOption = dropdown.getByTitle(
      new RegExp(escapeRegExp(input.anthropicModelName)),
    );
    await expect(openAIOption).toBeVisible();
    await expect(anthropicOption).toBeVisible();
    const [openAIGroupText, anthropicGroupText] = await Promise.all([
      openAIOption.evaluate((node) => {
        let current: Element | null = node;
        while (current?.previousElementSibling) {
          current = current.previousElementSibling;
          if (current.classList.contains("ant-select-item-group")) {
            return current.textContent || "";
          }
        }
        return "";
      }),
      anthropicOption.evaluate((node) => {
        let current: Element | null = node;
        while (current?.previousElementSibling) {
          current = current.previousElementSibling;
          if (current.classList.contains("ant-select-item-group")) {
            return current.textContent || "";
          }
        }
        return "";
      }),
    ]);
    expect(openAIGroupText).toMatch(/OpenAI/i);
    expect(anthropicGroupText).toMatch(/Anthropic/i);
    await this.cancelDrawer();
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
      tabsBg,
      contentHolderBg,
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
      // Card surface lives on the tabs root; content-holder is intentionally transparent.
      tabs.evaluate((node) => window.getComputedStyle(node).backgroundColor),
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
    expect(tabsBg).not.toBe("rgba(0, 0, 0, 0)");
    expect(contentHolderBg).toBe("rgba(0, 0, 0, 0)");
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

  async assertTierTypePage(capabilityType: string) {
    await expect(this.tierCapabilityTypeTab(capabilityType)).toHaveAttribute(
      "aria-selected",
      "true",
    );
    await expect(
      this.page.getByText(/默认参数 JSON|Default Params JSON/i),
    ).toHaveCount(0);
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
    await expect(this.dialog.getByText("渠道", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("模型", { exact: true })).toBeVisible();
    await expect(this.dialog.getByText("Thinking Effort")).toHaveCount(0);
    await this.cancelDrawer();
  }

  async assertTierDrawerDefaultConfig(tierName: RegExp) {
    await this.editTier(tierName);
    await expect(
      this.dialog.getByText("Thinking Effort", { exact: true }),
    ).toBeVisible();
    await expect(
      this.dialog.getByText(/模型默认|Model default/i).first(),
    ).toBeVisible();
    await expect(
      this.dialog.getByText(/默认参数 JSON|Default Params JSON/i),
    ).toHaveCount(0);
    await expect(
      this.dialog.getByTestId("ai-tier-default-params-editor"),
    ).toHaveCount(0);
  }

  async saveTierDrawer() {
    await this.confirmDrawer();
  }

  async openInvocationDetail(rowText?: string | string[]) {
    const detailButtons = this.page.getByRole("button", {
      name: /详\s*情|Detail/i,
    });
    if (!rowText) {
      await detailButtons.first().click();
      await waitForDialogReady(this.dialog);
      return;
    }

    const rowTexts = Array.isArray(rowText) ? rowText : [rowText];
    let targetRow = this.page.locator(".vxe-body--row:visible");
    for (const text of rowTexts) {
      targetRow = targetRow.filter({ hasText: text });
    }
    const row = targetRow.first();
    await expect(row).toBeVisible();
    const targetTop = await row.evaluate(
      (node) => node.getBoundingClientRect().top,
    );
    const buttonCount = await detailButtons.count();
    let matchedIndex = 0;
    let matchedDelta = Number.POSITIVE_INFINITY;
    for (let index = 0; index < buttonCount; index += 1) {
      const buttonTop = await detailButtons.nth(index).evaluate((node) => {
        const row = node.closest(".vxe-body--row");
        return (row ?? node).getBoundingClientRect().top;
      });
      const delta = Math.abs(buttonTop - targetTop);
      if (delta < matchedDelta) {
        matchedDelta = delta;
        matchedIndex = index;
      }
    }
    expect(matchedDelta).toBeLessThanOrEqual(2);
    await detailButtons.nth(matchedIndex).click();
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

  async filterInvocationsByCapabilityPurposeAndSource(
    capabilityKey: string,
    purpose: string,
    sourcePluginId: string,
  ) {
    const currentCapabilityKey = await this.openCapabilityMethodSelect();
    await this.selectCapabilityDropdownOption(
      capabilityKey,
      currentCapabilityKey,
    );
    await this.page.getByLabel(/用途|Purpose/i).fill(purpose);
    await this.page.getByLabel(/来源插件|Source Plugin/i).fill(sourcePluginId);
    await this.searchInvocations();
  }

  async selectInvocationCreatedAtTodayRange() {
    const { end, start } = todayToTomorrowPickerRange();
    const startValue = start;
    const endValue = end;
    const rangeInputs = this.page
      .locator(".ant-picker-range")
      .first()
      .locator("input");

    for (const [index, value] of [startValue, endValue].entries()) {
      const input = rangeInputs.nth(index);
      await input.evaluate((node) => node.removeAttribute("readonly"));
      await input.click();
      await input.fill(value);
      await input.press(index === 0 ? "Tab" : "Enter");
    }
    await this.page.keyboard.press("Escape");
    await expect(rangeInputs.first()).toHaveValue(startValue);
    await expect(rangeInputs.nth(1)).toHaveValue(endValue);
    await this.closeInvocationFloatingOverlays();
    await this.assertInvocationSearchFormLayout();
    await this.assertInvocationCreatedAtDateRangeUsesMonitorLogLayout();
  }

  async searchInvocations() {
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForTableReady(this.page);
    await this.closeInvocationFloatingOverlays();
  }

  async assertInvocationDeleteButtonStyle() {
    const button = this.page.getByTestId("ai-invocation-clear");
    await expect(button).toBeVisible();
    await expect(button).toHaveClass(/ant-btn-primary/);
    await expect(button).toHaveClass(/ant-btn-dangerous/);
  }

  async expectInvocationDeleteDialogRequiresRangeAndCancel() {
    const modal = await this.openInvocationDeleteDialog();
    await expect(
      modal.getByText("Delete Request Logs", { exact: true }),
    ).toBeVisible();
    await expect(
      modal.getByText(
        "Select how to delete request logs. By default, logs are cleaned by date and time range; selecting all logs cleans every request log in the current permission scope.",
      ),
    ).toBeVisible();
    await expect(
      modal.getByText("Delete all request logs", { exact: true }),
    ).toBeVisible();
    await expect(
      modal.getByText(
        "When selected, the date and time range is ignored and all request logs in the current permission scope are deleted.",
      ),
    ).toBeVisible();
    await expect(
      modal.getByTestId("ai-invocation-delete-range-section"),
    ).toBeVisible();
    await expect(modal.getByText(/plugin\.linapro-ai-core/)).toHaveCount(0);
    await this.assertInvocationDeleteDialogSpacing(modal);
    await this.captureEvidence("TC003-invocation-delete-dialog");
    await modal.getByRole("button", { name: /确\s*定|Confirm|OK/i }).click();
    await expect(
      this.page.getByText("Select a complete request log date and time range"),
    ).toBeVisible();
    await expect(modal).toBeVisible();
    await modal.getByRole("button", { name: /取\s*消|Cancel/i }).click();
    await modal.waitFor({ state: "hidden", timeout: 5_000 }).catch(() => {});
  }

  async expectInvocationDeleteAllModeUsesUnscopedClean() {
    const routePattern = "**/x/linapro-ai-core/api/v1/ai/invocations/clean**";
    const handler: Parameters<Page["route"]>[1] = async (route) => {
      await route.fulfill({
        body: JSON.stringify({
          code: 0,
          data: { deleted: 0 },
          message: "OK",
        }),
        contentType: "application/json",
        status: 200,
      });
    };

    await this.page.route(routePattern, handler);
    try {
      const modal = await this.openInvocationDeleteDialog();
      await modal
        .getByRole("checkbox", { name: "Delete all request logs" })
        .check();
      await expect(
        modal
          .getByTestId("ai-invocation-delete-range-section")
          .locator(".ant-picker-range"),
      ).toHaveClass(/ant-picker-disabled/);
      const requestPromise = this.page.waitForRequest((request) => {
        const url = new URL(request.url());
        return (
          request.method() === "DELETE" &&
          url.pathname.includes(
            "/x/linapro-ai-core/api/v1/ai/invocations/clean",
          ) &&
          !url.searchParams.has("startedAt") &&
          !url.searchParams.has("endedAt")
        );
      });
      await modal.getByRole("button", { name: /确\s*定|Confirm|OK/i }).click();
      await requestPromise;
      await modal.waitFor({ state: "hidden", timeout: 5_000 }).catch(() => {});
      await waitForTableReady(this.page);
    } finally {
      await this.page.unroute(routePattern, handler);
    }
  }

  async confirmInvocationCleanWithDialogRange() {
    const modal = await this.openInvocationDeleteDialog();
    const { end, start } = todayToTomorrowPickerRange();
    await this.fillInvocationDeleteRange(
      modal,
      `${start} 00:00:00`,
      `${end} 00:00:00`,
    );
    await modal.getByRole("button", { name: /确\s*定|Confirm|OK/i }).click();
    await modal.waitFor({ state: "hidden", timeout: 5_000 }).catch(() => {});
    await waitForTableReady(this.page);
  }

  private async openInvocationDeleteDialog() {
    await this.page.getByTestId("ai-invocation-clear").click();
    const modal = this.page.locator(".ant-modal-wrap:visible").filter({
      hasText: /Delete Request Logs|删除调用日志/,
    });
    await waitForDialogReady(modal);
    await this.page.waitForTimeout(150);
    return modal;
  }

  private async fillInvocationDeleteRange(
    modal: ReturnType<Page["locator"]>,
    startValue: string,
    endValue: string,
  ) {
    const rangeInputs = modal
      .getByTestId("ai-invocation-delete-range-section")
      .locator(".ant-picker-range")
      .locator("input");
    for (const [index, value] of [startValue, endValue].entries()) {
      const input = rangeInputs.nth(index);
      await input.evaluate((node) => node.removeAttribute("readonly"));
      await input.click();
      await input.fill(value);
      await input.press(index === 0 ? "Tab" : "Enter");
    }
    await expect(rangeInputs.first()).toHaveValue(startValue);
    await expect(rangeInputs.nth(1)).toHaveValue(endValue);
  }

  private async assertInvocationDeleteDialogSpacing(
    modal: ReturnType<Page["locator"]>,
  ) {
    const metrics = await modal.evaluate((node) => {
      const alert = node.querySelector<HTMLElement>(
        '[data-testid="ai-invocation-delete-alert"]',
      );
      const option = node.querySelector<HTMLElement>(
        '[data-testid="ai-invocation-delete-all-option"]',
      );
      const range = node.querySelector<HTMLElement>(
        '[data-testid="ai-invocation-delete-range-section"]',
      );
      if (!alert || !option || !range) {
        return null;
      }
      return {
        optionMarginTop: Number.parseFloat(
          window.getComputedStyle(option).marginTop,
        ),
        rangeMarginTop: Number.parseFloat(
          window.getComputedStyle(range).marginTop,
        ),
      };
    });
    expect(metrics).not.toBeNull();
    if (!metrics) {
      return;
    }
    expect(metrics.optionMarginTop).toBeGreaterThanOrEqual(16);
    expect(metrics.rangeMarginTop).toBeGreaterThanOrEqual(16);
  }

  private async assertInvocationCreatedAtDateRangeUsesMonitorLogLayout() {
    const methodSelect = await this.visibleAntSelectByLabel(
      /调用方法|Invocation Method/i,
    );
    const methodSelectWidth = await methodSelect.evaluate(
      (node) => node.getBoundingClientRect().width,
    );
    const metrics = await this.page
      .locator(".ant-picker-range")
      .first()
      .evaluate((node) => {
        const box = node.getBoundingClientRect();
        const field = node.closest(".ant-form-item");
        const fieldBox = field?.getBoundingClientRect();
        const clippingParent = node.closest(".overflow-hidden");
        const clippingParentBox = clippingParent?.getBoundingClientRect();
        const labels = Array.from(document.querySelectorAll("label"));
        function labelBox(pattern: RegExp) {
          return labels
            .find((item) => pattern.test(item.textContent || ""))
            ?.getBoundingClientRect();
        }
        const methodLabelBox = labelBox(/调用方法|Invocation Method/i);
        const sourcePluginLabelBox = labelBox(/来源插件|Source Plugin/i);
        return {
          clippingParentRight: clippingParentBox?.right ?? box.right,
          clippingParentWidth: clippingParentBox?.width ?? box.width,
          fieldRight: fieldBox?.right ?? box.right,
          fieldWidth: fieldBox?.width ?? box.width,
          createdLabelWidth:
            labels
              .find((item) => /创建时间|Created/i.test(item.textContent || ""))
              ?.getBoundingClientRect().width ?? 0,
          inputs: Array.from(
            node.querySelectorAll<HTMLInputElement>("input"),
          ).map((input) => ({
            clientWidth: input.clientWidth,
            fontSize: Number.parseFloat(
              window.getComputedStyle(input).fontSize,
            ),
            scrollWidth: input.scrollWidth,
            value: input.value,
          })),
          pickerRight: box.right,
          pickerWidth: box.width,
          methodLabelRight: methodLabelBox?.right ?? 0,
          sourcePluginLabelLeft: sourcePluginLabelBox?.left ?? 0,
          viewportRight: document.documentElement.clientWidth,
        };
      });

    expect(metrics.pickerWidth).toBeGreaterThan(0);
    expect(methodSelectWidth).toBeGreaterThan(0);
    expect(metrics.createdLabelWidth).toBeGreaterThanOrEqual(110);
    expect(metrics.createdLabelWidth).toBeLessThanOrEqual(114);
    expect(methodSelectWidth).toBeGreaterThanOrEqual(241);
    expect(methodSelectWidth).toBeLessThanOrEqual(243);
    expect(metrics.pickerWidth).toBeGreaterThanOrEqual(241);
    expect(metrics.pickerWidth).toBeLessThanOrEqual(243);
    expect(
      Math.abs(metrics.pickerWidth - methodSelectWidth),
    ).toBeLessThanOrEqual(1);
    expect(metrics.pickerWidth).toBeLessThanOrEqual(metrics.fieldWidth);
    expect(metrics.pickerRight).toBeLessThanOrEqual(
      metrics.clippingParentRight + 1,
    );
    expect(metrics.pickerRight).toBeLessThanOrEqual(metrics.fieldRight + 1);
    expect(
      metrics.sourcePluginLabelLeft - metrics.pickerRight,
    ).toBeGreaterThanOrEqual(5);
    expect(metrics.pickerRight).toBeLessThanOrEqual(metrics.viewportRight - 8);
    await expect(this.page.locator(".ant-picker-time-panel")).toHaveCount(0);
    expect(metrics.inputs).toHaveLength(2);
    for (const metric of metrics.inputs) {
      expect(metric.value).toMatch(/^\d{4}-\d{2}-\d{2}$/);
      expect(metric.clientWidth).toBeGreaterThan(0);
      expect(metric.fontSize).toBeGreaterThanOrEqual(13.5);
      expect(metric.scrollWidth).toBeLessThanOrEqual(metric.clientWidth + 4);
    }
  }

  private async assertInvocationSearchFormLayout() {
    const metrics = await this.page.evaluate(() => {
      function itemBox(pattern: RegExp) {
        const label = Array.from(document.querySelectorAll("label")).find(
          (node) => pattern.test(node.textContent || ""),
        );
        if (!label) {
          return null;
        }
        const labelBox = label.getBoundingClientRect();
        const controlId = label.getAttribute("for");
        const labelledControl = controlId
          ? document.querySelector<HTMLElement>(
              `[id="${CSS.escape(controlId)}"]`,
            )
          : null;
        const control =
          labelledControl?.closest<HTMLElement>(
            ".ant-select, .ant-picker, .ant-input-affix-wrapper",
          ) ??
          labelledControl?.closest<HTMLElement>(".ant-input") ??
          null;
        const controlBox = control?.getBoundingClientRect();
        const style = window.getComputedStyle(label);
        const range = document.createRange();
        range.selectNodeContents(label);
        const textBox = range.getBoundingClientRect();
        range.detach();
        const itemLeft = labelBox.left;
        const itemRight = controlBox?.right ?? labelBox.right;
        const itemTop = Math.min(labelBox.top, controlBox?.top ?? labelBox.top);
        return {
          controlLeft: controlBox?.left ?? 0,
          controlRight: controlBox?.right ?? 0,
          controlWidth: controlBox?.width ?? 0,
          itemLeft,
          itemRight,
          itemTop,
          itemWidth: itemRight - itemLeft,
          justifyContent: style.justifyContent,
          labelLeft: labelBox.left,
          labelRight: labelBox.right,
          labelWidth: labelBox.width,
          textLeft: textBox.left,
          textRight: textBox.right,
          textWidth: textBox.width,
        };
      }

      return {
        capability: itemBox(/调用方法|Invocation Method/i),
        created: itemBox(/创建时间|Created/i),
        searchFormLeft:
          document
            .querySelector(".vxe-grid--form-wrapper")
            ?.getBoundingClientRect().left ?? 0,
        searchFormRight:
          document
            .querySelector(".vxe-grid--form-wrapper")
            ?.getBoundingClientRect().right ??
          document.documentElement.clientWidth,
        purpose: itemBox(/用途|Purpose/i),
        sourcePlugin: itemBox(/来源插件|Source Plugin/i),
        status: itemBox(/状态|Status/i),
        tier: itemBox(/档位|Tier/i),
        viewportRight: document.documentElement.clientWidth,
      };
    });

    expect(metrics.capability).not.toBeNull();
    expect(metrics.created).not.toBeNull();
    expect(metrics.purpose).not.toBeNull();
    expect(metrics.sourcePlugin).not.toBeNull();
    expect(metrics.tier).not.toBeNull();
    expect(metrics.status).not.toBeNull();
    if (
      !metrics.capability ||
      !metrics.created ||
      !metrics.purpose ||
      !metrics.sourcePlugin ||
      !metrics.tier ||
      !metrics.status
    ) {
      return;
    }

    expect(
      Math.abs(metrics.purpose.itemTop - metrics.capability.itemTop),
    ).toBeLessThanOrEqual(4);
    expect(
      Math.abs(metrics.tier.itemTop - metrics.capability.itemTop),
    ).toBeLessThanOrEqual(4);
    expect(
      Math.abs(metrics.status.itemTop - metrics.capability.itemTop),
    ).toBeLessThanOrEqual(4);
    expect(
      Math.abs(metrics.status.itemTop - metrics.created.itemTop),
    ).toBeGreaterThan(20);
    expect(
      Math.abs(metrics.sourcePlugin.itemTop - metrics.created.itemTop),
    ).toBeLessThanOrEqual(4);
    expect(
      metrics.created.itemTop - metrics.capability.itemTop,
    ).toBeGreaterThan(20);
    expect(
      Math.abs(metrics.created.itemLeft - metrics.capability.itemLeft),
    ).toBeLessThanOrEqual(1);
    expect(
      Math.abs(metrics.created.textRight - metrics.capability.textRight),
    ).toBeLessThanOrEqual(1);
    expect(
      Math.abs(metrics.created.controlLeft - metrics.capability.controlLeft),
    ).toBeLessThanOrEqual(1);
    expect(
      metrics.capability.labelLeft - metrics.searchFormLeft,
    ).toBeGreaterThanOrEqual(15);
    expect(
      metrics.created.labelLeft - metrics.searchFormLeft,
    ).toBeGreaterThanOrEqual(15);
    expect(metrics.capability.textLeft).toBeGreaterThanOrEqual(
      metrics.searchFormLeft + 4,
    );
    expect(metrics.created.textLeft).toBeGreaterThanOrEqual(
      metrics.searchFormLeft + 4,
    );
    expect(
      Math.abs(metrics.status.itemLeft - metrics.purpose.itemLeft),
    ).toBeGreaterThan(120);
    expect(
      Math.abs(metrics.sourcePlugin.itemLeft - metrics.purpose.itemLeft),
    ).toBeLessThanOrEqual(4);
    expect(
      Math.abs(metrics.sourcePlugin.textRight - metrics.purpose.textRight),
    ).toBeLessThanOrEqual(1);
    expect(
      Math.abs(metrics.sourcePlugin.controlLeft - metrics.purpose.controlLeft),
    ).toBeLessThanOrEqual(1);
    expect(metrics.purpose.itemLeft).toBeGreaterThan(
      metrics.capability.itemLeft + 120,
    );
    expect(metrics.tier.itemLeft).toBeGreaterThan(
      metrics.purpose.itemLeft + 120,
    );
    expect(metrics.status.itemLeft).toBeGreaterThan(
      metrics.tier.itemLeft + 120,
    );
    expect(metrics.sourcePlugin.itemLeft).toBeGreaterThan(
      metrics.created.itemLeft + 120,
    );
    const measuredItems = {
      capability: metrics.capability,
      created: metrics.created,
      purpose: metrics.purpose,
      sourcePlugin: metrics.sourcePlugin,
      status: metrics.status,
      tier: metrics.tier,
    };
    const minControlWidths: Record<keyof typeof measuredItems, number> = {
      capability: 241,
      created: 241,
      purpose: 122,
      sourcePlugin: 122,
      status: 132,
      tier: 132,
    };
    const maxControlWidths: Record<keyof typeof measuredItems, number> = {
      capability: 243,
      created: 243,
      purpose: 124,
      sourcePlugin: 124,
      status: 136,
      tier: 136,
    };
    for (const [name, item] of Object.entries(measuredItems)) {
      const minControlWidth =
        minControlWidths[name as keyof typeof minControlWidths];
      const maxControlWidth =
        maxControlWidths[name as keyof typeof maxControlWidths];
      if (item.controlWidth < minControlWidth) {
        throw new Error(
          `${name} control width is too narrow: ${JSON.stringify(item)}`,
        );
      }
      if (item.controlWidth > maxControlWidth) {
        throw new Error(
          `${name} control width changed unexpectedly: ${JSON.stringify(item)}`,
        );
      }
      expect(item.controlRight).toBeLessThanOrEqual(item.itemRight + 1);
      expect(item.controlRight).toBeLessThanOrEqual(
        metrics.searchFormRight + 1,
      );
      expect(item.controlRight).toBeLessThanOrEqual(metrics.viewportRight - 8);
      expect(item.labelWidth).toBeLessThanOrEqual(120);
    }
    expect(
      Math.abs(metrics.created.itemWidth - metrics.capability.itemWidth),
    ).toBeLessThanOrEqual(1);
    expect(
      Math.abs(metrics.created.controlWidth - metrics.capability.controlWidth),
    ).toBeLessThanOrEqual(1);
    expect(metrics.capability.justifyContent).toBe("flex-end");
    expect(metrics.created.justifyContent).toBe("flex-end");
  }

  private async openCapabilityMethodSelect(root: Locator | Page = this.page) {
    const select = await this.visibleAntSelectByLabel(
      /调用方法|Invocation Method/i,
      root,
    );
    const currentCapabilityKey =
      (await select
        .locator(".ant-select-selection-item")
        .first()
        .textContent({ timeout: 500 })
        .catch(() => "")) || "text.generate";
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
    root: Locator | Page = this.page,
  ) {
    const targetIndex = capabilityMethodOptionOrder.indexOf(label);
    if (targetIndex < 0) {
      throw new Error(`未知调用方法选项: ${label}`);
    }

    const dropdown = this.visibleCapabilityDropdown();
    const visibleOptions = dropdown.locator(".ant-select-item-option:visible");
    await expect
      .poll(async () => visibleOptions.count(), { timeout: 5_000 })
      .toBeGreaterThan(0);

    if (await this.clickVisibleCapabilityOption(label)) {
      await this.expectCapabilityMethodSelected(label, root);
      return;
    }

    await this.ensureCapabilityDropdownOpen(root);
    const currentIndex = capabilityMethodOptionOrder.indexOf(currentLabel);
    const startIndex = currentIndex >= 0 ? currentIndex : 0;
    const key = targetIndex >= startIndex ? "ArrowDown" : "ArrowUp";
    const steps = Math.abs(targetIndex - startIndex);
    for (let index = 0; index < steps; index += 1) {
      await this.page.keyboard.press(key);
    }
    await this.page.keyboard.press("Enter");
    await this.expectCapabilityMethodSelected(label, root);
  }

  private visibleCapabilityDropdown() {
    return this.visibleSelectDropdown();
  }

  private visibleSelectDropdown() {
    return this.page
      .locator(".ant-select-dropdown:not(.ant-select-dropdown-hidden):visible")
      .last();
  }

  private async closeInvocationFloatingOverlays() {
    await this.page.keyboard.press("Escape");
    await expect(
      this.page.locator(
        ".ant-select-dropdown:not(.ant-select-dropdown-hidden):visible",
      ),
    ).toHaveCount(0, { timeout: 2_000 });
    await expect(this.page.locator(".ant-picker-dropdown:visible")).toHaveCount(
      0,
      { timeout: 2_000 },
    );
  }

  private async ensureCapabilityDropdownOpen(root: Locator | Page = this.page) {
    if ((await this.visibleCapabilityDropdown().count()) > 0) {
      return;
    }
    const select = await this.visibleAntSelectByLabel(
      /调用方法|Invocation Method/i,
      root,
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

  private async expectCapabilityMethodSelected(
    label: string,
    root: Locator | Page = this.page,
  ) {
    await expect
      .poll(async () => {
        const select = await this.visibleAntSelectByLabel(
          /调用方法|Invocation Method/i,
          root,
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

  private async selectModelProviderFilter(providerName: string) {
    const select = await this.visibleAntSelectByLabel(
      /渠道|Provider/i,
      this.modelPage(),
    );
    const providerTitle = new RegExp(`^${escapeRegExp(providerName)}$`);

    for (let attempt = 0; attempt < 3; attempt += 1) {
      await select.locator(".ant-select-selector").click();
      const option = this.visibleSelectDropdown().getByTitle(providerTitle);
      try {
        await expect(option).toBeVisible({ timeout: 5_000 });
        await option.click();
        await expect(
          select.locator(".ant-select-selection-item").first(),
        ).toContainText(providerName);
        return;
      } catch (error) {
        await this.page.keyboard.press("Escape");
        if (attempt === 2) {
          throw error;
        }
        await this.page.waitForTimeout(250);
      }
    }
  }

  private async visibleAntSelectByLabel(
    label: RegExp,
    root: Locator | Page = this.page,
  ) {
    const controls = root.getByLabel(label);
    const count = await controls.count();
    for (let index = 0; index < count; index += 1) {
      const select = controls
        .nth(index)
        .locator(
          'xpath=ancestor::*[contains(concat(" ", normalize-space(@class), " "), " ant-select ")][1]',
        );
      if (await select.isVisible().catch(() => false)) {
        return select;
      }
    }
    throw new Error(`未找到可见下拉框: ${label}`);
  }

  private splitCapabilityKey(value: string) {
    const [capabilityType = "text", ...methodParts] = value.split(".");
    return {
      capabilityMethod: methodParts.join(".") || "generate",
      capabilityType: capabilityType || "text",
    };
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
    throw new Error(`未找到渠道表列: ${header}`);
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

  private providerPage() {
    return this.page.getByTestId("ai-provider-management-page").first();
  }

  private modelPage() {
    return this.page.getByTestId("ai-model-management-page").first();
  }

  private async searchModel(name: string) {
    const page = this.modelPage();
    await page.getByRole("textbox").first().fill(name);
    await page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  private async modelHeaderIndex(header: RegExp) {
    const headers = this.modelPage().locator(
      ".vxe-table--main-wrapper .vxe-header--column:visible",
    );
    const count = await headers.count();
    for (let index = 0; index < count; index += 1) {
      const text = (await headers.nth(index).innerText()).trim();
      if (header.test(text)) {
        return index;
      }
    }
    throw new Error(`未找到模型表列: ${header}`);
  }

  private modelMainRow(modelName: string) {
    return this.modelPage()
      .locator(".vxe-table--main-wrapper .vxe-body--row:visible", {
        hasText: modelName,
      })
      .first();
  }

  private async modelActionRow(modelName: string) {
    const row = this.modelMainRow(modelName);
    await row.waitFor({ state: "visible", timeout: 10_000 });
    const rowID = await row.getAttribute("rowid");
    expect(rowID, `未找到模型行 rowid: ${modelName}`).toBeTruthy();
    const actionRow = this.modelPage()
      .locator(
        `.vxe-table--fixed-right-wrapper .vxe-body--row[rowid="${cssAttributeValue(rowID || "")}"]:visible`,
      )
      .first();
    await actionRow.waitFor({ state: "visible", timeout: 10_000 });
    return actionRow;
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
    expect(rowID, `未找到渠道行 rowid: ${providerName}`).toBeTruthy();
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
