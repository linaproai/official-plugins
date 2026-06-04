<script lang="ts">
export const pluginPageMeta = {
  routePath: "/ai/providers",
  title: "Providers",
};
</script>

<script setup lang="ts">
import type { Model, Provider, ProviderModelSummary } from "./ai-client";

import { ref } from "vue";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { message, Popconfirm, Space, Tabs } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import {
  modelDelete,
  modelList,
  modelSync,
  providerDelete,
  providerList,
} from "./ai-client";
import {
  buildModelColumns,
  buildModelQuerySchema,
  buildProviderColumns,
  buildProviderQuerySchema,
} from "./ai-data";
import ProviderDrawer from "./provider-drawer.vue";
import ModelDrawer from "./model-drawer.vue";

const [ProviderDrawerRef, providerDrawerApi] = useVbenDrawer({
  connectedComponent: ProviderDrawer,
});

const [ModelDrawerRef, modelDrawerApi] = useVbenDrawer({
  connectedComponent: ModelDrawer,
});

const TabPane = Tabs.TabPane;
const activeTab = ref("providers");
const providerTabIcons: Record<string, string> = {
  models: "lucide:box",
  providers: "lucide:building-2",
};

function providerTabLabel(tabKey: string, labelKey: string) {
  return {
    icon: providerTabIcons[tabKey] || "lucide:square",
    label: $t(labelKey),
  };
}

const [ProviderGrid, providerGridApi] = useVbenVxeGrid({
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
    height: "100%",
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

const [ModelGrid, modelGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildModelQuerySchema(),
    commonConfig: {
      labelWidth: 96,
      componentProps: { allowClear: true },
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
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
            ...formValues,
          }),
      },
    },
    rowConfig: { keyField: "id" },
    id: "linapro-ai-core-model-index",
  },
});

function handleAddProvider() {
  providerDrawerApi.setData({});
  providerDrawerApi.open();
}

function handleAddModel(row?: Provider) {
  modelDrawerApi.setData({ providerId: row?.id });
  modelDrawerApi.open();
}

function handleEdit(row: Provider) {
  providerDrawerApi.setData({ id: row.id });
  providerDrawerApi.open();
}

function handleEditModel(row: Model) {
  modelDrawerApi.setData({ model: row });
  modelDrawerApi.open();
}

async function handleDelete(row: Provider) {
  await providerDelete(row.id);
  message.success($t("pages.common.deleteSuccess"));
  await reloadGrids();
}

async function handleDeleteModel(model: Model | ProviderModelSummary) {
  await modelDelete(model.id);
  message.success($t("pages.common.deleteSuccess"));
  await reloadGrids();
}

async function handleSync(row: Provider) {
  const result = await modelSync(row.id);
  message.success(
    $t("plugin.linapro-ai-core.model.messages.syncDone", {
      created: result.created,
      kept: result.kept,
    }),
  );
  await reloadGrids();
}

async function reloadGrids() {
  await Promise.all([providerGridApi.query(), modelGridApi.query()]);
}
</script>

<template>
  <Page :auto-content-height="true" content-class="min-h-0 overflow-hidden">
    <Tabs
      v-model:active-key="activeTab"
      :tab-bar-gutter="28"
      class="ai-provider-tabs"
      data-testid="ai-provider-management-tabs"
    >
      <TabPane key="providers">
        <template #tab>
          <span class="ai-provider-tab-label">
            <IconifyIcon
              :icon="
                providerTabLabel(
                  'providers',
                  'plugin.linapro-ai-core.provider.tabs.providers',
                ).icon
              "
              class="ai-provider-tab-icon"
              data-testid="ai-provider-tab-icon-providers"
            />
            <span>
              {{
                providerTabLabel(
                  "providers",
                  "plugin.linapro-ai-core.provider.tabs.providers",
                ).label
              }}
            </span>
          </span>
        </template>

        <div
          v-if="activeTab === 'providers'"
          class="ai-provider-tab-content"
          data-testid="ai-provider-tab-content-providers"
        >
          <ProviderGrid
            :table-title="$t('plugin.linapro-ai-core.provider.tableTitle')"
          >
            <template #toolbar-tools>
              <Space>
                <a-button type="primary" @click="handleAddProvider">
                  <template #icon>
                    <IconifyIcon icon="lucide:plus" />
                  </template>
                  {{
                    $t("plugin.linapro-ai-core.provider.actions.addProvider")
                  }}
                </a-button>
                <a-button @click="handleAddModel()">
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
                <div class="ai-provider-action-row">
                  <ghost-button @click.stop="handleAddModel(row)">
                    {{ $t("plugin.linapro-ai-core.model.actions.addModel") }}
                  </ghost-button>
                </div>
                <div class="ai-provider-action-row">
                  <ghost-button @click.stop="handleSync(row)">
                    {{ $t("plugin.linapro-ai-core.model.actions.syncModels") }}
                  </ghost-button>
                </div>
              </div>
            </template>
          </ProviderGrid>
        </div>
      </TabPane>

      <TabPane key="models">
        <template #tab>
          <span class="ai-provider-tab-label">
            <IconifyIcon
              :icon="
                providerTabLabel(
                  'models',
                  'plugin.linapro-ai-core.provider.tabs.models',
                ).icon
              "
              class="ai-provider-tab-icon"
              data-testid="ai-provider-tab-icon-models"
            />
            <span>
              {{
                providerTabLabel(
                  "models",
                  "plugin.linapro-ai-core.provider.tabs.models",
                ).label
              }}
            </span>
          </span>
        </template>

        <div
          v-if="activeTab === 'models'"
          class="ai-provider-tab-content"
          data-testid="ai-provider-tab-content-models"
        >
          <ModelGrid
            :table-title="$t('plugin.linapro-ai-core.model.tableTitle')"
          >
            <template #toolbar-tools>
              <a-button type="primary" @click="handleAddModel()">
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
      </TabPane>
    </Tabs>

    <ProviderDrawerRef @reload="reloadGrids" />
    <ModelDrawerRef @reload="reloadGrids" />
  </Page>
</template>

<style scoped>
.ai-provider-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
  margin-bottom: 0;
  background: hsl(var(--background));
  min-height: 0;
}

.ai-provider-tabs :deep(.ant-tabs-nav) {
  flex: 0 0 auto;
  margin-bottom: 0;
}

.ai-provider-tabs :deep(.ant-tabs-nav-wrap) {
  padding: 0 20px;
}

.ai-provider-tabs :deep(.ant-tabs-nav::before) {
  border-bottom-color: hsl(var(--border));
}

.ai-provider-tabs :deep(.ant-tabs-content-holder) {
  flex: 1 1 auto;
  background: hsl(var(--background));
  border: 0;
  min-height: 0;
  overflow: hidden;
  padding: 16px 20px 0;
}

.ai-provider-tabs :deep(.ant-tabs-content),
.ai-provider-tabs :deep(.ant-tabs-tabpane) {
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.ai-provider-tabs :deep(.ant-tabs-tab) {
  margin: 0;
  border: 0 !important;
  border-radius: 0 !important;
  background: transparent !important;
  color: hsl(var(--muted-foreground));
  padding: 14px 0 12px;
  transition:
    color 0.16s ease,
    opacity 0.16s ease;
}

.ai-provider-tabs :deep(.ant-tabs-tab-active) {
  color: hsl(var(--primary)) !important;
}

.ai-provider-tabs :deep(.ant-tabs-ink-bar) {
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: hsl(var(--primary));
}

.ai-provider-tabs :deep(.ant-tabs-tab-btn) {
  color: inherit;
  text-align: left;
}

.ai-provider-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  line-height: 1;
}

.ai-provider-tab-icon {
  color: inherit;
  font-size: 16px;
}

.ai-provider-tab-content {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  background: hsl(var(--background));
}

.ai-provider-action-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  width: 100%;
  max-width: 100%;
  padding: 2px 0;
}

.ai-provider-action-primary,
.ai-provider-action-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
  max-width: 100%;
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
:deep(.ai-provider-endpoint-column .vxe-cell),
:deep(.ai-model-endpoint-column .vxe-cell) {
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
