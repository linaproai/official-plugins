<script lang="ts">
export const pluginPageMeta = {
  routePath: "/ai/models",
  title: "Models",
};
</script>

<script setup lang="ts">
import type { Model } from "./ai-client";

import { onMounted } from "vue";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { message, Popconfirm, Space } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { modelDelete, modelList, providerList } from "./ai-client";
import {
  buildModelColumns,
  buildModelQuerySchema,
} from "./ai-data";
import ModelDrawer from "./model-drawer.vue";

const [ModelDrawerRef, modelDrawerApi] = useVbenDrawer({
  connectedComponent: ModelDrawer,
});

onMounted(async () => {
  const out = await providerList({ pageNum: 1, pageSize: 100 });
  modelGridApi.formApi.updateSchema([
    {
      fieldName: "providerId",
      componentProps: {
        options: out.items.map((item) => ({
          label: item.name,
          value: item.id,
        })),
      },
    },
  ]);
});

const [ModelGrid, modelGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildModelQuerySchema(),
    commonConfig: {
      labelWidth: 80,
      componentProps: { allowClear: true },
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5",
  },
  gridOptions: {
    columns: buildModelColumns(),
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) =>
          await modelList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...normalizeModelFilters(formValues),
          }),
      },
    },
    rowConfig: { keyField: "id" },
    id: "linapro-ai-core-model-index",
  },
});

function hasFilterValue(value: unknown) {
  return value !== undefined && value !== null && value !== "";
}

function normalizeModelFilters(formValues: Record<string, any>) {
  const filters: Record<string, any> = {};

  for (const [key, value] of Object.entries(formValues)) {
    if (hasFilterValue(value)) {
      filters[key] = value;
    }
  }

  return filters;
}

function handleAddModel() {
  modelDrawerApi.setData({});
  modelDrawerApi.open();
}

function handleEditModel(row: Model) {
  modelDrawerApi.setData({ model: row });
  modelDrawerApi.open();
}

async function handleDeleteModel(row: Model) {
  await modelDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await reloadModels();
}

async function reloadModels() {
  await modelGridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true" content-class="min-h-0 overflow-hidden">
    <div class="ai-model-page" data-testid="ai-model-management-page">
      <ModelGrid :table-title="$t('plugin.linapro-ai-core.model.tableTitle')">
        <template #toolbar-tools>
          <a-button type="primary" @click="handleAddModel">
            <template #icon>
              <IconifyIcon icon="lucide:plus" />
            </template>
            {{ $t("plugin.linapro-ai-core.model.actions.addModel") }}
          </a-button>
        </template>

        <template #modelAction="{ row }">
          <Space>
            <ghost-button @click.stop="handleEditModel(row)">
              {{ $t("pages.common.edit") }}
            </ghost-button>
            <Popconfirm
              :title="$t('pages.common.deleteConfirm')"
              placement="left"
              @confirm="handleDeleteModel(row)"
            >
              <ghost-button danger @click.stop="">
                {{ $t("pages.common.delete") }}
              </ghost-button>
            </Popconfirm>
          </Space>
        </template>
      </ModelGrid>
    </div>

    <ModelDrawerRef @reload="reloadModels" />
  </Page>
</template>

<style scoped>
.ai-model-page {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: hsl(var(--background));
}

:deep(.ai-model-endpoint-column .vxe-cell) {
  max-height: none !important;
  overflow: visible !important;
  line-height: 1.4;
}
</style>
