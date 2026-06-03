import { expect, test } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';
import { DeptPage } from '../../pages/DeptPage';

test.describe('TC001 部门管理 CRUD', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(['linapro-org-core']);
  });

  function uniqueDeptNames() {
    const suffix = `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    return {
      parent: `测试部门_${suffix}`,
      child: `子部门A_${suffix}`,
      childRenamed: `子部门B_${suffix}`,
    };
  }

  async function deleteDeptIfPresent(deptPage: DeptPage, name: string) {
    try {
      if (await deptPage.hasDept(name)) {
        await deptPage.deleteDept(name);
      }
    } catch {
      // Cleanup should not hide the original assertion failure.
    }
  }

  async function cleanupDeptTree(
    deptPage: DeptPage,
    parentName: string,
    childNames: string[] = [],
  ) {
    for (const childName of childNames) {
      await deleteDeptIfPresent(deptPage, childName);
    }
    await deleteDeptIfPresent(deptPage, parentName);
  }

  test('TC001a: 在根部门下创建子部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    const names = uniqueDeptNames();
    await deptPage.goto();

    try {
      await deptPage.createSubDept('LinaPro.AI', names.parent);

      await expect(adminPage.getByText(/创建成功|success/i).first()).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await cleanupDeptTree(deptPage, names.parent);
    }
  });

  test('TC001b: 新创建的部门在列表中可见', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    const names = uniqueDeptNames();
    await deptPage.goto();

    try {
      await deptPage.createSubDept('LinaPro.AI', names.parent);

      const hasDept = await deptPage.hasDept(names.parent);
      expect(hasDept).toBeTruthy();
    } finally {
      await cleanupDeptTree(deptPage, names.parent);
    }
  });

  test('TC001c: 创建子部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    const names = uniqueDeptNames();
    await deptPage.goto();

    try {
      await deptPage.createSubDept('LinaPro.AI', names.parent);
      await deptPage.createSubDept(names.parent, names.child);

      await expect(adminPage.getByText(/创建成功|success/i).first()).toBeVisible({
        timeout: 5000,
      });
      expect(await deptPage.hasDept(names.child)).toBeTruthy();
    } finally {
      await cleanupDeptTree(deptPage, names.parent, [names.child]);
    }
  });

  test('TC001d: 编辑部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    const names = uniqueDeptNames();
    await deptPage.goto();

    try {
      await deptPage.createSubDept('LinaPro.AI', names.parent);
      await deptPage.createSubDept(names.parent, names.child);
      await deptPage.editDept(names.child, names.childRenamed);

      await expect(adminPage.getByText(/更新成功|success/i).first()).toBeVisible({
        timeout: 5000,
      });
      expect(await deptPage.hasDept(names.childRenamed)).toBeTruthy();
    } finally {
      await cleanupDeptTree(deptPage, names.parent, [
        names.childRenamed,
        names.child,
      ]);
    }
  });

  test('TC001e: 删除子部门后删除父部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    const names = uniqueDeptNames();
    await deptPage.goto();

    try {
      await deptPage.createSubDept('LinaPro.AI', names.parent);
      await deptPage.createSubDept(names.parent, names.child);
      await deptPage.deleteDept(names.child);
      await deptPage.deleteDept(names.parent);

      const hasDept = await deptPage.hasDept(names.parent);
      expect(hasDept).toBeFalsy();
    } finally {
      await cleanupDeptTree(deptPage, names.parent, [names.child]);
    }
  });
});
