<script lang="ts">
export const pluginPageMeta = {
  routePath: '/ai/invocations',
  title: 'Invocation Logs',
};
</script>

<script setup lang="ts">
import type { Invocation } from './ai-client';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { invocationList } from './ai-client';
import {
  buildInvocationColumns,
  buildInvocationQuerySchema,
  splitCapabilityMethod,
} from './ai-data';
import InvocationDetailDrawer from './invocation-detail-drawer.vue';

const [DetailDrawerRef, detailDrawerApi] = useVbenDrawer({
  connectedComponent: InvocationDetailDrawer,
});

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: buildInvocationQuerySchema(),
    commonConfig: {
      labelWidth: 112,
      componentProps: { allowClear: true },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  },
  gridOptions: {
    columns: buildInvocationColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          const capability = splitCapabilityMethod(formValues.capabilityKey || '');
          const { capabilityKey: _capabilityKey, ...filters } = formValues;
          return await invocationList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            capabilityMethod: capability.capabilityMethod,
            capabilityType: capability.capabilityType,
            ...filters,
          });
        },
      },
    },
    rowConfig: { keyField: 'id' },
    id: 'linapro-ai-core-invocation-index',
  },
});

function handleDetail(row: Invocation) {
  detailDrawerApi.setData({ record: row });
  detailDrawerApi.open();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-ai-core.invocation.tableTitle')">
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleDetail(row)">
            {{ $t('pages.common.detail') }}
          </ghost-button>
        </Space>
      </template>
    </Grid>

    <DetailDrawerRef />
  </Page>
</template>
