import type { Locator } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from '@host-tests/support/ui';

async function fillModalRange(dialog: Locator, start: string, end: string) {
  const rangeInputs = dialog.locator('.ant-picker-range input');
  for (const [index, value] of [start, end].entries()) {
    const input = rangeInputs.nth(index);
    await input.evaluate((node: HTMLInputElement) =>
      node.removeAttribute('readonly'),
    );
    await input.click();
    await input.fill(value);
    await input.press(index === 0 ? 'Tab' : 'Enter');
  }
  await expect(rangeInputs.first()).toHaveValue(start);
  await expect(rangeInputs.nth(1)).toHaveValue(end);
}

async function expectRangeSectionSeparated(
  dialog: Locator,
  alertTestId: string,
  rangeTestId: string,
) {
  const alertBox = await dialog.getByTestId(alertTestId).boundingBox();
  const rangeBox = await dialog.getByTestId(rangeTestId).boundingBox();
  expect(alertBox).not.toBeNull();
  expect(rangeBox).not.toBeNull();
  if (!alertBox || !rangeBox) {
    return;
  }
  expect(rangeBox.y - alertBox.y - alertBox.height).toBeGreaterThanOrEqual(12);
}

test.describe('TC005 操作日志范围删除', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'linapro-monitor-operlog');
  });

  test('TC005a: 空范围删除被阻止且范围删除入口独立可用', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    const deleteBtn = adminPage.getByTestId('operlog-range-delete');
    await expect(deleteBtn).toBeEnabled();

    await deleteBtn.click();
    const dialog = await waitForDialogReady(
      adminPage.locator('.ant-modal-wrap:visible').filter({
        hasText: '删除操作日志',
      }),
    );

    await expect(dialog).toContainText('删除操作日志');
    await expect(dialog).toContainText('请选择操作日志删除方式');
    await expect(dialog).toContainText('删除所有操作日志');
    await expect(dialog.locator('.ant-picker-range')).toBeVisible();
    await expectRangeSectionSeparated(
      dialog,
      'operlog-delete-alert',
      'operlog-delete-range-section',
    );

    await dialog.getByRole('button', { name: /确\s*(认|定)/ }).click();
    await expect(adminPage.getByText('请选择完整的操作日志日期范围')).toBeVisible();

    await dialog.getByRole('button', { name: /取\s*消/ }).click();
  });

  test('TC005b: 选择范围后按范围清理操作日志', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    await adminPage.getByTestId('operlog-range-delete').click();
    const dialog = await waitForDialogReady(
      adminPage.locator('.ant-modal-wrap:visible').filter({
        hasText: '删除操作日志',
      }),
    );

    await fillModalRange(dialog, '2099-01-01', '2099-01-02');

    const responsePromise = adminPage.waitForResponse((response) => {
      const url = new URL(response.url());
      return (
        url.pathname.includes(
          '/x/linapro-monitor-operlog/api/v1/operlog/clean',
        ) &&
        url.searchParams.get('beginTime') === '2099-01-01' &&
        url.searchParams.get('endTime') === '2099-01-02' &&
        response.request().method() === 'DELETE'
      );
    });

    await dialog.getByRole('button', { name: /确\s*(认|定)/ }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    await waitForRouteReady(adminPage);
  });

  test('TC005c: 选择全部日志后清理操作日志不传日期范围', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    let cleanRequestUrl = '';
    await adminPage.route(
      '**/x/linapro-monitor-operlog/api/v1/operlog/clean**',
      async (route) => {
        if (route.request().method() !== 'DELETE') {
          await route.continue();
          return;
        }
        cleanRequestUrl = route.request().url();
        await route.fulfill({
          body: JSON.stringify({ code: 0, data: { deleted: 0 }, message: 'OK' }),
          contentType: 'application/json',
          status: 200,
        });
      },
    );

    await adminPage.getByTestId('operlog-range-delete').click();
    const dialog = await waitForDialogReady(
      adminPage.locator('.ant-modal-wrap:visible').filter({
        hasText: '删除操作日志',
      }),
    );

    await dialog.getByText('删除所有操作日志').click();
    await expect(
      dialog.locator('.ant-picker-range input').first(),
    ).toBeDisabled();

    await dialog.getByRole('button', { name: /确\s*(认|定)/ }).click();
    await expect.poll(() => cleanRequestUrl).not.toBe('');

    const url = new URL(cleanRequestUrl);
    expect(url.searchParams.has('beginTime')).toBe(false);
    expect(url.searchParams.has('endTime')).toBe(false);
    await waitForRouteReady(adminPage);
  });
});
