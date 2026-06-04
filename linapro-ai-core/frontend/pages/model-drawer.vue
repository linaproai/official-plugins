<script setup lang="ts">
import type {
  Model,
  ModelCapability,
  ModelCapabilityInput,
} from "./ai-client";

import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import {
  modelAdd,
  modelCapabilities,
  modelCapabilitiesSave,
  modelDelete,
  modelUpdate,
  providerList,
} from "./ai-client";
import {
  buildModelFormSchema,
  joinCapabilityMethod,
  splitCapabilityMethod,
} from "./ai-data";

const emit = defineEmits<{ reload: [] }>();

const currentModel = ref<Model>();
const currentCapabilities = ref<ModelCapability[]>([]);
const providerOptions = ref<Array<{ label: string; value: number }>>([]);
const providers = ref<Awaited<ReturnType<typeof providerList>>["items"]>([]);
const title = computed(modelDrawerTitle);
const isEdit = computed(() => Boolean(currentModel.value?.id));

function modelDrawerTitle() {
  return isEdit.value
    ? $t("plugin.linapro-ai-core.model.drawer.editTitle")
    : $t("plugin.linapro-ai-core.model.drawer.createTitle");
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

function normalizeSelectedEndpointIds(value: unknown) {
  const values = Array.isArray(value) ? value : value ? [value] : [];
  return [
    ...new Set(
      values.map((item) => Number(item || 0)).filter((item) => item > 0),
    ),
  ];
}

function providerEndpointById(providerId: number, endpointId: number) {
  return providers.value
    .find((item) => item.id === providerId)
    ?.endpoints?.find((endpoint) => endpoint.id === endpointId);
}

function toModelCapabilityInput(
  item: ModelCapabilityInput,
): ModelCapabilityInput {
  return {
    capabilityMethod: item.capabilityMethod,
    capabilityType: item.capabilityType,
    defaultParamsJson: item.defaultParamsJson,
    enabled: item.enabled,
    endpointId: item.endpointId,
    inputModalities: item.inputModalities,
    maxAssetBytes: item.maxAssetBytes,
    maxInputAssets: item.maxInputAssets,
    maxInputTokens: item.maxInputTokens,
    maxOutputAssets: item.maxOutputAssets,
    maxOutputTokens: item.maxOutputTokens,
    outputModalities: item.outputModalities,
    supportedEfforts: item.supportedEfforts,
    supportsOperation: item.supportsOperation,
    supportsStreaming: item.supportsStreaming,
    supportsThinking: item.supportsThinking,
  };
}

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
        disabled: isEdit.value,
        onChange: (value: number) =>
          refreshEndpointOptions(Number(value), true),
        options: providerOptions.value,
        showSearch: true,
      },
    },
    {
      fieldName: "capabilityKey",
      componentProps: {
        disabled: isEdit.value,
      },
    },
  ]);
}

async function refreshEndpointOptions(
  providerId: number,
  resetEndpoint = false,
) {
  const provider = providers.value.find((item) => item.id === providerId);
  const endpoints = (provider?.endpoints || []).filter(
    (endpoint) =>
      !isEdit.value || endpoint.protocol === currentModel.value?.protocol,
  );
  const options = endpoints.map((endpoint) => ({
    label: `${endpointProtocolLabel(endpoint.protocol)} / ${endpoint.baseUrl}`,
    value: endpoint.id,
  }));
  modelFormApi.updateSchema([
    {
      fieldName: "endpointIds",
      componentProps: {
        allowClear: false,
        maxTagCount: "responsive",
        mode: isEdit.value ? undefined : "multiple",
        optionFilterProp: "label",
        options,
        showSearch: true,
      },
    },
  ]);
  if (resetEndpoint) {
    await modelFormApi.setValues({
      endpointIds: isEdit.value
        ? options[0]?.value
        : options.length === 1
          ? [options[0]?.value]
          : [],
    });
  }
  return options;
}

function currentModelCapability(model?: Model) {
  if (!model) {
    return undefined;
  }
  return currentCapabilities.value.find(
    (item) =>
      item.capabilityType === model.capabilityType &&
      item.capabilityMethod === model.capabilityMethod,
  );
}

async function resetModelForm(providerId = 0, model?: Model) {
  currentModel.value = model;
  const capability = currentModelCapability(model);
  await modelFormApi.resetForm();
  const endpointOptions = await refreshEndpointOptions(providerId, false);
  await modelFormApi.setValues({
    capabilityKey: model
      ? joinCapabilityMethod(model.capabilityType, model.capabilityMethod)
      : "text.generate",
    enabled: model?.enabled ?? 1,
    endpointIds:
      capability?.endpointId ||
      model?.endpointId ||
      (providerId > 0 && endpointOptions.length === 1 && !isEdit.value
        ? [endpointOptions[0]?.value]
        : []),
    maxInputTokens: capability?.maxInputTokens ?? model?.maxInputTokens ?? 0,
    maxOutputTokens: capability?.maxOutputTokens ?? model?.maxOutputTokens ?? 0,
    modelName: model?.modelName,
    providerId: providerId || undefined,
    supportsThinking: capability?.supportsThinking ?? model?.supportsThinking ?? 0,
    supportedEfforts: capability?.supportedEfforts ?? model?.supportedEfforts ?? [],
  });
}

async function saveModel() {
  const { valid } = await modelFormApi.validate();
  if (!valid) {
    return false;
  }
  const values = await modelFormApi.getValues();
  const providerId = Number(values.providerId || 0);
  const endpointIds = normalizeSelectedEndpointIds(values.endpointIds);
  const capability = splitCapabilityMethod(values.capabilityKey);
  const supportsThinking = Number(values.supportsThinking || 0);
  if (isEdit.value) {
    const endpointId = endpointIds[0] || 0;
    const endpoint = providerEndpointById(providerId, endpointId);
    await modelUpdate(currentModel.value!.id, {
      enabled: values.enabled,
      endpointId,
      modelName: values.modelName,
      protocol: endpoint?.protocol || currentModel.value!.protocol,
    });
    const updatedCapability: ModelCapabilityInput = {
      ...currentModelCapability(currentModel.value),
      ...capability,
      enabled: values.enabled,
      endpointId,
      maxInputTokens: values.maxInputTokens,
      maxOutputTokens: values.maxOutputTokens,
      supportedEfforts:
        supportsThinking === 1 ? values.supportedEfforts || [] : [],
      supportsThinking,
    };
    const nextCapabilities = currentCapabilities.value.some(
      (item) =>
        item.capabilityType === capability.capabilityType &&
        item.capabilityMethod === capability.capabilityMethod,
    )
      ? currentCapabilities.value.map((item) =>
          item.capabilityType === capability.capabilityType &&
          item.capabilityMethod === capability.capabilityMethod
            ? updatedCapability
            : toModelCapabilityInput(item),
        )
      : [updatedCapability];
    await modelCapabilitiesSave(currentModel.value!.id, nextCapabilities);
    message.success($t("pages.common.updateSuccess"));
    emit("reload");
    return true;
  }
  const createdModelIds: number[] = [];
  try {
    for (const endpointId of endpointIds) {
      const endpoint = providerEndpointById(providerId, endpointId);
      const created = await modelAdd(providerId, {
        ...capability,
        enabled: values.enabled,
        endpointId,
        maxInputTokens: values.maxInputTokens,
        maxOutputTokens: values.maxOutputTokens,
        modelName: values.modelName,
        protocol: endpoint?.protocol || "openai",
        supportedEfforts:
          supportsThinking === 1 ? values.supportedEfforts || [] : [],
        supportsThinking,
      });
      if (created?.id) {
        createdModelIds.push(Number(created.id));
      }
    }
  } catch (error) {
    await Promise.all(
      createdModelIds.map((id) => modelDelete(id).catch(() => undefined)),
    );
    throw error;
  }
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
    const data = drawerApi.getData<{ model?: Model; providerId?: number }>();
    currentModel.value = data?.model;
    await loadProviderOptions();
    currentCapabilities.value = data?.model
      ? await modelCapabilities(data.model.id)
      : [];
    await resetModelForm(
      Number(data?.model?.providerId || data?.providerId || 0),
      data?.model,
    );
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
    currentModel.value = undefined;
    currentCapabilities.value = [];
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
