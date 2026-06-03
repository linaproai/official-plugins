import type { APIRequestContext } from "@playwright/test";

import { pluginApiPath } from "@host-tests/fixtures/config";
import { createAdminApiContext } from "@host-tests/fixtures/plugin";
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from "@host-tests/support/postgres";

const pluginId = "linapro-ai-core";

export type AiProviderModelFixture = {
  anthropicEndpointUrl?: string;
  capabilityMethod?: string;
  capabilityType?: string;
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
) {
  if (!response.ok()) {
    throw new Error(
      `${message}, status=${response.status()}, body=${await response.text()}`,
    );
  }
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
    supportedEfforts?: string[];
    supportsThinking?: number;
    websiteUrl?: string;
  },
): Promise<AiProviderModelFixture> {
  const openaiEndpointUrl = input.openaiEndpointUrl ?? "http://127.0.0.1:65535/v1";
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
  await assertOk(providerResponse, "创建 AI 供应商失败");
  const provider = unwrapApiData(await providerResponse.json());
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
        capabilityMethod: "generate",
        capabilityType: "text",
        enabled: 1,
        endpointId,
        maxInputTokens: 4096,
        maxOutputTokens: 512,
        modelName: input.modelName,
        protocol: "openai",
        supportedEfforts: input.supportedEfforts ?? ["low", "medium", "high"],
        supportsThinking: input.supportsThinking ?? 1,
      },
    },
  );
  await assertOk(modelResponse, "创建 AI 模型失败");
  const model = unwrapApiData(await modelResponse.json());

  return {
    anthropicEndpointUrl: input.anthropicEndpointUrl,
    capabilityMethod: "generate",
    capabilityType: "text",
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
  await assertOk(response, "创建 AI 供应商端点失败");
  const endpoint = unwrapApiData(await response.json());
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
  await assertOk(response, "查询 AI 供应商端点失败");
  const out = unwrapApiData(await response.json());
  return out?.list ?? [];
}

export function deleteProviderEndpointRaw(
  api: APIRequestContext,
  providerId: number,
  endpointId: number,
) {
  return api.delete(
    pluginApiPath(pluginId, `ai/providers/${providerId}/endpoints/${endpointId}`),
  );
}

export async function createProviderModel(
  api: APIRequestContext,
  providerId: number,
  input: {
    capabilityMethod: string;
    capabilityType: string;
    modelName: string;
    protocol: string;
    enabled?: number;
    endpointId: number;
    maxInputTokens?: number;
    maxOutputTokens?: number;
    supportedEfforts?: string[];
    supportsThinking?: number;
  },
) {
  const response = await api.post(
    pluginApiPath(pluginId, `ai/providers/${providerId}/models`),
    {
      data: {
        capabilityMethod: input.capabilityMethod,
        capabilityType: input.capabilityType,
        enabled: input.enabled ?? 1,
        endpointId: input.endpointId,
        maxInputTokens: input.maxInputTokens ?? 0,
        maxOutputTokens: input.maxOutputTokens ?? 0,
        modelName: input.modelName,
        protocol: input.protocol,
        supportedEfforts: input.supportedEfforts ?? [],
        supportsThinking: input.supportsThinking ?? 0,
      },
    },
  );
  await assertOk(response, "创建 AI 模型失败");
  const model = unwrapApiData(await response.json());
  return Number(model?.id || 0);
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
  await assertOk(response, "查询 AI 模型能力失败");
  const out = unwrapApiData(await response.json());
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
  defaultEffort = "low",
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

export function deleteProviderRaw(api: APIRequestContext, providerId: number) {
  return api.delete(pluginApiPath(pluginId, `ai/providers/${providerId}`));
}

export async function listProviderModels(
  api: APIRequestContext,
  providerId: number,
  capabilityType = "text",
  capabilityMethod = "generate",
) {
  const response = await api.get(
    pluginApiPath(pluginId, `ai/providers/${providerId}/models`),
    {
      params: {
        capabilityMethod,
        capabilityType,
        pageNum: 1,
        pageSize: 100,
      },
    },
  );
  await assertOk(response, "查询 AI 模型失败");
  const out = unwrapApiData(await response.json());
  return out?.list ?? [];
}

export async function clearTier(
  _api: APIRequestContext,
  code: "advanced" | "basic" | "standard",
  capabilityType = "text",
  capabilityMethod = "generate",
) {
  const defaultEffort =
    capabilityType === "text" && capabilityMethod === "generate"
      ? code === "basic"
        ? "low"
        : code === "standard"
          ? "medium"
          : "high"
      : "";
  const escapedCode = pgEscapeLiteral(code);
  const escapedEffort = pgEscapeLiteral(defaultEffort);
  const escapedType = pgEscapeLiteral(capabilityType);
  const escapedMethod = pgEscapeLiteral(capabilityMethod);
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_tier_binding WHERE tier_id IN (SELECT id FROM plugin_linapro_ai_tier WHERE capability_type = '${escapedType}' AND capability_method = '${escapedMethod}' AND code = '${escapedCode}') AND priority = 0;`,
    `UPDATE plugin_linapro_ai_tier SET enabled = 1, default_effort = '${escapedEffort}', last_test_status = '', last_test_latency_ms = 0, last_test_error_summary = '', last_test_at = NULL WHERE capability_type = '${escapedType}' AND capability_method = '${escapedMethod}' AND code = '${escapedCode}';`,
  ]);
}

export function insertInvocationLog(input: {
  purpose: string;
  requestId: string;
  assetSummaryJson?: string;
  capabilityMethod?: string;
  capabilityType?: string;
  metadataSummaryJson?: string;
  operationSummaryJson?: string;
  status?: string;
}): AiInvocationFixture {
  const purpose = input.purpose.trim();
  const requestId = input.requestId.trim();
  const capabilityType = input.capabilityType ?? "text";
  const capabilityMethod = input.capabilityMethod ?? "generate";
  const assetSummaryJson = input.assetSummaryJson ?? "{}";
  const operationSummaryJson = input.operationSummaryJson ?? "{}";
  const metadataSummaryJson = input.metadataSummaryJson ?? "{}";
  const status = input.status ?? "failed";
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
      'e2e-ai-core',
      0,
      1,
      0,
      0,
      'E2E Provider',
      'e2e-model',
      'openai',
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
      NOW()
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
  };
}

export function deleteInvocationLog(requestId: string) {
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_invocation WHERE request_id = '${pgEscapeLiteral(requestId)}';`,
  ]);
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
  params: Record<string, unknown>,
) {
  const response = await api.get(pluginApiPath(pluginId, "ai/provider-operations"), {
    params,
  });
  await assertOk(response, "查询 AI provider operation 失败");
  const out = unwrapApiData(await response.json());
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
      `删除 AI 供应商失败, status=${response.status()}, body=${await response.text()}`,
    );
  }
}
