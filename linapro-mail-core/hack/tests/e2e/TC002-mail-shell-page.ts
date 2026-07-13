import { test, expect, type Page, type Locator } from '@host-tests/fixtures/auth';
import { workspacePath } from '@host-tests/fixtures/config';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { PluginPage } from '@host-tests/pages/PluginPage';
import { waitForRouteReady } from '@host-tests/support/ui';
import path from 'node:path';
import { mkdirSync } from 'node:fs';

/**
 * TC002 mail-core single-account settings page.
 * Full visual coverage via screenshots + key form/save assertions.
 */
const pluginID = 'linapro-mail-core';
const mailSettingsPath = /linapro-mail-core-settings/i;
const mailSettingsWorkspace = workspacePath('/setting/linapro-mail-core-settings');

test.describe.configure({ timeout: 180_000 });

function evidenceDir() {
  const day = new Date().toISOString().slice(0, 10).replaceAll('-', '');
  const dir = path.join(process.cwd(), '../../temp', day);
  mkdirSync(dir, { recursive: true });
  return dir;
}

function evidencePath(name: string) {
  const stamp = new Date()
    .toISOString()
    .replaceAll(/[-:]/g, '')
    .replace('T', '-')
    .slice(0, 15);
  return path.join(evidenceDir(), `${stamp}-${name}.png`);
}

async function shot(page: Page, name: string) {
  const file = evidencePath(name);
  await page.screenshot({ path: file, fullPage: true });
  return file;
}

/** Resolve the editable control under ant-design Input / InputNumber wrappers. */
function fieldInput(page: Page, testId: string): Locator {
  const root = page.getByTestId(testId);
  return root.locator('input').first().or(root);
}

/** Select an Ant Design option by visible label under the inbound-kind control. */
async function selectInboundKind(page: Page, label: 'IMAP' | 'POP3' | '无（仅发信）') {
  const select = page.getByTestId('mail-settings-inbound-kind');
  await select.click();
  const option = page.locator('.ant-select-dropdown:visible .ant-select-item-option', {
    hasText: new RegExp(`^${label}$`),
  });
  await expect(option.first()).toBeVisible({ timeout: 5_000 });
  await option.first().click();
  // Ensure the selection closed and value stuck before continuing.
  await expect(page.locator('.ant-select-dropdown:visible')).toHaveCount(0, {
    timeout: 5_000,
  }).catch(() => undefined);
  await expect(select).toContainText(label, { timeout: 5_000 });
}

async function openMailSettings(page: Page) {
  await page.goto(mailSettingsWorkspace);
  await waitForRouteReady(page, 30_000);
  await expect(page).toHaveURL(mailSettingsPath);
  await expect(page.getByTestId('mail-settings-page')).toBeVisible({
    timeout: 30_000,
  });
  // Wait for settings GET to finish (card loading spinner gone).
  await expect(page.locator('[data-testid="mail-settings-page"] .ant-spin-spinning')).toHaveCount(
    0,
    { timeout: 30_000 },
  );
  await expect(page.getByTestId('mail-settings-username')).toBeVisible({
    timeout: 15_000,
  });
}

test.describe('TC-2 mail-core management settings page', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
  });

  test('TC-2a: full-page screenshots and layout review points', async ({
    adminPage,
  }) => {
    await openMailSettings(adminPage);

    // 1) Initial empty / loaded layout
    const initialShot = await shot(adminPage, 'mail-settings-01-initial');

    const pageRoot = adminPage.getByTestId('mail-settings-page');
    await expect(pageRoot.locator('.ant-card').first()).toBeVisible();
    await expect(pageRoot.locator('.ant-form').first()).toBeVisible();

    // Page tip: short description only, zh-CN translated (not i18n key / not old long copy).
    const pageTip = pageRoot.getByTestId('mail-settings-tip');
    await expect(pageTip).toBeVisible();
    await expect(pageTip).toHaveText('配置平台唯一邮件账号，用于系统通知与邮件能力。');
    await expect(pageTip).not.toContainText('顶部填写账号与密码');

    // Core controls — account/password/from at top; credentials shared (not per protocol).
    await expect(adminPage.getByTestId('mail-settings-username')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-password')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-from')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-name')).toHaveCount(0);
    await expect(adminPage.getByTestId('mail-settings-smtp-host')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-smtp-port')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-smtp-tls')).toBeVisible();
    // SMTP TLS label must expose Ant Design help tooltip (question-mark icon).
    const smtpTlsItem = adminPage.locator('.ant-form-item', {
      has: adminPage.getByTestId('mail-settings-smtp-tls'),
    });
    await expect(smtpTlsItem.locator('.ant-form-item-tooltip').first()).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-inbound-kind')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-test')).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-save')).toBeVisible();
    // No per-protocol credential fields or "当前协议" hint.
    await expect(adminPage.getByTestId('mail-settings-inbound-username')).toHaveCount(0);
    await expect(adminPage.getByTestId('mail-settings-inbound-password')).toHaveCount(0);
    await expect(adminPage.getByTestId('mail-settings-inbound-kind-hint')).toHaveCount(0);
    await expect(adminPage.getByText(/当前协议/)).toHaveCount(0);

    // zh-CN labels (i18n resolved, not raw keys)
    await expect(adminPage.getByText('账号').first()).toBeVisible();
    await expect(adminPage.getByText('密码').first()).toBeVisible();
    await expect(adminPage.getByText('发件地址').first()).toBeVisible();
    await expect(adminPage.getByText('账号名称')).toHaveCount(0);
    await expect(adminPage.getByText('用户名')).toHaveCount(0);
    await expect(adminPage.getByText('SMTP（发信）').first()).toBeVisible();
    await expect(adminPage.getByText('收信（可选）').first()).toBeVisible();
    // Prefer testids for Ant Design buttons (role name matching can be flaky).
    // Ant Design inserts spacing between CJK chars on primary buttons (保 存 设 置).
    await expect(adminPage.getByTestId('mail-settings-test')).toContainText(
      /测\s*试\s*连\s*接|Test connection/i,
    );
    await expect(adminPage.getByTestId('mail-settings-save')).toContainText(
      /保\s*存\s*设\s*置|Save settings/i,
    );

    // 2) Fill shared credentials + SMTP
    // Account required; From left empty to default to account; optional override asserted later.
    const suffix = Date.now();
    const account = `user-${suffix}@example.com`;
    await fieldInput(adminPage, 'mail-settings-username').fill(account);
    await fieldInput(adminPage, 'mail-settings-password').fill('secret-password');
    // Leave from empty — backend defaults From to account.
    await fieldInput(adminPage, 'mail-settings-from').fill('');
    await fieldInput(adminPage, 'mail-settings-smtp-host').fill('smtp.example.com');
    await fieldInput(adminPage, 'mail-settings-smtp-port').fill('587');
    const filledOutboundShot = await shot(adminPage, 'mail-settings-02-filled-outbound');

    // 3) Expand inbound IMAP fields (host/port/TLS only; credentials stay shared).
    await selectInboundKind(adminPage, 'IMAP');
    await expect(adminPage.getByTestId('mail-settings-inbound-host')).toBeVisible();
    await fieldInput(adminPage, 'mail-settings-inbound-host').fill('imap.example.com');
    await expect
      .poll(async () => fieldInput(adminPage, 'mail-settings-inbound-port').inputValue(), {
        timeout: 5_000,
      })
      .toBe('993');
    // Still no inbound credential fields after enabling IMAP.
    await expect(adminPage.getByTestId('mail-settings-inbound-username')).toHaveCount(0);
    await expect(adminPage.getByTestId('mail-settings-inbound-password')).toHaveCount(0);
    const filledInboundShot = await shot(adminPage, 'mail-settings-03-filled-inbound');

    // 3b) Switch to POP3 — port must update (regression for stuck IMAP UI).
    await selectInboundKind(adminPage, 'POP3');
    await expect
      .poll(async () => fieldInput(adminPage, 'mail-settings-inbound-port').inputValue(), {
        timeout: 5_000,
      })
      .toBe('995');
    await fieldInput(adminPage, 'mail-settings-inbound-host').fill('pop.example.com');
    const filledPop3Shot = await shot(adminPage, 'mail-settings-03b-filled-pop3');

    // Switch back to IMAP for the rest of the save flow.
    await selectInboundKind(adminPage, 'IMAP');
    await expect
      .poll(async () => fieldInput(adminPage, 'mail-settings-inbound-port').inputValue(), {
        timeout: 5_000,
      })
      .toBe('993');
    await fieldInput(adminPage, 'mail-settings-inbound-host').fill('imap.example.com');

    // 4) Save with SMTP + IMAP configured (single account)
    const saveResponsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/mail/settings') &&
        res.request().method() === 'PUT' &&
        res.status() === 200,
      { timeout: 30_000 },
    );
    await adminPage.getByTestId('mail-settings-save').click();
    await saveResponsePromise;
    // Toast may auto-dismiss quickly; tolerate either toast or silent success.
    await adminPage
      .getByText(/邮件设置已保存|Mail settings saved/i)
      .first()
      .waitFor({ state: 'visible', timeout: 5_000 })
      .catch(() => undefined);
    const afterSaveShot = await shot(adminPage, 'mail-settings-04-after-save');

    // 5) Reload — values persist, password stays empty (masked)
    // From defaults to account: UI clears From when it equals account.
    await adminPage.reload();
    await openMailSettings(adminPage);
    await expect(fieldInput(adminPage, 'mail-settings-username')).toHaveValue(account);
    await expect(fieldInput(adminPage, 'mail-settings-from')).toHaveValue('');
    await expect(fieldInput(adminPage, 'mail-settings-smtp-host')).toHaveValue(
      'smtp.example.com',
    );
    await expect(fieldInput(adminPage, 'mail-settings-inbound-host')).toHaveValue(
      'imap.example.com',
    );
    await expect(fieldInput(adminPage, 'mail-settings-password')).toHaveValue('');

    // 5b) Optional From override persists as a distinct value.
    const overrideFrom = `noreply-${suffix}@example.com`;
    await fieldInput(adminPage, 'mail-settings-from').fill(overrideFrom);
    const overrideSavePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/mail/settings') &&
        res.request().method() === 'PUT' &&
        res.status() === 200,
      { timeout: 30_000 },
    );
    await adminPage.getByTestId('mail-settings-save').click();
    await overrideSavePromise;
    await adminPage.reload();
    await openMailSettings(adminPage);
    await expect(fieldInput(adminPage, 'mail-settings-username')).toHaveValue(account);
    await expect(fieldInput(adminPage, 'mail-settings-from')).toHaveValue(overrideFrom);
    const afterReloadShot = await shot(adminPage, 'mail-settings-05-after-reload');

    // 6) Test connection against example host — failure must open error modal (not page-top alert).
    await expect(adminPage.getByTestId('mail-test-result-alert')).toHaveCount(0);
    const testResponsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/mail/settings/test') &&
        res.request().method() === 'POST',
      { timeout: 30_000 },
    );
    await adminPage.getByTestId('mail-settings-test').click();
    await testResponsePromise;
    // Modal.error title is zh-CN "连接测试失败".
    await expect(
      adminPage.getByText('连接测试失败', { exact: true }).first(),
    ).toBeVisible({ timeout: 15_000 });
    await expect(adminPage.getByTestId('mail-error-modal-detail')).toBeVisible();
    await expect(adminPage.getByTestId('mail-test-result-alert')).toHaveCount(0);
    const afterTestShot = await shot(adminPage, 'mail-settings-06-after-test');
    // Dismiss error modal before the send-test flow.
    await adminPage.locator('.ant-modal-confirm-btns .ant-btn').last().click();
    await expect(adminPage.getByText('连接测试失败', { exact: true })).toHaveCount(0, {
      timeout: 5_000,
    });

    // 7) Send-test modal: layout (help vs first input) + form SMTP send.
    // Button label is "测试发送" (renamed from "测试邮件").
    await expect(adminPage.getByTestId('mail-settings-send-test')).toBeVisible();
    await expect(adminPage.getByText('测试发送').first()).toBeVisible();
    await adminPage.getByTestId('mail-settings-send-test').click();
    await expect(adminPage.getByTestId('mail-send-test-to')).toBeVisible({
      timeout: 10_000,
    });
    await expect(
      adminPage.locator('.ant-modal:visible').getByText('测试发送').first(),
    ).toBeVisible();

    // Help tip must be translated zh-CN (not i18n key).
    const sendHelp = adminPage.getByTestId('mail-send-test-help');
    await expect(sendHelp).toBeVisible();
    await expect(sendHelp).toContainText('使用当前页面上的 SMTP 配置发送');

    // Alert bottom must not overlap the recipient input top border.
    // Prefer layout container + inline gap so spacing does not depend on Tailwind scanning.
    const layout = adminPage.getByTestId('mail-send-test-body-layout');
    await expect(layout).toBeVisible();
    const toInput = fieldInput(adminPage, 'mail-send-test-to');
    await expect
      .poll(
        async () => {
          const helpBox = await sendHelp.boundingBox();
          const toBox = await toInput.boundingBox();
          if (!helpBox || !toBox) {
            return -999;
          }
          return toBox.y - (helpBox.y + helpBox.height);
        },
        { timeout: 5_000 },
      )
      .toBeGreaterThanOrEqual(12);
    const sendTestModalShot = await shot(
      adminPage,
      'mail-settings-07-send-test-modal-open',
    );

    await fieldInput(adminPage, 'mail-send-test-to').fill(`ops-${suffix}@example.com`);
    // Body is prefilled on TextArea; ensure non-empty.
    const bodyBox = adminPage
      .getByTestId('mail-send-test-body')
      .locator('textarea')
      .first()
      .or(adminPage.getByTestId('mail-send-test-body'));
    await expect(bodyBox).not.toHaveValue('');
    const sendTestPromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/mail/settings/send-test') &&
        res.request().method() === 'POST',
      { timeout: 30_000 },
    );
    await adminPage
      .locator('.ant-modal:visible .ant-modal-footer .ant-btn-primary')
      .click();
    const sendRes = await sendTestPromise;
    expect(sendRes.ok() || sendRes.status() < 500).toBeTruthy();
    // Example SMTP host should fail; failure is modal detail (or success toast if somehow delivered).
    await expect
      .poll(
        async () => {
          const failTitle = await adminPage
            .getByText(/测试发送失败|Test send failed/i)
            .count();
          const okToast = await adminPage
            .getByText(/测试发送成功|Test send succeeded/i)
            .count();
          const detail = await adminPage.getByTestId('mail-error-modal-detail').count();
          return failTitle + okToast + detail;
        },
        { timeout: 15_000 },
      )
      .toBeGreaterThan(0);
    const afterSendTestShot = await shot(adminPage, 'mail-settings-08-after-send-test');

    // Dismiss error confirm (if any), then close the send-test form modal.
    // Failure keeps the form dialog open; it must not intercept later page actions.
    const confirmBtn = adminPage.locator('.ant-modal-confirm-btns .ant-btn').last();
    if ((await confirmBtn.count()) > 0 && (await confirmBtn.isVisible().catch(() => false))) {
      await confirmBtn.click();
    }
    const sendTestCancel = adminPage.locator(
      '.ant-modal:visible .ant-modal-footer .ant-btn:not(.ant-btn-primary)',
    );
    if ((await sendTestCancel.count()) > 0) {
      await sendTestCancel.last().click();
    }
    await expect(adminPage.getByTestId('mail-send-test-to')).toHaveCount(0, {
      timeout: 5_000,
    });

    // 8) Receive-test uses already-saved IMAP form values (example host fails).
    await expect(adminPage.getByTestId('mail-settings-receive-test')).toBeVisible();
    await expect(adminPage.getByText('测试接收').first()).toBeVisible();
    await expect(adminPage.getByTestId('mail-settings-inbound-host')).toBeVisible();
    const receiveTestPromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/mail/settings/receive-test') &&
        res.request().method() === 'POST',
      { timeout: 30_000 },
    );
    await adminPage.getByTestId('mail-settings-receive-test').click();
    const receiveRes = await receiveTestPromise;
    expect(receiveRes.ok() || receiveRes.status() < 500).toBeTruthy();
    await expect
      .poll(
        async () => {
          const failTitle = await adminPage
            .getByText(/测试接收失败|Test receive failed/i)
            .count();
          const okToast = await adminPage
            .getByText(/测试接收成功|Test receive succeeded/i)
            .count();
          const detail = await adminPage.getByTestId('mail-error-modal-detail').count();
          return failTitle + okToast + detail;
        },
        { timeout: 15_000 },
      )
      .toBeGreaterThan(0);
    const afterReceiveTestShot = await shot(
      adminPage,
      'mail-settings-09-after-receive-test',
    );

    // Attach paths in report for manual multimodal review.
    test.info().annotations.push(
      { type: 'screenshot', description: initialShot },
      { type: 'screenshot', description: filledOutboundShot },
      { type: 'screenshot', description: filledInboundShot },
      { type: 'screenshot', description: filledPop3Shot },
      { type: 'screenshot', description: afterSaveShot },
      { type: 'screenshot', description: afterReloadShot },
      { type: 'screenshot', description: afterTestShot },
      { type: 'screenshot', description: sendTestModalShot },
      { type: 'screenshot', description: afterSendTestShot },
      { type: 'screenshot', description: afterReceiveTestShot },
    );
  });

  test('TC-2b: plugin management Manage navigates to mail settings', async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();

    const manage = pluginPage.pluginManageAction(pluginID);
    await expect(manage).toBeEnabled();
    await pluginPage.openPluginManagement(pluginID);

    await waitForRouteReady(adminPage, 30_000);
    await expect(adminPage).toHaveURL(mailSettingsPath);
    await expect(adminPage.getByTestId('mail-settings-page')).toBeVisible({
      timeout: 30_000,
    });
    await shot(adminPage, 'mail-settings-07-from-manage-button');
  });
});
