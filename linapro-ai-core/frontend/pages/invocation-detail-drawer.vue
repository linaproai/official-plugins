<script setup lang="ts">
import type { Invocation } from './ai-client';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';
import { joinCapabilityMethod, tierCodeLabel } from './ai-data';

const record = ref<Invocation>();

const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(open) {
    if (!open) {
      return;
    }
    const data = drawerApi.getData<{ record?: Invocation }>();
    record.value = data?.record;
  },
  onClosed() {
    record.value = undefined;
  },
});
</script>

<template>
  <Drawer
    :title="$t('plugin.linapro-ai-core.invocation.drawer.detailTitle')"
    class="w-[640px] max-w-[calc(100vw-32px)]"
  >
    <Descriptions v-if="record" :column="1" bordered size="small">
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.requestId')">
        {{ record.requestId || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.purpose')">
        {{ record.purpose || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.capability.method')">
        {{ joinCapabilityMethod(record.capabilityType, record.capabilityMethod) }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.status')">
        <Tag :color="record.status === 'success' ? 'success' : 'error'">
          {{
            record.status === 'success'
              ? $t('plugin.linapro-ai-core.common.success')
              : $t('plugin.linapro-ai-core.common.failed')
          }}
        </Tag>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.tierCode')">
        {{ record.tierCode ? tierCodeLabel(record.tierCode) : '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.providerName')">
        {{ record.providerName || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.modelName')">
        {{ record.modelName || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.model.fields.protocol')">
        {{ record.protocol || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.tier.fields.defaultEffort')">
        {{ record.thinkingEffort || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.tokens')">
        {{ record.inputTokens }} / {{ record.outputTokens }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.latencyMs')">
        {{ record.latencyMs }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.assetSummaryJson')">
        <pre class="m-0 whitespace-pre-wrap break-all text-xs">{{ record.assetSummaryJson || '{}' }}</pre>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.operationSummaryJson')">
        <pre class="m-0 whitespace-pre-wrap break-all text-xs">{{ record.operationSummaryJson || '{}' }}</pre>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.metadataSummaryJson')">
        <pre class="m-0 whitespace-pre-wrap break-all text-xs">{{ record.metadataSummaryJson || '{}' }}</pre>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.errorCode')">
        {{ record.errorCode || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.linapro-ai-core.invocation.fields.errorSummary')">
        {{ record.errorSummary || '-' }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('pages.common.createdAt')">
        {{ formatTimestamp(record.createdAt) }}
      </DescriptionsItem>
    </Descriptions>
  </Drawer>
</template>
