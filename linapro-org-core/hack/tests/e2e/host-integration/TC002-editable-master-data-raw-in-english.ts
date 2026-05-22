import type { APIRequestContext, Page } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import { pluginApiPath } from '@host-tests/fixtures/config';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { DeptPage } from '../../pages/DeptPage';
import { PostPage } from '../../pages/PostPage';
import { RolePage } from '@host-tests/pages/RolePage';
import { UserPage } from '@host-tests/pages/UserPage';
import { createAdminApiContext, expectSuccess } from '@host-tests/support/api/job';

const pluginID = 'linapro-org-core';

test.describe('TC002 可编辑主数据退出 i18n 投影专项回归', () => {
  test('TC-2a: 英文环境下用户与组织管理页面中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const userPage = new UserPage(adminPage);
    const deptPage = new DeptPage(adminPage);
    const postPage = new PostPage(adminPage);
    const rolePage = new RolePage(adminPage);

    await ensureOrgRawData(adminPage);
    await mainLayout.switchLanguage('English');

    await userPage.goto();
    await expect(await userPage.hasDeptTreeNode('研发部门')).toBe(true);

    await deptPage.goto();
    await expect(await deptPage.hasDeptInExpandedTree('研发部门')).toBe(true);

    await postPage.goto();
    await expect(await postPage.hasPostName('总经理')).toBe(true);

    await rolePage.goto();
    await expect(await rolePage.hasRole('User')).toBe(true);
  });

  test('TC-2b: 英文环境下组织插件角色关联数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const rolePage = new RolePage(adminPage);

    await ensureOrgRawData(adminPage);
    await mainLayout.switchLanguage('English');

    await rolePage.goto();
    await expect(await rolePage.hasRole('User')).toBe(true);
  });
});

async function ensureOrgRawData(page: Page) {
  await ensureSourcePluginEnabled(page, 'linapro-org-core');
  const api = await createAdminApiContext();
  try {
    const dept = await ensureDept(api);
    await ensurePost(api, dept.id);
  } finally {
    await api.dispose();
  }
}

async function ensureDept(api: APIRequestContext) {
  const existing = await expectSuccess<{
    list: Array<{ id: number; name: string }>;
  }>(
    await api.get(
      pluginApiPath(pluginID, `dept?name=${encodeURIComponent('研发部门')}`),
    ),
  );
  const dept = existing.list.find((item) => item.name === '研发部门');
  if (dept) {
    return dept;
  }
  return expectSuccess<{ id: number }>(
    await api.post(pluginApiPath(pluginID, 'dept'), {
      data: {
        code: 'e2e-raw-dev',
        name: '研发部门',
        orderNum: 1,
        parentId: 0,
        status: 1,
      },
    }),
  );
}

async function ensurePost(api: APIRequestContext, deptId: number) {
  const existing = await expectSuccess<{
    list: Array<{ id: number; name: string }>;
  }>(
    await api.get(
      pluginApiPath(
        pluginID,
        `post?pageNum=1&pageSize=100&name=${encodeURIComponent('总经理')}`,
      ),
    ),
  );
  if (existing.list.some((item) => item.name === '总经理')) {
    return;
  }
  await expectSuccess(
    await api.post(pluginApiPath(pluginID, 'post'), {
      data: {
        code: 'E2E_RAW_CEO',
        deptId,
        name: '总经理',
        sort: 1,
        status: 1,
      },
    }),
  );
}
