<script setup lang="ts">
import type { ProviderEndpoint } from "./ai-client";

import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import { message, Popconfirm, Space, Tag } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import {
  providerEndpointAdd,
  providerEndpointDelete,
  providerEndpoints,
  providerEndpointUpdate,
} from "./ai-client";
import { buildEndpointFormSchema, protocolOptions } from "./ai-data";

const emit = defineEmits<{ reload: [] }>();

const providerId = ref(0);
const providerName = ref("");
const editingId = ref(0);
const endpoints = ref<ProviderEndpoint[]>([]);
const title = computed(() =>
  $t("plugin.linapro-ai-core.endpoint.drawer.title", {
    name: providerName.value || "-",
  }),
);

const [EndpointForm, endpointFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    formItemClass: "col-span-1",
    labelWidth: 132,
  },
  schema: buildEndpointFormSchema(),
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

function protocolLabel(value: string) {
  return protocolOptions.find((item) => item.value === value)?.label || value;
}

function resetForm() {
  editingId.value = 0;
  endpointFormApi.resetForm();
  endpointFormApi.setValues({
    enabled: 1,
    metadataJson: "{}",
    protocol: "openai-compatible",
  });
}

async function reloadEndpoints() {
  endpoints.value = providerId.value
    ? await providerEndpoints(providerId.value)
    : [];
}

async function editEndpoint(endpoint: ProviderEndpoint) {
  editingId.value = endpoint.id;
  await endpointFormApi.setValues({
    ...endpoint,
    secretRef: "",
  });
  endpointFormApi.updateSchema([
    {
      fieldName: "secretRef",
      componentProps: {
        autocomplete: "new-password",
        placeholder: $t("plugin.linapro-ai-core.endpoint.placeholders.keepSecret"),
      },
    },
  ]);
}

async function removeEndpoint(endpoint: ProviderEndpoint) {
  await providerEndpointDelete(providerId.value, endpoint.id);
  message.success($t("pages.common.deleteSuccess"));
  await reloadEndpoints();
  emit("reload");
}

async function saveEndpoint() {
  const { valid } = await endpointFormApi.validate();
  if (!valid || providerId.value <= 0) {
    return false;
  }
  const values = await endpointFormApi.getValues();
  if (editingId.value) {
    await providerEndpointUpdate(providerId.value, editingId.value, values);
    message.success($t("pages.common.updateSuccess"));
  } else {
    await providerEndpointAdd(providerId.value, values);
    message.success($t("pages.common.createSuccess"));
  }
  resetForm();
  await reloadEndpoints();
  emit("reload");
  return true;
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });
    const data = drawerApi.getData<{ providerId?: number; providerName?: string }>();
    providerId.value = Number(data?.providerId || 0);
    providerName.value = data?.providerName || "";
    resetForm();
    await reloadEndpoints();
    drawerApi.setState({ loading: false });
  },
  onClosed() {
    providerId.value = 0;
    providerName.value = "";
    endpoints.value = [];
    resetForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[760px] max-w-[calc(100vw-32px)]">
    <div class="flex flex-col gap-4">
      <div class="rounded border border-border">
        <div
          v-for="endpoint in endpoints"
          :key="endpoint.id"
          class="flex items-start justify-between gap-3 border-b border-border px-3 py-2 last:border-b-0"
        >
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <Tag :color="endpoint.enabled === 1 ? 'blue' : 'default'">
                {{ protocolLabel(endpoint.protocol) }}
              </Tag>
              <span class="font-mono text-xs text-muted-foreground">
                {{ endpoint.secretRef || '-' }}
              </span>
            </div>
            <div class="mt-1 break-all font-mono text-xs">
              {{ endpoint.baseUrl }}
            </div>
          </div>
          <Space class="shrink-0">
            <ghost-button @click.stop="editEndpoint(endpoint)">
              {{ $t('pages.common.edit') }}
            </ghost-button>
            <Popconfirm
              :title="$t('pages.common.deleteConfirm')"
              placement="left"
              @confirm="removeEndpoint(endpoint)"
            >
              <ghost-button danger @click.stop="">
                {{ $t('pages.common.delete') }}
              </ghost-button>
            </Popconfirm>
          </Space>
        </div>
        <div
          v-if="endpoints.length === 0"
          class="px-3 py-4 text-sm text-muted-foreground"
        >
          {{ $t('plugin.linapro-ai-core.endpoint.empty') }}
        </div>
      </div>

      <EndpointForm />

      <div class="flex justify-end">
        <Space>
          <a-button @click="resetForm">
            {{ $t('pages.common.reset') }}
          </a-button>
          <a-button type="primary" @click="saveEndpoint">
            <template #icon>
              <IconifyIcon icon="lucide:save" />
            </template>
            {{
              editingId
                ? $t('pages.common.save')
                : $t('plugin.linapro-ai-core.endpoint.actions.add')
            }}
          </a-button>
        </Space>
      </div>
    </div>
  </Drawer>
</template>
