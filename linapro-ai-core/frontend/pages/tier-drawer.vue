<script setup lang="ts">
import type { Model, Provider, Tier } from './ai-client';
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message, Space } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { $t } from '#/locales';
import {
  providerList,
  providerModels,
  tierTest,
  tierUpdate,
} from './ai-client';
import { buildEffortOptions, buildEnabledOptions, tierDisplayName } from './ai-data';

const emit = defineEmits<{ reload: [] }>();

const tier = ref<Tier>();
const providers = ref<Provider[]>([]);
const models = ref<Model[]>([]);
const testing = ref(false);
const title = computed(tierDrawerTitle);

function tierDrawerTitle() {
  return $t('plugin.linapro-ai-core.tier.drawer.editTitle', {
    name: tierDisplayName(tier.value),
  });
}

function modelLabel(model: Model) {
  const efforts = model.supportedEfforts?.length
    ? model.supportedEfforts.join(',')
    : $t('plugin.linapro-ai-core.effort.empty');
  return `${model.modelName} / ${model.protocol} / ${efforts}`;
}

function supportsThinkingEffort() {
  return (
    tier.value?.capabilityType === 'text' &&
    tier.value?.capabilityMethod === 'generate'
  );
}

function buildSchema(): VbenFormSchema[] {
  return [
    {
      component: 'RadioGroup',
      fieldName: 'enabled',
      label: $t('pages.common.status'),
      defaultValue: 1,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: buildEnabledOptions(),
      },
    },
    {
      component: 'Select',
      fieldName: 'defaultEffort',
      label: $t('plugin.linapro-ai-core.tier.fields.defaultEffort'),
      componentProps: { options: buildEffortOptions() },
    },
    {
      component: 'Select',
      fieldName: 'providerId',
      label: $t('plugin.linapro-ai-core.tier.fields.provider'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'Select',
      fieldName: 'modelId',
      label: $t('plugin.linapro-ai-core.tier.fields.model'),
      formItemClass: 'col-span-2',
    },
  ];
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
    formItemClass: 'col-span-1',
    labelWidth: 132,
  },
  schema: buildSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

async function refreshModelOptions(providerId: number, resetModel = false) {
  models.value = providerId
    ? await providerModels(
        providerId,
        1,
        tier.value?.capabilityType || 'text',
        tier.value?.capabilityMethod || 'generate',
      )
    : [];
  formApi.updateSchema([
    {
      fieldName: 'modelId',
      componentProps: {
        options: models.value.map((item) => ({
          label: modelLabel(item),
          value: item.id,
        })),
      },
    },
  ]);
  if (resetModel) {
    await formApi.setValues({ modelId: undefined });
  }
}

async function refreshProviderOptions() {
  const out = await providerList({ pageNum: 1, pageSize: 100, enabled: 1 });
  providers.value = out.items;
  formApi.updateSchema([
    {
      fieldName: 'providerId',
      componentProps: {
        onChange: (value: number) => refreshModelOptions(Number(value), true),
        options: providers.value.map((item) => ({
          label: item.name,
          value: item.id,
        })),
      },
    },
  ]);
}

async function currentValues() {
  const values = await formApi.getValues();
  return {
    capabilityMethod: tier.value?.capabilityMethod || 'generate',
    capabilityType: tier.value?.capabilityType || 'text',
    enabled: Number(values.enabled ?? 0),
    defaultEffort: supportsThinkingEffort() ? values.defaultEffort || '' : '',
    providerId: Number(values.providerId || 0),
    modelId: Number(values.modelId || 0),
  };
}

function effortSupported(model: Model | undefined, effort: string) {
  if (!effort) {
    return true;
  }
  return model?.supportsThinking === 1 && model.supportedEfforts?.includes(effort);
}

function validateBindingValues(
  values: Awaited<ReturnType<typeof currentValues>>,
  requireBinding: boolean,
) {
  const hasProvider = values.providerId > 0;
  const hasModel = values.modelId > 0;
  const bindingRequired = requireBinding || values.enabled === 1 || hasProvider || hasModel;
  if (bindingRequired && (!hasProvider || !hasModel)) {
    message.error($t('plugin.linapro-ai-core.tier.messages.bindingRequired'));
    return false;
  }
  if (hasModel) {
    const model = models.value.find((item) => item.id === values.modelId);
    if (supportsThinkingEffort() && !effortSupported(model, values.defaultEffort)) {
      message.error($t('plugin.linapro-ai-core.tier.messages.unsupportedEffort'));
      return false;
    }
  }
  return true;
}

async function handleTest() {
  if (testing.value) {
    return;
  }
  const values = await currentValues();
  if (!validateBindingValues(values, true)) {
    return;
  }
  testing.value = true;
  try {
    const result = await tierTest(tier.value?.code || '', {
      ...values,
      thinkingEffort: values.defaultEffort,
      maxOutputTokens: 128,
    });
    if (result.status === 'success') {
      message.success($t('plugin.linapro-ai-core.tier.messages.testSuccess'));
    } else {
      message.error(result.errorSummary || $t('plugin.linapro-ai-core.tier.messages.testFailed'));
    }
    emit('reload');
  } finally {
    testing.value = false;
  }
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });
    const data = drawerApi.getData<{ tier?: Tier }>();
    tier.value = data?.tier;
    formApi.updateSchema([
      {
        fieldName: 'defaultEffort',
        hide: !supportsThinkingEffort(),
      },
    ]);
    await formApi.resetForm();
    await refreshProviderOptions();
    const providerId = tier.value?.binding?.providerId || undefined;
    await refreshModelOptions(Number(providerId || 0), false);
    await formApi.setValues({
      enabled: tier.value?.enabled ?? 0,
      defaultEffort: tier.value?.defaultEffort || '',
      providerId,
      modelId: tier.value?.binding?.modelId || undefined,
    });
    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const { valid } = await formApi.validate();
      if (!valid || !tier.value) {
        return;
      }
      const values = await currentValues();
      if (!validateBindingValues(values, false)) {
        return;
      }
      await tierUpdate(tier.value.code, values);
      message.success($t('pages.common.updateSuccess'));
      emit('reload');
      drawerApi.close();
    } finally {
      drawerApi.lock(false);
    }
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[720px] max-w-[calc(100vw-32px)]">
    <div class="flex flex-col gap-[16px]">
      <Form />
      <div class="flex justify-end">
        <Space>
          <a-button :disabled="testing" :loading="testing" @click="handleTest">
            {{ $t('plugin.linapro-ai-core.tier.actions.testDraft') }}
          </a-button>
        </Space>
      </div>
    </div>
  </Drawer>
</template>
