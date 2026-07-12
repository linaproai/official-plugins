import { pluginApiPath, requestClient } from '#/api/request';

const pluginID = 'linapro-content-notice';

function noticeApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface Notice {
  id: number;
  title: string;
  type: number;
  content: string;
  fileIds: string;
  status: number;
  remark: string;
  createdBy: number;
  createdByName: string;
  updatedBy: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface NoticeListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  type?: number;
  createdBy?: string;
}

export async function noticeList(params?: NoticeListParams) {
  const res = await requestClient.get<{ list: Notice[]; total: number }>(
    noticeApi('notice'),
    { params },
  );
  return { items: res.list, total: res.total };
}

export function noticeAdd(data: Partial<Notice>) {
  return requestClient.post(noticeApi('notice'), data);
}

export function noticeUpdate(id: number, data: Partial<Notice>) {
  return requestClient.put(noticeApi(`notice/${id}`), data);
}

export function noticeDelete(ids: number[] | string) {
  const list =
    typeof ids === 'string'
      ? ids
          .split(',')
          .map((part) => Number(part.trim()))
          .filter((id) => Number.isFinite(id) && id > 0)
      : ids;
  return requestClient.delete(noticeApi('notice'), {
    params: { ids: list },
  });
}

export function noticeInfo(id: number) {
  return requestClient.get<Notice>(noticeApi(`notice/${id}`));
}
