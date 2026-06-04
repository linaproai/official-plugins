<script lang="ts">
export const pluginPageMeta = {
  routePath: "/ai/providers",
  title: "Providers",
};
</script>

<script setup lang="ts">
import type { Provider, ProviderModelSummary } from "./ai-client";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { message, Popconfirm, Space } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import {
  modelDelete,
  modelSync,
  providerDelete,
  providerList,
} from "./ai-client";
import { buildProviderColumns, buildProviderQuerySchema } from "./ai-data";
import ProviderDrawer from "./provider-drawer.vue";
import ModelDrawer from "./model-drawer.vue";

const [ProviderDrawerRef, providerDrawerApi] = useVbenDrawer({
  connectedComponent: ProviderDrawer,
});

const [ModelDrawerRef, modelDrawerApi] = useVbenDrawer({
  connectedComponent: ModelDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildProviderQuerySchema(),
    commonConfig: {
      labelWidth: 96,
      componentProps: { allowClear: true },
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: buildProviderColumns({
      onDeleteModel: handleDeleteModel,
      providerIcon: IconifyIcon,
    }),
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) =>
          await providerList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          }),
      },
    },
    rowConfig: { keyField: "id" },
    id: "linapro-ai-core-provider-index",
  },
});

function handleAddProvider() {
  providerDrawerApi.setData({});
  providerDrawerApi.open();
}

function handleAddModel() {
  modelDrawerApi.setData({});
  modelDrawerApi.open();
}

function handleEdit(row: Provider) {
  providerDrawerApi.setData({ id: row.id });
  providerDrawerApi.open();
}

async function handleDelete(row: Provider) {
  await providerDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await gridApi.query();
}

async function handleDeleteModel(model: ProviderModelSummary) {
  await modelDelete(model.id);
  message.success($t("pages.common.deleteSuccess"));
  await gridApi.query();
}

async function handleSync(row: Provider, protocol: string) {
  const result = await modelSync(row.id, protocol);
  message.success(
    $t("plugin.linapro-ai-core.model.messages.syncDone", {
      created: result.created,
      kept: result.kept,
    }),
  );
  await gridApi.query();
}

function syncableProtocols(row: Provider) {
  const enabledProtocols = new Set(
    (row.endpoints || [])
      .filter((item) => item.enabled === 1)
      .map((item) => item.protocol),
  );
  return [
    enabledProtocols.has("openai") ? "openai" : "",
    enabledProtocols.has("anthropic") ? "anthropic" : "",
  ].filter(Boolean);
}

function syncActionLabel(protocol: string) {
  return protocol === "anthropic"
    ? $t("plugin.linapro-ai-core.model.actions.syncAnthropic")
    : $t("plugin.linapro-ai-core.model.actions.syncOpenAI");
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-ai-core.provider.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button type="primary" @click="handleAddProvider">
            <template #icon>
              <IconifyIcon icon="lucide:plus" />
            </template>
            {{ $t("plugin.linapro-ai-core.provider.actions.addProvider") }}
          </a-button>
          <a-button @click="handleAddModel">
            <template #icon>
              <IconifyIcon icon="lucide:box" />
            </template>
            {{ $t("plugin.linapro-ai-core.model.actions.addModel") }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <div class="ai-provider-action-list">
          <div class="ai-provider-action-primary">
            <ghost-button @click.stop="handleEdit(row)">
              {{ $t("pages.common.edit") }}
            </ghost-button>
            <Popconfirm
              :title="$t('pages.common.deleteConfirm')"
              placement="left"
              @confirm="handleDelete(row)"
            >
              <ghost-button danger @click.stop="">
                {{ $t("pages.common.delete") }}
              </ghost-button>
            </Popconfirm>
          </div>
          <div class="ai-provider-action-sync">
            <ghost-button
              v-for="protocol in syncableProtocols(row)"
              :key="protocol"
              @click.stop="handleSync(row, protocol)"
            >
              {{ syncActionLabel(protocol) }}
            </ghost-button>
          </div>
        </div>
      </template>
    </Grid>

    <ProviderDrawerRef @reload="() => gridApi.query()" />
    <ModelDrawerRef @reload="() => gridApi.query()" />
  </Page>
</template>

<style scoped>
.ai-provider-action-list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  max-width: 100%;
  padding: 2px 0;
}

.ai-provider-action-primary,
.ai-provider-action-sync {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  max-width: 100%;
}

.ai-provider-action-sync {
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

:deep(.ai-provider-action-list .ant-btn) {
  min-height: 24px;
  padding-inline: 4px;
  white-space: nowrap;
}

:deep(.ai-provider-action-column .vxe-cell) {
  max-height: none !important;
  overflow: visible !important;
  line-height: 1.4;
}

:deep(.ai-provider-model-column .vxe-cell),
:deep(.ai-provider-endpoint-column .vxe-cell) {
  max-height: none !important;
  overflow: visible !important;
  line-height: 1.4;
}

:deep(.ai-model-delete-icon) {
  display: block;
  width: 0.875rem;
  height: 0.875rem;
  background-color: currentcolor;
  -webkit-mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='black' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M18 6 6 18'/%3E%3Cpath d='m6 6 12 12'/%3E%3C/svg%3E")
    center / contain no-repeat;
  mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='24' height='24' viewBox='0 0 24 24' fill='none' stroke='black' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M18 6 6 18'/%3E%3Cpath d='m6 6 12 12'/%3E%3C/svg%3E")
    center / contain no-repeat;
}
</style>
