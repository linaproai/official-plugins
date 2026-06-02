import type { APIRequestContext } from '@playwright/test';

import { pluginApiPath } from '@host-tests/fixtures/config';
import { createAdminApiContext } from '@host-tests/fixtures/plugin';
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from '@host-tests/support/postgres';

const pluginId = 'linapro-ai-core';

export type AiProviderModelFixture = {
  modelId: number;
  modelName: string;
  providerId: number;
  providerName: string;
};

export type AiInvocationFixture = {
  purpose: string;
  requestId: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

async function assertOk(response: Awaited<ReturnType<APIRequestContext['get']>>, message: string) {
  if (!response.ok()) {
    throw new Error(`${message}, status=${response.status()}, body=${await response.text()}`);
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
    supportedEfforts?: string[];
    supportsThinking?: number;
  },
): Promise<AiProviderModelFixture> {
  const providerResponse = await api.post(pluginApiPath(pluginId, 'ai/providers'), {
    data: {
      apiKeySecretRef: 'sk-e2e-placeholder',
      enabled: 1,
      name: input.providerName,
      openaiBaseUrl: 'http://127.0.0.1:65535/v1',
    },
  });
  await assertOk(providerResponse, '创建 AI 供应商失败');
  const provider = unwrapApiData(await providerResponse.json());
  const providerId = Number(provider?.id || 0);

  const modelResponse = await api.post(pluginApiPath(pluginId, `ai/providers/${providerId}/models`), {
    data: {
      capabilityType: 'text',
      enabled: 1,
      maxInputTokens: 4096,
      maxOutputTokens: 512,
      modelName: input.modelName,
      protocol: 'openai',
      supportedEfforts: input.supportedEfforts ?? ['low', 'medium', 'high'],
      supportsThinking: input.supportsThinking ?? 1,
    },
  });
  await assertOk(modelResponse, '创建 AI 模型失败');
  const model = unwrapApiData(await modelResponse.json());

  return {
    modelId: Number(model?.id || 0),
    modelName: input.modelName,
    providerId,
    providerName: input.providerName,
  };
}

export async function bindTier(
  api: APIRequestContext,
  code: 'advanced' | 'basic' | 'standard',
  fixture: AiProviderModelFixture,
  defaultEffort = 'low',
) {
  const response = await api.put(pluginApiPath(pluginId, `ai/tiers/${code}`), {
    data: {
      defaultEffort,
      enabled: 1,
      modelId: fixture.modelId,
      providerId: fixture.providerId,
    },
  });
  await assertOk(response, `绑定 AI 档位失败: ${code}`);
}

export function updateTierRaw(
  api: APIRequestContext,
  code: 'advanced' | 'basic' | 'standard',
  data: Record<string, unknown>,
) {
  return api.put(pluginApiPath(pluginId, `ai/tiers/${code}`), { data });
}

export function deleteProviderRaw(api: APIRequestContext, providerId: number) {
  return api.delete(pluginApiPath(pluginId, `ai/providers/${providerId}`));
}

export async function listProviderModels(api: APIRequestContext, providerId: number) {
  const response = await api.get(pluginApiPath(pluginId, `ai/providers/${providerId}/models`), {
    params: { capabilityType: 'text', pageNum: 1, pageSize: 100 },
  });
  await assertOk(response, '查询 AI 模型失败');
  const out = unwrapApiData(await response.json());
  return out?.list ?? [];
}

export async function clearTier(_api: APIRequestContext, code: 'advanced' | 'basic' | 'standard') {
  const defaultEffort = code === 'basic' ? 'low' : code === 'standard' ? 'medium' : 'high';
  const escapedCode = pgEscapeLiteral(code);
  const escapedEffort = pgEscapeLiteral(defaultEffort);
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_tier_binding WHERE tier_id IN (SELECT id FROM plugin_linapro_ai_tier WHERE capability_type = 'text' AND code = '${escapedCode}') AND priority = 0;`,
    `UPDATE plugin_linapro_ai_tier SET enabled = 1, default_effort = '${escapedEffort}' WHERE capability_type = 'text' AND code = '${escapedCode}';`,
  ]);
}

export function insertInvocationLog(input: { purpose: string; requestId: string }): AiInvocationFixture {
  const purpose = input.purpose.trim();
  const requestId = input.requestId.trim();
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_invocation WHERE request_id = '${pgEscapeLiteral(requestId)}';`,
    `INSERT INTO plugin_linapro_ai_invocation (
      request_id,
      capability_type,
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
      error_code,
      error_summary,
      created_at
    ) VALUES (
      '${pgEscapeLiteral(requestId)}',
      'text',
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
      'failed',
      11,
      7,
      123,
      'AI_CORE_PROVIDER_HTTP_ERROR',
      'Provider returned a redacted error summary',
      NOW()
    );`,
  ]);
  return { purpose, requestId };
}

export function deleteInvocationLog(requestId: string) {
  execPgSQLStatements([
    `DELETE FROM plugin_linapro_ai_invocation WHERE request_id = '${pgEscapeLiteral(requestId)}';`,
  ]);
}

export async function deleteProvider(api: APIRequestContext, providerId: number) {
  const response = await api.delete(pluginApiPath(pluginId, `ai/providers/${providerId}`));
  if (!response.ok() && response.status() !== 404) {
    throw new Error(`删除 AI 供应商失败, status=${response.status()}, body=${await response.text()}`);
  }
}
