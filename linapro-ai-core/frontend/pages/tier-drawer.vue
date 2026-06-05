<script setup lang="ts">
import type { Model, Provider, Tier, TierTestResult } from "./ai-client";
import type { VbenFormSchema } from "#/adapter/form";

import { computed, ref } from "vue";

import { useVbenDrawer } from "@vben/common-ui";

import { message, Space } from "ant-design-vue";

import { useVbenForm } from "#/adapter/form";
import { $t } from "#/locales";
import {
  methodDefaultList,
  methodDefaultUpdate,
  providerList,
  providerModels,
  tierTest,
  tierUpdate,
} from "./ai-client";
import {
  buildEffortOptions,
  buildEnabledOptions,
  tierDisplayName,
} from "./ai-data";
import JsonHighlightEditor from "./json-highlight-editor.vue";

const emit = defineEmits<{ reload: [] }>();

const tier = ref<Tier>();
const providers = ref<Provider[]>([]);
const models = ref<Model[]>([]);
const testing = ref(false);
const currentTestResult = ref<TierTestResult>();
const defaultParamsJson = ref("{}");
const defaultParamsCompact = ref("{}");
const defaultParamsInvalid = ref(false);
const title = computed(tierDrawerTitle);
const currentTestLatency = computed(() =>
  formatLatencyMs(currentTestResult.value?.latencyMs),
);

function tierDrawerTitle() {
  return $t("plugin.linapro-ai-core.tier.drawer.editTitle", {
    name: tierDisplayName(tier.value),
  });
}

function modelLabel(model: Model) {
  return model.modelName;
}

function modelProtocolLabel(protocol: string) {
  return protocol?.includes("anthropic")
    ? "Anthropic"
    : protocol?.includes("voyage")
      ? "Voyage"
      : "OpenAI";
}

function modelProtocolGroupKey(protocol: string) {
  return protocol?.includes("anthropic")
    ? "anthropic"
    : protocol?.includes("voyage")
      ? "voyage"
      : "openai";
}

function buildModelOptionGroups(items: Model[]) {
  const groups = new Map<string, Array<{ label: string; value: number }>>();
  for (const item of items) {
    const key = modelProtocolGroupKey(item.protocol);
    const group = groups.get(key) || [];
    group.push({ label: modelLabel(item), value: item.id });
    groups.set(key, group);
  }
  const order = ["openai", "anthropic", "voyage"];
  return [
    ...order.filter((key) => groups.has(key)),
    ...[...groups.keys()].filter((key) => !order.includes(key)),
  ].map((key) => ({
    label: modelProtocolLabel(key),
    options: groups.get(key) || [],
  }));
}

function formatLatencyMs(value: number | undefined) {
  return `${Math.max(0, Math.round(Number(value || 0)))}ms`;
}

function resultMessage(result: TierTestResult, fallbackKey: string) {
  const text =
    result.errorSummary ||
    $t(fallbackKey) ||
    $t("plugin.linapro-ai-core.tier.messages.testFailed");
  return `${text} (${formatLatencyMs(result.latencyMs)})`;
}

function supportsThinkingEffort() {
  return (
    tier.value?.capabilityType === "text" &&
    tier.value?.capabilityMethod === "generate"
  );
}

function formatDefaultParams(raw: string) {
  try {
    return JSON.stringify(JSON.parse(raw || "{}"), null, 2);
  } catch {
    return raw || "{}";
  }
}

function normalizeDefaultParams(raw: string) {
  try {
    const parsed = JSON.parse(raw || "{}");
    if (
      parsed === null ||
      Array.isArray(parsed) ||
      typeof parsed !== "object"
    ) {
      return "";
    }
    return JSON.stringify(parsed);
  } catch {
    return "";
  }
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
  models.value = providerId ? await providerModels(providerId, 1) : [];
  formApi.updateSchema([
    {
      fieldName: "modelId",
      componentProps: {
        optionFilterProp: "label",
        options: buildModelOptionGroups(models.value),
        showSearch: true,
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

async function refreshMethodDefault() {
  const capabilityType = tier.value?.capabilityType || "text";
  const capabilityMethod = tier.value?.capabilityMethod || "generate";
  const items = await methodDefaultList();
  const current = items.find(
    (item) =>
      item.capabilityType === capabilityType &&
      item.capabilityMethod === capabilityMethod,
  );
  defaultParamsCompact.value =
    normalizeDefaultParams(current?.defaultParamsJson || "{}") || "{}";
  defaultParamsJson.value = formatDefaultParams(defaultParamsCompact.value);
  defaultParamsInvalid.value = false;
}

async function saveMethodDefaultIfChanged() {
  const normalized = normalizeDefaultParams(defaultParamsJson.value);
  if (!normalized) {
    defaultParamsInvalid.value = true;
    message.error(
      $t("plugin.linapro-ai-core.methodDefault.messages.invalidJson"),
    );
    return false;
  }
  defaultParamsInvalid.value = false;
  if (normalized === defaultParamsCompact.value) {
    return true;
  }
  await methodDefaultUpdate(
    tier.value?.capabilityType || "text",
    tier.value?.capabilityMethod || "generate",
    normalized,
  );
  defaultParamsCompact.value = normalized;
  defaultParamsJson.value = formatDefaultParams(normalized);
  return true;
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
    currentTestResult.value = result;
    if (result.status === "success") {
      message.success(
        resultMessage(
          result,
          "plugin.linapro-ai-core.tier.messages.testSuccess",
        ),
      );
    } else {
      message.error(
        resultMessage(
          result,
          "plugin.linapro-ai-core.tier.messages.testFailed",
        ),
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
    currentTestResult.value = undefined;
    formApi.updateSchema([
      {
        fieldName: "defaultEffort",
        hide: !supportsThinkingEffort(),
      },
    ]);
    await formApi.resetForm();
    await refreshProviderOptions();
    await refreshMethodDefault();
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
      if (!(await saveMethodDefaultIfChanged())) {
        return;
      }
      await tierUpdate(tier.value.code, values);
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
      <Form />
      <div class="tier-default-params-item relative flex">
        <label class="tier-default-params-label">
          {{
            $t("plugin.linapro-ai-core.methodDefault.fields.defaultParamsJson")
          }}
        </label>
        <div class="tier-default-params-control flex-auto">
          <JsonHighlightEditor
            v-model="defaultParamsJson"
            :invalid="defaultParamsInvalid"
            :placeholder="
              $t('plugin.linapro-ai-core.methodDefault.placeholders.json')
            "
            testid="ai-tier-default-params-editor"
            @update:model-value="defaultParamsInvalid = false"
          />
        </div>
      </div>
      <div class="flex justify-end">
        <Space>
          <div
            v-if="currentTestResult"
            class="tier-current-test-result"
            data-testid="ai-tier-current-test-result"
          >
            <span>{{
              $t("plugin.linapro-ai-core.invocation.fields.latencyMs")
            }}</span>
            <span>{{ currentTestLatency }}</span>
          </div>
          <a-button :disabled="testing" :loading="testing" @click="handleTest">
            {{ $t("plugin.linapro-ai-core.tier.actions.testDraft") }}
          </a-button>
        </Space>
      </div>
    </div>
  </Drawer>
</template>

<style scoped>
.tier-current-test-result {
  align-items: center;
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  color: hsl(var(--muted-foreground));
  display: inline-flex;
  font-size: 12px;
  gap: 8px;
  line-height: 22px;
  padding: 0 10px;
}

.tier-current-test-result span:last-child {
  color: hsl(var(--foreground));
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono",
    "Courier New", monospace;
}

.tier-default-params-item {
  align-items: flex-start;
  column-gap: 8px;
  width: 100%;
}

.tier-default-params-label {
  flex: 0 0 132px;
  width: 132px;
  min-height: 32px;
  padding-right: 12px;
  color: hsl(var(--foreground));
  font-size: 14px;
  line-height: 32px;
  text-align: right;
}

.tier-default-params-control {
  min-width: 0;
}
</style>
