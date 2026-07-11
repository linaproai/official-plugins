import type { Page } from "@playwright/test";

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
}
