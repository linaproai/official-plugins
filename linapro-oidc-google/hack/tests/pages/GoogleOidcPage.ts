import type { Page } from "@playwright/test";
import { expect } from "@host-tests/fixtures/auth";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { waitForRouteReady } from "@host-tests/support/ui";
import { workspacePath } from "@host-tests/fixtures/config";

/**
 * Google OIDC plugin page object. Host LoginPage must not hard-code this plugin.
 */
export class GoogleOidcPage {
  constructor(private page: Page) {}

  get loginEntry() {
    return this.page.getByTestId("linapro-oidc-google-entry");
  }

  get loginEntryButton() {
    return this.page.getByTestId("linapro-oidc-google-entry-button");
  }

  /**
   * Layout metrics for the platform social icon entry (not full-width button).
   * Compares icon size against the host social region / form width.
   */
  async getLoginEntryLayout() {
    return this.loginEntry.evaluate((entry) => {
      const button = entry.querySelector(
        '[data-testid="linapro-oidc-google-entry-button"]',
      ) as HTMLElement | null;
      if (!button) {
        throw new Error("Google social login entry button is missing");
      }
      const region =
        entry.closest('[data-testid="login-social-auth-region"]') ??
        entry.closest("form")?.parentElement ??
        entry.parentElement;
      const regionBox = region?.getBoundingClientRect();
      const buttonBox = button.getBoundingClientRect();
      return {
        buttonHeight: buttonBox.height,
        buttonWidth: buttonBox.width,
        regionWidth: regionBox?.width ?? 0,
        isIconSized: buttonBox.width < 72 && buttonBox.height < 72,
      };
    });
  }

  /** Settings form uses host horizontal layout (label left / control right). */
  get settingsForm() {
    return this.page.locator("form.ant-form-horizontal").first();
  }

  /** Page intro tip uses host-standard info Alert above the form. */
  get settingsIntroAlert() {
    return this.page.locator(".ant-alert-info").first();
  }

  get fieldHelpIcons() {
    return this.page.locator(".ant-form-item-tooltip");
  }

  get requiredFieldLabels() {
    return this.settingsForm.locator(".ant-form-item-required");
  }

  async openSettingsPage() {
    const layout = new MainLayout(this.page);
    await this.page.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(this.page);
    await layout.expandSidebarGroup(/授权登录|Auth Login/i);
    await layout.sidebarMenuItem(/Google 登录|Google Login/i).click();
    await waitForRouteReady(this.page);
    await expect(this.settingsForm).toBeVisible({ timeout: 15000 });
  }

  async expectFieldHelpTooltip(helpText: string | RegExp) {
    const icons = this.fieldHelpIcons;
    await expect(icons.first()).toBeVisible();
    expect(await icons.count()).toBeGreaterThan(0);

    const preferred = this.page
      .locator(".ant-form-item")
      .filter({ hasText: /客户端 ID|Client ID/i })
      .locator(".ant-form-item-tooltip")
      .first();
    const target = (await preferred.count()) > 0 ? preferred : icons.first();
    await target.hover();
    const tooltip = this.page.locator(".ant-tooltip:visible").last();
    await expect(tooltip).toBeVisible({ timeout: 5000 });
    await expect(tooltip).toContainText(helpText);
  }
}
