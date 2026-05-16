import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0182 } from '../../support/multi-tenant-scenarios';

test.describe('TC-182 租户删除生命周期前置条件', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-182a: active tenant delete passes lifecycle precondition before soft delete', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0182();
  });
});
