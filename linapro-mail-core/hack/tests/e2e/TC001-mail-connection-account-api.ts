import { test, expect } from '@host-tests/fixtures/auth';
import { pluginApiPath } from '@host-tests/fixtures/config';
import {
  createAdminApiContext,
  ensureSourcePluginEnabledViaAPI,
} from '@host-tests/fixtures/plugin';

/**
 * TC001 mail-core Connection/Account API smoke.
 * Covers outbound-only account creation and connection list after plugin enable.
 */
test.describe('TC-1 mail-core connection and account API', () => {
  test('TC-1a: enable mail-core and create smtp connection + outbound-only account', async () => {
    const adminApi = await createAdminApiContext();
    try {
      await ensureSourcePluginEnabledViaAPI(adminApi, 'linapro-mail-core');
      await ensureSourcePluginEnabledViaAPI(adminApi, 'linapro-mail-smtp');

      const connectionsPath = pluginApiPath('linapro-mail-core', 'mail/connections');
      const accountsPath = pluginApiPath('linapro-mail-core', 'mail/accounts');
      const suffix = Date.now();

      const createConn = await adminApi.post(connectionsPath, {
        data: {
          name: `e2e-smtp-${suffix}`,
          kind: 'smtp',
          host: 'smtp.example.com',
          port: 587,
          username: 'noreply@example.com',
          secretRef: 'e2e-secret',
          tlsMode: 'starttls',
          authMode: 'password',
          status: 1,
          remark: 'e2e',
        },
      });
      expect(createConn.ok()).toBeTruthy();
      const connBody = await createConn.json();
      const connectionId = connBody?.data?.id ?? connBody?.id;
      expect(connectionId).toBeTruthy();

      const createAccount = await adminApi.post(accountsPath, {
        data: {
          name: `e2e-account-${suffix}`,
          fromAddress: 'noreply@example.com',
          outboundConnectionId: connectionId,
          inboundConnectionId: 0,
          isDefault: false,
          status: 1,
          remark: 'outbound-only',
        },
      });
      expect(createAccount.ok()).toBeTruthy();
      const accountBody = await createAccount.json();
      const accountId = accountBody?.data?.id ?? accountBody?.id;
      expect(accountId).toBeTruthy();

      const listConn = await adminApi.get(`${connectionsPath}?pageNum=1&pageSize=20`);
      expect(listConn.ok()).toBeTruthy();
      const listBody = await listConn.json();
      const list = listBody?.data?.list ?? listBody?.list ?? [];
      expect(Array.isArray(list)).toBeTruthy();
      expect(list.some((item: { id?: number }) => item.id === connectionId)).toBeTruthy();

      // Cleanup
      await adminApi.delete(accountsPath, { data: { ids: [accountId] } });
      await adminApi.delete(connectionsPath, { data: { ids: [connectionId] } });
    } finally {
      await adminApi.dispose();
    }
  });
});
