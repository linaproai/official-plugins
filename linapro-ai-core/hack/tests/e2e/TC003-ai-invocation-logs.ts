import { expect, test } from "@host-tests/fixtures/auth";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  deleteInvocationLog,
  insertInvocationLog,
  listInvocations,
  withAdminApi,
} from "../support/ai-core-api";

test.describe("TC-3 AI 调用日志", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-3a~d: 调用日志支持来源列、调用方法/时间筛选、详情脱敏展示和范围清理", async ({
    adminPage,
    mainLayout,
  }) => {
    const suffix = Date.now();
    const sourcePluginId = `e2e-source-${suffix}`;
    const fixture = insertInvocationLog({
      purpose: `e2e.invocation.${suffix}`,
      protocol: "ANTHROPIC",
      requestId: `e2e-invocation-${suffix}`,
      sourcePluginId,
    });
    const otherSourceFixture = insertInvocationLog({
      purpose: fixture.purpose,
      requestId: `e2e-invocation-other-source-${suffix}`,
      sourcePluginId: `e2e-other-source-${suffix}`,
    });
    const oldFixture = insertInvocationLog({
      createdAtSql: "NOW() - INTERVAL '3 days'",
      purpose: `e2e.invocation.old.${suffix}`,
      requestId: `e2e-invocation-old-${suffix}`,
      sourcePluginId,
    });
    const smartCenter = new SmartCenterPage(adminPage);
    try {
      await mainLayout.switchLanguage("English");
      await smartCenter.gotoInvocations();
      await mainLayout.expandSidebarGroup("AI Hub");
      await expect(mainLayout.sidebarMenuItem("AI Hub")).toBeVisible();
      await expect(mainLayout.sidebarMenuItem("AI Tiers")).toBeVisible();
      await expect(mainLayout.sidebarMenuItem("Request Logs")).toBeVisible();

      const mainContent = adminPage.locator("#__vben_main_content");
      await expect(mainContent.getByText("Request Logs").first()).toBeVisible();
      await expect(mainContent.getByText("Invocation Logs")).toHaveCount(0);
      await expect(
        mainContent.getByText(/来源插件|Source Plugin/i).first(),
      ).toBeVisible();
      await expect(
        mainContent.getByText(/创建时间|Created/i).first(),
      ).toBeVisible();
      await expect(
        mainContent.getByText(/调用方法|Invocation Method/i).first(),
      ).toBeVisible();
      await expect(
        mainContent.getByText(/能力方法|Capability Method/i),
      ).toHaveCount(0);
      await smartCenter.assertInvocationMethodLabelSingleLine();
      await smartCenter.captureEvidence("TC003-invocation-search-labels");
      await expect(adminPage.getByTestId("ai-invocation-clear")).toContainText(
        /删\s*除|Delete/i,
      );
      await smartCenter.assertInvocationDeleteButtonStyle();

      await smartCenter.expectInvocationDeleteDialogRequiresRangeAndCancel();
      await smartCenter.expectInvocationDeleteAllModeUsesUnscopedClean();

      const filterRequestPromise = adminPage.waitForRequest((req) => {
        const url = new URL(req.url());
        return (
          req.method() === "GET" &&
          url.pathname.includes("/x/linapro-ai-core/api/v1/ai/invocations") &&
          url.searchParams.get("capabilityType") === "text" &&
          url.searchParams.get("capabilityMethod") === "generate" &&
          url.searchParams.get("purpose") === fixture.purpose &&
          url.searchParams.get("sourcePluginId") === fixture.sourcePluginId
        );
      });
      await smartCenter.filterInvocationsByCapabilityPurposeAndSource(
        "text.generate",
        fixture.purpose,
        fixture.sourcePluginId,
      );
      await filterRequestPromise;

      const row = adminPage
        .locator(".vxe-body--row:visible", {
          hasText: fixture.sourcePluginId,
        })
        .filter({ hasText: fixture.purpose })
        .first();
      await expect(row).toBeVisible();
      await expect(row).toContainText("Anthropic");
      await expect(row).toContainText(fixture.sourcePluginId);
      await expect(row.getByText("anthropic", { exact: true })).toHaveCount(0);
      await expect(row.getByText("ANTHROPIC", { exact: true })).toHaveCount(0);
      await expect(
        adminPage.locator(".vxe-body--row:visible", {
          hasText: otherSourceFixture.sourcePluginId,
        }),
      ).toHaveCount(0);
      await smartCenter.captureEvidence("TC003-invocation-list-protocol-label");
      await smartCenter.openInvocationDetail([
        fixture.purpose,
        fixture.sourcePluginId,
      ]);
      const detail = adminPage.locator('[role="dialog"]').last();
      await expect(
        detail.getByText(/调用详情|Invocation Detail/i),
      ).toBeVisible();
      await expect(detail.getByText(fixture.requestId)).toBeVisible();
      await expect(detail.getByText(fixture.purpose)).toBeVisible();
      await expect(
        detail.getByText("Anthropic", { exact: true }),
      ).toBeVisible();
      await expect(detail.getByText("anthropic", { exact: true })).toHaveCount(
        0,
      );
      await expect(detail.getByText("ANTHROPIC", { exact: true })).toHaveCount(
        0,
      );
      await smartCenter.captureEvidence(
        "TC003-invocation-detail-protocol-label",
      );
      await expect(detail.getByText(/redacted error summary/i)).toBeVisible();
      await expect(adminPage.getByText(/完整 prompt|full prompt/i)).toHaveCount(
        0,
      );
      await expect(adminPage.getByText(/sk-/i)).toHaveCount(0);
      await smartCenter.cancelDrawer();

      await smartCenter.selectInvocationCreatedAtTodayRange();
      const timeRequestPromise = adminPage.waitForRequest((req) => {
        const url = new URL(req.url());
        return (
          req.method() === "GET" &&
          url.pathname.includes("/x/linapro-ai-core/api/v1/ai/invocations") &&
          url.searchParams.has("startedAt") &&
          url.searchParams.has("endedAt") &&
          url.searchParams.get("sourcePluginId") === fixture.sourcePluginId
        );
      });
      await smartCenter.searchInvocations();
      const timeRequest = await timeRequestPromise;
      const timeUrl = new URL(timeRequest.url());
      expect(Number(timeUrl.searchParams.get("startedAt"))).toBeGreaterThan(0);
      expect(Number(timeUrl.searchParams.get("endedAt"))).toBeGreaterThan(0);
      await expect(row).toBeVisible();
      await smartCenter.captureEvidence("TC003-invocation-filters-source-time");

      const cleanResponsePromise = adminPage.waitForResponse((res) => {
        const url = new URL(res.url());
        return (
          res.request().method() === "DELETE" &&
          url.pathname.includes(
            "/x/linapro-ai-core/api/v1/ai/invocations/clean",
          ) &&
          url.searchParams.has("startedAt") &&
          url.searchParams.has("endedAt")
        );
      });
      await smartCenter.confirmInvocationCleanWithDialogRange();
      const cleanResponse = await cleanResponsePromise;
      expect(cleanResponse.status()).toBe(200);
      const cleanPayload = await cleanResponse.json();
      const cleanData = cleanPayload?.data ?? cleanPayload;
      expect(Number(cleanData?.deleted ?? 0)).toBeGreaterThan(0);

      await withAdminApi(async (api) => {
        const targetLogs = await listInvocations(api, {
          capabilityMethod: "generate",
          capabilityType: "text",
          pageNum: 1,
          pageSize: 10,
          purpose: fixture.purpose,
          sourcePluginId: fixture.sourcePluginId,
        });
        expect(Number(targetLogs?.total ?? 0)).toBe(0);

        const oldLogs = await listInvocations(api, {
          capabilityMethod: "generate",
          capabilityType: "text",
          pageNum: 1,
          pageSize: 10,
          purpose: oldFixture.purpose,
          sourcePluginId: oldFixture.sourcePluginId,
        });
        expect(Number(oldLogs?.total ?? 0)).toBeGreaterThanOrEqual(1);
      });
      await smartCenter.captureEvidence("TC003-invocation-range-clean");
    } finally {
      deleteInvocationLog(fixture.requestId);
      deleteInvocationLog(otherSourceFixture.requestId);
      deleteInvocationLog(oldFixture.requestId);
    }
  });
});
