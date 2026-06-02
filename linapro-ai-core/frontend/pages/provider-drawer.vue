<script setup lang="ts">
import type { Model } from './ai-client';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message, Popconfirm, Space, Tag } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { $t } from '#/locales';
import {
  modelAdd,
  modelDelete,
  modelSync,
  modelUpdate,
  providerAdd,
  providerInfo,
  providerModels,
  providerUpdate,
} from './ai-client';
import { buildModelFormSchema, buildProviderFormSchema } from './ai-data';

const emit = defineEmits<{ reload: [] }>();

const providerId = ref(0);
const models = ref<Model[]>([]);
const editingModelId = ref(0);
const title = computed(providerDrawerTitle);
const modelColumns = computed(buildModelColumns);

function providerDrawerTitle() {
  return providerId.value
    ? $t('plugin.linapro-ai-core.provider.drawer.editTitle')
    : $t('plugin.linapro-ai-core.provider.drawer.createTitle');
}

function buildModelColumns() {
  return [
    {
      dataIndex: 'modelName',
      title: $t('plugin.linapro-ai-core.model.fields.modelName'),
    },
    {
      dataIndex: 'protocol',
      title: $t('plugin.linapro-ai-core.model.fields.protocol'),
    },
    {
      dataIndex: 'supportsThinking',
      title: $t('plugin.linapro-ai-core.model.fields.supportsThinking'),
    },
    {
      dataIndex: 'enabled',
      title: $t('pages.common.status'),
    },
    {
      dataIndex: 'action',
      fixed: 'right',
      title: $t('pages.common.actions'),
      width: 150,
    },
  ];
}

const [ProviderForm, providerFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
    formItemClass: 'col-span-1',
    labelWidth: 132,
  },
  schema: buildProviderFormSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

const [ModelForm, modelFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
    formItemClass: 'col-span-1',
    labelWidth: 132,
  },
  schema: buildModelFormSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

async function reloadModels() {
  if (!providerId.value) {
    models.value = [];
    return;
  }
  models.value = await providerModels(providerId.value);
}

async function resetModelForm() {
  editingModelId.value = 0;
  await modelFormApi.resetForm();
  await modelFormApi.setValues({
    capabilityType: 'text',
    enabled: 1,
    maxInputTokens: 0,
    maxOutputTokens: 0,
    protocol: 'openai',
    supportsThinking: 0,
    supportedEfforts: [],
  });
}

async function saveProvider() {
  const { valid } = await providerFormApi.validate();
  if (!valid) {
    return false;
  }
  const values = await providerFormApi.getValues();
  if (providerId.value) {
    await providerUpdate(providerId.value, values);
    message.success($t('pages.common.updateSuccess'));
  } else {
    const created = await providerAdd(values) as { id?: number };
    providerId.value = Number(created?.id || 0);
    message.success($t('pages.common.createSuccess'));
  }
  emit('reload');
  return true;
}

async function handleModelSave() {
  if (!providerId.value) {
    const ok = await saveProvider();
    if (!ok || !providerId.value) {
      return;
    }
  }
  const { valid } = await modelFormApi.validate();
  if (!valid) {
    return;
  }
  const values = await modelFormApi.getValues();
  if (editingModelId.value) {
    await modelUpdate(editingModelId.value, values);
    message.success($t('pages.common.updateSuccess'));
  } else {
    await modelAdd(providerId.value, values);
    message.success($t('pages.common.createSuccess'));
  }
  await resetModelForm();
  await reloadModels();
  emit('reload');
}

async function handleModelEdit(row: Model) {
  editingModelId.value = row.id;
  await modelFormApi.setValues({ ...row });
}

async function handleModelDelete(row: Model) {
  await modelDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  await reloadModels();
  emit('reload');
}

async function handleSync(protocol: string) {
  if (!providerId.value) {
    const ok = await saveProvider();
    if (!ok || !providerId.value) {
      return;
    }
  }
  const result = await modelSync(providerId.value, protocol);
  message.success(
    $t('plugin.linapro-ai-core.model.messages.syncDone', {
      created: result.created,
      kept: result.kept,
    }),
  );
  await reloadModels();
  emit('reload');
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });
    const data = drawerApi.getData<{ id?: number }>();
    providerId.value = Number(data?.id || 0);
    await providerFormApi.resetForm();
    await resetModelForm();
    if (providerId.value) {
      const detail = await providerInfo(providerId.value);
      await providerFormApi.setValues(detail);
    }
    await reloadModels();
    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const ok = await saveProvider();
      if (ok) {
        drawerApi.close();
      }
    } finally {
      drawerApi.lock(false);
    }
  },
  onClosed() {
    providerId.value = 0;
    models.value = [];
    providerFormApi.resetForm();
    resetModelForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[920px] max-w-[calc(100vw-32px)]">
    <div class="flex flex-col gap-[16px]">
      <ProviderForm />

      <a-divider class="!my-0" />

      <div class="flex items-center justify-between gap-[12px]">
        <div class="text-base font-medium">
          {{ $t('plugin.linapro-ai-core.model.tableTitle') }}
        </div>
        <Space>
          <a-button @click="handleSync('openai')">
            {{ $t('plugin.linapro-ai-core.model.actions.syncOpenAI') }}
          </a-button>
          <a-button @click="handleSync('anthropic')">
            {{ $t('plugin.linapro-ai-core.model.actions.syncAnthropic') }}
          </a-button>
        </Space>
      </div>

      <a-table
        :columns="modelColumns"
        :data-source="models"
        :pagination="false"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'supportsThinking'">
            <Tag :color="record.supportsThinking === 1 ? 'success' : 'default'">
              {{
                record.supportsThinking === 1
                  ? $t('plugin.linapro-ai-core.common.yes')
                  : $t('plugin.linapro-ai-core.common.no')
              }}
            </Tag>
          </template>
          <template v-else-if="column.dataIndex === 'enabled'">
            <Tag :color="record.enabled === 1 ? 'success' : 'default'">
              {{
                record.enabled === 1
                  ? $t('plugin.linapro-ai-core.common.enabled')
                  : $t('plugin.linapro-ai-core.common.disabled')
              }}
            </Tag>
          </template>
          <template v-else-if="column.dataIndex === 'action'">
            <Space>
              <ghost-button @click="handleModelEdit(record)">
                {{ $t('pages.common.edit') }}
              </ghost-button>
              <Popconfirm
                :title="$t('pages.common.deleteConfirm')"
                placement="left"
                @confirm="handleModelDelete(record)"
              >
                <ghost-button danger @click.stop="">
                  {{ $t('pages.common.delete') }}
                </ghost-button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </a-table>

      <a-divider class="!my-0" />

      <div class="flex items-center justify-between gap-[12px]">
        <div class="text-base font-medium">
          {{
            editingModelId
              ? $t('plugin.linapro-ai-core.model.drawer.editTitle')
              : $t('plugin.linapro-ai-core.model.drawer.createTitle')
          }}
        </div>
        <Space>
          <a-button @click="resetModelForm">
            {{ $t('pages.common.reset') }}
          </a-button>
          <a-button type="primary" @click="handleModelSave">
            {{ $t('pages.common.save') }}
          </a-button>
        </Space>
      </div>
      <ModelForm />
    </div>
  </Drawer>
</template>
