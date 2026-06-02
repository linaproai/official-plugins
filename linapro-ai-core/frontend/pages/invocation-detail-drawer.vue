<script setup lang="ts">
import type { Invocation } from './ai-client';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';
import { tierCodeLabel } from './ai-data';

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
    <a-descriptions v-if="record" :column="1" bordered size="small">
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.requestId')">
        {{ record.requestId || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.purpose')">
        {{ record.purpose || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.status')">
        <Tag :color="record.status === 'success' ? 'success' : 'error'">
          {{
            record.status === 'success'
              ? $t('plugin.linapro-ai-core.common.success')
              : $t('plugin.linapro-ai-core.common.failed')
          }}
        </Tag>
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.tierCode')">
        {{ record.tierCode ? tierCodeLabel(record.tierCode) : '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.providerName')">
        {{ record.providerName || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.modelName')">
        {{ record.modelName || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.model.fields.protocol')">
        {{ record.protocol || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.tier.fields.defaultEffort')">
        {{ record.thinkingEffort || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.tokens')">
        {{ record.inputTokens }} / {{ record.outputTokens }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.latencyMs')">
        {{ record.latencyMs }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.errorCode')">
        {{ record.errorCode || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('plugin.linapro-ai-core.invocation.fields.errorSummary')">
        {{ record.errorSummary || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="$t('pages.common.createdAt')">
        {{ formatTimestamp(record.createdAt) }}
      </a-descriptions-item>
    </a-descriptions>
  </Drawer>
</template>
