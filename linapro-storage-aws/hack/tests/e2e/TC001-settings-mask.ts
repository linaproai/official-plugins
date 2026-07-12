/**
 * TC001 linapro-storage-aws 设置页密钥掩码与保存
 */
import { expect, test } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  prepareSourcePluginsBaseline,
} from "@host-tests/fixtures/plugin";
import { pluginApiPath } from "@host-tests/fixtures/config";

const pluginID = "linapro-storage-aws";

test.describe("TC001 linapro-storage-aws 设置保存与掩码", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([pluginID]);
  });

  test("TC001a: 保存后密钥掩码回显且空提交保持", async () => {
    const api = await createAdminApiContext();
    try {
      const putRes = await api.put(pluginApiPath(pluginID, "settings"), {
        data: {
          accessKeyID: "AKIA_TEST",
          secretAccessKey: "super-secret-value",
          region: "us-east-1",
          bucket: "demo-bucket",
          pathPrefix: "pfx",
        },
      });
      expect(putRes.ok()).toBeTruthy();
      const putBody = await putRes.json();
      const saved = putBody?.data?.settings ?? putBody?.settings;
      expect(saved?.accessKeyID).toBe("AKIA_TEST");
      expect(saved?.secretAccessKeyConfigured).toBeTruthy();
      expect(saved?.secretAccessKeyMasked).toMatch(/\*+/);
      expect(saved?.bucket).toBe("demo-bucket");
      expect(saved?.region).toBe("us-east-1");
      expect(saved?.endpoint).toBeUndefined();
      expect(saved?.forcePathStyle).toBeUndefined();

      const keepRes = await api.put(pluginApiPath(pluginID, "settings"), {
        data: {
          accessKeyID: "AKIA_TEST",
          secretAccessKey: "",
          region: "us-west-2",
          bucket: "demo-bucket-2",
          pathPrefix: "pfx",
        },
      });
      expect(keepRes.ok()).toBeTruthy();

      const getRes = await api.get(pluginApiPath(pluginID, "settings"));
      expect(getRes.ok()).toBeTruthy();
      const getBody = await getRes.json();
      const loaded = getBody?.data?.settings ?? getBody?.settings;
      expect(loaded?.bucket).toBe("demo-bucket-2");
      expect(loaded?.region).toBe("us-west-2");
      expect(loaded?.secretAccessKeyConfigured).toBeTruthy();
      expect(loaded?.secretAccessKeyMasked).toMatch(/\*+/);
      expect(JSON.stringify(loaded)).not.toContain("super-secret-value");
    } finally {
      await api.dispose();
    }
  });
});
