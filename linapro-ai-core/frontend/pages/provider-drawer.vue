<script setup lang="ts">
import type {
  ProviderEndpoint,
  ProviderEndpointSaveInput,
  ProviderSaveInput,
} from "./ai-client";

import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";

import { message } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import { providerAdd, providerInfo, providerUpdate } from "./ai-client";
import { buildProviderFormSchema } from "./ai-data";

const emit = defineEmits<{ reload: [] }>();

const providerId = ref(0);
const openaiEndpoint = ref<ProviderEndpoint>();
const anthropicEndpoint = ref<ProviderEndpoint>();
const title = computed(providerDrawerTitle);

function providerDrawerTitle() {
  return providerId.value
    ? $t("plugin.linapro-ai-core.provider.drawer.editTitle")
    : $t("plugin.linapro-ai-core.provider.drawer.createTitle");
}

const [ProviderForm, providerFormApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    formItemClass: "col-span-1",
    labelWidth: 132,
  },
  schema: buildProviderFormSchema(),
  showDefaultActions: false,
  wrapperClass: "grid-cols-1",
});

function asString(value: unknown) {
  return String(value ?? "").trim();
}

function endpointPayload(
  endpoint: ProviderEndpoint | undefined,
  protocol: "anthropic" | "openai",
  baseUrl: string,
  secretRef: string,
): ProviderEndpointSaveInput | undefined {
  if (!endpoint?.id && !baseUrl) {
    return undefined;
  }
  return {
    baseUrl,
    enabled: endpoint?.enabled ?? 1,
    id: endpoint?.id,
    metadataJson: endpoint?.metadataJson || "{}",
    protocol,
    secretRef,
  };
}

function providerPayload(values: Record<string, unknown>): ProviderSaveInput {
  const secretRef = asString(values.secretRef);
  const endpoints = [
    endpointPayload(
      openaiEndpoint.value,
      "openai",
      asString(values.openaiBaseUrl),
      secretRef,
    ),
    endpointPayload(
      anthropicEndpoint.value,
      "anthropic",
      asString(values.anthropicBaseUrl),
      secretRef,
    ),
  ].filter((item): item is ProviderEndpointSaveInput => Boolean(item));

  return {
    enabled: Number(values.enabled ?? 1),
    endpoints,
    name: asString(values.name),
    remark: asString(values.remark),
    websiteUrl: asString(values.websiteUrl),
  };
}

function resetEndpointState() {
  openaiEndpoint.value = undefined;
  anthropicEndpoint.value = undefined;
}

function updateSecretPlaceholder() {
  providerFormApi.updateSchema([
    {
      fieldName: "secretRef",
      componentProps: {
        autocomplete: "new-password",
        placeholder: providerId.value
          ? $t("plugin.linapro-ai-core.provider.placeholders.keepSecret")
          : $t("plugin.linapro-ai-core.provider.placeholders.apiKeyCreate"),
      },
    },
  ]);
}

async function applyCreateDefaults() {
  resetEndpointState();
  await providerFormApi.setValues({
    anthropicBaseUrl: "https://api.anthropic.com/v1",
    enabled: 1,
    openaiBaseUrl: "",
    secretRef: "",
  });
}

async function applyProviderDetail(id: number) {
  const detail = await providerInfo(id);
  const endpoints = detail.endpoints || [];
  openaiEndpoint.value = endpoints.find((item) => item.protocol === "openai");
  anthropicEndpoint.value = endpoints.find(
    (item) => item.protocol === "anthropic",
  );
  await providerFormApi.setValues({
    enabled: detail.enabled,
    name: detail.name,
    remark: detail.remark,
    websiteUrl: detail.websiteUrl,
    openaiBaseUrl: openaiEndpoint.value?.baseUrl || "",
    anthropicBaseUrl: anthropicEndpoint.value?.baseUrl || "",
    secretRef: "",
  });
}

async function saveProvider() {
  const { valid } = await providerFormApi.validate();
  if (!valid) {
    return false;
  }
  const values = await providerFormApi.getValues();
  const payload = providerPayload(values);
  if (providerId.value) {
    await providerUpdate(providerId.value, payload);
    message.success($t("pages.common.updateSuccess"));
  } else {
    const created = (await providerAdd(payload)) as { id?: number };
    providerId.value = Number(created?.id || 0);
    message.success($t("pages.common.createSuccess"));
  }
  emit("reload");
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
    updateSecretPlaceholder();
    if (providerId.value) {
      await applyProviderDetail(providerId.value);
    } else {
      await applyCreateDefaults();
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
    resetEndpointState();
    providerFormApi.resetForm();
    updateSecretPlaceholder();
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
