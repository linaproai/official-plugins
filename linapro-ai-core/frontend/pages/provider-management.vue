<script lang="ts">
export const pluginPageMeta = {
  routePath: '/ai/providers',
  title: 'Provider Management',
};
</script>

<script setup lang="ts">
import type { Provider } from './ai-client';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { providerDelete, providerList } from './ai-client';
import { buildProviderColumns, buildProviderQuerySchema } from './ai-data';
import ProviderDrawer from './provider-drawer.vue';

const [ProviderDrawerRef, providerDrawerApi] = useVbenDrawer({
  connectedComponent: ProviderDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildProviderQuerySchema(),
    commonConfig: {
      labelWidth: 96,
      componentProps: { allowClear: true },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: buildProviderColumns(),
    height: 'auto',
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
    rowConfig: { keyField: 'id' },
    id: 'linapro-ai-core-provider-index',
  },
});

function handleAdd() {
  providerDrawerApi.setData({});
  providerDrawerApi.open();
}

function handleEdit(row: Provider) {
  providerDrawerApi.setData({ id: row.id });
  providerDrawerApi.open();
}

async function handleDelete(row: Provider) {
  await providerDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  await gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-ai-core.provider.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button type="primary" @click="handleAdd">
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <Popconfirm
            :title="$t('pages.common.deleteConfirm')"
            placement="left"
            @confirm="handleDelete(row)"
          >
            <ghost-button danger @click.stop="">
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <ProviderDrawerRef @reload="() => gridApi.query()" />
  </Page>
</template>
