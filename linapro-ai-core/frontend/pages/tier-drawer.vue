<script setup lang="ts">
import type { Model, Provider, Tier } from "./ai-client";
import type { VbenFormSchema } from "#/adapter/form";

import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";

import { message, Space } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import {
  methodDefaults,
  methodDefaultUpdate,
  providerList,
  providerModels,
  tierTest,
  tierUpdate,
} from "./ai-client";
import {
  buildEffortOptions,
  buildEnabledOptions,
  capabilityTypeLabel,
  tierDisplayName,
} from "./ai-data";
import JsonHighlightEditor from "./json-highlight-editor.vue";

const emit = defineEmits<{ reload: [] }>();

const tier = ref<Tier>();
const providers = ref<Provider[]>([]);
const models = ref<Model[]>([]);
const testing = ref(false);
const defaultParamsInvalid = ref(false);
const methodDefaultParamsJson = ref("{}");
const title = computed(tierDrawerTitle);
const methodDefaultScope = computed(() => {
  const capabilityType = tier.value?.capabilityType || "text";
  const capabilityMethod = tier.value?.capabilityMethod || "generate";
  return `${capabilityTypeLabel(capabilityType)} / ${capabilityType}.${capabilityMethod}`;
});

function tierDrawerTitle() {
  return $t("plugin.linapro-ai-core.tier.drawer.editTitle", {
    name: tierDisplayName(tier.value),
  });
}

function modelLabel(model: Model) {
  const efforts = model.supportedEfforts?.length
    ? model.supportedEfforts.join(",")
    : $t("plugin.linapro-ai-core.effort.empty");
  return `${model.modelName} / ${model.protocol} / ${efforts}`;
}

function supportsThinkingEffort() {
  return (
    tier.value?.capabilityType === "text" &&
    tier.value?.capabilityMethod === "generate"
  );
}

function formatDefaultParamsJson(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "{}";
  }
  try {
    return JSON.stringify(JSON.parse(trimmed), null, 2);
  } catch {
    return trimmed;
  }
}

function normalizeDefaultParamsJson(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "{}";
  }
  const parsed = JSON.parse(trimmed);
  if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
    throw new Error("default params must be a JSON object");
  }
  return JSON.stringify(parsed);
}

function buildSchema(): VbenFormSchema[] {
  return [
    {
      component: "RadioGroup",
      fieldName: "enabled",
      label: $t("pages.common.status"),
      defaultValue: 1,
      componentProps: {
        buttonStyle: "solid",
        optionType: "button",
        options: buildEnabledOptions(),
      },
    },
    {
      component: "Select",
      fieldName: "defaultEffort",
      label: $t("plugin.linapro-ai-core.tier.fields.defaultEffort"),
      componentProps: { options: buildEffortOptions() },
    },
    {
      component: "Select",
      fieldName: "providerId",
      label: $t("plugin.linapro-ai-core.tier.fields.provider"),
      formItemClass: "col-span-2",
    },
    {
      component: "Select",
      fieldName: "modelId",
      label: $t("plugin.linapro-ai-core.tier.fields.model"),
      formItemClass: "col-span-2",
    },
    {
      component: "Input",
      fieldName: "defaultParamsJson",
      label: $t(
        "plugin.linapro-ai-core.methodDefault.fields.defaultParamsJson",
      ),
      formItemClass: "col-span-2 items-start",
    },
  ];
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: { class: "w-full" },
    formItemClass: "col-span-1",
    labelWidth: 132,
  },
  schema: buildSchema(),
  showDefaultActions: false,
  wrapperClass: "grid-cols-2",
});

async function refreshModelOptions(providerId: number, resetModel = false) {
  models.value = providerId
    ? await providerModels(
        providerId,
        1,
        tier.value?.capabilityType || "text",
        tier.value?.capabilityMethod || "generate",
      )
    : [];
  formApi.updateSchema([
    {
      fieldName: "modelId",
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
      fieldName: "providerId",
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

async function refreshMethodDefaultParams() {
  const capabilityType = tier.value?.capabilityType || "text";
  const capabilityMethod = tier.value?.capabilityMethod || "generate";
  const rows = await methodDefaults();
  const row = rows.find(
    (item) =>
      item.capabilityType === capabilityType &&
      item.capabilityMethod === capabilityMethod,
  );
  methodDefaultParamsJson.value = formatDefaultParamsJson(
    row?.defaultParamsJson || "{}",
  );
  defaultParamsInvalid.value = false;
}

async function currentValues() {
  const values = await formApi.getValues();
  return {
    capabilityMethod: tier.value?.capabilityMethod || "generate",
    capabilityType: tier.value?.capabilityType || "text",
    enabled: Number(values.enabled ?? 0),
    defaultEffort: supportsThinkingEffort() ? values.defaultEffort || "" : "",
    providerId: Number(values.providerId || 0),
    modelId: Number(values.modelId || 0),
  };
}

function effortSupported(model: Model | undefined, effort: string) {
  if (!effort) {
    return true;
  }
  return (
    model?.supportsThinking === 1 && model.supportedEfforts?.includes(effort)
  );
}

function validateBindingValues(
  values: Awaited<ReturnType<typeof currentValues>>,
  requireBinding: boolean,
) {
  const hasProvider = values.providerId > 0;
  const hasModel = values.modelId > 0;
  const bindingRequired =
    requireBinding || values.enabled === 1 || hasProvider || hasModel;
  if (bindingRequired && (!hasProvider || !hasModel)) {
    message.error($t("plugin.linapro-ai-core.tier.messages.bindingRequired"));
    return false;
  }
  if (hasModel) {
    const model = models.value.find((item) => item.id === values.modelId);
    if (
      supportsThinkingEffort() &&
      !effortSupported(model, values.defaultEffort)
    ) {
      message.error(
        $t("plugin.linapro-ai-core.tier.messages.unsupportedEffort"),
      );
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
    const result = await tierTest(tier.value?.code || "", {
      ...values,
      thinkingEffort: values.defaultEffort,
      maxOutputTokens: 128,
    });
    if (result.status === "success") {
      message.success($t("plugin.linapro-ai-core.tier.messages.testSuccess"));
    } else {
      message.error(
        result.errorSummary ||
          $t("plugin.linapro-ai-core.tier.messages.testFailed"),
      );
    }
    emit("reload");
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
        fieldName: "defaultEffort",
        hide: !supportsThinkingEffort(),
      },
    ]);
    await formApi.resetForm();
    await Promise.all([refreshProviderOptions(), refreshMethodDefaultParams()]);
    const providerId = tier.value?.binding?.providerId || undefined;
    await refreshModelOptions(Number(providerId || 0), false);
    await formApi.setValues({
      enabled: tier.value?.enabled ?? 0,
      defaultEffort: tier.value?.defaultEffort || "",
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
      let defaultParamsJson = "{}";
      try {
        defaultParamsJson = normalizeDefaultParamsJson(
          methodDefaultParamsJson.value,
        );
      } catch {
        defaultParamsInvalid.value = true;
        message.error(
          $t("plugin.linapro-ai-core.methodDefault.messages.invalidJson"),
        );
        return;
      }
      await Promise.all([
        tierUpdate(tier.value.code, values),
        methodDefaultUpdate(values.capabilityType, values.capabilityMethod, {
          defaultParamsJson,
          enabled: 1,
        }),
      ]);
      message.success($t("pages.common.updateSuccess"));
      emit("reload");
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
      <Form>
        <template #defaultParamsJson>
          <div class="tier-default-params">
            <div class="tier-default-params__scope">
              {{ methodDefaultScope }}
            </div>
            <JsonHighlightEditor
              v-model="methodDefaultParamsJson"
              :invalid="defaultParamsInvalid"
              :placeholder="
                $t('plugin.linapro-ai-core.methodDefault.placeholders.json')
              "
              testid="ai-tier-default-params-editor"
              @update:model-value="defaultParamsInvalid = false"
            />
          </div>
        </template>
      </Form>
      <div class="flex justify-end">
        <Space>
          <a-button :disabled="testing" :loading="testing" @click="handleTest">
            {{ $t("plugin.linapro-ai-core.tier.actions.testDraft") }}
          </a-button>
        </Space>
      </div>
    </div>
  </Drawer>
</template>

<style scoped>
.tier-default-params {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 8px;
}

.tier-default-params__scope {
  color: hsl(var(--muted-foreground));
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono",
    "Courier New", monospace;
  font-size: 12px;
  line-height: 20px;
}
</style>
