import type { Page } from "@playwright/test";

/**
 * Discord OIDC plugin page object. Host LoginPage must not hard-code this plugin.
 */
export class DiscordOidcPage {
  constructor(private page: Page) {}

  get loginEntry() {
    return this.page.getByTestId("linapro-oidc-discord-entry");
  }

  get loginEntryButton() {
    return this.page.getByTestId("linapro-oidc-discord-entry-button");
  }
}
