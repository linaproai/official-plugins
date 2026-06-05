import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-monitor-loginlog';

function loginLogApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface LoginLog {
  id: number;
  userName: string;
  status: number;
  ip: string;
  browser: string;
  os: string;
  msg: string;
  loginTime: number | null;
}

export interface LoginLogListParams {
  pageNum?: number;
  pageSize?: number;
  userName?: string;
  ip?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface LoginLogListResult {
  items: LoginLog[];
  total: number;
}

export interface LoginLogCleanResult {
  deleted: number;
}

export async function loginLogList(params?: LoginLogListParams) {
  return await requestClient.get<LoginLogListResult>(loginLogApi('loginlog'), {
    params,
  });
}

export function loginLogDetail(id: number) {
  return requestClient.get<LoginLog>(loginLogApi(`loginlog/${id}`));
}

export function loginLogClean(params?: { beginTime?: string; endTime?: string }) {
  return requestClient.delete<LoginLogCleanResult>(
    loginLogApi('loginlog/clean'),
    { params },
  );
}

export function loginLogDelete(ids: number[]) {
  return requestClient.delete(loginLogApi(`loginlog/${ids.join(',')}`));
}

export function loginLogExport(params?: LoginLogListParams) {
  return requestClient.download<Blob>(loginLogApi('loginlog/export'), {
    params,
  });
}
