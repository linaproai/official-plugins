/**
 * TC003 linapro-storage-s3 连接测试失败错误展示
 *
 * 失败时以 Modal.error 弹窗展示完整 SDK 详情（与邮件管理页一致），
 * 不再使用页面顶部 Alert，也不再把长错误塞进 Toast。
 */
import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { MainLayout } from "@host-tests/pages/MainLayout";
import { workspacePath } from "@host-tests/fixtures/config";
import { waitForRouteReady } from "@host-tests/support/ui";

const pluginID = "linapro-storage-s3";
const longErrorDetail =
  "AccessDenied: The request was denied. RequestId=req-e2e-test-connection-error-display-abcdefghijklmnopqrstuvwxyz " +
  "StatusCode=403 Bucket=demo-bucket Endpoint=http://minio:9000 " +
  "Detail=" +
  "x".repeat(180);

test.describe("TC003 linapro-storage-s3 连接测试失败错误展示", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([pluginID]);
  });

  test("TC003a: 失败时以弹窗展示完整详情，页面顶部无 Alert", async ({
    adminPage,
  }) => {
    await adminPage.route(
      `**/x/${pluginID}/api/v1/settings/test-connection**`,
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            code: 0,
            message: "success",
            data: {
              ok: false,
              message: longErrorDetail,
            },
          }),
        });
      },
    );

    const layout = new MainLayout(adminPage);
    await adminPage.goto(workspacePath("/dashboard/workspace"));
    await waitForRouteReady(adminPage);
    await layout.expandSidebarGroup(/系统设置|Settings/i);
    await layout.sidebarMenuItem(/存储管理-S3|Storage Management - S3/i).click();
    await waitForRouteReady(adminPage);

    const form = adminPage.locator("form.ant-form-horizontal").first();
    await expect(form).toBeVisible({ timeout: 15000 });

    await adminPage
      .getByRole("button", { name: /测试连接|Test connection/i })
      .click();

    // Modal.error title uses the short i18n string.
    await expect(
      adminPage.getByText("连接测试失败", { exact: true }).first(),
    ).toBeVisible({ timeout: 10000 });

    const detail = adminPage.getByTestId("storage-error-modal-detail");
    await expect(detail).toBeVisible();
    await expect(detail).toContainText("AccessDenied");
    await expect(detail).toContainText(
      "RequestId=req-e2e-test-connection-error-display",
    );
    await expect(detail).toContainText(longErrorDetail.slice(-40));

    // Must not render the old page-top error Alert.
    await expect(adminPage.getByTestId("storage-test-result-alert")).toHaveCount(
      0,
    );
    await expect(adminPage.getByTestId("storage-test-error-detail")).toHaveCount(
      0,
    );

    // Toast (if any) must not carry the long SDK body.
    const toast = adminPage.locator(".ant-message-notice:visible");
    if ((await toast.count()) > 0) {
      await expect(toast.first()).not.toContainText("AccessDenied");
      await expect(toast.first()).not.toContainText("RequestId=");
    }
  });
});
