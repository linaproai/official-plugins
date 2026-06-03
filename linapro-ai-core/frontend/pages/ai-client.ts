import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-ai-core';
const tierTestRequestTimeout = 65_000;
export const defaultCapabilityType = 'text';
export const defaultCapabilityMethod = 'generate';

function aiApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface Provider {
  id: number;
  name: string;
  websiteUrl: string;
  remark: string;
  enabled: number;
  modelCount: number;
  enabledModelCount: number;
  endpointCount: number;
  enabledEndpointCount: number;
  models: ProviderModelSummary[];
  endpoints: ProviderEndpoint[];
  createdAt: number;
  updatedAt: number;
}

export interface ProviderEndpoint {
  id: number;
  providerId: number;
  protocol: string;
  baseUrl: string;
  secretRef: string;
  enabled: number;
  metadataJson: string;
  createdAt: number;
  updatedAt: number;
}

export interface ProviderModelSummary {
  id: number;
  capabilityType: string;
  capabilityMethod: string;
  modelName: string;
  protocol: string;
  enabled: number;
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
  endpointId: number;
  capabilityType: string;
  capabilityMethod: string;
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
  capabilityMethod: string;
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
  capabilityMethod: string;
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
  assetSummaryJson: string;
  operationSummaryJson: string;
  metadataSummaryJson: string;
  errorCode: string;
  errorSummary: string;
  createdAt: number;
}

export interface ProviderOperation {
  id: number;
  operationRef: string;
  capabilityType: string;
  capabilityMethod: string;
  purpose: string;
  sourcePluginId: string;
  providerId: number;
  modelId: number;
  providerName: string;
  modelName: string;
  protocol: string;
  status: string;
  nextPollAfterMs: number;
  expiresAt: number;
  assetSummaryJson: string;
  errorCode: string;
  errorSummary: string;
  createdAt: number;
  updatedAt: number;
}

export interface InvocationListParams {
  pageNum?: number;
  pageSize?: number;
  capabilityType?: string;
  capabilityMethod?: string;
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

export async function providerEndpoints(
  providerId: number,
  params?: { enabled?: number; protocol?: string },
) {
  const res = await requestClient.get<{ list: ProviderEndpoint[] }>(
    aiApi(`ai/providers/${providerId}/endpoints`),
    { params },
  );
  return res.list;
}

export function providerEndpointAdd(
  providerId: number,
  data: Partial<ProviderEndpoint>,
) {
  return requestClient.post(aiApi(`ai/providers/${providerId}/endpoints`), data);
}

export function providerEndpointUpdate(
  providerId: number,
  id: number,
  data: Partial<ProviderEndpoint>,
) {
  return requestClient.put(aiApi(`ai/providers/${providerId}/endpoints/${id}`), data);
}

export function providerEndpointDelete(providerId: number, id: number) {
  return requestClient.delete(aiApi(`ai/providers/${providerId}/endpoints/${id}`));
}

export async function providerModels(
  providerId: number,
  enabled?: number,
  capabilityType = defaultCapabilityType,
  capabilityMethod = defaultCapabilityMethod,
) {
  const res = await requestClient.get<{ list: Model[]; total: number }>(
    aiApi(`ai/providers/${providerId}/models`),
    {
      params: {
        capabilityMethod,
        capabilityType,
        enabled,
        pageNum: 1,
        pageSize: 100,
      },
    },
  );
  return res.list;
}

export function modelAdd(providerId: number, data: Partial<Model>) {
  return requestClient.post(aiApi(`ai/providers/${providerId}/models`), {
    capabilityMethod: defaultCapabilityMethod,
    capabilityType: defaultCapabilityType,
    ...data,
  });
}

export function modelUpdate(id: number, data: Partial<Model>) {
  return requestClient.put(aiApi(`ai/models/${id}`), data);
}

export function modelDelete(id: number) {
  return requestClient.delete(aiApi(`ai/models/${id}`));
}

export interface ModelCapabilityInput {
  endpointId?: number;
  capabilityType: string;
  capabilityMethod: string;
  inputModalities?: string[];
  outputModalities?: string[];
  maxInputTokens?: number;
  maxOutputTokens?: number;
  maxInputAssets?: number;
  maxOutputAssets?: number;
  maxAssetBytes?: number;
  supportsStreaming?: number;
  supportsOperation?: number;
  supportsThinking?: number;
  supportedEfforts?: string[];
  defaultParamsJson?: string;
  enabled?: number;
}

export interface ModelCapability extends ModelCapabilityInput {
  id: number;
  modelId: number;
  endpointId: number;
  maxInputTokens: number;
  maxOutputTokens: number;
  maxInputAssets: number;
  maxOutputAssets: number;
  maxAssetBytes: number;
  supportsStreaming: number;
  supportsOperation: number;
  supportsThinking: number;
  supportedEfforts: string[];
  defaultParamsJson: string;
  enabled: number;
  createdAt: number;
  updatedAt: number;
}

export async function modelCapabilities(id: number) {
  const res = await requestClient.get<{ list: ModelCapability[] }>(
    aiApi(`ai/models/${id}/capabilities`),
  );
  return res.list;
}

export function modelCapabilitiesSave(id: number, items: ModelCapabilityInput[]) {
  return requestClient.put(aiApi(`ai/models/${id}/capabilities`), { items });
}

export function modelSync(providerId: number, protocol: string) {
  return requestClient.post<{ created: number; kept: number }>(
    aiApi(`ai/providers/${providerId}/models/sync`),
    { protocol },
  );
}

export interface MethodDefaultParam {
  id: number;
  capabilityType: string;
  capabilityMethod: string;
  defaultParamsJson: string;
  enabled: number;
  createdAt: number;
  updatedAt: number;
}

export async function methodDefaults() {
  const res = await requestClient.get<{ list: MethodDefaultParam[] }>(
    aiApi('ai/method-defaults'),
  );
  return res.list;
}

export function methodDefaultUpdate(
  capabilityType: string,
  capabilityMethod: string,
  data: Partial<MethodDefaultParam>,
) {
  return requestClient.put(
    aiApi(`ai/method-defaults/${capabilityType}/${capabilityMethod}`),
    data,
  );
}

export async function tierList(
  capabilityType = defaultCapabilityType,
  capabilityMethod = defaultCapabilityMethod,
) {
  const res = await requestClient.get<{ list: Tier[] }>(aiApi('ai/tiers'), {
    params: {
      capabilityMethod,
      capabilityType,
    },
  });
  return res.list;
}

export function tierUpdate(code: string, data: Partial<Tier>) {
  return requestClient.put(aiApi(`ai/tiers/${code}`), {
    capabilityMethod: data.capabilityMethod || defaultCapabilityMethod,
    capabilityType: data.capabilityType || defaultCapabilityType,
    ...data,
  });
}

export function tierTest(code: string, data: Record<string, any>) {
  return requestClient.post<TierTestResult>(
    aiApi(`ai/tiers/${code}/test`),
    {
      capabilityMethod: data.capabilityMethod || defaultCapabilityMethod,
      capabilityType: data.capabilityType || defaultCapabilityType,
      ...data,
    },
    { timeout: tierTestRequestTimeout },
  );
}

export async function invocationList(params?: InvocationListParams) {
  const res = await requestClient.get<{ list: Invocation[]; total: number }>(
    aiApi('ai/invocations'),
    { params },
  );
  return { items: res.list, total: res.total };
}

export async function providerOperationList(params?: InvocationListParams) {
  const res = await requestClient.get<{ list: ProviderOperation[]; total: number }>(
    aiApi('ai/provider-operations'),
    { params },
  );
  return { items: res.list, total: res.total };
}
