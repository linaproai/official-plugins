import type { APIRequestContext } from "@host-tests/support/playwright";

import { pluginApiPath } from "@host-tests/fixtures/config";
import { createAdminApiContext } from "@host-tests/fixtures/plugin";
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from "@host-tests/support/postgres";

const pluginId = "linapro-ai-core";

export type AiProviderModelFixture = {
  anthropicEndpointUrl?: string;
  endpointId: number;
  maskedApiKey: string;
  modelId: number;
  modelName: string;
  openaiEndpointUrl: string;
  providerId: number;
  providerName: string;
  websiteUrl: string;
};

export type AiInvocationFixture = {
  assetSummaryJson?: string;
  capabilityMethod?: string;
  capabilityType?: string;
  metadataSummaryJson?: string;
  operationSummaryJson?: string;
  purpose: string;
  requestId: string;
  sourcePluginId: string;
};

export type AiProviderOperationFixture = {
  operationRef: string;
  purpose: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

async function assertOk(
  response: Awaited<ReturnType<APIRequestContext["get"]>>,
  message: string,
): Promise<any> {
  const body = await response.text();
  let payload: any;
  if (body.trim()) {
    try {
      payload = JSON.parse(body);
    } catch {
      if (!response.ok()) {
        throw new Error(`${message}, status=${response.status()}, body=${body}`);
      }
    }
  }
  if (!response.ok()) {
    throw new Error(
      `${message}, status=${response.status()}, body=${body}`,
    );
  }
  if (
    payload &&
    typeof payload === "object" &&
    typeof payload.code === "number" &&
    payload.code !== 0
  ) {
    throw new Error(`${message}, businessCode=${payload.code}, body=${body}`);
  }
  return payload;
}

export async function withAdminApi<T>(
  run: (api: APIRequestContext) => Promise<T>,
): Promise<T> {
  const api = await createAdminApiContext();
  try {
    return await run(api);
  } finally {
    await api.dispose();
  }
}

export async function createProviderWithModel(
  api: APIRequestContext,
  input: {
    modelName: string;
    providerName: string;
    anthropicEndpointUrl?: string;
    openaiEndpointUrl?: string;
    secretRef?: string;
    websiteUrl?: string;
  },
): Promise<AiProviderModelFixture> {
  const openaiEndpointUrl =
    input.openaiEndpointUrl ?? "http://127.0.0.1:65535/v1";
  const secretRef = input.secretRef ?? "sk-e2e-placeholder";
  const websiteUrl =
    input.websiteUrl ??
    `https://example.com/${encodeURIComponent(input.providerName)}`;
  const providerResponse = await api.post(
    pluginApiPath(pluginId, "ai/providers"),
    {
      data: {
        enabled: 1,
        name: input.providerName,
        websiteUrl,
      },
    },
  );
  const provider = unwrapApiData(
    await assertOk(providerResponse, "创建 AI 渠道失败"),
  );
  const providerId = Number(provider?.id || 0);
  const endpointId = await createProviderEndpoint(api, providerId, {
    baseUrl: openaiEndpointUrl,
    protocol: "openai",
    secretRef,
  });
  if (input.anthropicEndpointUrl) {
    await createProviderEndpoint(api, providerId, {
      baseUrl: input.anthropicEndpointUrl,
      protocol: "anthropic",
      secretRef,
    });
  }

  const modelResponse = await api.post(
    pluginApiPath(pluginId, `ai/providers/${providerId}/models`),
    {
      data: {
        enabled: 1,
        endpointId,
        modelName: input.modelName,
        protocol: "openai",
      },
    },
  );
  const model = unwrapApiData(
    await assertOk(modelResponse, "创建 AI 模型失败"),
  );

  return {
    anthropicEndpointUrl: input.anthropicEndpointUrl,
    endpointId,
    maskedApiKey: maskApiKeyForExpectation(secretRef),
    modelId: Number(model?.id || 0),
    modelName: input.modelName,
    openaiEndpointUrl,
    providerId,
    providerName: input.providerName,
    websiteUrl,
  };
}

export async function createProviderEndpoint(
  api: APIRequestContext,
  providerId: number,
  input: {
    baseUrl: string;
    protocol: string;
    enabled?: number;
    metadataJson?: string;
    secretRef?: string;
  },
) {
  const response = await api.post(
    pluginApiPath(pluginId, `ai/providers/${providerId}/endpoints`),
    {
      data: {
        baseUrl: input.baseUrl,
        enabled: input.enabled ?? 1,
        metadataJson: input.metadataJson ?? "{}",
        protocol: input.protocol,
        secretRef: input.secretRef ?? "sk-e2e-endpoint",
      },
    },
  );
  const endpoint = unwrapApiData(
    await assertOk(response, "创建 AI 渠道端点失败"),
  );
  return Number(endpoint?.id || 0);
}

export async function listProviderEndpoints(
  api: APIRequestContext,
  providerId: number,
  params?: { enabled?: number; protocol?: string },
) {
  const response = await api.get(
    pluginApiPath(pluginId, `ai/providers/${providerId}/endpoints`),
    { params },
  );
  const out = unwrapApiData(
    await assertOk(response, "查询 AI 渠道端点失败"),
  );
  return out?.list ?? [];
}

export function deleteProviderEndpointRaw(
  api: APIRequestContext,
  providerId: number,
  endpointId: number,
) {
  return api.delete(
    pluginApiPath(
      pluginId,
      `ai/providers/${providerId}/endpoints/${endpointId}`,
    ),
  );
}

export async function createProviderModel(
  api: APIRequestContext,
  providerId: number,
  input: {
    modelName: string;
    protocol: string;
    enabled?: number;
    endpointId: number;
  },
) {
  const response = await api.post(
    pluginApiPath(pluginId, `ai/providers/${providerId}/models`),
    {
      data: {
        enabled: input.enabled ?? 1,
        endpointId: input.endpointId,
        modelName: input.modelName,
        protocol: input.protocol,
      },
    },
  );
  const model = unwrapApiData(await assertOk(response, "创建 AI 模型失败"));
  return Number(model?.id || 0);
}

export function insertProviderModelIdentityOnly(input: {
  endpointId: number;
  modelName: string;
  protocol: string;
  providerId: number;
}) {
  const modelName = input.modelName.trim();
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_model_capability WHERE model_id IN (SELECT id FROM plugin_linapro_ai_model WHERE provider_id = ${Number(input.providerId)} AND protocol = '${pgEscapeLiteral(input.protocol)}' AND model_name = '${pgEscapeLiteral(modelName)}');`,
    `DELETE FROM plugin_linapro_ai_model WHERE provider_id = ${Number(input.providerId)} AND protocol = '${pgEscapeLiteral(input.protocol)}' AND model_name = '${pgEscapeLiteral(modelName)}';`,
    `INSERT INTO plugin_linapro_ai_model (
      provider_id,
      endpoint_id,
      model_name,
      protocol,
      source,
      enabled,
      created_at,
      updated_at
    ) VALUES (
      ${Number(input.providerId)},
      ${Number(input.endpointId)},
      '${pgEscapeLiteral(modelName)}',
      '${pgEscapeLiteral(input.protocol)}',
      'api',
      1,
      NOW(),
      NOW()
    );`,
  ]);
}

export async function saveModelCapabilities(
  api: APIRequestContext,
  modelId: number,
  items: Array<Record<string, unknown>>,
) {
  const response = await api.put(
    pluginApiPath(pluginId, `ai/models/${modelId}/capabilities`),
    { data: { items } },
  );
  await assertOk(response, "保存 AI 模型能力失败");
}

export async function listModelCapabilities(
  api: APIRequestContext,
  modelId: number,
) {
  const response = await api.get(
    pluginApiPath(pluginId, `ai/models/${modelId}/capabilities`),
  );
  const out = unwrapApiData(
    await assertOk(response, "查询 AI 模型能力失败"),
  );
  return out?.list ?? [];
}

function maskApiKeyForExpectation(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.includes("***")) {
    return trimmed;
  }
  if (trimmed.length <= 2) {
    return "*".repeat(trimmed.length);
  }
  const separator = trimmed.indexOf("-");
  const prefix =
    separator >= 0 && separator < trimmed.length - 2
      ? trimmed.slice(0, separator + 1)
      : "";
  return `${prefix}**********${trimmed.slice(-2)}`;
}

export async function bindTier(
  api: APIRequestContext,
  code: "advanced" | "basic" | "standard",
  fixture: AiProviderModelFixture,
  defaultEffort = "",
) {
  const response = await api.put(pluginApiPath(pluginId, `ai/tiers/${code}`), {
    data: {
      defaultEffort,
      enabled: 1,
      capabilityMethod: "generate",
      capabilityType: "text",
      modelId: fixture.modelId,
      providerId: fixture.providerId,
    },
  });
  await assertOk(response, `绑定 AI 档位失败: ${code}`);
}

export async function bindCapabilityTier(
  api: APIRequestContext,
  code: "advanced" | "basic" | "standard",
  fixture: Pick<AiProviderModelFixture, "modelId" | "providerId">,
  input: {
    capabilityMethod: string;
    capabilityType: string;
    defaultEffort?: string;
    enabled?: number;
  },
) {
  const response = await api.put(pluginApiPath(pluginId, `ai/tiers/${code}`), {
    data: {
      capabilityMethod: input.capabilityMethod,
      capabilityType: input.capabilityType,
      defaultEffort: input.defaultEffort ?? "",
      enabled: input.enabled ?? 1,
      modelId: fixture.modelId,
      providerId: fixture.providerId,
    },
  });
  await assertOk(response, `绑定 AI 能力档位失败: ${code}`);
}

export function updateTierRaw(
  api: APIRequestContext,
  code: "advanced" | "basic" | "standard",
  data: Record<string, unknown>,
) {
  return api.put(pluginApiPath(pluginId, `ai/tiers/${code}`), {
    data: {
      capabilityMethod: "generate",
      capabilityType: "text",
      ...data,
    },
  });
}

export async function listTiers(
  api: APIRequestContext,
  capabilityType = "text",
  capabilityMethod = "generate",
) {
  const response = await api.get(pluginApiPath(pluginId, "ai/tiers"), {
    params: {
      capabilityMethod,
      capabilityType,
    },
  });
  const out = unwrapApiData(await assertOk(response, "查询 AI 档位失败"));
  return out?.list ?? [];
}

export function deleteProviderRaw(api: APIRequestContext, providerId: number) {
  return api.delete(pluginApiPath(pluginId, `ai/providers/${providerId}`));
}

export async function findProviderByName(api: APIRequestContext, name: string) {
  const response = await api.get(pluginApiPath(pluginId, "ai/providers"), {
    params: {
      keyword: name,
      pageNum: 1,
      pageSize: 10,
    },
  });
  const out = unwrapApiData(await assertOk(response, "查询 AI 渠道失败"));
  return (out?.list || []).find((item: any) => item?.name === name);
}

export async function listProviderModels(
  api: APIRequestContext,
  providerId: number,
) {
  const response = await api.get(
    pluginApiPath(pluginId, `ai/providers/${providerId}/models`),
    {
      params: {
        pageNum: 1,
        pageSize: 100,
      },
    },
  );
  const out = unwrapApiData(await assertOk(response, "查询 AI 模型失败"));
  return out?.list ?? [];
}

export async function clearTier(
  _api: APIRequestContext,
  code: "advanced" | "basic" | "standard",
  capabilityType = "text",
  capabilityMethod = "generate",
) {
  const defaultEffort = "";
  const escapedCode = pgEscapeLiteral(code);
  const escapedEffort = pgEscapeLiteral(defaultEffort);
  const escapedType = pgEscapeLiteral(capabilityType);
  const escapedMethod = pgEscapeLiteral(capabilityMethod);
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_tier_binding WHERE tier_id IN (SELECT id FROM plugin_linapro_ai_tier WHERE capability_type = '${escapedType}' AND capability_method = '${escapedMethod}' AND code = '${escapedCode}') AND priority = 0;`,
    `UPDATE plugin_linapro_ai_tier SET enabled = 1, default_effort = '${escapedEffort}', last_test_status = '', last_test_latency_ms = 0, last_test_error_summary = '', last_test_at = NULL WHERE capability_type = '${escapedType}' AND capability_method = '${escapedMethod}' AND code = '${escapedCode}';`,
  ]);
}

export function clearTierUpdatedAt(
  code: "advanced" | "basic" | "standard",
  capabilityType = "text",
  capabilityMethod = "generate",
) {
  const escapedCode = pgEscapeLiteral(code);
  const escapedType = pgEscapeLiteral(capabilityType);
  const escapedMethod = pgEscapeLiteral(capabilityMethod);
  execPgSQLStatements([
    `UPDATE plugin_linapro_ai_tier SET updated_at = NULL WHERE capability_type = '${escapedType}' AND capability_method = '${escapedMethod}' AND code = '${escapedCode}';`,
  ]);
}

export function insertInvocationLog(input: {
  purpose: string;
  requestId: string;
  assetSummaryJson?: string;
  capabilityMethod?: string;
  capabilityType?: string;
  createdAtSql?: string;
  metadataSummaryJson?: string;
  operationSummaryJson?: string;
  protocol?: string;
  sourcePluginId?: string;
  status?: string;
}): AiInvocationFixture {
  const purpose = input.purpose.trim();
  const requestId = input.requestId.trim();
  const capabilityType = input.capabilityType ?? "text";
  const capabilityMethod = input.capabilityMethod ?? "generate";
  const assetSummaryJson = input.assetSummaryJson ?? "{}";
  const operationSummaryJson = input.operationSummaryJson ?? "{}";
  const metadataSummaryJson = input.metadataSummaryJson ?? "{}";
  const protocol = input.protocol ?? "openai";
  const sourcePluginId = input.sourcePluginId ?? "e2e-ai-core";
  const status = input.status ?? "failed";
  const createdAtSql = input.createdAtSql ?? "NOW()";
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_invocation WHERE request_id = '${pgEscapeLiteral(requestId)}';`,
    `INSERT INTO plugin_linapro_ai_invocation (
      request_id,
      capability_type,
      capability_method,
      purpose,
      tier_code,
      source_plugin_id,
      tenant_id,
      user_id,
      provider_id,
      model_id,
      provider_name,
      model_name,
      protocol,
      thinking_effort,
      status,
      input_tokens,
      output_tokens,
      latency_ms,
      asset_summary_json,
      operation_summary_json,
      metadata_summary_json,
      error_code,
      error_summary,
      created_at
    ) VALUES (
      '${pgEscapeLiteral(requestId)}',
      '${pgEscapeLiteral(capabilityType)}',
      '${pgEscapeLiteral(capabilityMethod)}',
      '${pgEscapeLiteral(purpose)}',
      'standard',
      '${pgEscapeLiteral(sourcePluginId)}',
      0,
      1,
      0,
      0,
      'E2E Provider',
      'e2e-model',
      '${pgEscapeLiteral(protocol)}',
      'medium',
      '${pgEscapeLiteral(status)}',
      11,
      7,
      123,
      '${pgEscapeLiteral(assetSummaryJson)}',
      '${pgEscapeLiteral(operationSummaryJson)}',
      '${pgEscapeLiteral(metadataSummaryJson)}',
      'AI_CORE_PROVIDER_HTTP_ERROR',
      'Provider returned a redacted error summary',
      ${createdAtSql}
    );`,
  ]);
  return {
    assetSummaryJson,
    capabilityMethod,
    capabilityType,
    metadataSummaryJson,
    operationSummaryJson,
    purpose,
    requestId,
    sourcePluginId,
  };
}

export function deleteInvocationLog(requestId: string) {
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_invocation WHERE request_id = '${pgEscapeLiteral(requestId)}';`,
  ]);
}

export async function listInvocations(
  api: APIRequestContext,
  params: Record<string, boolean | number | string>,
) {
  const response = await api.get(pluginApiPath(pluginId, "ai/invocations"), {
    params,
  });
  return unwrapApiData(await assertOk(response, "查询 AI 调用日志失败"));
}

export function insertProviderOperation(input: {
  operationRef: string;
  purpose: string;
  assetSummaryJson?: string;
  capabilityMethod?: string;
  capabilityType?: string;
  errorSummary?: string;
  status?: string;
}): AiProviderOperationFixture {
  const operationRef = input.operationRef.trim();
  const purpose = input.purpose.trim();
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_provider_operation WHERE operation_ref = '${pgEscapeLiteral(operationRef)}';`,
    `INSERT INTO plugin_linapro_ai_provider_operation (
      operation_ref,
      capability_type,
      capability_method,
      purpose,
      source_plugin_id,
      provider_name,
      model_name,
      protocol,
      status,
      next_poll_after_ms,
      expires_at,
      asset_summary_json,
      error_code,
      error_summary,
      created_at,
      updated_at
    ) VALUES (
      '${pgEscapeLiteral(operationRef)}',
      '${pgEscapeLiteral(input.capabilityType ?? "video")}',
      '${pgEscapeLiteral(input.capabilityMethod ?? "generate")}',
      '${pgEscapeLiteral(purpose)}',
      'e2e-ai-core',
      'E2E Provider',
      'e2e-video-model',
      'openai',
      '${pgEscapeLiteral(input.status ?? "running")}',
      3000,
      NOW() + INTERVAL '1 hour',
      '${pgEscapeLiteral(input.assetSummaryJson ?? "{}")}',
      'AI_CORE_PROVIDER_HTTP_ERROR',
      '${pgEscapeLiteral(input.errorSummary ?? "Provider operation failed with redacted summary")}',
      NOW(),
      NOW()
    );`,
  ]);
  return { operationRef, purpose };
}

export function deleteProviderOperation(operationRef: string) {
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_provider_operation WHERE operation_ref = '${pgEscapeLiteral(operationRef)}';`,
  ]);
}

export async function listProviderOperations(
  api: APIRequestContext,
  params: Record<string, boolean | number | string>,
) {
  const response = await api.get(
    pluginApiPath(pluginId, "ai/provider-operations"),
    {
      params,
    },
  );
  const out = unwrapApiData(
    await assertOk(response, "查询 AI provider operation 失败"),
  );
  return out;
}

export async function deleteProvider(
  api: APIRequestContext,
  providerId: number,
) {
  const response = await api.delete(
    pluginApiPath(pluginId, `ai/providers/${providerId}`),
  );
  if (!response.ok() && response.status() !== 404) {
    throw new Error(
      `删除 AI 渠道失败, status=${response.status()}, body=${await response.text()}`,
    );
  }
}
