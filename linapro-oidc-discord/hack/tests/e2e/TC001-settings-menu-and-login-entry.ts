/**
 * TC001 linapro-oidc-discord 菜单挂载与登录入口
 *
 * - 依赖 linapro-extlogin-core 的授权登录目录
 * - 设置菜单挂在该目录下
 * - 未配置 Client 凭证时登录入口 fail-closed 回登录页
 */
import type { APIRequestContext } from "@playwright/test";

import { expect, test } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  prepareSourcePluginsBaseline,
} from "@host-tests/fixtures/plugin";
import {
  config,
  pluginApiPath,
  workspacePath,
} from "@host-tests/fixtures/config";
import { LoginPage } from "@host-tests/pages/LoginPage";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { waitForRouteReady } from "@host-tests/support/ui";

import { DiscordOidcPage } from "../pages/DiscordOidcPage";

const ownerPluginID = "linapro-extlogin-core";
const pluginID = "linapro-oidc-discord";

type MenuNode = {
  children?: MenuNode[];
  name?: string;
  path?: string;
  perms?: string;
};

function findByName(list: MenuNode[], name: string | RegExp): MenuNode | null {
  for (const item of list) {
    const title = String(item.name ?? "");
    if (
      (typeof name === "string" && title === name) ||
      (name instanceof RegExp && name.test(title))
    ) {
      return item;
    }
    const nested = findByName(item.children ?? [], name);
    if (nested) {
      return nested;
    }
  }
  return null;
}

function findByPath(list: MenuNode[], path: string): MenuNode | null {
  for (const item of list) {
    if (String(item.path ?? "") === path) {
      return item;
    }
    const nested = findByPath(item.children ?? [], path);
    if (nested) {
      return nested;
    }
  }
  return null;
}

async function fetchMenuTree(
  api: APIRequestContext,
  locale?: string,
): Promise<MenuNode[]> {
  const response = await api.get(
    "menu",
    locale ? { headers: { "Accept-Language": locale } } : undefined,
  );
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  const data = body?.data ?? body;
  if (Array.isArray(data?.list)) {
    return data.list as MenuNode[];
  }
  if (Array.isArray(data)) {
    return data as MenuNode[];
  }
  return [];
}

test.describe("TC001 linapro-oidc-discord 设置菜单与登录入口", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([ownerPluginID, pluginID]);
  });

  test("TC001a: 设置菜单挂在授权登录目录下", async () => {
    const api = await createAdminApiContext();
    try {
      const tree = await fetchMenuTree(api);
      const authLogin = findByName(tree, /授权登录|Auth Login/i);
      expect(authLogin, "依赖的授权登录目录应存在").not.toBeNull();

      const discord = (authLogin?.children ?? []).find(
        (node) =>
          node.path === "linapro-oidc-discord-settings" ||
          node.perms === "linapro-oidc-discord:settings:view" ||
          /Discord 登录|Discord Login/i.test(String(node.name ?? "")),
      );
      expect(discord, "Discord 登录应挂在授权登录下").toBeTruthy();
    } finally {
      await api.dispose();
    }
  });

  test("TC001b: 侧边栏可从授权登录进入 Discord 设置入口", async ({
    adminPage,
  }) => {
    const layout = new MainLayout(adminPage);
    await adminPage.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(adminPage);
    await layout.expandSidebarGroup(/授权登录|Auth Login/i);
    await expect(
      layout.sidebarMenuItem(/Discord 登录|Discord Login/i),
    ).toBeVisible();
  });

  test("TC001d: 菜单标题随 Accept-Language 本地化", async () => {
    const api = await createAdminApiContext();
    const settingsPath = "linapro-oidc-discord-settings";
    try {
      const enTree = await fetchMenuTree(api, "en-US");
      const enNode = findByPath(enTree, settingsPath);
      expect(enNode, "en-US 应返回 Discord 设置菜单").not.toBeNull();
      expect(enNode?.name).toBe("Discord Login");

      const zhTree = await fetchMenuTree(api, "zh-CN");
      const zhNode = findByPath(zhTree, settingsPath);
      expect(zhNode, "zh-CN 应返回 Discord 设置菜单").not.toBeNull();
      expect(zhNode?.name).toBe("Discord 登录");
    } finally {
      await api.dispose();
    }
  });

  test("TC001e: 登录入口为「其他登录方式」下的平台图标", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    if ((await page.locator("html").getAttribute("lang")) !== "zh-CN") {
      await loginPage.switchLanguage("简体中文");
    }

    const discordPage = new DiscordOidcPage(page);
    await expect(discordPage.loginEntry).toBeVisible();
    // Social region divider + icon-sized control (not full-width protocol button).
    await expect(
      page.getByText("其他登录方式", { exact: true }).first(),
    ).toBeVisible();
    await expect(loginPage.socialAuthRegion).toBeVisible();
    await expect(loginPage.socialAuthSlot).toBeVisible();

    const layout = await discordPage.getLoginEntryLayout();
    expect(layout.isIconSized).toBe(true);
    // Icon must not stretch to the social region width (full-width protocol buttons do).
    expect(layout.regionWidth).toBeGreaterThan(200);
    expect(layout.buttonWidth).toBeLessThan(layout.regionWidth * 0.25);
  });

  test("TC001c: 未配置凭证时登录入口 fail-closed 回登录页", async ({
    browser,
  }) => {
    const api = await createAdminApiContext();
    const settingsURL = `${config.publicBaseURL}${pluginApiPath(pluginID, "settings")}`;
    const original = await api.get(settingsURL);
    expect(original.ok()).toBeTruthy();
    const originalBody = await original.json();
    const previousClientId = String(
      originalBody?.data?.settings?.clientId ?? "",
    );

    const clearResponse = await api.put(settingsURL, {
      data: {
        clientId: "",
        clientSecret: "",
        redirectUrl: "",
        enableBackendRedirect: false,
        defaultBackendRedirect: "",
        backendRedirects: "",
        allowAutoProvision: false,
      },
    });
    expect(clearResponse.ok()).toBeTruthy();

    const context = await browser.newContext({
      baseURL: config.baseURL,
    });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);
    const discordPage = new DiscordOidcPage(page);
    try {
      await loginPage.goto();
      await expect(discordPage.loginEntry).toBeVisible();

      await Promise.all([
        page.waitForURL(/\/admin\/auth\/login/, { timeout: 15000 }),
        discordPage.loginEntryButton.click(),
      ]);

      await expect(page).toHaveURL(/\/admin\/auth\/login/);
      expect(page.url()).not.toContain("discord.com");
      await expect(
        page.getByText("第三方登录尚未完成配置", { exact: false }),
      ).toBeVisible({ timeout: 10000 });
    } finally {
      await context.close();
      await api.put(settingsURL, {
        data: {
          clientId: previousClientId,
          clientSecret: "",
          redirectUrl: "",
          enableBackendRedirect: false,
          defaultBackendRedirect: "",
          backendRedirects: "",
          allowAutoProvision: false,
        },
      });
      await api.dispose();
    }
  });
});
