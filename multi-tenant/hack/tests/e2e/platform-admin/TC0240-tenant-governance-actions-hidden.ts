import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0240 } from '../../support/multi-tenant-scenarios';

test.describe('TC-240 租户态平台治理动作隐藏', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-240a: tenant context cannot use platform menu or plugin governance actions', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0240();
  });
});
