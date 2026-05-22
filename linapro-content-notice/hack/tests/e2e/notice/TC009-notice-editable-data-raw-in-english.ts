import type { APIRequestContext, Page } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import { pluginApiPath } from '@host-tests/fixtures/config';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { NoticePage } from '../../pages/NoticePage';
import { createAdminApiContext, expectSuccess } from '@host-tests/support/api/job';

const pluginID = 'linapro-content-notice';

test.describe('TC009 通知公告可编辑数据退出 i18n 投影专项回归', () => {
  test('TC-5a: 英文环境下通知管理页中的可编辑业务记录保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const noticePage = new NoticePage(adminPage);

    await ensureNoticeRawData(adminPage);
    await mainLayout.switchLanguage('English');

    await noticePage.goto();
    await expect(await noticePage.hasNotice('系统升级通知')).toBe(true);
  });
});

async function ensureNoticeRawData(page: Page) {
  await ensureSourcePluginEnabled(page, 'linapro-content-notice');
  const api = await createAdminApiContext();
  try {
    await ensureNotice(api);
  } finally {
    await api.dispose();
  }
}

async function ensureNotice(api: APIRequestContext) {
  const existing = await expectSuccess<{
    list: Array<{ id: number; title: string }>;
  }>(
    await api.get(
      pluginApiPath(
        pluginID,
        `notice?pageNum=1&pageSize=100&title=${encodeURIComponent('系统升级通知')}`,
      ),
    ),
  );
  if (existing.list.some((item) => item.title === '系统升级通知')) {
    return;
  }
  await expectSuccess(
    await api.post(pluginApiPath(pluginID, 'notice'), {
      data: {
        content:
          '<p>系统将于本周六凌晨2:00-4:00进行升级维护，届时系统将暂停服务。</p>',
        status: 1,
        title: '系统升级通知',
        type: 1,
      },
    }),
  );
}
