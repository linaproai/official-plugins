import { expect, type Page } from '@host-tests/support/playwright';

import { workspacePath } from '@host-tests/fixtures/config';
import {
  waitForBusyIndicatorsToClear,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from '@host-tests/support/ui';

export class SmartCenterPage {
  constructor(private page: Page) {}

  private get dialog() {
    return this.page.locator('[role="dialog"]').last();
  }

  async gotoProviders() {
    await this.page.goto(workspacePath('/ai/providers'));
    await waitForTableReady(this.page);
  }

  async gotoTiers() {
    await this.page.goto(workspacePath('/ai/tiers'));
    await waitForTableReady(this.page);
  }

  async gotoInvocations() {
    await this.page.goto(workspacePath('/ai/invocations'));
    await waitForTableReady(this.page);
  }

  async openCreateProvider() {
    await this.page
      .getByRole('button', { name: /新\s*增|Add/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async assertCreateProviderDrawerChineseTranslations() {
    await this.openCreateProvider();
    await expect(this.dialog.getByText('新增供应商')).toBeVisible();
    await expect(this.dialog.getByText('供应商名称')).toBeVisible();
    await expect(this.dialog.getByText('启用', { exact: true }).first()).toBeVisible();
    await expect(this.dialog.getByText('停用', { exact: true }).first()).toBeVisible();
    await expect(this.dialog.getByText('plugin.linapro-ai-core.common.enabled')).toHaveCount(0);
    await expect(this.dialog.getByText('plugin.linapro-ai-core.common.disabled')).toHaveCount(0);
    await expect(this.dialog.getByText('plugin.linapro-ai-core.effort.empty')).toHaveCount(0);
    await this.cancelDrawer();
  }

  async cancelDrawer() {
    await this.dialog
      .getByRole('button', { name: /取\s*消|Cancel/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async openProvider(name: string) {
    const row = this.page.locator('.vxe-body--row:visible', { hasText: name });
    await row.first().waitFor({ state: 'visible', timeout: 10_000 });
    await this.page
      .getByRole('button', { name: /编\s*辑|Edit/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async fillProvider(data: {
    name: string;
    openaiBaseUrl: string;
    apiKey: string;
  }) {
    await this.dialog
      .getByRole('textbox', { name: /供应商名称|Provider Name/i })
      .fill(data.name);
    await this.dialog
      .getByRole('textbox', { name: /OpenAI Base URL/i })
      .fill(data.openaiBaseUrl);
    await this.dialog
      .getByLabel(/API Key|Secret/i)
      .fill(data.apiKey);
  }

  async fillModel(data: { modelName: string; maxOutputTokens?: string }) {
    await this.dialog
      .getByRole('textbox', { name: /模型名称|Model Name/i })
      .fill(data.modelName);
    const maxOutput = this.dialog
      .getByLabel(/最大输出 Tokens|Max Output Tokens/i)
      .first();
    await maxOutput.fill(data.maxOutputTokens || '256');
  }

  async saveModel() {
    await this.dialog
      .getByRole('button', { name: /保\s*存|Save/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async confirmDrawer() {
    await this.dialog
      .getByRole('button', { name: /确\s*认|Confirm/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  async searchProvider(name: string) {
    await this.page.getByLabel(/供应商|Provider/i).first().fill(name);
    await this.page
      .getByRole('button', { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async deleteProvider(name: string) {
    const row = this.page.locator('.vxe-body--row:visible', { hasText: name });
    await row.first().waitFor({ state: 'visible', timeout: 10_000 });
    await this.page
      .getByRole('button', { name: /删\s*除|Delete/i })
      .first()
      .click();
    await this.page
      .locator('.ant-popover')
      .getByRole('button', { name: /确\s*定|OK/i })
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
      .getByRole('button', { name: /编\s*辑|Edit/i })
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
      .getByRole('button', { name: /编\s*辑|Edit/i })
      .nth(rowIndex)
      .click();
    await waitForDialogReady(this.dialog);
    return this.dialog;
  }

  async openInvocationDetail() {
    await this.page
      .getByRole('button', { name: /详\s*情|Detail/i })
      .first()
      .click();
    await waitForDialogReady(this.dialog);
  }

  async filterInvocationsByPurpose(purpose: string) {
    await this.page.getByLabel(/用途|Purpose/i).fill(purpose);
    await this.page
      .getByRole('button', { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForTableReady(this.page);
  }

  private async tierRowIndex(tierName: RegExp) {
    const rows = this.page.locator('.vxe-table--body .vxe-body--row:visible');
    const count = await rows.count();
    for (let index = 0; index < count; index += 1) {
      const text = await rows.nth(index).textContent();
      if (text && tierName.test(text)) {
        return index;
      }
    }
    throw new Error(`未找到档位行: ${tierName}`);
  }
}
