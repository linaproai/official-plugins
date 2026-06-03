import { test, expect } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  ensureSourcePluginEnabledViaAPI,
} from "@host-tests/fixtures/plugin";

test.describe("TC-3 Server monitor visibility-aware polling", () => {
  test.beforeEach(async () => {
    const adminApi = await createAdminApiContext();
    try {
      await ensureSourcePluginEnabledViaAPI(adminApi, "linapro-monitor-server");
    } finally {
      await adminApi.dispose();
    }
  });

  test("TC-3a: hidden tab pauses polling and visible tab refreshes immediately", async ({
    authenticatedPage,
  }) => {
    const page = authenticatedPage;
    const installVisibilityMock = () => {
      let state: DocumentVisibilityState = "visible";
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => state,
      });
      Object.defineProperty(document, "hidden", {
        configurable: true,
        get: () => state !== "visible",
      });
      (window as any).__setE2EVisibility = (
        nextState: DocumentVisibilityState,
      ) => {
        state = nextState;
        document.dispatchEvent(new Event("visibilitychange"));
      };
    };

    await page.addInitScript(installVisibilityMock);
    await page.evaluate(installVisibilityMock);

    const monitorRequests: string[] = [];
    page.on("request", (request) => {
      if (
        request.method() === "GET" &&
        request.url().includes("/x/linapro-monitor-server/api/v1/monitor/server")
      ) {
        monitorRequests.push(request.url());
      }
    });

    await page.clock.install();
    const initialMonitorResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        response.url().includes("/x/linapro-monitor-server/api/v1/monitor/server") &&
        response.status() === 200,
    );
    await page.goto("/monitor/server", { waitUntil: "domcontentloaded" });
    await initialMonitorResponse;

    await page.evaluate(() => {
      (window as any).__setE2EVisibility("hidden");
    });
    const requestsAfterHidden = monitorRequests.length;
    await page.clock.runFor(31_000);
    expect(monitorRequests).toHaveLength(requestsAfterHidden);
    await page.clock.resume();

    const visibleResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        response.url().includes("/x/linapro-monitor-server/api/v1/monitor/server") &&
        response.status() === 200,
    );
    await page.evaluate(() => {
      (window as any).__setE2EVisibility("visible");
    });
    await visibleResponse;

    expect(monitorRequests.length).toBe(requestsAfterHidden + 1);
  });
});
