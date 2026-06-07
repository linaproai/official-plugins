import { execFileSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse, Page } from "@host-tests/support/playwright";

import { request as playwrightRequest, expect } from "@host-tests/support/playwright";

import { test } from '@host-tests/fixtures/auth';
import { config } from '@host-tests/fixtures/config';
import { refreshPluginProjection } from '@host-tests/fixtures/plugin';
import { LoginPage } from '@host-tests/pages/LoginPage';
import { DemoDynamicPage } from '../../pages/DemoDynamicPage';
import {
  execPgSQLStatements,
  pgEscapeLiteral,
  pgIdentifier,
  queryPgScalar,
} from '@host-tests/support/postgres';
import { waitForUploadReady } from '@host-tests/support/ui';

const apiBaseURL = config.apiBaseURL;
const publicBaseURL = config.publicBaseURL;
const pluginID = "plugin-dev-dynamic-e2e";
const pluginName = "Runtime E2E Plugin";
const pluginVersion = "v0.1.0";
const hostedAssetPath = `/x-assets/${pluginID}/${pluginVersion}/index.html`;
const embeddedAssetPath = `/x-assets/${pluginID}/${pluginVersion}/mount.js`;
const iframeMenuKey = "plugin:plugin-dev-dynamic-e2e:iframe-entry";
const embeddedMenuKey = "plugin:plugin-dev-dynamic-e2e:embedded-entry";
const newWindowMenuKey = "plugin:plugin-dev-dynamic-e2e:new-window-entry";
const iframeMenuName = "运行时 iframe 示例";
const embeddedMenuName = "运行时内嵌示例";
const newWindowMenuName = "运行时新标签页示例";
const bundledRuntimePluginID = "linapro-demo-dynamic";
const bundledRuntimeDependencyPluginID = "linapro-demo-source";
const bundledRuntimeRecordTable = "plugin_linapro_demo_dynamic_record";
const bundledRuntimeAttachmentPath = "demo-record-files/";
const bundledRuntimeCronHandlerRef = `plugin:${bundledRuntimePluginID}/cron:heartbeat`;
const bundledRuntimeCronStateKey = "cron_heartbeat_count";
const bundledRuntimeLegacyArtifactPath = path.join(
  repoRoot(),
  "apps",
  "lina-plugins",
  bundledRuntimePluginID,
  "runtime",
  `${bundledRuntimePluginID}.wasm`,
);
const bundledRuntimeMenuName = "动态插件示例";
const bundledRuntimeStandalonePath =
  "/x-assets/linapro-demo-dynamic/v0.1.0/standalone.html";
const bytesPerMegabyte = 1024 * 1024;
const defaultRequestBodyLimitBytes = 8 * bytesPerMegabyte;
const fallbackUploadMaxSizeMB = 20;
const multipartRequestProbeBytes = defaultRequestBodyLimitBytes + 256 * 1024;

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
};

type PluginDynamicStateItem = {
  id: string;
  enabled?: number;
  installed?: number;
  runtimeState?: string;
};

type JobListItem = {
  id: number;
  handlerRef?: string;
  isBuiltin?: number;
  name?: string;
  status?: string;
};

type UserRouteNode = {
  component?: string;
  path?: string;
  children?: UserRouteNode[];
  meta?: {
    title?: string;
    iframeSrc?: string;
    link?: string;
    openInNewWindow?: boolean;
    query?: Record<string, string>;
  };
};

type BundledRuntimeDemoRecordListPayload = {
  list?: Array<{ title?: string }>;
  total?: number;
};

type ConfigItem = {
  id: number;
  key: string;
  value: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function expectApiSuccess<T = unknown>(
  response: APIResponse,
  message: string,
): Promise<T> {
  assertOk(response, message);
  const payload = await response.json();
  expect(payload?.code ?? 0, `${message}: ${payload?.message ?? ""}`).toBe(0);
  return unwrapApiData(payload) as T;
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
      clientType: "web",
    },
  });
  assertOk(loginResponse, "管理员登录 API 失败");

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, "未获取到 accessToken").toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function getConfigByKey(adminApi: APIRequestContext, key: string) {
  return expectApiSuccess<ConfigItem>(
    await adminApi.get(`config/key/${encodeURIComponent(key)}`),
    `查询系统参数失败: ${key}`,
  );
}

async function updateConfigValue(
  adminApi: APIRequestContext,
  configID: number,
  value: string,
) {
  await expectApiSuccess(
    await adminApi.put(`config/${configID}`, { data: { value } }),
    `更新系统参数失败: ${configID}`,
  );
}

async function listPlugins(
  adminApi: APIRequestContext,
): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  assertOk(response, "查询插件列表失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function findPlugin(adminApi: APIRequestContext, id = pluginID) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === id) ?? null;
}

async function listPluginDynamicStates(
  adminApi: APIRequestContext,
): Promise<PluginDynamicStateItem[]> {
  const response = await adminApi.get("plugins/dynamic");
  assertOk(response, "查询插件运行时状态失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function findPluginDynamicState(
  adminApi: APIRequestContext,
  id = pluginID,
) {
  const list = await listPluginDynamicStates(adminApi);
  return list.find((item) => item.id === id) ?? null;
}

async function expectPluginDynamicStateReady(
  adminApi: APIRequestContext,
  id: string,
  enabled: boolean,
) {
  await expect
    .poll(
      async () => {
        const state = await findPluginDynamicState(adminApi, id);
        return [
          state?.installed ?? 0,
          state?.enabled ?? 0,
          state?.runtimeState ?? "",
        ].join(":");
      },
      {
        message: enabled
          ? `${id} dynamic state should allow frontend entries`
          : `${id} dynamic state should hide frontend entries`,
        timeout: 20_000,
      },
    )
    .toBe(enabled ? "1:1:normal" : "1:0:normal");
}

async function waitForPluginDiscovered(
  adminApi: APIRequestContext,
  id = pluginID,
): Promise<PluginListItem> {
  await expect
    .poll(
      async () => {
        const discoveredPlugin = await findPlugin(adminApi, id);
        return discoveredPlugin?.id ?? "";
      },
      {
        message: `上传后应发现动态插件: ${id}`,
        timeout: 15_000,
      },
    )
    .toBe(id);

  const discoveredPlugin = await findPlugin(adminApi, id);
  expect(discoveredPlugin, `上传后应发现动态插件: ${id}`).toBeTruthy();
  return discoveredPlugin!;
}

async function fetchCurrentUserRoutes(
  adminApi: APIRequestContext,
): Promise<UserRouteNode[]> {
  const response = await adminApi.get("menus/all");
  assertOk(response, "查询当前用户动态路由失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

function findRouteByTitle(
  routes: UserRouteNode[],
  title: string,
): UserRouteNode | null {
  for (const route of routes) {
    if (route.meta?.title === title) {
      return route;
    }
    const matchedChild = findRouteByTitle(route.children ?? [], title);
    if (matchedChild) {
      return matchedChild;
    }
  }
  return null;
}

async function expectCurrentUserRouteVisible(
  adminApi: APIRequestContext,
  title: string,
  visible: boolean,
) {
  await expect
    .poll(
      async () => {
        const routes = await fetchCurrentUserRoutes(adminApi);
        return findRouteByTitle(routes, title) !== null;
      },
      {
        message: visible
          ? `菜单路由应包含 ${title}`
          : `菜单路由不应包含 ${title}`,
        timeout: 15_000,
      },
    )
    .toBe(visible);
}

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function runtimePluginDir() {
  return path.join(repoRoot(), "apps", "lina-plugins", pluginID);
}

function tempDir() {
  return path.join(repoRoot(), "temp");
}

function runtimeFixtureDir() {
  return path.join(tempDir(), "e2e-runtime-wasm", String(process.pid));
}

function runtimeStorageDir() {
  return path.join(tempDir(), "output");
}

function tempWasmPath() {
  return path.join(runtimeFixtureDir(), `${pluginID}.wasm`);
}

function runtimeStorageArtifactPath() {
  return path.join(runtimeStorageDir(), `${pluginID}.wasm`);
}

function bundledRuntimeStorageArtifactPath() {
  return path.join(runtimeStorageDir(), `${bundledRuntimePluginID}.wasm`);
}

function bundledRuntimeStorageRootDir() {
  return path.join(
    runtimeStorageDir(),
    ".host-services",
    "storage",
    bundledRuntimePluginID,
  );
}

function bundledRuntimeAttachmentFixtureDir() {
  return path.join(tempDir(), "e2e-linapro-demo-dynamic", String(process.pid));
}

function bundledRuntimeAttachmentFixturePath() {
  return path.join(
    bundledRuntimeAttachmentFixtureDir(),
    "linapro-demo-dynamic-note.txt",
  );
}

function cleanupBundledRuntimeAttachmentFixture() {
  rmSync(bundledRuntimeAttachmentFixtureDir(), { force: true, recursive: true });
}

function bundledRuntimeUploadProbePath() {
  return path.join(tempDir(), `${bundledRuntimePluginID}-upload-probe.wasm`);
}

function pluginHostedAssetPath(relativePath = "index.html") {
  return `/x-assets/${pluginID}/${pluginVersion}/${relativePath}`;
}

function pluginAssetURL(relativePath = "index.html") {
  return `${publicBaseURL}${pluginHostedAssetPath(relativePath)}`;
}

function cleanupRuntimePluginWorkspace() {
  rmSync(runtimePluginDir(), { force: true, recursive: true });
  rmSync(runtimeFixtureDir(), { force: true, recursive: true });
  rmSync(runtimeStorageArtifactPath(), { force: true });
}

function cleanupRuntimePluginRows() {
  const escapedId = pgEscapeLiteral(pluginID);
  execPgSQLStatements([
    `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT id FROM sys_menu WHERE menu_key IN ('${pgEscapeLiteral(iframeMenuKey)}', '${pgEscapeLiteral(embeddedMenuKey)}', '${pgEscapeLiteral(newWindowMenuKey)}'));`,
    `DELETE FROM sys_menu WHERE menu_key IN ('${pgEscapeLiteral(iframeMenuKey)}', '${pgEscapeLiteral(embeddedMenuKey)}', '${pgEscapeLiteral(newWindowMenuKey)}');`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedId}';`,
  ]);
}

function cleanupBundledRuntimeDemoData() {
  execPgSQLStatements([
    `DROP TABLE IF EXISTS ${pgIdentifier(bundledRuntimeRecordTable)};`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${pgEscapeLiteral(bundledRuntimePluginID)}';`,
    `DELETE FROM sys_plugin_state WHERE plugin_id = '${pgEscapeLiteral(bundledRuntimePluginID)}' AND state_key = '${pgEscapeLiteral(bundledRuntimeCronStateKey)}';`,
  ]);
  rmSync(bundledRuntimeStorageRootDir(), { force: true, recursive: true });
  cleanupBundledRuntimeAttachmentFixture();
}

function runtimeUploadMaxSizeMB() {
  const rawValue = queryPgScalar(
    [
      'SELECT "value"',
      "FROM sys_config",
      "WHERE \"key\" = 'sys.upload.maxSize'",
      "AND deleted_at IS NULL",
      "ORDER BY id DESC",
      "LIMIT 1;",
    ].join(" "),
  );
  const parsedValue = Number.parseInt(rawValue, 10);
  if (Number.isFinite(parsedValue) && parsedValue > 0) {
    return parsedValue;
  }
  return fallbackUploadMaxSizeMB;
}

function bundledRuntimeRecordTableExists() {
  const count = queryPgScalar(
    [
      "SELECT COUNT(*)",
      "FROM information_schema.tables",
      "WHERE table_schema = 'public'",
      `AND table_name = '${pgEscapeLiteral(bundledRuntimeRecordTable)}'`,
      ";",
    ].join(" "),
  );
  return count === "1";
}

function bundledRuntimeRecordCountByTitle(title: string) {
  if (!bundledRuntimeRecordTableExists()) {
    return 0;
  }
  const escapedTitle = pgEscapeLiteral(title);
  return Number(
    queryPgScalar(
      `SELECT COUNT(*) FROM ${pgIdentifier(bundledRuntimeRecordTable)} WHERE title = '${escapedTitle}';`,
    ),
  );
}

function bundledRuntimeCronStateCount() {
  return Number(
    queryPgScalar(
      [
        "SELECT COALESCE(MAX(state_value), '0')",
        "FROM sys_plugin_state",
        `WHERE plugin_id = '${pgEscapeLiteral(bundledRuntimePluginID)}'`,
        `AND state_key = '${pgEscapeLiteral(bundledRuntimeCronStateKey)}'`,
        ";",
      ].join(" "),
    ) || "0",
  );
}

function seedBundledRuntimePaginationRecords(recordKey: string, count: number) {
  if (!bundledRuntimeRecordTableExists()) {
    throw new Error(`${bundledRuntimeRecordTable} is missing before seeding`);
  }

  const titles = Array.from({ length: count }, (_value, index) => {
    return `动态插件分页记录-${recordKey}-${String(index + 1).padStart(2, "0")}`;
  });
  const tableName = pgIdentifier(bundledRuntimeRecordTable);
  const statements = [`DELETE FROM ${tableName};`];
  for (const [index, title] of titles.entries()) {
    const sequence = String(index + 1).padStart(2, "0");
    const escapedID = pgEscapeLiteral(`${recordKey}-${sequence}`);
    const escapedTitle = pgEscapeLiteral(title);
    const escapedContent = pgEscapeLiteral(
      `用于分页验证的动态插件示例记录 ${sequence}`,
    );
    statements.push(
      [
        `INSERT INTO ${tableName} (`,
        "id, tenant_id, title, content, attachment_name, attachment_path, created_at, updated_at",
        ") VALUES (",
        `'${escapedID}',`,
        "0,",
        `'${escapedTitle}',`,
        `'${escapedContent}',`,
        "'', '',",
        `TIMESTAMP '2026-04-17 09:00:00' + make_interval(mins => ${index}),`,
        `TIMESTAMP '2026-04-17 09:00:00' + make_interval(mins => ${index})`,
        ");",
      ].join(" "),
    );
  }
  execPgSQLStatements(statements);
  return titles;
}

async function bundledRuntimeDemoRecordListSnapshot(
  adminApi: APIRequestContext,
  pageSize = 20,
) {
  try {
    const response = await adminApi.get(
      `${publicBaseURL}/x/${bundledRuntimePluginID}/api/v1/demo-records`,
      {
        params: {
          pageNum: 1,
          pageSize,
        },
      },
    );
    if (!response.ok()) {
      return {
        ok: false,
        status: response.status(),
        titles: [] as string[],
        total: 0,
      };
    }

    const payload = unwrapApiData(
      (await response.json()) as BundledRuntimeDemoRecordListPayload,
    ) as BundledRuntimeDemoRecordListPayload;
    const records = Array.isArray(payload?.list) ? payload.list : [];
    return {
      ok: true,
      status: response.status(),
      titles: records
        .map((item) => item.title ?? "")
        .filter((title) => title.length > 0),
      total: Number(payload?.total ?? records.length),
    };
  } catch (error) {
    return {
      error: error instanceof Error ? error.message : String(error),
      ok: false,
      status: 0,
      titles: [] as string[],
      total: 0,
    };
  }
}

async function waitForBundledRuntimeDemoRecord(
  adminApi: APIRequestContext,
  title: string,
  pageSize = 20,
) {
  await expect
    .poll(
      async () => {
        const snapshot = await bundledRuntimeDemoRecordListSnapshot(
          adminApi,
          pageSize,
        );
        return snapshot.ok && snapshot.titles.includes(title);
      },
      {
        message: `等待 ${bundledRuntimePluginID} 动态路由返回记录: ${title}`,
        timeout: 20_000,
      },
    )
    .toBe(true);
}

async function waitForBundledRuntimeDemoRecordTotal(
  adminApi: APIRequestContext,
  total: number,
  pageSize = total,
) {
  await expect
    .poll(
      async () => {
        const snapshot = await bundledRuntimeDemoRecordListSnapshot(
          adminApi,
          pageSize,
        );
        return snapshot.ok ? snapshot.total : -1;
      },
      {
        message: `等待 ${bundledRuntimePluginID} 动态路由返回 ${total} 条记录`,
        timeout: 20_000,
      },
    )
    .toBe(total);
}

function countFilesRecursive(targetPath: string): number {
  if (!existsSync(targetPath)) {
    return 0;
  }
  const fileInfo = statSync(targetPath);
  if (!fileInfo.isDirectory()) {
    return 1;
  }
  return readdirSync(targetPath).reduce((total, item) => {
    return total + countFilesRecursive(path.join(targetPath, item));
  }, 0);
}

function bundledRuntimeStoredFileCount() {
  return countFilesRecursive(bundledRuntimeStorageRootDir());
}

function ensureBundledRuntimeAttachmentFixture() {
  mkdirSync(bundledRuntimeAttachmentFixtureDir(), { recursive: true });
  writeFileSync(
    bundledRuntimeAttachmentFixturePath(),
    "linapro-demo-dynamic attachment fixture",
  );
  return bundledRuntimeAttachmentFixturePath();
}

function writeULEB128(buffer: number[], value: number) {
  let current = value >>> 0;
  while (true) {
    let byte = current & 0x7f;
    current >>>= 7;
    if (current !== 0) {
      byte |= 0x80;
    }
    buffer.push(byte);
    if (current === 0) {
      return;
    }
  }
}

function buildRuntimeInstallSQL() {
  return [
    "CREATE TABLE IF NOT EXISTS plugin_runtime_e2e_log (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, created_at TIMESTAMP NULL);",
  ].join("\n");
}

function buildRuntimeUninstallSQL() {
  return ["DROP TABLE IF EXISTS plugin_runtime_e2e_log;"].join("\n");
}

function buildRuntimeManifestMenus() {
  return [
    {
      key: iframeMenuKey,
      name: iframeMenuName,
      path: hostedAssetPath,
      perms: "plugin-dev-dynamic-e2e:iframe:view",
      icon: "ant-design:appstore-outlined",
      type: "M",
      sort: -3,
      remark: "Runtime-hosted iframe entry used by Playwright verification.",
    },
    {
      key: embeddedMenuKey,
      name: embeddedMenuName,
      path: embeddedAssetPath,
      component: "system/plugin/dynamic-page",
      perms: "plugin-dev-dynamic-e2e:embedded:view",
      icon: "ant-design:deployment-unit-outlined",
      type: "M",
      sort: -2,
      query: {
        pluginAccessMode: "embedded-mount",
      },
      remark:
        "Runtime-hosted embedded mount entry used by Playwright verification.",
    },
    {
      key: newWindowMenuKey,
      name: newWindowMenuName,
      path: hostedAssetPath,
      perms: "plugin-dev-dynamic-e2e:new-window:view",
      icon: "ant-design:link-outlined",
      type: "M",
      sort: -1,
      is_frame: 1,
      remark:
        "Runtime-hosted new-window entry used by Playwright verification.",
    },
  ];
}

function appendCustomSection(buffer: number[], name: string, payload: Buffer) {
  const section: number[] = [];
  writeULEB128(section, Buffer.byteLength(name));
  section.push(...Buffer.from(name));
  section.push(...payload);

  buffer.push(0x00);
  writeULEB128(buffer, section.length);
  buffer.push(...section);
}

function buildRuntimeWasmFixture() {
  const frontendAssetPayload = Buffer.from(
    JSON.stringify([
      {
        path: "frontend/pages/index.html",
        contentBase64: Buffer.from(
          `<html><body><h1>${pluginName}</h1><p>runtime frontend asset</p></body></html>`,
        ).toString("base64"),
        contentType: "text/html; charset=utf-8",
      },
      {
        path: "frontend/pages/mount.js",
        contentBase64: Buffer.from(
          `
            export function mount(context) {
              const wrapper = document.createElement('section');
              wrapper.setAttribute('data-testid', 'runtime-embedded-root');
              const heading = document.createElement('h1');
              heading.textContent = '${pluginName}';
              const description = document.createElement('p');
              description.textContent = 'runtime embedded mount';
              const detail = document.createElement('small');
              detail.textContent = 'route=' + context.routePath;
              wrapper.append(heading, description, detail);
              context.container.replaceChildren(wrapper);
              return {
                unmount(nextContext) {
                  nextContext.container.replaceChildren();
                },
              };
            }
          `,
        ).toString("base64"),
        contentType: "text/javascript; charset=utf-8",
      },
    ]),
  );
  const manifestPayload = Buffer.from(
    JSON.stringify({
      id: pluginID,
      name: pluginName,
      version: pluginVersion,
      type: "dynamic",
      scopeNature: "tenant_aware",
      supportsMultiTenant: false,
      defaultInstallMode: "global",
      description: "Runtime plugin used by Playwright lifecycle verification.",
      public_assets: [
        {
          source: "frontend/pages",
          mount: "/",
          index: "index.html",
        },
      ],
      menus: buildRuntimeManifestMenus(),
    }),
  );
  const runtimePayload = Buffer.from(
    JSON.stringify({
      runtimeKind: "wasm",
      abiVersion: "v1",
      frontendAssetCount: 2,
      sqlAssetCount: 2,
    }),
  );
  const installSQLPayload = Buffer.from(
    JSON.stringify([
      {
        key: "001-plugin-dev-dynamic-e2e.sql",
        content: buildRuntimeInstallSQL(),
      },
    ]),
  );
  const uninstallSQLPayload = Buffer.from(
    JSON.stringify([
      {
        key: "001-plugin-dev-dynamic-e2e.sql",
        content: buildRuntimeUninstallSQL(),
      },
    ]),
  );

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(bytes, "lina.plugin.manifest", manifestPayload);
  appendCustomSection(bytes, "lina.plugin.dynamic", runtimePayload);
  appendCustomSection(
    bytes,
    "lina.plugin.frontend.assets",
    frontendAssetPayload,
  );
  appendCustomSection(bytes, "lina.plugin.install.sql", installSQLPayload);
  appendCustomSection(bytes, "lina.plugin.uninstall.sql", uninstallSQLPayload);
  return Buffer.from(bytes);
}

function ensureRuntimeWasmFixture() {
  mkdirSync(runtimeFixtureDir(), { recursive: true });
  writeFileSync(tempWasmPath(), buildRuntimeWasmFixture());
  return tempWasmPath();
}

async function loginAsAdmin(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  enabled: boolean,
  id = pluginID,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${id}/enable` : `plugins/${id}/disable`,
  );
  assertOk(response, `更新插件状态失败: enabled=${enabled}`);
}

async function ensureBundledRuntimeDependencyInstalled(
  adminApi: APIRequestContext,
) {
  const dependency = await findPlugin(adminApi, bundledRuntimeDependencyPluginID);
  if (dependency?.installed === 1) {
    return;
  }
  await installPlugin(adminApi, bundledRuntimeDependencyPluginID);
  await expect
    .poll(
      async () =>
        (await findPlugin(adminApi, bundledRuntimeDependencyPluginID))
          ?.installed ?? 0,
      {
        message: `${bundledRuntimeDependencyPluginID} should be installed before ${bundledRuntimePluginID}`,
      },
    )
    .toBe(1);
}

async function installPlugin(adminApi: APIRequestContext, id = pluginID) {
  if (id === bundledRuntimePluginID) {
    await ensureBundledRuntimeDependencyInstalled(adminApi);
  }
  await expectApiSuccess(
    await adminApi.post(`plugins/${id}/install`),
    `安装动态插件失败: ${id}`,
  );
}

async function uninstallPlugin(adminApi: APIRequestContext, id = pluginID) {
  const response = await adminApi.delete(`plugins/${id}`);
  assertOk(response, "卸载动态插件失败");
}

async function listJobs(adminApi: APIRequestContext): Promise<JobListItem[]> {
  const pageSize = 200;
  let pageNum = 1;
  const items: JobListItem[] = [];

  while (true) {
    const response = await adminApi.get("job", {
      params: {
        pageNum,
        pageSize,
      },
    });
    assertOk(response, "查询定时任务列表失败");
    const payload = unwrapApiData(await response.json());
    const currentItems = (payload?.list ?? []) as JobListItem[];
    items.push(...currentItems);

    const total = Number(payload?.total ?? 0);
    if (items.length >= total || currentItems.length === 0) {
      return items;
    }
    pageNum += 1;
  }
}

async function findJobByHandlerRef(
  adminApi: APIRequestContext,
  handlerRef: string,
  keyword?: string,
): Promise<JobListItem | null> {
  if (keyword) {
    const response = await adminApi.get("job", {
      params: {
        keyword,
        pageNum: 1,
        pageSize: 20,
      },
    });
    assertOk(response, `按关键字查询定时任务失败: ${keyword}`);
    const payload = unwrapApiData(await response.json());
    const jobs = (payload?.list ?? []) as JobListItem[];
    return jobs.find((item) => item.handlerRef === handlerRef) ?? null;
  }

  const jobs = await listJobs(adminApi);
  return jobs.find((item) => item.handlerRef === handlerRef) ?? null;
}

async function triggerJob(adminApi: APIRequestContext, jobID: number) {
  const response = await adminApi.post(`job/${jobID}/trigger`);
  assertOk(response, `立即执行定时任务失败: ${jobID}`);
  return unwrapApiData(await response.json());
}

async function uploadDynamicPluginViaAPI(
  adminApi: APIRequestContext,
  artifactPath: string,
  overwrite = false,
  paddingBytes = 0,
) {
  const multipart: Record<string, any> = {
    overwriteSupport: overwrite ? "1" : "0",
    file: {
      name: path.basename(artifactPath),
      mimeType: "application/wasm",
      buffer: readFileSync(artifactPath),
    },
  };
  if (paddingBytes > 0) {
    multipart.transportPadding = "x".repeat(paddingBytes);
  }

  const response = await adminApi.post("plugins/dynamic/package", { multipart });
  assertOk(response, `上传动态插件失败: ${artifactPath}`);
  return unwrapApiData(await response.json());
}

async function resetBundledRuntimePlugin(adminApi: APIRequestContext) {
  const plugin = await findPlugin(adminApi, bundledRuntimePluginID);
  if (!plugin) {
    return;
  }
  if (plugin.enabled === 1 || plugin.installed === 1) {
    ensureBundledRuntimePluginArtifact();
  }
  if (plugin.enabled === 1) {
    await setPluginEnabled(adminApi, false, bundledRuntimePluginID);
  }
  if (plugin.installed === 1) {
    await uninstallPlugin(adminApi, bundledRuntimePluginID);
  }
}

function ensureBundledRuntimePluginArtifact() {
  execFileSync(
    "make",
    ["wasm", `p=${bundledRuntimePluginID}`, "out=../../temp/output"],
    {
      cwd: repoRoot(),
      stdio: "inherit",
    },
  );
  rmSync(bundledRuntimeLegacyArtifactPath, { force: true });
  return bundledRuntimeStorageArtifactPath();
}

async function ensureBundledRuntimePluginSynchronized(
  adminApi: APIRequestContext,
) {
  ensureBundledRuntimePluginArtifact();
  const response = await adminApi.post("plugins/sync");
  assertOk(response, `同步 ${bundledRuntimePluginID} 插件失败`);
  await waitForPluginDiscovered(adminApi, bundledRuntimePluginID);
}

function ensureBundledRuntimeUploadFixture() {
  const sourcePath = ensureBundledRuntimePluginArtifact();
  const uploadPath = bundledRuntimeUploadProbePath();
  copyFileSync(sourcePath, uploadPath);
  return uploadPath;
}

function bundledRuntimeMultipartPaddingBytes(artifactPath: string) {
  const artifactSize = statSync(artifactPath).size;
  if (artifactSize >= multipartRequestProbeBytes) {
    return 0;
  }
  return multipartRequestProbeBytes - artifactSize;
}

async function expectPluginAssetStatus(
  page: Page,
  expectedStatus: number,
): Promise<APIResponse> {
  const response = await page.request.get(pluginAssetURL());
  expect(response.status()).toBe(expectedStatus);
  return response;
}

test.describe("TC-1 运行时 wasm 插件生命周期", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    cleanupRuntimePluginWorkspace();
    cleanupRuntimePluginRows();
    cleanupBundledRuntimeDemoData();
    rmSync(bundledRuntimeUploadProbePath(), { force: true });
    rmSync(bundledRuntimeStorageArtifactPath(), { force: true });
    rmSync(bundledRuntimeLegacyArtifactPath, { force: true });
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test.beforeEach(async () => {
    cleanupRuntimePluginWorkspace();
    cleanupRuntimePluginRows();
    cleanupBundledRuntimeDemoData();
    rmSync(bundledRuntimeUploadProbePath(), { force: true });
    await resetBundledRuntimePlugin(adminApi!);
    rmSync(bundledRuntimeStorageArtifactPath(), { force: true });
  });

  test.afterEach(async () => {
    cleanupRuntimePluginWorkspace();
    cleanupRuntimePluginRows();
    cleanupBundledRuntimeDemoData();
    rmSync(bundledRuntimeUploadProbePath(), { force: true });
    await resetBundledRuntimePlugin(adminApi!);
    rmSync(bundledRuntimeStorageArtifactPath(), { force: true });
  });

  test("TC-1a: 上传入口展示非白底主按钮和精简文案", async ({ page }) => {
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.dynamicUploadTriggerLabel()).toBeVisible();
    await expect(pluginPage.dynamicUploadTrigger).toHaveClass(
      /ant-btn-primary/,
    );

    await pluginPage.dynamicUploadTrigger.click();
    await expect(pluginPage.dynamicUploadDialog()).toBeVisible();
    await expect(pluginPage.dynamicUploadHint()).toBeVisible();
    await expect(pluginPage.dynamicOverwriteHint()).toBeVisible();
  });

  test("TC-1b~g: runtime wasm 上传、安装启用、菜单挂载和卸载资源状态完整链路", async ({
    page,
  }) => {
    test.setTimeout(120_000);
    const wasmPath = ensureRuntimeWasmFixture();
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();

    await test.step("TC-1b: 上传后宿主立即识别插件并进入可安装状态", async () => {
      await pluginPage.uploadDynamicPlugin(
        wasmPath,
        false,
        "插件包上传成功",
      );

      const pluginAfterUpload = await waitForPluginDiscovered(
        adminApi!,
        pluginID,
      );
      expect(pluginAfterUpload, "上传后应发现动态插件").toBeTruthy();
      expect(pluginAfterUpload?.installed, "上传后默认仍未安装").toBe(0);
      expect(pluginAfterUpload?.enabled ?? 0, "上传后默认仍未启用").toBe(0);
      await expect(
        page.getByRole("button", { name: /安\s*装/ }).last(),
      ).toBeVisible();
    });

    await test.step("TC-1c: 安装并启用后状态切换到已安装和已启用", async () => {
      // The action column is rendered by a detached fixed-table layer, so the
      // install/uninstall state transitions are driven through API setup while
      // the UI still validates the resulting registry and switch status.
      await installPlugin(adminApi!, pluginID);
      await page.reload();
      await pluginPage.searchByPluginId(pluginID);
      await pluginPage.setPluginEnabled(pluginID, true);

      const pluginAfterEnable = await findPlugin(adminApi!);
      expect(pluginAfterEnable?.installed).toBe(1);
      expect(pluginAfterEnable?.enabled).toBe(1);
      await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
        "aria-checked",
        "true",
      );

      const assetResponse = await expectPluginAssetStatus(page, 200);
      expect(await assetResponse.text()).toContain(pluginName);
      expect(assetResponse.headers()["content-type"]).toContain("text/html");
    });

    await page.reload();
    const routes = await fetchCurrentUserRoutes(adminApi!);

    await test.step("TC-1e: iframe 菜单在宿主内容区内嵌打开运行时托管页面", async () => {
      const iframeRoute = findRouteByTitle(routes, iframeMenuName);
      expect(iframeRoute, "启用后应生成 iframe 动态路由").toBeTruthy();
      expect(iframeRoute?.component).toBe("IFrameView");
      expect(iframeRoute?.meta?.iframeSrc).toBe(pluginHostedAssetPath());

      await pluginPage.clickSidebarMenuItem(iframeMenuName);
      await expect(
        pluginPage.pluginIframeFrame().getByRole("heading", { name: pluginName }),
      ).toBeVisible();
      await expect(
        pluginPage
          .pluginIframeFrame()
          .getByText("runtime frontend asset", { exact: true }),
      ).toBeVisible();
      expect(page.url(), "iframe 模式应保持在宿主路由下").not.toContain(
        "/plugin-assets/",
      );
    });

    await test.step("TC-1f: 新标签页菜单直接打开运行时托管页面", async () => {
      const newWindowRoute = findRouteByTitle(routes, newWindowMenuName);
      expect(newWindowRoute, "启用后应生成新标签页动态路由").toBeTruthy();
      expect(newWindowRoute?.component).toBe("BasicLayout");
      expect(newWindowRoute?.meta?.link).toBe(pluginHostedAssetPath());
      expect(newWindowRoute?.meta?.openInNewWindow).toBeTruthy();

      await pluginPage.gotoManage();
      const popupPromise = page.waitForEvent("popup");
      await pluginPage.clickSidebarMenuItem(newWindowMenuName);
      const popup = await popupPromise;
      await popup.waitForLoadState("domcontentloaded");

      expect(
        new URL(popup.url()).pathname,
        "新标签页应落到稳定的运行时托管资源路径",
      ).toBe(pluginHostedAssetPath());
      await expect(
        popup.getByRole("heading", { name: pluginName }),
      ).toBeVisible();
      await expect(
        popup.getByText("runtime frontend asset", { exact: true }),
      ).toBeVisible();
      await expect(page).toHaveURL(/\/system\/plugin(?:\/)?$/);
      await popup.close();
    });

    await test.step("TC-1g: 宿主内嵌菜单通过 runtime-page 壳挂载 ESM 入口", async () => {
      const embeddedRoute = findRouteByTitle(routes, embeddedMenuName);
      expect(embeddedRoute, "启用后应生成宿主内嵌动态路由").toBeTruthy();
      expect(embeddedRoute?.component).toBe("#/views/system/plugin/dynamic-page");
      expect(embeddedRoute?.meta?.query?.pluginAccessMode).toBe("embedded-mount");
      expect(embeddedRoute?.meta?.query?.embeddedSrc).toBe(embeddedAssetPath);

      await pluginPage.clickSidebarMenuItem(embeddedMenuName);
      await expect(pluginPage.pluginDynamicEmbeddedHost()).toBeVisible();
      await expect(page.getByRole("heading", { name: pluginName })).toBeVisible();
      await expect(
        page.getByText("runtime embedded mount", { exact: true }),
      ).toBeVisible();
      await expect(page.getByText("route=", { exact: false })).toBeVisible();
      expect(
        new URL(page.url()).pathname,
        "宿主内嵌模式应保持在宿主动态路由下",
      ).not.toContain("/plugin-assets/");
    });

    await test.step("TC-1d: 禁用并卸载后回到未安装状态且资源不可访问", async () => {
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginID);
      await pluginPage.setPluginEnabled(pluginID, false);
      await expectPluginAssetStatus(page, 404);
      await uninstallPlugin(adminApi!, pluginID);
      await expectPluginAssetStatus(page, 404);

      const pluginAfterUninstall = await findPlugin(adminApi!);
      if (pluginAfterUninstall) {
        expect(pluginAfterUninstall.installed).toBe(0);
        expect(pluginAfterUninstall.enabled).toBe(0);
      } else {
        expect(pluginAfterUninstall).toBeNull();
      }
    });
  });

  test("TC-1h: 独立的 linapro-demo-dynamic 菜单页会展示按钮并打开纯静态独立页面", async ({
    page,
  }) => {
    await ensureBundledRuntimePluginSynchronized(adminApi!);
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(bundledRuntimePluginID);

    await ensureBundledRuntimeDependencyInstalled(adminApi!);
    await pluginPage.openInstallAuthorization(bundledRuntimePluginID);
    const hostServiceAuthModal = pluginPage.hostServiceAuthModal();
    await expect(hostServiceAuthModal).toContainText("Cron");
    await expect(hostServiceAuthModal).not.toContainText("申请存储路径");
    await expect(hostServiceAuthModal).not.toContainText("申请数据表名");
    await expect(hostServiceAuthModal).not.toContainText("申请访问地址");
    await expect(hostServiceAuthModal).toContainText("动态插件心跳");
    await expect(hostServiceAuthModal).toContainText("heartbeat");
    await expect(hostServiceAuthModal).toContainText("# */10 * * * *");
    await expect(hostServiceAuthModal).toContainText("所有节点执行");
    await expect(hostServiceAuthModal).toContainText("单例执行");
    await expect(
      hostServiceAuthModal.getByTestId(
        "plugin-host-service-summary-label-cron-cron-review",
      ),
    ).toHaveCount(0);
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-list-${bundledRuntimePluginID}-cron`,
      ),
    ).toBeVisible();
    const cronItem = hostServiceAuthModal.getByTestId(
      `plugin-host-service-auth-item-${bundledRuntimePluginID}-cron-heartbeat`,
    );
    await expect(cronItem).toContainText("动态插件心跳");
    await expect(cronItem).toContainText("表达式：");
    await expect(cronItem).toContainText("调度范围：");
    await expect(cronItem).toContainText("并发策略：");
    const cronLabelFontWeight = await cronItem
      .locator("span", { hasText: "表达式：" })
      .first()
      .evaluate((node) =>
        Number.parseInt(getComputedStyle(node).fontWeight, 10),
      );
    expect(cronLabelFontWeight).toBeGreaterThanOrEqual(600);
    const hostServiceAuthText = await hostServiceAuthModal.innerText();
    expect(hostServiceAuthText.indexOf("Cron")).toBeLessThan(
      hostServiceAuthText.indexOf("运行时"),
    );
    await pluginPage.confirmHostServiceAuthorization();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);
    await page.reload();
    await pluginPage.setPluginEnabled(bundledRuntimePluginID, true);
    await page.reload();

    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(pluginPage.pluginDynamicEmbeddedHost()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicTitle()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicDescription()).toBeVisible();
    await expect(page.getByText("动态加载").first()).toBeVisible();
    await expect(
      pluginPage.pluginDemoDynamicOpenStandaloneButton(),
    ).toBeVisible();
    await pluginPage.pluginDemoDynamicOpenStandaloneButton().hover();
    await expect
      .poll(async () => {
        return pluginPage
          .pluginDemoDynamicOpenStandaloneButton()
          .evaluate((node) => window.getComputedStyle(node).cursor);
      })
      .toBe("pointer");

    const popupPromise = page.waitForEvent("popup");
    await pluginPage.pluginDemoDynamicOpenStandaloneButton().click();
    const popup = await popupPromise;
    await popup.waitForLoadState("domcontentloaded");

    expect(
      new URL(popup.url()).pathname,
      "独立页面应落到动态插件托管的静态资源地址",
    ).toBe(bundledRuntimeStandalonePath);
    await expect(
      popup.getByTestId("linapro-demo-dynamic-standalone"),
    ).toBeVisible();
    await expect(
      popup.getByRole("heading", {
        name: /动态插件独立页面|Dynamic Plugin Standalone Page/,
      }),
    ).toBeVisible();
    await expect(
      popup.getByText(
        /当前页面由 linapro-demo-dynamic 直接以托管静态资源形式提供|This page is served directly by linapro-demo-dynamic as a hosted static asset/,
      ),
    ).toBeVisible();
    await popup.close();
  });

  test("TC-1i: 运行时产物被手动删除后列表仍保留条目、菜单隐藏且拒绝同版本重新上传", async ({
    page,
  }) => {
    const wasmPath = ensureRuntimeWasmFixture();
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.uploadDynamicPlugin(wasmPath);
    await waitForPluginDiscovered(adminApi!, pluginID);
    await installPlugin(adminApi!, pluginID);
    await page.reload();
    await pluginPage.setPluginEnabled(pluginID, true);
    await page.reload();

    rmSync(runtimeStorageArtifactPath(), { force: true });

    await page.reload();
    await pluginPage.searchByPluginId(pluginID);
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await pluginPage.expectSidebarMenuHidden(iframeMenuName);
    await pluginPage.expectSidebarMenuHidden(embeddedMenuName);
    await pluginPage.expectSidebarMenuHidden(newWindowMenuName);

    const pluginAfterArtifactRemoval = await findPlugin(adminApi!);
    expect(
      pluginAfterArtifactRemoval,
      "删除运行时产物后插件列表仍应保留该 runtime 条目",
    ).toBeTruthy();
    expect(pluginAfterArtifactRemoval?.installed).toBe(1);
    expect(pluginAfterArtifactRemoval?.enabled).toBe(1);

    const duplicateUploadResponse = await adminApi!.post(
      "plugins/dynamic/package",
      {
        multipart: {
          overwriteSupport: "0",
          file: {
            name: path.basename(wasmPath),
            mimeType: "application/wasm",
            buffer: readFileSync(wasmPath),
          },
        },
      },
    );
    assertOk(duplicateUploadResponse, "同版本动态插件重新上传请求失败");
    const duplicateUploadPayload = await duplicateUploadResponse.json();
    expect(duplicateUploadPayload?.code).not.toBe(0);
    expect(duplicateUploadPayload?.message).toContain("higher version");

    const pluginAfterDuplicateUpload = await findPlugin(adminApi!);
    expect(
      pluginAfterDuplicateUpload,
      "拒绝同版本上传后仍应保留已安装插件条目",
    ).toBeTruthy();
    expect(pluginAfterDuplicateUpload?.installed).toBe(1);
    expect(pluginAfterDuplicateUpload?.enabled).toBe(1);
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
  });

  test("TC-1j: 启用 linapro-demo-dynamic 后固定前缀动态路由返回真实 Wasm bridge 响应", async ({
    page,
  }) => {
    await ensureBundledRuntimePluginSynchronized(adminApi!);
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(bundledRuntimePluginID);

    await installPlugin(adminApi!, bundledRuntimePluginID);
    await page.reload();
    await pluginPage.setPluginEnabled(bundledRuntimePluginID, true);

    const response = await adminApi!.get(
      `/x/${bundledRuntimePluginID}/api/v1/backend-summary`,
    );
    assertOk(response, "请求动态插件固定前缀路由失败");
    expect(response.status()).toBe(200);
    expect(response.headers()["x-lina-plugin-bridge"]).toBe(
      bundledRuntimePluginID,
    );
    expect(response.headers()["x-lina-plugin-middleware"]).toBe(
      "backend-summary",
    );

    const payload = await response.json();
    expect(payload.message).toContain(
      "linapro-demo-dynamic Wasm bridge runtime",
    );
    expect(payload.pluginId).toBe(bundledRuntimePluginID);
    expect(payload.publicPath).toBe(
      `/x/${bundledRuntimePluginID}/api/v1/backend-summary`,
    );
    expect(payload.access).toBe("login");
    expect(payload.permission).toBe("linapro-demo-dynamic:backend:view");
    expect(payload.authenticated).toBeTruthy();
    expect(payload.username).toBe(config.adminUser);
    expect(payload.isSuperAdmin).toBeTruthy();
  });

  test("TC-1o: linapro-demo-dynamic 安装后其内置定时任务立即出现在任务管理中，启用后可手动执行", async ({
    page,
  }) => {
    await ensureBundledRuntimePluginSynchronized(adminApi!);
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(bundledRuntimePluginID);

    await ensureBundledRuntimeDependencyInstalled(adminApi!);
    await pluginPage.openInstallAuthorization(bundledRuntimePluginID);
    await pluginPage.confirmHostServiceAuthorization();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);

    await expect
      .poll(async () => {
        const job = await findJobByHandlerRef(
          adminApi!,
          bundledRuntimeCronHandlerRef,
          "动态插件心跳",
        );
        return job?.status ?? "";
      })
      .toBe("paused_by_plugin");

    const installedJob = await findJobByHandlerRef(
      adminApi!,
      bundledRuntimeCronHandlerRef,
      "动态插件心跳",
    );
    expect(installedJob, "安装后应投影出动态插件内置定时任务").toBeTruthy();
    expect(installedJob?.isBuiltin).toBe(1);

    await page.reload();
    await pluginPage.setPluginEnabled(bundledRuntimePluginID, true);

    await expect
      .poll(async () => {
        const job = await findJobByHandlerRef(
          adminApi!,
          bundledRuntimeCronHandlerRef,
          "动态插件心跳",
        );
        return job?.status ?? "";
      })
      .toBe("enabled");

    const enabledJob = await findJobByHandlerRef(
      adminApi!,
      bundledRuntimeCronHandlerRef,
      "动态插件心跳",
    );
    expect(enabledJob, "启用后应保留动态插件内置定时任务").toBeTruthy();
    expect(enabledJob?.status).toBe("enabled");

    const cronStateBeforeTrigger = bundledRuntimeCronStateCount();
    await triggerJob(adminApi!, enabledJob!.id);
    await expect
      .poll(
        () => bundledRuntimeCronStateCount(),
        {
          message: "手动执行后动态插件心跳计数应增长",
          timeout: 20_000,
        },
      )
      .toBeGreaterThan(cronStateBeforeTrigger);
  });

  test("TC-1k: linapro-demo-dynamic 示例记录支持 CRUD，并在禁用与卸载时按选项保留或清理数据附件", async ({
    page,
  }) => {
    // The CRUD + dual-uninstall lifecycle runs three full install/enable
    // cycles plus a file upload, well past the default 60s test budget.
    test.setTimeout(300_000);
    await ensureBundledRuntimePluginSynchronized(adminApi!);
    const recordTitle = `动态插件示例记录-${Date.now()}`;
    const updatedRecordTitle = `${recordTitle}-更新`;
    const cleanupRecordTitle = `${recordTitle}-清理`;
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(bundledRuntimePluginID);

    const confirmBundledRuntimeInstall = async () => {
      await ensureBundledRuntimeDependencyInstalled(adminApi!);
      await pluginPage.openInstallAuthorization(bundledRuntimePluginID);
      await pluginPage.confirmHostServiceAuthorization();
    };

    const setBundledRuntimeEnabled = async (enabled: boolean) => {
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(bundledRuntimePluginID);
      await pluginPage.setPluginEnabled(bundledRuntimePluginID, enabled);
      await expect
        .poll(
          async () =>
            (await findPlugin(adminApi!, bundledRuntimePluginID))?.enabled ?? 0,
        )
        .toBe(enabled ? 1 : 0);
      await expectPluginDynamicStateReady(
        adminApi!,
        bundledRuntimePluginID,
        enabled,
      );
      await expectCurrentUserRouteVisible(
        adminApi!,
        bundledRuntimeMenuName,
        enabled,
      );
      await refreshPluginProjection(page);
    };

    await confirmBundledRuntimeInstall();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);
    await setBundledRuntimeEnabled(true);
    await waitForBundledRuntimeDemoRecord(
      adminApi!,
      "Dynamic Plugin SQL Demo Record",
    );

    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(pluginPage.pluginDemoDynamicTitle()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicRecordGrid()).toBeVisible();
    await expect(
      pluginPage.pluginDemoDynamicRecordRow("Dynamic Plugin SQL Demo Record"),
    ).toBeVisible();

    await pluginPage.createPluginDemoDynamicRecord({
      title: recordTitle,
      content: "首次创建的动态插件示例记录",
      attachmentPath: ensureBundledRuntimeAttachmentFixture(),
    });
    expect(bundledRuntimeRecordCountByTitle(recordTitle)).toBe(1);
    expect(bundledRuntimeStoredFileCount()).toBeGreaterThan(0);

    await pluginPage.updatePluginDemoDynamicRecord(recordTitle, {
      title: updatedRecordTitle,
      content: "更新后的动态插件示例记录内容",
    });
    expect(bundledRuntimeRecordCountByTitle(recordTitle)).toBe(0);
    expect(bundledRuntimeRecordCountByTitle(updatedRecordTitle)).toBe(1);

    await pluginPage.gotoManage();
    await setBundledRuntimeEnabled(false);
    await pluginPage.expectSidebarMenuHidden(bundledRuntimeMenuName);
    expect(bundledRuntimeRecordCountByTitle(updatedRecordTitle)).toBe(1);

    await setBundledRuntimeEnabled(true);
    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(
      pluginPage.pluginDemoDynamicRecordRow(updatedRecordTitle),
    ).toBeVisible();

    await pluginPage.gotoManage();
    await pluginPage.uninstallPluginWithOptions(bundledRuntimePluginID, false);
    expect(bundledRuntimeRecordCountByTitle(updatedRecordTitle)).toBe(1);
    expect(bundledRuntimeStoredFileCount()).toBeGreaterThan(0);

    await confirmBundledRuntimeInstall();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);
    await setBundledRuntimeEnabled(true);
    await waitForBundledRuntimeDemoRecord(adminApi!, updatedRecordTitle);
    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(
      pluginPage.pluginDemoDynamicRecordRow(updatedRecordTitle),
    ).toBeVisible();

    await pluginPage.deletePluginDemoDynamicRecord(updatedRecordTitle);
    expect(bundledRuntimeRecordCountByTitle(updatedRecordTitle)).toBe(0);
    expect(bundledRuntimeStoredFileCount()).toBe(0);

    await pluginPage.createPluginDemoDynamicRecord({
      title: cleanupRecordTitle,
      content: "用于验证卸载清理的数据与附件",
      attachmentPath: ensureBundledRuntimeAttachmentFixture(),
    });
    expect(bundledRuntimeRecordCountByTitle(cleanupRecordTitle)).toBe(1);
    expect(bundledRuntimeStoredFileCount()).toBeGreaterThan(0);

    await pluginPage.gotoManage();
    await pluginPage.uninstallPluginWithOptions(bundledRuntimePluginID, true);
    expect(bundledRuntimeRecordTableExists()).toBeFalsy();
    expect(bundledRuntimeStoredFileCount()).toBe(0);

    await confirmBundledRuntimeInstall();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);
    await setBundledRuntimeEnabled(true);
    await waitForBundledRuntimeDemoRecord(
      adminApi!,
      "Dynamic Plugin SQL Demo Record",
    );
    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(
      pluginPage.pluginDemoDynamicRecordRow("Dynamic Plugin SQL Demo Record"),
    ).toBeVisible();
    await expect(
      pluginPage.pluginDemoDynamicRecordRow(cleanupRecordTitle),
    ).toHaveCount(0);
  });

  test("TC-1l: linapro-demo-dynamic 示例记录列表支持分页浏览并同步更新区间摘要", async ({
    page,
  }) => {
    await ensureBundledRuntimePluginSynchronized(adminApi!);
    const paginationRecordKey = `${Date.now()}`;
    let seededTitles: string[] = [];
    await loginAsAdmin(page);

    const pluginPage = new DemoDynamicPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(bundledRuntimePluginID);

    await ensureBundledRuntimeDependencyInstalled(adminApi!);
    await pluginPage.openInstallAuthorization(bundledRuntimePluginID);
    await pluginPage.confirmHostServiceAuthorization();
    await expect
      .poll(
        async () =>
          (await findPlugin(adminApi!, bundledRuntimePluginID))?.installed ?? 0,
      )
      .toBe(1);
    await page.reload();
    await pluginPage.setPluginEnabled(bundledRuntimePluginID, true);
    await expectCurrentUserRouteVisible(
      adminApi!,
      bundledRuntimeMenuName,
      true,
    );
    await page.reload();

    seededTitles = seedBundledRuntimePaginationRecords(paginationRecordKey, 12);
    await waitForBundledRuntimeDemoRecordTotal(adminApi!, seededTitles.length);
    const newestTitle = seededTitles[seededTitles.length - 1];
    const oldestTitle = seededTitles[0];

    await pluginPage.clickSidebarMenuItem(bundledRuntimeMenuName);
    await expect(pluginPage.pluginDemoDynamicRecordGrid()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicRecordPagination()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicPaginationSummary()).toHaveText(
      "第 1 / 2 页，显示第 1-10 条，共 12 条",
    );
    await expect(pluginPage.pluginDemoDynamicPaginationPage(1)).toBeDisabled();
    await expect(pluginPage.pluginDemoDynamicPaginationPage(2)).toBeEnabled();
    await expect(pluginPage.pluginDemoDynamicRecordRow(newestTitle)).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicRecordRow(oldestTitle)).toHaveCount(
      0,
    );

    await pluginPage.pluginDemoDynamicPaginationPage(2).click();
    await expect(pluginPage.pluginDemoDynamicPaginationSummary()).toHaveText(
      "第 2 / 2 页，显示第 11-12 条，共 12 条",
    );
    await expect(pluginPage.pluginDemoDynamicPaginationPage(2)).toBeDisabled();
    await expect(pluginPage.pluginDemoDynamicRecordRow(oldestTitle)).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicRecordRow(newestTitle)).toHaveCount(
      0,
    );

    await pluginPage.pluginDemoDynamicPaginationPrevButton().click();
    await expect(pluginPage.pluginDemoDynamicPaginationSummary()).toHaveText(
      "第 1 / 2 页，显示第 1-10 条，共 12 条",
    );
    await expect(pluginPage.pluginDemoDynamicRecordRow(newestTitle)).toBeVisible();
  });

  test("TC-1m: linapro-demo-dynamic 在 multipart 请求体超过默认 8MB 时仍按上传参数上限完成上传", async () => {
    const artifactPath = ensureBundledRuntimeUploadFixture();
    const paddingBytes = bundledRuntimeMultipartPaddingBytes(artifactPath);
    const minimumUploadMaxSizeMB =
      Math.ceil(statSync(artifactPath).size / bytesPerMegabyte) + 1;
    const uploadMaxSizeConfig = await getConfigByKey(
      adminApi!,
      "sys.upload.maxSize",
    );
    expect(
      statSync(artifactPath).size + paddingBytes,
      "探针请求体应超过默认 8MB 门槛，才能覆盖本次回归场景",
    ).toBeGreaterThan(defaultRequestBodyLimitBytes);

    if (runtimeUploadMaxSizeMB() < minimumUploadMaxSizeMB) {
      await updateConfigValue(
        adminApi!,
        uploadMaxSizeConfig.id,
        `${minimumUploadMaxSizeMB}`,
      );
      await expect
        .poll(() => runtimeUploadMaxSizeMB(), {
          timeout: 10000,
          message: "sys.upload.maxSize should accept the runtime artifact",
        })
        .toBeGreaterThanOrEqual(minimumUploadMaxSizeMB);
    }

    try {
      const uploadPayload = await uploadDynamicPluginViaAPI(
        adminApi!,
        artifactPath,
        true,
        paddingBytes,
      );

      expect(uploadPayload?.id).toBe(bundledRuntimePluginID);
      expect(uploadPayload?.installed ?? 0).toBe(0);
      expect(uploadPayload?.enabled ?? 0).toBe(0);

      const pluginAfterUpload = await findPlugin(adminApi!, bundledRuntimePluginID);
      expect(pluginAfterUpload, "上传后应保留 linapro-demo-dynamic 记录").toBeTruthy();
      expect(pluginAfterUpload?.installed ?? 0).toBe(0);
      expect(pluginAfterUpload?.enabled ?? 0).toBe(0);
    } finally {
      if (`${runtimeUploadMaxSizeMB()}` !== uploadMaxSizeConfig.value) {
        await updateConfigValue(
          adminApi!,
          uploadMaxSizeConfig.id,
          uploadMaxSizeConfig.value,
        );
        await expect
          .poll(() => runtimeUploadMaxSizeMB(), {
            timeout: 10000,
            message: "sys.upload.maxSize should be restored",
          })
          .toBe(Number.parseInt(uploadMaxSizeConfig.value, 10));
      }
    }
  });

  test("TC-1n: 超过上传上限的 multipart 请求返回文件过大提示而不是 500", async ({
    page,
  }) => {
    await loginAsAdmin(page);

    const uploadMaxSizeMB = runtimeUploadMaxSizeMB();
    const oversizedProbeBytes = uploadMaxSizeMB * bytesPerMegabyte + 2 * bytesPerMegabyte;
    const expectedMessage = `文件大小不能超过${uploadMaxSizeMB}MB`;
    const oversizedBuffer = Buffer.alloc(oversizedProbeBytes, 0x61);
    const oversizedFilePath = path.join(
      repoRoot(),
      "temp",
      "runtime-e2e",
      "linapro-demo-dynamic-oversized.wasm",
    );
    mkdirSync(path.dirname(oversizedFilePath), { recursive: true });
    writeFileSync(oversizedFilePath, oversizedBuffer);

    try {
      const apiResponse = await adminApi!.post("plugins/dynamic/package", {
        multipart: {
          overwriteSupport: "0",
          file: {
            name: "linapro-demo-dynamic-oversized.wasm",
            mimeType: "application/wasm",
            buffer: oversizedBuffer,
          },
        },
      });
      const apiPayload = (await apiResponse.json()) as {
        code?: number;
        message?: string;
      };

      expect(apiResponse.status(), "超限上传不应再返回服务器 500").toBe(200);
      expect(apiPayload.code ?? 0, "超限上传应返回业务错误码").not.toBe(0);
      expect(apiPayload.message ?? "").toContain(expectedMessage);

      const pluginPage = new DemoDynamicPage(page);
      await pluginPage.gotoManage();
      await pluginPage.dynamicUploadTrigger.click();
      await expect(pluginPage.dynamicUploadDialog()).toBeVisible();

      const [fileChooser] = await Promise.all([
        page.waitForEvent("filechooser"),
        pluginPage.dynamicUploadDragger.click(),
      ]);
      await fileChooser.setFiles(oversizedFilePath);
      await waitForUploadReady(pluginPage.dynamicUploadDialog());

      const uploadResponsePromise = page.waitForResponse(
        (response) =>
          response.url().includes("/plugins/dynamic/package") &&
          response.request().method() === "POST",
        { timeout: 30000 },
      );

      await pluginPage.dynamicUploadConfirmButton().click();

      const uploadResponse = await uploadResponsePromise;
      expect(uploadResponse.status(), "超限上传不应再返回服务器 500").toBe(200);
      await expect(pluginPage.messageNotice(expectedMessage)).toBeVisible();
      await expect(pluginPage.dynamicUploadDialog()).toBeVisible();
      await expect(pluginPage.uploadSuccessDialog()).toHaveCount(0);
    } finally {
      rmSync(oversizedFilePath, { force: true });
    }
  });
});
