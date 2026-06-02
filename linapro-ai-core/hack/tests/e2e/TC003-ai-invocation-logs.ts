import { expect, test } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';
import { SmartCenterPage } from '../pages/SmartCenterPage';
import { deleteInvocationLog, insertInvocationLog } from '../support/ai-core-api';

test.describe('TC-3 AI 调用日志', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(['linapro-ai-core']);
  });

  test('TC-3a: 调用日志支持分页筛选和详情脱敏展示', async ({ adminPage }) => {
    const suffix = Date.now();
    const fixture = insertInvocationLog({
      purpose: `e2e.invocation.${suffix}`,
      requestId: `e2e-invocation-${suffix}`,
    });
    const smartCenter = new SmartCenterPage(adminPage);
    try {
      await smartCenter.gotoInvocations();
      await smartCenter.filterInvocationsByPurpose(fixture.purpose);

      const mainContent = adminPage.locator('#__vben_main_content');
      await expect(mainContent.getByText(/调用日志|Invocation Logs/i).first()).toBeVisible();
      await expect(adminPage.locator('.vxe-body--row:visible', { hasText: fixture.purpose })).toBeVisible();
      await smartCenter.openInvocationDetail();
      const detail = adminPage.locator('[role="dialog"]').last();
      await expect(detail.getByText(/调用详情|Invocation Detail/i)).toBeVisible();
      await expect(detail.getByText(fixture.requestId)).toBeVisible();
      await expect(detail.getByText(fixture.purpose)).toBeVisible();
      await expect(detail.getByText(/redacted error summary/i)).toBeVisible();
      await expect(adminPage.getByText(/完整 prompt|full prompt/i)).toHaveCount(0);
      await expect(adminPage.getByText(/sk-/i)).toHaveCount(0);
    } finally {
      deleteInvocationLog(fixture.requestId);
    }
  });
});
