<script setup lang="ts">
import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import { modelAdd, providerList } from "./ai-client";
import { buildModelFormSchema, splitCapabilityMethod } from "./ai-data";

const emit = defineEmits<{ reload: [] }>();

const providerOptions = ref<Array<{ label: string; value: number }>>([]);
const providers = ref<Awaited<ReturnType<typeof providerList>>["items"]>([]);
const title = computed(modelDrawerTitle);

function modelDrawerTitle() {
  return $t("plugin.linapro-ai-core.model.drawer.createTitle");
}

function endpointProtocolLabel(protocol: string) {
  return protocol.includes("anthropic") ? "Anthropic" : "OpenAI";
}

const [ModelForm, modelFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    formItemClass: "col-span-1",
    labelClass: "whitespace-nowrap",
    labelWidth: 176,
  },
  schema: buildModelFormSchema(),
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

async function loadProviderOptions() {
  const res = await providerList({ pageNum: 1, pageSize: 100 });
  providers.value = res.items;
  providerOptions.value = res.items.map((provider) => ({
    label: provider.name,
    value: provider.id,
  }));
  modelFormApi.updateSchema([
    {
      fieldName: "providerId",
      componentProps: {
        onChange: (value: number) =>
          refreshEndpointOptions(Number(value), true),
        options: providerOptions.value,
        showSearch: true,
      },
    },
  ]);
}

async function refreshEndpointOptions(
  providerId: number,
  resetEndpoint = false,
) {
  const provider = providers.value.find((item) => item.id === providerId);
  const options = (provider?.endpoints || []).map((endpoint) => ({
    label: `${endpointProtocolLabel(endpoint.protocol)} / ${endpoint.baseUrl}`,
    value: endpoint.id,
  }));
  modelFormApi.updateSchema([
    {
      fieldName: "endpointId",
      componentProps: {
        allowClear: false,
        options,
        showSearch: true,
      },
    },
  ]);
  if (resetEndpoint) {
    await modelFormApi.setValues({
      endpointId: options.length === 1 ? options[0]?.value : undefined,
    });
  }
  return options;
}

async function resetModelForm(providerId = 0) {
  await modelFormApi.resetForm();
  const endpointOptions = await refreshEndpointOptions(providerId, false);
  await modelFormApi.setValues({
    capabilityKey: "text.generate",
    enabled: 1,
    endpointId:
      providerId > 0 && endpointOptions.length === 1
        ? endpointOptions[0]?.value
        : undefined,
    maxInputTokens: 0,
    maxOutputTokens: 0,
    protocol: "openai",
    providerId: providerId || undefined,
    supportsThinking: 0,
    supportedEfforts: [],
  });
}

async function saveModel() {
  const { valid } = await modelFormApi.validate();
  if (!valid) {
    return false;
  }
  const values = await modelFormApi.getValues();
  const providerId = Number(values.providerId || 0);
  const capability = splitCapabilityMethod(values.capabilityKey);
  const supportsThinking = Number(values.supportsThinking || 0);
  await modelAdd(providerId, {
    ...capability,
    enabled: values.enabled,
    endpointId: Number(values.endpointId || 0),
    maxInputTokens: values.maxInputTokens,
    maxOutputTokens: values.maxOutputTokens,
    modelName: values.modelName,
    protocol: values.protocol,
    supportedEfforts:
      supportsThinking === 1 ? values.supportedEfforts || [] : [],
    supportsThinking,
  });
  message.success($t("pages.common.createSuccess"));
  emit("reload");
  return true;
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });
    const data = drawerApi.getData<{ providerId?: number }>();
    await loadProviderOptions();
    await resetModelForm(Number(data?.providerId || 0));
    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const ok = await saveModel();
      if (ok) {
        drawerApi.close();
      }
    } finally {
      drawerApi.lock(false);
    }
  },
  onClosed() {
    providerOptions.value = [];
    providers.value = [];
    modelFormApi.resetForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[760px] max-w-[calc(100vw-32px)]">
    <ModelForm />
  </Drawer>
</template>
