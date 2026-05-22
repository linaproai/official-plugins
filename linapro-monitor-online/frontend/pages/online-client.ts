import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-monitor-online';

function onlineApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface OnlineUser {
  tokenId: string;
  username: string;
  deptName: string;
  ip: string;
  browser: string;
  os: string;
  loginTime: number | null;
}

export interface OnlineListResult {
  items: OnlineUser[];
  total: number;
}

export interface OnlineListParams {
  pageNum?: number;
  pageSize?: number;
  username?: string;
  ip?: string;
}

export function onlineList(params?: OnlineListParams) {
  return requestClient.get<OnlineListResult>(
    onlineApi('monitor/online/list'),
    { params },
  );
}

export function forceLogout(tokenId: string) {
  return requestClient.delete(onlineApi(`monitor/online/${tokenId}`));
}
