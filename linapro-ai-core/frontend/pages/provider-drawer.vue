<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { $t } from '#/locales';
import {
  providerAdd,
  providerInfo,
  providerUpdate,
} from './ai-client';
import { buildProviderFormSchema } from './ai-data';

const emit = defineEmits<{ reload: [] }>();

const providerId = ref(0);
const title = computed(providerDrawerTitle);

function providerDrawerTitle() {
  return providerId.value
    ? $t('plugin.linapro-ai-core.provider.drawer.editTitle')
    : $t('plugin.linapro-ai-core.provider.drawer.createTitle');
}

const [ProviderForm, providerFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: 'w-full' },
    formItemClass: 'col-span-1',
    labelWidth: 132,
  },
  schema: buildProviderFormSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-1',
});

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

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });
    const data = drawerApi.getData<{ id?: number }>();
    providerId.value = Number(data?.id || 0);
    await providerFormApi.resetForm();
    if (providerId.value) {
      const detail = await providerInfo(providerId.value);
      await providerFormApi.setValues(detail);
    }
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
    providerFormApi.resetForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[720px] max-w-[calc(100vw-32px)]">
    <div class="flex flex-col gap-[16px]">
      <ProviderForm />
    </div>
  </Drawer>
</template>
