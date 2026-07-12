import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-org-core';

function orgApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface Post {
  id: number;
  deptId: number;
  code: string;
  name: string;
  sort: number;
  status: number;
  remark: string;
  createdAt: number | null;
}

export interface PostListParams {
  pageNum?: number;
  pageSize?: number;
  deptId?: number;
  code?: string;
  name?: string;
  status?: number;
}

export interface PostOption {
  postId: number;
  postName: string;
}

export interface DeptTreeNode {
  id: number;
  label: string;
  children?: DeptTreeNode[];
}

export async function postList(params?: PostListParams) {
  const res = await requestClient.get<{ list: Post[]; total: number }>(
    orgApi('post'),
    { params },
  );
  return { items: res.list, total: res.total };
}

export function postAdd(data: Partial<Post>) {
  return requestClient.post(orgApi('post'), data);
}

export function postUpdate(id: number, data: Partial<Post>) {
  return requestClient.put(orgApi(`post/${id}`), data);
}

export function postDelete(ids: number[] | string) {
  const list =
    typeof ids === 'string'
      ? ids
          .split(',')
          .map((part) => Number(part.trim()))
          .filter((id) => Number.isFinite(id) && id > 0)
      : ids;
  return requestClient.delete(orgApi('post'), {
    params: { ids: list },
  });
}

export function postInfo(id: number) {
  return requestClient.get<Post>(orgApi(`post/${id}`));
}

export function postExport(params?: PostListParams) {
  return requestClient.download<Blob>(orgApi('post/export'), { params });
}

export async function postDeptTree() {
  const res = await requestClient.get<{ list: DeptTreeNode[] }>(
    orgApi('post/dept-tree'),
  );
  return res.list;
}

export async function postOptionSelect(deptId?: number) {
  const res = await requestClient.get<{ list: PostOption[] }>(
    orgApi('post/option-select'),
    {
      params: deptId ? { deptId } : {},
    },
  );
  return res.list;
}
