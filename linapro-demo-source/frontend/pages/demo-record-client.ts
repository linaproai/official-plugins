import { pluginApiPath, requestClient } from "#/api/request";

const pluginID = "linapro-demo-source";

function demoSourceApi(pathName: string) {
  return pluginApiPath(pluginID, pathName);
}

export interface DemoRecordItem {
  id: number;
  title: string;
  content: string;
  attachmentName: string;
  hasAttachment: number;
  createdAt: number | null;
  updatedAt: number | null;
}

export interface DemoRecordDetail {
  id: number;
  title: string;
  content: string;
  attachmentName: string;
  hasAttachment: number;
}

export interface DemoRecordListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
}

export interface DemoSummary {
  message: string;
}

export interface DemoRecordSaveInput {
  title: string;
  content: string;
  removeAttachment?: boolean;
}

export async function getDemoSummary() {
  return requestClient.get<DemoSummary>(
    demoSourceApi("plugins/linapro-demo-source/summary"),
  );
}

export async function listDemoRecords(params?: DemoRecordListParams) {
  const res = await requestClient.get<{
    list: DemoRecordItem[];
    total: number;
  }>(demoSourceApi("plugins/linapro-demo-source/records"), { params });
  return {
    items: res.list,
    total: res.total,
  };
}

export async function getDemoRecord(id: number) {
  return requestClient.get<DemoRecordDetail>(
    demoSourceApi(`plugins/linapro-demo-source/records/${id}`),
  );
}

export async function createDemoRecord(
  values: DemoRecordSaveInput,
  file?: File | null,
) {
  return requestClient.post<{ id: number }>(
    demoSourceApi("plugins/linapro-demo-source/records"),
    buildRecordFormData(values, file),
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );
}

export async function updateDemoRecord(
  id: number,
  values: DemoRecordSaveInput,
  file?: File | null,
) {
  return requestClient.put<{ id: number }>(
    demoSourceApi(`plugins/linapro-demo-source/records/${id}`),
    buildRecordFormData(values, file),
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );
}

export async function deleteDemoRecord(id: number) {
  return requestClient.delete(
    demoSourceApi(`plugins/linapro-demo-source/records/${id}`),
  );
}

export async function downloadDemoRecordAttachment(id: number) {
  return requestClient.download<Blob>(
    demoSourceApi(`plugins/linapro-demo-source/records/${id}/attachment`),
  );
}

function buildRecordFormData(values: DemoRecordSaveInput, file?: File | null) {
  const formData = new FormData();
  formData.append("title", values.title);
  formData.append("content", values.content ?? "");
  if (values.removeAttachment) {
    formData.append("removeAttachment", "1");
  }
  if (file) {
    formData.append("file", file, file.name);
  }
  return formData;
}
