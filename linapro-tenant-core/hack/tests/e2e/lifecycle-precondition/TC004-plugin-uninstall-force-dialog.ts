import {
  createTenant,
  deleteTenant,
  ensureMultiTenantPluginEnabled,
  expect,
  test,
} from '../../support/linapro-tenant-core';
import { createAdminApiContext, getPlugin } from '@host-tests/support/api/job';
import { PluginPage } from '@host-tests/pages/PluginPage';

test.describe('TC-4 多租户插件卸载前置条件弹窗', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-4a: lifecycle precondition veto opens localized force confirmation dialog and resubmits with force', async ({
    adminPage,
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');

    const api = await createAdminApiContext();
    const pluginId = 'linapro-tenant-core';
    let tenantId = 0;
    try {
      await ensureMultiTenantPluginEnabled(api);
      const tenant = await createTenant(api, {
        code: `tc227-${Date.now()}`.slice(0, 32),
        name: 'TC227 Tenant',
      });
      tenantId = tenant.id;

      const pluginPage = new PluginPage(adminPage);
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginId);

      await pluginPage.openUninstallDialog(pluginId);
      await expect(pluginPage.uninstallPurgeCheckboxWrapper()).toBeVisible();
      await expect(pluginPage.uninstallPurgeWarning()).toBeVisible();
      if (await pluginPage.uninstallPurgeCheckbox().isChecked()) {
        await pluginPage.uninstallPurgeCheckbox().click();
      }
      await expect(pluginPage.uninstallPurgeCheckbox()).not.toBeChecked();

      const vetoResponsePromise = adminPage.waitForResponse(
        (response) =>
          response.url().includes(`/plugins/${pluginId}`) &&
          response.request().method() === 'DELETE',
      );
      await pluginPage.uninstallConfirmButton().click();

      const vetoResponse = await vetoResponsePromise;
      expect(vetoResponse.url()).not.toContain('force=true');
      expect((await vetoResponse.json()).errorCode).toBe(
        'PLUGIN_LIFECYCLE_PRECONDITION_VETOED',
      );

      await expect(pluginPage.lifecyclePreconditionDialog()).toBeVisible();
      await expect(pluginPage.uninstallDialog()).toHaveCount(0);
      await expect(pluginPage.lifecyclePreconditionDialog()).toHaveCSS('gap', '10px');
      await expect(pluginPage.lifecyclePreconditionReasonAlert()).not.toContainText(
        '插件返回了阻止当前操作的原因。',
      );
      await expect(pluginPage.lifecyclePreconditionReasonText()).toContainText(
        '当前插件阻止操作，原因：',
      );
      await expect(pluginPage.lifecyclePreconditionReasonText()).toContainText(
        '仍存在租户。请先删除租户，再卸载多租户插件。',
      );
      await expect(pluginPage.lifecyclePreconditionForceAlert()).toContainText(
        '强制卸载会绕过上述前置条件并清理插件数据，请确认你理解该风险。',
      );
      await expect(pluginPage.lifecyclePreconditionForceAlert()).toContainText(
        '输入插件 ID "linapro-tenant-core" 以启用强制卸载。',
      );
      await expect(pluginPage.lifecyclePreconditionDialog()).toContainText(pluginId);
      await expect(pluginPage.lifecyclePreconditionConfirmButton()).toBeDisabled();

      await pluginPage.lifecyclePreconditionForcePluginIdInput().fill(pluginId);
      await expect(pluginPage.lifecyclePreconditionConfirmButton()).toBeEnabled();

      const forceResponsePromise = adminPage.waitForResponse(
        (response) =>
          response.url().includes(`/plugins/${pluginId}`) &&
          response.url().includes('force=true') &&
          response.request().method() === 'DELETE',
      );
      await pluginPage.lifecyclePreconditionConfirmButton().click();
      const forceResponse = await forceResponsePromise;
      const forcePayload = await forceResponse.json();
      expect(forcePayload.code).toBe(0);
      await expect(pluginPage.lifecyclePreconditionDialog()).toHaveCount(0);

      const pluginAfterForce = await getPlugin(api, pluginId);
      expect(pluginAfterForce.installed).toBe(0);
    } finally {
      await ensureMultiTenantPluginEnabled(api).catch(() => {});
      const refreshed = await getPlugin(api, pluginId).catch(() => null);
      if (tenantId > 0) {
        await deleteTenant(api, tenantId).catch(() => {});
      }
      if (refreshed?.enabled !== 1) {
        await ensureMultiTenantPluginEnabled(api).catch(() => {});
      }
      await api.dispose();
    }
  });
});
