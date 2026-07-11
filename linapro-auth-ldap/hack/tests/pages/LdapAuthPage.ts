import type { Page } from '@host-tests/support/playwright';
import { expect } from '@host-tests/fixtures/auth';
import { MainLayout } from '@host-tests/pages/MainLayout';
import { waitForRouteReady } from '@host-tests/support/ui';
import { workspacePath } from '@host-tests/fixtures/config';

export class LdapAuthPage {
  constructor(private page: Page) {}

  get loginEntry() {
    return this.page.getByTestId('linapro-auth-ldap-entry');
  }

  get loginEntryButton() {
    return this.page.getByTestId('linapro-auth-ldap-entry-button');
  }

  /** Settings form uses host horizontal layout (label left / control right). */
  get settingsForm() {
    return this.page.locator('form.ant-form-horizontal').first();
  }

  /** Page intro tip uses host-standard info Alert above the form. */
  get settingsIntroAlert() {
    return this.page.locator('.ant-alert-info').first();
  }

  get fieldHelpIcons() {
    return this.page.locator('.ant-form-item-tooltip');
  }

  get requiredFieldLabels() {
    return this.settingsForm.locator('.ant-form-item-required');
  }

  async openSettingsPage() {
    const layout = new MainLayout(this.page);
    await this.page.goto(workspacePath('/dashboard/workspace'));
    await waitForRouteReady(this.page);
    await layout.expandSidebarGroup(/授权登录|Auth Login/i);
    await layout.sidebarMenuItem(/LDAP 设置|LDAP Settings|LDAP/i).click();
    await waitForRouteReady(this.page);
    await expect(this.settingsForm).toBeVisible({ timeout: 15000 });
  }

  async expectFieldHelpTooltip(helpText: string | RegExp) {
    const icons = this.fieldHelpIcons;
    await expect(icons.first()).toBeVisible();
    expect(await icons.count()).toBeGreaterThan(0);

    const preferred = this.page
      .locator('.ant-form-item')
      .filter({ hasText: /Base DN|用户过滤器|User filter|Subject/i })
      .locator('.ant-form-item-tooltip')
      .first();
    const target = (await preferred.count()) > 0 ? preferred : icons.first();
    await target.hover();
    const tooltip = this.page.locator('.ant-tooltip:visible').last();
    await expect(tooltip).toBeVisible({ timeout: 5000 });
    await expect(tooltip).toContainText(helpText);
  }

  get loginModal() {
    return this.page.getByRole('dialog', {
      name: /LDAP login|LDAP 账号登录/,
    });
  }

  get loginModalTitle() {
    return this.loginModal.getByText('LDAP 账号登录', { exact: true });
  }

  get usernameInput() {
    return this.page.getByTestId('linapro-auth-ldap-username');
  }

  get passwordInput() {
    return this.page.getByTestId('linapro-auth-ldap-password');
  }

  get confirmButton() {
    return this.loginModal.getByRole('button', { name: /登\s*录/ });
  }

  get cancelButton() {
    return this.loginModal.getByRole('button', { name: /取\s*消/ });
  }

  get credentialsRequiredMessages() {
    return this.loginModal.getByText('请输入用户名和密码', { exact: true });
  }

  async openLoginModal() {
    await this.loginEntryButton.click();
  }

  async getLoginEntryLayout() {
    return this.loginEntry.evaluate((entry) => {
      const button = entry.querySelector('button');
      if (!(button instanceof HTMLElement)) {
        throw new Error('LDAP login entry button is missing');
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

  async getLoginModalLayout() {
    return this.loginModal.evaluate((modal) => {
      const box = modal.getBoundingClientRect();
      return {
        left: box.left,
        right: box.right,
        viewportWidth: window.innerWidth,
        width: box.width,
      };
    });
  }
}
