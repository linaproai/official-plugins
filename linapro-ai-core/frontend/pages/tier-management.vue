<script lang="ts">
export const pluginPageMeta = {
  routePath: '/ai/tiers',
  title: 'Tier Management',
};
</script>

<script setup lang="ts">
import type { Tier } from './ai-client';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { tierList, tierTest } from './ai-client';
import { buildTierColumns } from './ai-data';
import TierDrawer from './tier-drawer.vue';

const [TierDrawerRef, tierDrawerApi] = useVbenDrawer({
  connectedComponent: TierDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: buildTierColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: { enabled: false },
    proxyConfig: {
      ajax: {
        query: async () => ({ items: await tierList(), total: 3 }),
      },
    },
    rowConfig: { keyField: 'code' },
    id: 'linapro-ai-core-tier-index',
  },
});

function handleEdit(row: Tier) {
  tierDrawerApi.setData({ tier: row });
  tierDrawerApi.open();
}

async function handleTest(row: Tier) {
  const result = await tierTest(row.code, { maxOutputTokens: 128 });
  if (result.status === 'success') {
    message.success($t('plugin.linapro-ai-core.tier.messages.testSuccess'));
  } else {
    message.error(result.errorSummary || $t('plugin.linapro-ai-core.tier.messages.testFailed'));
  }
  await gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-ai-core.tier.tableTitle')">
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <ghost-button @click.stop="handleTest(row)">
            {{ $t('plugin.linapro-ai-core.tier.actions.testSaved') }}
          </ghost-button>
        </Space>
      </template>
    </Grid>

    <TierDrawerRef @reload="() => gridApi.query()" />
  </Page>
</template>
