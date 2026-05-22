import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-org-core';

function orgApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface Dept {
  id: number;
  parentId: number;
  ancestors: string;
  name: string;
  code: string;
  orderNum: number;
  leader: number;
  phone: string;
  email: string;
  status: number;
  remark: string;
  createdAt: number | null;
}

export interface DeptTree {
  id: number;
  label: string;
  userCount?: number;
  children?: DeptTree[];
}

export interface DeptUser {
  id: number;
  username: string;
  nickname: string;
}

export async function deptList(params?: Record<string, any>) {
  const res = await requestClient.get<{ list: Dept[] }>(orgApi('dept'), {
    params,
  });
  return res.list;
}

export function deptAdd(data: Partial<Dept>) {
  return requestClient.post(orgApi('dept'), data);
}

export function deptUpdate(id: number, data: Partial<Dept>) {
  return requestClient.put(orgApi(`dept/${id}`), data);
}

export function deptDelete(id: number) {
  return requestClient.delete(orgApi(`dept/${id}`));
}

export function deptInfo(id: number) {
  return requestClient.get<Dept>(orgApi(`dept/${id}`));
}

export async function deptTree() {
  const res = await requestClient.get<{ list: DeptTree[] }>(orgApi('dept/tree'));
  return res.list;
}

export async function deptExclude(id: number) {
  const res = await requestClient.get<{ list: DeptTree[] }>(
    orgApi(`dept/exclude/${id}`),
  );
  return res.list;
}

export async function deptUsers(id: number, params?: { keyword?: string; limit?: number }) {
  const res = await requestClient.get<{ list: DeptUser[] }>(
    orgApi(`dept/${id}/users`),
    { params },
  );
  return res.list;
}
