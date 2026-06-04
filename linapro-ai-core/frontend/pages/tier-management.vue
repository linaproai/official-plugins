<script lang="ts">
export const pluginPageMeta = {
  routePath: "/ai/tiers",
  title: "Tier Management",
};
</script>

<script setup lang="ts">
import type { Tier } from "./ai-client";

import { computed, nextTick, ref } from "vue";

import { Page, useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { message, Space, Tabs } from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { $t } from "#/locales";
import { tierList, tierTest } from "./ai-client";
import {
  buildTierColumns,
  capabilityTypeLabel,
  defaultTierCapabilityMethod,
  splitCapabilityMethod,
  tierCapabilityTypeKeys,
} from "./ai-data";
import TierDrawer from "./tier-drawer.vue";

const [TierDrawerRef, tierDrawerApi] = useVbenDrawer({
  connectedComponent: TierDrawer,
});

const TabPane = Tabs.TabPane;
const testingTierCodes = ref<Record<string, boolean>>({});
const selectedCapabilityType = ref("text");
const selectedCapabilityKey = computed(() =>
  defaultTierCapabilityMethod(selectedCapabilityType.value),
);
const capabilityTypeIcons: Record<string, string> = {
  audio: "lucide:audio-lines",
  document: "lucide:file-search",
  embedding: "lucide:network",
  image: "lucide:image",
  safety: "lucide:shield-check",
  text: "lucide:file-text",
  video: "lucide:video",
  vision: "lucide:eye",
};

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: buildTierColumns(),
    height: 300,
    keepSource: true,
    pagerConfig: { enabled: false },
    proxyConfig: {
      ajax: {
        query: async () => {
          const capability = splitCapabilityMethod(selectedCapabilityKey.value);
          const tiers = await tierList(
            capability.capabilityType,
            capability.capabilityMethod,
          );
          return { items: tiers, total: tiers.length };
        },
      },
    },
    rowConfig: { keyField: "code" },
    id: "linapro-ai-core-tier-index",
  },
});

async function handleCapabilityTypeChange(activeKey: string | number) {
  selectedCapabilityType.value = String(activeKey || "text");
  await nextTick();
  await gridApi.query();
}

function capabilityTypeIcon(type: string) {
  return capabilityTypeIcons[type] || "lucide:square";
}

function handleEdit(row: Tier) {
  tierDrawerApi.setData({ tier: row });
  tierDrawerApi.open();
}

function isTierTesting(code: string) {
  return (
    testingTierCodes.value[`${selectedCapabilityKey.value}:${code}`] === true
  );
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

function formatLatencyMs(value: number | undefined) {
  return `${Math.max(0, Math.round(Number(value || 0)))}ms`;
}

function resultMessage(result: Awaited<ReturnType<typeof tierTest>>) {
  const text =
    result.status === "success"
      ? $t("plugin.linapro-ai-core.tier.messages.testSuccess")
      : result.errorSummary ||
        $t("plugin.linapro-ai-core.tier.messages.testFailed");
  return `${text} (${formatLatencyMs(result.latencyMs)})`;
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
    if (result.status === "success") {
      message.success(resultMessage(result));
    } else {
      message.error(resultMessage(result));
    }
    await gridApi.query();
  } finally {
    setTierTesting(row.code, false);
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Tabs
      v-model:active-key="selectedCapabilityType"
      :tab-bar-gutter="28"
      class="tier-capability-tabs"
      data-testid="ai-tier-capability-tabs"
      @change="handleCapabilityTypeChange"
    >
      <TabPane
        v-for="capabilityType in tierCapabilityTypeKeys"
        :key="capabilityType"
      >
        <template #tab>
          <span class="tier-capability-tab-label">
            <IconifyIcon
              :data-testid="`ai-tier-capability-tab-icon-${capabilityType}`"
              :icon="capabilityTypeIcon(capabilityType)"
              class="tier-capability-tab-icon"
            />
            <span>{{ capabilityTypeLabel(capabilityType) }}</span>
          </span>
        </template>

        <div
          v-if="capabilityType === selectedCapabilityType"
          class="tier-capability-content"
          data-testid="ai-tier-capability-content"
        >
          <Grid :table-title="$t('plugin.linapro-ai-core.tier.tableTitle')">
            <template #action="{ row }">
              <Space>
                <ghost-button @click.stop="handleEdit(row)">
                  {{ $t("pages.common.edit") }}
                </ghost-button>
                <ghost-button
                  :disabled="isTierTesting(row.code)"
                  :loading="isTierTesting(row.code)"
                  @click.stop="handleTest(row)"
                >
                  {{ $t("plugin.linapro-ai-core.tier.actions.testSaved") }}
                </ghost-button>
              </Space>
            </template>
          </Grid>
        </div>
      </TabPane>
    </Tabs>

    <TierDrawerRef @reload="() => gridApi.query()" />
  </Page>
</template>

<style scoped>
.tier-capability-tabs {
  margin-bottom: 12px;
  background: hsl(var(--background));
}

.tier-capability-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 0;
}

.tier-capability-tabs :deep(.ant-tabs-nav-wrap) {
  padding: 0 20px;
}

.tier-capability-tabs :deep(.ant-tabs-nav::before) {
  border-bottom-color: hsl(var(--border));
}

.tier-capability-tabs :deep(.ant-tabs-content-holder) {
  background: hsl(var(--background));
  border: 0;
  padding: 16px 20px 0;
}

.tier-capability-tabs :deep(.ant-tabs-tab) {
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

.tier-capability-tabs :deep(.ant-tabs-tab:hover),
.tier-capability-tabs :deep(.ant-tabs-tab-active),
.tier-capability-tabs :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
  color: hsl(var(--primary)) !important;
}

.tier-capability-tabs :deep(.ant-tabs-ink-bar) {
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: hsl(var(--primary));
}

.tier-capability-tabs :deep(.ant-tabs-tab-btn) {
  color: inherit;
  text-align: left;
}

.tier-capability-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  line-height: 1;
}

.tier-capability-tab-icon {
  color: inherit;
  font-size: 16px;
}

.tier-capability-content {
  background: hsl(var(--background));
  min-height: 320px;
}
</style>
