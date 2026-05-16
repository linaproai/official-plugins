import { test } from '@host-tests/fixtures/auth';
import { MultiTenantPage } from '@host-tests/pages/MultiTenantPage';

test.describe('TC-222 多租户管理工作台页面路由', () => {
  test('TC-222a: platform tenant management stays visible', async ({
    page,
  }) => {
    test.setTimeout(180_000);
    const multiTenantPage = new MultiTenantPage(page);

    await multiTenantPage.gotoPlatformTenants();
    await multiTenantPage.expectPlatformTenantWorkbench();
  });

  test('TC-222b: platform user management exposes tenant controls', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);

    await multiTenantPage.gotoSystemUsers();
    await multiTenantPage.expectSystemUserTenantWorkbench();
  });

  test('TC-222c: tenant member management uses the user page', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);
    await multiTenantPage.expectTenantMemberManagementUsesUserPage();
  });

  test('TC-222d: tenant switch enters the tenant workbench', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);
    await multiTenantPage.exerciseTenantSwitch();
  });

  test('TC-222e: platform impersonation can enter and exit a tenant', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);
    await multiTenantPage.exerciseImpersonation();
  });

  test('TC-222f: obsolete tenant management routes fall back', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);
    await multiTenantPage.expectRemovedManagementRoutesFallback();
  });
});
