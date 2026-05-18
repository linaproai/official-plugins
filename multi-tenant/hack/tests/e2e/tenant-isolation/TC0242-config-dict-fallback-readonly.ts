import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0242 } from '../../support/multi-tenant-scenarios';

test.describe('TC-242 参数和字典 fallback 只读', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-242a: fallback rows hide direct edit and do not request missing details', async ({
    multiTenantMode,
    page,
  }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0242(page);
  });
});
