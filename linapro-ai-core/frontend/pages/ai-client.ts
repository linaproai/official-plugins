import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-ai-core';

function aiApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface Provider {
  id: number;
  name: string;
  websiteUrl: string;
  remark: string;
  openaiBaseUrl: string;
  anthropicBaseUrl: string;
  apiKeySecretRef: string;
  enabled: number;
  modelCount: number;
  enabledModelCount: number;
  createdAt: number;
  updatedAt: number;
}

export interface ProviderListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  enabled?: number;
}

export interface Model {
  id: number;
  providerId: number;
  capabilityType: string;
  modelName: string;
  protocol: string;
  source: string;
  supportsThinking: number;
  supportedEfforts: string[];
  maxInputTokens: number;
  maxOutputTokens: number;
  enabled: number;
  createdAt: number;
  updatedAt: number;
}

export interface TierBinding {
  providerId: number;
  providerName: string;
  modelId: number;
  modelName: string;
  protocol: string;
  supportsThinking: number;
  supportedEfforts: string[];
  enabled: number;
}

export interface Tier {
  id: number;
  capabilityType: string;
  code: string;
  displayName: string;
  description: string;
  defaultEffort: string;
  enabled: number;
  sortOrder: number;
  binding?: TierBinding;
  lastTestStatus: string;
  lastTestLatencyMs: number;
  lastTestErrorSummary: string;
  lastTestAt: number;
  updatedAt: number;
}

export interface TierTestResult {
  status: string;
  latencyMs: number;
  providerName: string;
  modelName: string;
  protocol: string;
  thinkingEffort: string;
  errorSummary: string;
  testedAt: number;
}

export interface Invocation {
  id: number;
  requestId: string;
  capabilityType: string;
  purpose: string;
  tierCode: string;
  sourcePluginId: string;
  tenantId: number;
  userId: number;
  providerId: number;
  modelId: number;
  providerName: string;
  modelName: string;
  protocol: string;
  thinkingEffort: string;
  status: string;
  inputTokens: number;
  outputTokens: number;
  latencyMs: number;
  errorCode: string;
  errorSummary: string;
  createdAt: number;
}

export interface InvocationListParams {
  pageNum?: number;
  pageSize?: number;
  capabilityType?: string;
  purpose?: string;
  tierCode?: string;
  status?: string;
  providerId?: number;
  modelId?: number;
  sourcePluginId?: string;
  startedAt?: number;
  endedAt?: number;
}

export async function providerList(params?: ProviderListParams) {
  const res = await requestClient.get<{ list: Provider[]; total: number }>(
    aiApi('ai/providers'),
    { params },
  );
  return { items: res.list, total: res.total };
}

export function providerInfo(id: number) {
  return requestClient.get<Provider>(aiApi(`ai/providers/${id}`));
}

export function providerAdd(data: Partial<Provider>) {
  return requestClient.post(aiApi('ai/providers'), data);
}

export function providerUpdate(id: number, data: Partial<Provider>) {
  return requestClient.put(aiApi(`ai/providers/${id}`), data);
}

export function providerDelete(id: number) {
  return requestClient.delete(aiApi(`ai/providers/${id}`));
}

export async function providerModels(providerId: number, enabled?: number) {
  const res = await requestClient.get<{ list: Model[]; total: number }>(
    aiApi(`ai/providers/${providerId}/models`),
    { params: { capabilityType: 'text', enabled, pageNum: 1, pageSize: 100 } },
  );
  return res.list;
}

export function modelAdd(providerId: number, data: Partial<Model>) {
  return requestClient.post(aiApi(`ai/providers/${providerId}/models`), data);
}

export function modelUpdate(id: number, data: Partial<Model>) {
  return requestClient.put(aiApi(`ai/models/${id}`), data);
}

export function modelDelete(id: number) {
  return requestClient.delete(aiApi(`ai/models/${id}`));
}

export function modelSync(providerId: number, protocol: string) {
  return requestClient.post<{ created: number; kept: number }>(
    aiApi(`ai/providers/${providerId}/models/sync`),
    { protocol },
  );
}

export async function tierList() {
  const res = await requestClient.get<{ list: Tier[] }>(aiApi('ai/tiers'), {
    params: { capabilityType: 'text' },
  });
  return res.list;
}

export function tierUpdate(code: string, data: Partial<Tier>) {
  return requestClient.put(aiApi(`ai/tiers/${code}`), data);
}

export function tierTest(code: string, data: Record<string, any>) {
  return requestClient.post<TierTestResult>(aiApi(`ai/tiers/${code}/test`), data);
}

export async function invocationList(params?: InvocationListParams) {
  const res = await requestClient.get<{ list: Invocation[]; total: number }>(
    aiApi('ai/invocations'),
    { params },
  );
  return { items: res.list, total: res.total };
}
