import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0239 } from '../../support/multi-tenant-scenarios';

test.describe('TC-239 租户角色平台权限阻断', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-239a: dirty platform grants are hidden and rejected in tenant context', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0239();
  });
});
