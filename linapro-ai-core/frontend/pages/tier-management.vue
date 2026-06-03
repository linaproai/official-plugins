<script lang="ts">
export const pluginPageMeta = {
  routePath: '/ai/tiers',
  title: 'Tier Management',
};
</script>

<script setup lang="ts">
import type { MethodDefaultParam, Tier } from './ai-client';

import { computed, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { methodDefaults, tierList, tierTest } from './ai-client';
import {
  buildTierColumns,
  buildTierQuerySchema,
  splitCapabilityMethod,
} from './ai-data';
import TierDrawer from './tier-drawer.vue';

const [TierDrawerRef, tierDrawerApi] = useVbenDrawer({
  connectedComponent: TierDrawer,
});

const testingTierCodes = ref<Record<string, boolean>>({});
const selectedCapabilityKey = ref('text.generate');
const methodDefaultRows = ref<MethodDefaultParam[]>([]);
const currentMethodDefault = computed(() => {
  const capability = splitCapabilityMethod(selectedCapabilityKey.value);
  return methodDefaultRows.value.find(
    (item) =>
      item.capabilityType === capability.capabilityType &&
      item.capabilityMethod === capability.capabilityMethod,
  );
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildTierQuerySchema(),
    commonConfig: {
      labelWidth: 112,
      componentProps: { allowClear: false },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
  },
  gridOptions: {
    columns: buildTierColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: { enabled: false },
    proxyConfig: {
      ajax: {
        query: async (_params: unknown, formValues: Record<string, any> = {}) => {
          selectedCapabilityKey.value = formValues.capabilityKey || 'text.generate';
          const capability = splitCapabilityMethod(selectedCapabilityKey.value);
          const [tiers, defaults] = await Promise.all([
            tierList(capability.capabilityType, capability.capabilityMethod),
            methodDefaults(),
          ]);
          methodDefaultRows.value = defaults;
          return { items: tiers, total: tiers.length };
        },
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

function isTierTesting(code: string) {
  return testingTierCodes.value[`${selectedCapabilityKey.value}:${code}`] === true;
}

function setTierTesting(code: string, testing: boolean) {
  const key = `${selectedCapabilityKey.value}:${code}`;
  if (testing) {
    testingTierCodes.value = { ...testingTierCodes.value, [key]: true };
    return;
  }
  const next = { ...testingTierCodes.value };
  delete next[key];
  testingTierCodes.value = next;
}

async function handleTest(row: Tier) {
  if (isTierTesting(row.code)) {
    return;
  }
  setTierTesting(row.code, true);
  try {
    const result = await tierTest(row.code, {
      capabilityMethod: row.capabilityMethod,
      capabilityType: row.capabilityType,
      maxOutputTokens: 128,
    });
    if (result.status === 'success') {
      message.success($t('plugin.linapro-ai-core.tier.messages.testSuccess'));
    } else {
      message.error(result.errorSummary || $t('plugin.linapro-ai-core.tier.messages.testFailed'));
    }
    await gridApi.query();
  } finally {
    setTierTesting(row.code, false);
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.linapro-ai-core.tier.tableTitle')">
      <template #toolbar-tools>
        <a-tag color="blue">
          {{
            $t('plugin.linapro-ai-core.methodDefault.current', {
              method: selectedCapabilityKey,
            })
          }}
        </a-tag>
        <span class="ml-2 font-mono text-xs text-muted-foreground">
          {{ currentMethodDefault?.defaultParamsJson || '{}' }}
        </span>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <ghost-button
            :disabled="isTierTesting(row.code)"
            :loading="isTierTesting(row.code)"
            @click.stop="handleTest(row)"
          >
            {{ $t('plugin.linapro-ai-core.tier.actions.testSaved') }}
          </ghost-button>
        </Space>
      </template>
    </Grid>

    <TierDrawerRef @reload="() => gridApi.query()" />
  </Page>
</template>
