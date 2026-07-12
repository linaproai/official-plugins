/**
 * TC003 linapro-storage-s3 连接测试失败错误展示
 *
 * 失败时顶部 Toast 仅短文案，长 SDK 详情在页内 Alert 中展示并可复制。
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

  test("TC003a: 失败时 Toast 为短文案且页内 Alert 展示完整详情", async ({
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
    await layout.expandSidebarGroup(/存储管理|Storage/i);
    await layout.sidebarMenuItem(/S3 存储|S3存储|S3 Storage|^S3$/i).click();
    await waitForRouteReady(adminPage);

    const form = adminPage.locator("form.ant-form-horizontal").first();
    await expect(form).toBeVisible({ timeout: 15000 });

    await adminPage
      .getByRole("button", { name: /测试连接|Test connection/i })
      .click();

    const toast = adminPage.locator(".ant-message-notice:visible").filter({
      hasText: /连接测试失败|Connection test failed/i,
    });
    await expect(toast.first()).toBeVisible({ timeout: 10000 });
    await expect(toast.first()).not.toContainText("AccessDenied");
    await expect(toast.first()).not.toContainText("RequestId=");

    const alert = adminPage.getByTestId("storage-test-result-alert");
    await expect(alert).toBeVisible();
    await expect(alert).toContainText(/连接测试失败|Connection test failed/i);

    const title = adminPage.getByTestId("storage-test-error-title");
    await expect(title).toBeVisible();
    await expect(title).toHaveText(/连接测试失败|Connection test failed/i);
    // Title should use normal body font size (not Alert's larger heading size).
    const titleFontSize = await title.evaluate(
      (el) => Number.parseFloat(getComputedStyle(el).fontSize),
    );
    const labelFontSize = await form
      .locator(".ant-form-item-label label")
      .first()
      .evaluate((el) => Number.parseFloat(getComputedStyle(el).fontSize));
    expect(titleFontSize).toBeLessThanOrEqual(labelFontSize + 0.5);

    const detail = adminPage.getByTestId("storage-test-error-detail");
    await expect(detail).toBeVisible();
    await expect(detail).toContainText("AccessDenied");
    await expect(detail).toContainText("RequestId=req-e2e-test-connection-error-display");
    await expect(detail).toContainText(longErrorDetail.slice(-40));

    await expect(
      alert.getByRole("button", { name: /复制详情|Copy details/i }),
    ).toBeVisible();
  });
});
