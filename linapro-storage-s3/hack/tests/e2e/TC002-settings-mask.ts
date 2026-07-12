/**
 * TC002 linapro-storage-s3 设置页密钥掩码与保存（S3 协议字段）
 */
import { expect, test } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  prepareSourcePluginsBaseline,
} from "@host-tests/fixtures/plugin";
import { pluginApiPath } from "@host-tests/fixtures/config";

const pluginID = "linapro-storage-s3";

test.describe("TC002 linapro-storage-s3 设置保存与掩码", () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline([pluginID]);
  });

  test("TC002a: 保存后密钥掩码回显且空提交保持", async () => {
    const api = await createAdminApiContext();
    try {
      const putRes = await api.put(pluginApiPath(pluginID, "settings"), {
        data: {
          accessKeyID: "AKIA_TEST",
          secretAccessKey: "super-secret-value",
          region: "",
          bucket: "demo-bucket",
          endpoint: "http://minio:9000",
          pathPrefix: "pfx",
          forcePathStyle: true,
        },
      });
      expect(putRes.ok()).toBeTruthy();
      const putBody = await putRes.json();
      const saved = putBody?.data?.settings ?? putBody?.settings;
      expect(saved?.accessKeyID).toBe("AKIA_TEST");
      expect(saved?.secretAccessKeyConfigured).toBeTruthy();
      expect(saved?.secretAccessKeyMasked).toMatch(/\*+/);
      expect(saved?.bucket).toBe("demo-bucket");
      expect(saved?.endpoint).toBe("http://minio:9000");
      expect(saved?.forcePathStyle).toBeTruthy();

      const keepRes = await api.put(pluginApiPath(pluginID, "settings"), {
        data: {
          accessKeyID: "AKIA_TEST",
          secretAccessKey: "",
          region: "",
          bucket: "demo-bucket-2",
          endpoint: "http://minio:9000",
          pathPrefix: "pfx",
          forcePathStyle: true,
        },
      });
      expect(keepRes.ok()).toBeTruthy();

      const getRes = await api.get(pluginApiPath(pluginID, "settings"));
      expect(getRes.ok()).toBeTruthy();
      const getBody = await getRes.json();
      const loaded = getBody?.data?.settings ?? getBody?.settings;
      expect(loaded?.bucket).toBe("demo-bucket-2");
      expect(loaded?.secretAccessKeyConfigured).toBeTruthy();
      expect(loaded?.secretAccessKeyMasked).toMatch(/\*+/);
      expect(JSON.stringify(loaded)).not.toContain("super-secret-value");
    } finally {
      await api.dispose();
    }
  });
});
