import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-monitor-operlog';

function operLogApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface OperLog {
  id: number;
  title: string;
  operSummary: string;
  operType: string;
  method: string;
  requestMethod: string;
  operName: string;
  operUrl: string;
  operIp: string;
  operParam: string;
  jsonResult: string;
  status: number;
  errorMsg: string;
  costTime: number;
  operTime: number | null;
}

export interface OperLogListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  operName?: string;
  operType?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface OperLogListResult {
  items: OperLog[];
  total: number;
}

export async function operLogList(params?: OperLogListParams) {
  return await requestClient.get<OperLogListResult>(operLogApi('operlog'), {
    params,
  });
}

export function operLogDetail(id: number) {
  return requestClient.get<OperLog>(operLogApi(`operlog/${id}`));
}

export function operLogClean(params?: { beginTime?: string; endTime?: string }) {
  return requestClient.delete(operLogApi('operlog/clean'), { params });
}

export function operLogDelete(ids: number[]) {
  return requestClient.delete(operLogApi(`operlog/${ids.join(',')}`));
}

export function operLogExport(params?: OperLogListParams) {
  return requestClient.download<Blob>(operLogApi('operlog/export'), {
    params,
  });
}
