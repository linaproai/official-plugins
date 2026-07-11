import type { Page } from '@host-tests/support/playwright';
import { expect } from '@host-tests/fixtures/auth';
import { MainLayout } from '@host-tests/pages/MainLayout';
import { waitForRouteReady } from '@host-tests/support/ui';
import { workspacePath } from '@host-tests/fixtures/config';

/**
 * Generic OIDC plugin page object. Host LoginPage must not hard-code this plugin.
 */
export class GenericOidcPage {
  constructor(private page: Page) {}

  get loginEntry() {
    return this.page.getByTestId('linapro-oidc-generic-entry');
  }

  get loginEntryButton() {
    return this.page.getByTestId('linapro-oidc-generic-entry-button');
  }

  get settingsCardTitle() {
    // Card title may equal menu title ("OIDC 设置"); scope to form card head.
    return this.page
      .locator('.ant-card-head-title')
      .filter({ hasText: /OIDC 设置|OIDC Settings/i })
      .first();
  }

  get fieldHelpIcons() {
    return this.page.locator('.ant-form-item-tooltip');
  }

  async openSettingsPage() {
    const layout = new MainLayout(this.page);
    await this.page.goto(workspacePath('/dashboard/workspace'));
    await waitForRouteReady(this.page);
    await layout.expandSidebarGroup(/授权登录|Auth Login/i);
    await layout
      .sidebarMenuItem(/OIDC 设置|OIDC Settings/i)
      .click();
    await waitForRouteReady(this.page);
    await expect(this.settingsCardTitle).toBeVisible({ timeout: 15000 });
  }

  async expectFieldHelpTooltip(helpText: string | RegExp) {
    const icons = this.fieldHelpIcons;
    await expect(icons.first()).toBeVisible();
    const count = await icons.count();
    expect(count).toBeGreaterThan(0);

    // Prefer the Issuer/host-style hard field when present.
    const preferred = this.page
      .locator('.ant-form-item')
      .filter({ hasText: /Issuer|Client ID|客户端/i })
      .locator('.ant-form-item-tooltip')
      .first();
    const target = (await preferred.count()) > 0 ? preferred : icons.first();
    await target.hover();
    const tooltip = this.page.locator('.ant-tooltip:visible').last();
    await expect(tooltip).toBeVisible({ timeout: 5000 });
    await expect(tooltip).toContainText(helpText);
  }

  async getLoginEntryLayout() {
    return this.loginEntry.evaluate((entry) => {
      const button = entry.querySelector('button');
      if (!(button instanceof HTMLElement)) {
        throw new Error('Generic OIDC login entry button is missing');
      }
      const entryBox = entry.getBoundingClientRect();
      const buttonBox = button.getBoundingClientRect();
      return {
        buttonHeight: buttonBox.height,
        buttonRight: buttonBox.right,
        buttonWidth: buttonBox.width,
        clientHeight: button.clientHeight,
        clientWidth: button.clientWidth,
        entryWidth: entryBox.width,
        scrollHeight: button.scrollHeight,
        scrollWidth: button.scrollWidth,
        viewportWidth: window.innerWidth,
      };
    });
  }
}
