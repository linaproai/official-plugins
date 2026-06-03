import { expect, test } from "@host-tests/fixtures/auth";
import { config, pluginApiPath } from "@host-tests/fixtures/config";
import { prepareSourcePluginsBaseline } from "@host-tests/fixtures/plugin";
import { request as playwrightRequest } from "@host-tests/support/playwright";
import { SmartCenterPage } from "../pages/SmartCenterPage";
import {
  deleteInvocationLog,
  deleteProviderOperation,
  insertInvocationLog,
  insertProviderOperation,
  listProviderOperations,
  withAdminApi,
} from "../support/ai-core-api";

test.describe("TC-6 智能中心多模态调用日志", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(["linapro-ai-core"]);
  });

  test("TC-6a: 多模态日志展示资产和 operation 摘要并保持脱敏", async ({ adminPage }) => {
    const suffix = Date.now();
    const purpose = `e2e.multimodal.${suffix}`;
    const requestId = `e2e-multimodal-${suffix}`;
    const operationRef = `op-e2e-${suffix}`;
    const invocation = insertInvocationLog({
      assetSummaryJson: '{"assetRef":"asset://e2e/video","mimeType":"video/mp4"}',
      capabilityMethod: "generate",
      capabilityType: "video",
      metadataSummaryJson: '{"frames":12}',
      operationSummaryJson: `{"operationRef":"${operationRef}","status":"running"}`,
      purpose,
      requestId,
      status: "failed",
    });
    const operation = insertProviderOperation({
      assetSummaryJson: '{"assetRef":"asset://e2e/video"}',
      capabilityMethod: "generate",
      capabilityType: "video",
      operationRef,
      purpose,
      status: "running",
    });

    const smartCenter = new SmartCenterPage(adminPage);
    try {
      await smartCenter.gotoInvocations();
      await smartCenter.filterInvocationsByCapabilityAndPurpose(
        "video.generate",
        invocation.purpose,
      );
      const row = adminPage.locator(".vxe-body--row:visible", {
        hasText: invocation.purpose,
      });
      await expect(row).toBeVisible();
      await expect(row).toContainText("video");
      await expect(row).toContainText("generate");
      await smartCenter.captureEvidence("TC006-video-invocation-list");

      await smartCenter.openInvocationDetail();
      const detail = adminPage.locator('[role="dialog"]').last();
      await expect(detail.getByText("调用详情")).toBeVisible();
      await expect(detail.getByText(requestId)).toBeVisible();
      await expect(detail.getByText("video.generate")).toBeVisible();
      await expect(detail.getByText("资产摘要")).toBeVisible();
      await expect(detail.getByText("asset://e2e/video")).toBeVisible();
      await expect(detail.getByText("Operation 摘要")).toBeVisible();
      await expect(detail.getByText(operationRef)).toBeVisible();
      await expect(detail.getByText("元数据摘要")).toBeVisible();
      await expect(detail.getByText('"frames":12')).toBeVisible();
      await expect(detail.getByText(/完整 prompt|full prompt/i)).toHaveCount(0);
      await expect(detail.getByText(/sk-/i)).toHaveCount(0);
      await smartCenter.captureEvidence("TC006-video-invocation-detail");

      await withAdminApi(async (api) => {
        const out = await listProviderOperations(api, {
          capabilityMethod: "generate",
          capabilityType: "video",
          operationRef: operation.operationRef,
          purpose: operation.purpose,
          status: "running",
        });
        expect(out.total).toBe(1);
        expect(out.list[0].operationRef).toBe(operationRef);
        expect(out.list[0].assetSummaryJson).toContain("asset://e2e/video");
      });

      const unauthenticated = await playwrightRequest.newContext({
        baseURL: config.apiBaseURL,
      });
      try {
        const response = await unauthenticated.get(
          pluginApiPath("linapro-ai-core", "ai/provider-operations"),
        );
        expect([401, 403]).toContain(response.status());
      } finally {
        await unauthenticated.dispose();
      }
    } finally {
      await smartCenter.cancelDrawer().catch(() => {});
      deleteInvocationLog(requestId);
      deleteProviderOperation(operationRef);
    }
  });
});
