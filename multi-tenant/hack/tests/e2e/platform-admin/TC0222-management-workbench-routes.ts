import { test } from '@host-tests/fixtures/auth';
import { MultiTenantPage } from '@host-tests/pages/MultiTenantPage';

test.describe('TC-222 多租户管理工作台页面路由', () => {
  test('TC-222a: platform management stays visible and obsolete tenant menus are pruned', async ({
    page,
  }) => {
    test.setTimeout(180_000);
    const multiTenantPage = new MultiTenantPage(page);

    await multiTenantPage.gotoPlatformTenants();
    await multiTenantPage.expectPlatformTenantWorkbench();

    await multiTenantPage.gotoSystemUsers();
    await multiTenantPage.expectSystemUserTenantWorkbench();

    await multiTenantPage.expectTenantMemberManagementUsesUserPage();
    await multiTenantPage.exerciseTenantSwitch();
    await multiTenantPage.exerciseImpersonation();

    await multiTenantPage.expectRemovedManagementRoutesFallback();
  });
});
