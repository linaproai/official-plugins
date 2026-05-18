import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0241 } from '../../support/multi-tenant-scenarios';

test.describe('TC-241 租户任务分组隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-241a: tenant job groups are listed, created, updated, and migrated in tenant scope', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0241();
  });
});
