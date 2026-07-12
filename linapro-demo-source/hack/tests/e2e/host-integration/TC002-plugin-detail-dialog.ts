import { test, expect } from '@host-tests/fixtures/auth';
import { DemoSourcePage } from '../../pages/DemoSourcePage';

const pluginID = "linapro-demo-source";
const pluginName = "示例插件-源码插件";
const pluginVersion = "v0.1.0";
const pluginDescription = "提供左侧菜单页面与公开/受保护路由示例的源码插件";

test.describe("TC-2 插件详情弹窗", () => {
  let pluginPage: DemoSourcePage;

  test.beforeEach(async ({ adminPage }) => {
    pluginPage = new DemoSourcePage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
  });

  test("TC-2a: 点击详情按钮展示插件基础治理信息", async () => {
    await pluginPage.openPluginDetail(pluginID);

    const modal = pluginPage.pluginDetailModal();
    await expect(modal).toContainText(pluginName);
    await expect(modal).toContainText(pluginID);
    await expect(modal).toContainText("源码插件");
    await expect(modal).toContainText(pluginVersion);
    // Detail collapses installed/enabled/runtime into one status field, and
    // scopeNature/installMode into one plugin-scope field.
    await expect(modal).toContainText("插件状态");
    await expect(modal).not.toContainText("安装状态");
    await expect(modal).not.toContainText("作用域性质");
    await expect(modal).not.toContainText("安装模式");
    await expect(modal).toContainText("启动管理");
    await expect(modal).not.toContainText("授权状态");
    await expect(modal).toContainText("示例数据");
    await expect(modal).toContainText("支持多租户");
    await expect(modal).toContainText("新租户启用");
    await expect(modal).toContainText("插件作用域");
    await expect(modal).toContainText("安装时间");
    await expect(modal).toContainText("更新时间");
    await expect(modal).not.toContainText("授权要求");
    await expect(pluginPage.pluginDetailHasMockData()).toContainText("是");
    await expect(pluginPage.pluginDetailSupportsMultiTenant()).toContainText(
      "是",
    );
    await expect(pluginPage.pluginDetailTenantProvisioning()).toContainText(
      "否",
    );
    await expect(pluginPage.pluginDetailPluginScope()).toContainText("租户级");
    await expect(pluginPage.pluginDetailDescriptionRow()).toBeVisible();
    await expect(pluginPage.pluginDetailDescriptionRow()).toContainText(
      pluginDescription,
    );
  });

  test("TC-2b: 源码插件详情页不展示多余的宿主服务空状态提示", async () => {
    await pluginPage.openPluginDetail(pluginID);
    await expect(pluginPage.pluginDetailEmptyHostServices()).toHaveCount(0);
    await expect(pluginPage.pluginDetailModal()).not.toContainText(
      "当前动态插件未声明额外宿主服务。",
    );
  });
});
