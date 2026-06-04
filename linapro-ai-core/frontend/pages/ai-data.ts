import type { VbenFormSchema } from "#/adapter/form";
import type { VxeGridProps } from "#/adapter/vxe-table";
import type { Provider, ProviderModelSummary, Tier } from "./ai-client";
import type { Component } from "vue";

import { h } from "vue";

import { Popconfirm, Tag } from "ant-design-vue";

import { $t } from "#/locales";
import { formatTimestamp } from "#/utils/time";

function statusTag(value: number | string) {
  const enabled = Number(value) === 1 || value === "success";
  const failed = value === "failed";
  const color = failed ? "error" : enabled ? "success" : "default";
  const label =
    value === "success"
      ? $t("plugin.linapro-ai-core.common.success")
      : value === "failed"
        ? $t("plugin.linapro-ai-core.common.failed")
        : Number(value) === 1
          ? $t("plugin.linapro-ai-core.common.enabled")
          : $t("plugin.linapro-ai-core.common.disabled");
  return h(Tag, { color }, () => label);
}

export function buildEnabledOptions() {
  return [
    { label: $t("plugin.linapro-ai-core.common.enabled"), value: 1 },
    { label: $t("plugin.linapro-ai-core.common.disabled"), value: 0 },
  ];
}

export const protocolOptions = [
  { label: "OpenAI", value: "openai" },
  { label: "Anthropic", value: "anthropic" },
];

export const endpointProtocolOptions = protocolOptions;

const protocolDisplayLabels: Record<string, string> = {
  anthropic: "Anthropic",
  "anthropic-compatible": "Anthropic",
  openai: "OpenAI",
  "openai-compatible": "OpenAI",
  voyage: "Voyage",
};

function protocolLabel(value: string) {
  return protocolDisplayLabels[value] || value || "OpenAI";
}

function protocolBadgeMeta(value: string) {
  if (value === "anthropic" || value === "anthropic-compatible") {
    return {
      icon: "simple-icons:anthropic",
      iconClass: "text-orange-700 dark:text-orange-100",
      styleClass:
        "border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-400/30 dark:bg-orange-500/15 dark:text-orange-200",
      type: "anthropic",
    };
  }
  if (value === "voyage") {
    return {
      icon: "simple-icons:voyage",
      iconClass: "text-cyan-700 dark:text-cyan-100",
      styleClass:
        "border-cyan-200 bg-cyan-50 text-cyan-700 dark:border-cyan-400/30 dark:bg-cyan-500/15 dark:text-cyan-200",
      type: "voyage",
    };
  }
  return {
    icon: "simple-icons:openai",
    iconClass: "text-emerald-700 dark:text-emerald-100",
    styleClass:
      "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-400/30 dark:bg-emerald-500/15 dark:text-emerald-200",
    type: "openai",
  };
}

function formatOptionalTimestamp(value: null | number | string | undefined) {
  if (value === 0 || value === "0") {
    return "";
  }
  return formatTimestamp(value, "");
}

function formatLatencyMs(value: unknown) {
  return `${Math.max(0, Math.round(Number(value || 0)))}ms`;
}

function externalHref(url: string) {
  const trimmed = url.trim();
  if (!trimmed) {
    return "";
  }
  return /^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`;
}

function providerNameCell(row: Provider) {
  const websiteUrl = row.websiteUrl?.trim();
  const children = [
    h(
      "span",
      { class: "truncate font-medium text-foreground" },
      row.name || "-",
    ),
  ];
  if (websiteUrl) {
    children.push(
      h(
        "a",
        {
          class:
            "min-w-0 break-all text-xs leading-4 text-primary hover:underline",
          href: externalHref(websiteUrl),
          onClick: (event: MouseEvent) => event.stopPropagation(),
          rel: "noopener noreferrer",
          target: "_blank",
        },
        websiteUrl,
      ),
    );
  }
  return h("div", { class: "flex min-w-0 flex-col gap-0.5 py-1" }, children);
}

function endpointRows(row: Provider) {
  const endpoints = (row.endpoints || []).filter((item) => item.baseUrl);
  return endpoints.map((endpoint) => ({
    enabled: endpoint.enabled,
    label: protocolLabel(endpoint.protocol),
    secretRef: endpoint.secretRef,
    type: endpoint.protocol,
    url: endpoint.baseUrl,
  }));
}

function endpointCell(row: Provider, providerIcon?: Component) {
  const endpoints = endpointRows(row);
  if (endpoints.length === 0) {
    return h(
      "span",
      { class: "text-muted-foreground text-xs" },
      $t("plugin.linapro-ai-core.provider.empty.noEndpoint"),
    );
  }
  return h(
    "div",
    { class: "flex min-w-0 flex-col gap-1.5 py-1" },
    endpoints.map((endpoint) => {
      const badgeMeta = protocolBadgeMeta(endpoint.type);
      return h(
        "div",
        {
          class:
            "ai-provider-endpoint-item relative min-w-0 rounded border border-border bg-muted/30 px-2 py-1.5 text-left",
        },
        [
          h(
            "span",
            {
              class:
                "ai-provider-endpoint-url block min-w-0 break-all pr-[108px] text-left font-mono text-xs leading-5 whitespace-normal",
              title: endpoint.url,
            },
            endpoint.url,
          ),
          h(
            "span",
            {
              class: [
                "ai-provider-endpoint-badge absolute right-3 top-1.5 inline-flex h-5 max-w-[92px] items-center gap-1 rounded border px-1.5 text-[9px] font-medium leading-none shadow-sm",
                badgeMeta.styleClass,
              ].join(" "),
              "data-protocol": endpoint.type,
              "data-provider-icon": badgeMeta.type,
            },
            [
              h(
                "span",
                {
                  "aria-hidden": true,
                  class: [
                    "ai-provider-endpoint-icon-mark inline-flex size-3 shrink-0 items-center justify-center leading-none",
                    badgeMeta.iconClass,
                  ].join(" "),
                  "data-provider-icon": badgeMeta.type,
                },
                providerIcon
                  ? h(providerIcon, { class: "size-3", icon: badgeMeta.icon })
                  : undefined,
              ),
              h("span", { class: "truncate" }, endpoint.label),
            ],
          ),
          endpoint.enabled === 0
            ? h(Tag, { class: "!m-0 shrink-0", color: "default" }, () =>
                $t("plugin.linapro-ai-core.common.disabled"),
              )
            : undefined,
        ],
      );
    }),
  );
}

function secretCell(row: Provider) {
  const secrets = (row.endpoints || [])
    .map((endpoint) => endpoint.secretRef)
    .filter(Boolean);
  const text =
    secrets.length > 0
      ? [...new Set(secrets)].join("\n")
      : $t("plugin.linapro-ai-core.provider.empty.noKey");
  return h(
    "span",
    {
      class:
        secrets.length > 0
          ? "font-mono text-xs text-foreground"
          : "text-muted-foreground text-xs",
      style: "white-space: pre-line",
    },
    text,
  );
}

type ProviderColumnOptions = {
  onDeleteModel?: (model: ProviderModelSummary) => Promise<void> | void;
  providerIcon?: Component;
};

function modelDeleteButton(
  model: ProviderModelSummary,
  onDeleteModel?: ProviderColumnOptions["onDeleteModel"],
) {
  if (!onDeleteModel) {
    return undefined;
  }
  const label = `${$t("pages.common.delete")} ${model.modelName}`;
  return h(
    Popconfirm,
    {
      onConfirm: () => onDeleteModel(model),
      placement: "top",
      title: $t("pages.common.deleteConfirm"),
    },
    {
      default: () =>
        h(
          "button",
          {
            "aria-label": label,
            class:
              "inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
            onClick: (event: MouseEvent) => event.stopPropagation(),
            title: label,
            type: "button",
          },
          h("span", {
            "aria-hidden": true,
            class: "ai-model-delete-icon",
          }),
        ),
    },
  );
}

function modelTag(
  model: ProviderModelSummary,
  onDeleteModel?: ProviderColumnOptions["onDeleteModel"],
) {
  return h(
    "span",
    {
      class:
        "ai-provider-model-tag inline-flex min-h-8 max-w-full items-start gap-1.5 rounded-full border border-border bg-background px-2.5 py-1.5 text-sm shadow-sm",
    },
    [
      h(
        "span",
        {
          class:
            "ai-provider-model-name min-w-0 break-all text-foreground whitespace-normal leading-5",
        },
        model.modelName,
      ),
      h(
        "span",
        { class: "shrink-0 text-xs text-muted-foreground leading-5" },
        protocolLabel(model.protocol),
      ),
      modelDeleteButton(model, onDeleteModel),
    ].filter(Boolean),
  );
}

function groupModels(models: ProviderModelSummary[]) {
  const order = [
    "openai",
    "openai-compatible",
    "anthropic",
    "anthropic-compatible",
    "voyage",
  ];
  const groups = new Map<string, ProviderModelSummary[]>();
  for (const model of models) {
    const key = model.protocol || "openai";
    const current = groups.get(key) || [];
    current.push(model);
    groups.set(key, current);
  }
  return [
    ...order.filter((key) => groups.has(key)),
    ...[...groups.keys()].filter((key) => !order.includes(key)),
  ].map((key) => ({ key, models: groups.get(key) || [] }));
}

function modelsCell(
  row: Provider,
  onDeleteModel?: ProviderColumnOptions["onDeleteModel"],
) {
  const models = row.models || [];
  if (models.length === 0) {
    return h(
      "span",
      { class: "text-muted-foreground text-xs" },
      $t("plugin.linapro-ai-core.provider.empty.noModels"),
    );
  }
  const groups = groupModels(models);
  return h(
    "div",
    {
      class:
        "ai-provider-model-list flex min-h-[48px] min-w-0 max-w-full flex-col justify-center gap-2 overflow-visible py-2",
    },
    groups.map((group, index) =>
      h(
        "div",
        {
          class: [
            "ai-provider-model-row flex min-w-0 flex-wrap gap-2 overflow-visible",
            index > 0 ? "border-t border-border pt-2" : "",
          ]
            .filter(Boolean)
            .join(" "),
        },
        group.models.map((model) => modelTag(model, onDeleteModel)),
      ),
    ),
  );
}

export function buildCapabilityMethodOptions() {
  return [
    { label: "text.generate", value: "text.generate" },
    { label: "image.generate", value: "image.generate" },
    { label: "image.edit", value: "image.edit" },
    { label: "embedding.create", value: "embedding.create" },
    { label: "audio.transcribe", value: "audio.transcribe" },
    { label: "audio.synthesize", value: "audio.synthesize" },
    { label: "vision.analyze", value: "vision.analyze" },
    { label: "document.analyze", value: "document.analyze" },
    { label: "document.cite", value: "document.cite" },
    { label: "safety.moderate", value: "safety.moderate" },
    { label: "video.generate", value: "video.generate" },
    { label: "video.edit", value: "video.edit" },
    { label: "video.extend", value: "video.extend" },
    { label: "video.operation.get", value: "video.operation.get" },
    { label: "video.operation.cancel", value: "video.operation.cancel" },
  ];
}

export const tierCapabilityTypeKeys = [
  "text",
  "image",
  "embedding",
  "audio",
  "vision",
  "document",
  "safety",
  "video",
] as const;

const tierCapabilityDefaultMethods: Record<string, string> = {
  audio: "audio.transcribe",
  document: "document.analyze",
  embedding: "embedding.create",
  image: "image.generate",
  safety: "safety.moderate",
  text: "text.generate",
  video: "video.generate",
  vision: "vision.analyze",
};

function titleCaseCapabilityType(type: string) {
  return type ? `${type.charAt(0).toUpperCase()}${type.slice(1)}` : "Text";
}

export function capabilityTypeLabel(type = "text") {
  const normalized = type || "text";
  const key = `plugin.linapro-ai-core.capability.types.${normalized}`;
  const label = $t(key);
  return label && label !== key ? label : titleCaseCapabilityType(normalized);
}

export function defaultTierCapabilityMethod(type = "text") {
  return (
    tierCapabilityDefaultMethods[type] || tierCapabilityDefaultMethods.text
  );
}

export function splitCapabilityMethod(value = "text.generate") {
  const [capabilityType = "text", ...methodParts] = String(
    value || "text.generate",
  ).split(".");
  return {
    capabilityMethod: methodParts.join(".") || "generate",
    capabilityType: capabilityType || "text",
  };
}

export function joinCapabilityMethod(type = "text", method = "generate") {
  return `${type || "text"}.${method || "generate"}`;
}

export function buildCapabilityQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: "Select",
      fieldName: "capabilityKey",
      label: $t("plugin.linapro-ai-core.capability.method"),
      defaultValue: "text.generate",
      componentProps: {
        options: buildCapabilityMethodOptions(),
      },
    },
  ];
}

export function buildEffortOptions() {
  return [
    { label: $t("plugin.linapro-ai-core.effort.empty"), value: "" },
    { label: "low", value: "low" },
    { label: "medium", value: "medium" },
    { label: "high", value: "high" },
    { label: "xhigh", value: "xhigh" },
    { label: "max", value: "max" },
  ];
}

export function tierDisplayName(
  tier: Pick<Tier, "code" | "displayName"> | undefined,
) {
  const code = tier?.code?.trim();
  if (!code) {
    return tier?.displayName || "";
  }
  const key = `plugin.linapro-ai-core.tier.names.${code}`;
  const label = $t(key);
  return label && label !== key ? label : tier?.displayName || code;
}

export function tierCodeLabel(code: string) {
  return tierDisplayName({ code, displayName: code });
}

export function buildProviderQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: "Input",
      fieldName: "keyword",
      label: $t("plugin.linapro-ai-core.provider.fields.keyword"),
    },
    {
      component: "Select",
      fieldName: "enabled",
      label: $t("pages.common.status"),
      componentProps: { options: buildEnabledOptions() },
    },
  ];
}

export function buildProviderColumns(
  options: ProviderColumnOptions = {},
): VxeGridProps["columns"] {
  return [
    {
      field: "name",
      title: $t("plugin.linapro-ai-core.provider.fields.name"),
      minWidth: 200,
      showOverflow: false,
      slots: { default: ({ row }) => providerNameCell(row as Provider) },
    },
    {
      field: "models",
      title: $t("plugin.linapro-ai-core.provider.fields.models"),
      className: "ai-provider-model-column",
      minWidth: 300,
      showOverflow: false,
      slots: {
        default: ({ row }) =>
          modelsCell(row as Provider, options.onDeleteModel),
      },
    },
    {
      field: "endpoint",
      title: $t("plugin.linapro-ai-core.provider.fields.endpoint"),
      className: "ai-provider-endpoint-column",
      minWidth: 300,
      showOverflow: false,
      slots: {
        default: ({ row }) =>
          endpointCell(row as Provider, options.providerIcon),
      },
    },
    {
      field: "enabled",
      title: $t("pages.common.status"),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.enabled) },
    },
    {
      field: "endpointSecrets",
      title: $t("plugin.linapro-ai-core.provider.fields.secret"),
      minWidth: 160,
      slots: { default: ({ row }) => secretCell(row as Provider) },
    },
    {
      field: "updatedAt",
      title: $t("pages.common.updatedAt"),
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: "action",
      className: "ai-provider-action-column",
      fixed: "right",
      showOverflow: false,
      resizable: false,
      slots: { default: "action" },
      title: $t("pages.common.actions"),
      width: 190,
    },
  ];
}

export function buildProviderFormSchema(): VbenFormSchema[] {
  return [
    {
      component: "Input",
      fieldName: "name",
      label: $t("plugin.linapro-ai-core.provider.fields.name"),
      rules: "required",
    },
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
      component: "Input",
      fieldName: "websiteUrl",
      label: $t("plugin.linapro-ai-core.provider.fields.websiteUrl"),
    },
    {
      component: "InputPassword",
      fieldName: "secretRef",
      label: $t("plugin.linapro-ai-core.endpoint.fields.secretRef"),
      componentProps: {
        autocomplete: "new-password",
        placeholder: $t(
          "plugin.linapro-ai-core.provider.placeholders.apiKeyCreate",
        ),
      },
    },
    {
      component: "Input",
      fieldName: "openaiBaseUrl",
      label: `${$t("plugin.linapro-ai-core.endpoint.names.openai")} ${$t("plugin.linapro-ai-core.endpoint.fields.baseUrl")}`,
      componentProps: {
        placeholder: "https://api.openai.com/v1",
      },
    },
    {
      component: "Input",
      fieldName: "anthropicBaseUrl",
      label: `${$t("plugin.linapro-ai-core.endpoint.names.anthropic")} ${$t("plugin.linapro-ai-core.endpoint.fields.baseUrl")}`,
      defaultValue: "https://api.anthropic.com/v1",
      componentProps: {
        placeholder: "https://api.anthropic.com/v1",
      },
    },
    {
      component: "Textarea",
      fieldName: "remark",
      label: $t("pages.common.remark"),
      componentProps: { rows: 3 },
    },
  ];
}

export function buildEndpointFormSchema(): VbenFormSchema[] {
  return [
    {
      component: "Select",
      fieldName: "protocol",
      label: $t("plugin.linapro-ai-core.endpoint.fields.protocol"),
      rules: "selectRequired",
      componentProps: { options: endpointProtocolOptions },
    },
    {
      component: "Input",
      fieldName: "baseUrl",
      label: $t("plugin.linapro-ai-core.endpoint.fields.baseUrl"),
      rules: "required",
    },
    {
      component: "InputPassword",
      fieldName: "secretRef",
      label: $t("plugin.linapro-ai-core.endpoint.fields.secretRef"),
      componentProps: {
        autocomplete: "new-password",
      },
    },
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
      component: "Textarea",
      fieldName: "metadataJson",
      label: $t("plugin.linapro-ai-core.endpoint.fields.metadataJson"),
      componentProps: { rows: 3 },
    },
  ];
}

export function buildModelFormSchema(
  providerOptions: Array<{ label: string; value: number }> = [],
  endpointOptions: Array<{ label: string; value: number }> = [],
): VbenFormSchema[] {
  return [
    {
      component: "Select",
      fieldName: "providerId",
      label: $t("plugin.linapro-ai-core.model.fields.provider"),
      rules: "selectRequired",
      componentProps: {
        options: providerOptions,
        showSearch: true,
      },
    },
    {
      component: "Select",
      fieldName: "endpointIds",
      label: `${$t("plugin.linapro-ai-core.model.fields.endpoint")} / ${$t("plugin.linapro-ai-core.model.fields.protocol")}`,
      rules: "selectRequired",
      componentProps: {
        allowClear: false,
        maxTagCount: "responsive",
        mode: "multiple",
        optionFilterProp: "label",
        options: endpointOptions,
        showSearch: true,
      },
    },
    {
      component: "Select",
      fieldName: "capabilityKey",
      label: $t("plugin.linapro-ai-core.capability.method"),
      rules: "selectRequired",
      componentProps: {
        options: buildCapabilityMethodOptions(),
      },
    },
    {
      component: "Input",
      fieldName: "modelName",
      label: $t("plugin.linapro-ai-core.model.fields.modelName"),
      rules: "required",
    },
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
      component: "RadioGroup",
      fieldName: "supportsThinking",
      label: $t("plugin.linapro-ai-core.model.fields.supportsThinking"),
      defaultValue: 0,
      componentProps: {
        buttonStyle: "solid",
        optionType: "button",
        options: buildEnabledOptions(),
      },
    },
    {
      component: "Select",
      fieldName: "supportedEfforts",
      label: $t("plugin.linapro-ai-core.model.fields.supportedEfforts"),
      componentProps: {
        mode: "multiple",
        options: buildEffortOptions().filter((item) => item.value),
      },
      dependencies: {
        show: (values) => Number(values.supportsThinking || 0) === 1,
        trigger: (values, actions) => {
          if (Number(values.supportsThinking || 0) === 1) {
            return;
          }
          if ((values.supportedEfforts || []).length > 0) {
            actions.setFieldValue("supportedEfforts", []);
          }
        },
        triggerFields: ["supportsThinking"],
      },
    },
    {
      component: "InputNumber",
      fieldName: "maxInputTokens",
      label: $t("plugin.linapro-ai-core.model.fields.maxInputTokens"),
    },
    {
      component: "InputNumber",
      fieldName: "maxOutputTokens",
      label: $t("plugin.linapro-ai-core.model.fields.maxOutputTokens"),
    },
  ];
}

export function buildTierColumns(): VxeGridProps["columns"] {
  return [
    {
      field: "displayName",
      title: $t("plugin.linapro-ai-core.tier.fields.displayName"),
      formatter: ({ row }) => tierDisplayName(row),
      minWidth: 130,
    },
    {
      field: "enabled",
      title: $t("pages.common.status"),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.enabled) },
    },
    {
      field: "binding.providerName",
      title: $t("plugin.linapro-ai-core.tier.fields.provider"),
      minWidth: 150,
    },
    {
      field: "binding.modelName",
      title: $t("plugin.linapro-ai-core.tier.fields.model"),
      minWidth: 180,
    },
    {
      field: "binding.protocol",
      title: $t("plugin.linapro-ai-core.model.fields.protocol"),
      minWidth: 100,
    },
    {
      field: "lastTestStatus",
      title: $t("plugin.linapro-ai-core.tier.fields.lastTestStatus"),
      minWidth: 120,
      slots: {
        default: ({ row }) =>
          row.lastTestStatus
            ? h(
                "div",
                { class: "flex items-center gap-2" },
                [
                  statusTag(row.lastTestStatus),
                  h(
                    "span",
                    { class: "font-mono text-xs text-muted-foreground" },
                    formatLatencyMs(row.lastTestLatencyMs),
                  ),
                ],
              )
            : "-",
      },
    },
    {
      field: "updatedAt",
      title: $t("pages.common.updatedAt"),
      formatter: ({ cellValue }) => formatOptionalTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: "action",
      fixed: "right",
      resizable: false,
      slots: { default: "action" },
      title: $t("pages.common.actions"),
      width: 190,
    },
  ];
}

export function buildTierQuerySchema(): VbenFormSchema[] {
  return buildCapabilityQuerySchema();
}

export function buildInvocationQuerySchema(): VbenFormSchema[] {
  return [
    ...buildCapabilityQuerySchema(),
    {
      component: "Input",
      fieldName: "purpose",
      label: $t("plugin.linapro-ai-core.invocation.fields.purpose"),
    },
    {
      component: "Select",
      fieldName: "tierCode",
      label: $t("plugin.linapro-ai-core.invocation.fields.tierCode"),
      componentProps: {
        options: ["basic", "standard", "advanced"].map((value) => ({
          label: tierCodeLabel(value),
          value,
        })),
      },
    },
    {
      component: "Select",
      fieldName: "status",
      label: $t("plugin.linapro-ai-core.invocation.fields.status"),
      componentProps: {
        options: [
          {
            label: $t("plugin.linapro-ai-core.common.success"),
            value: "success",
          },
          {
            label: $t("plugin.linapro-ai-core.common.failed"),
            value: "failed",
          },
        ],
      },
    },
    {
      component: "Input",
      fieldName: "sourcePluginId",
      label: $t("plugin.linapro-ai-core.invocation.fields.sourcePluginId"),
    },
  ];
}

export function buildInvocationColumns(): VxeGridProps["columns"] {
  return [
    {
      field: "createdAt",
      title: $t("pages.common.createdAt"),
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: "purpose",
      title: $t("plugin.linapro-ai-core.invocation.fields.purpose"),
      minWidth: 180,
    },
    {
      field: "capabilityType",
      title: $t("plugin.linapro-ai-core.capability.method"),
      formatter: ({ row }) =>
        joinCapabilityMethod(row.capabilityType, row.capabilityMethod),
      minWidth: 150,
    },
    {
      field: "tierCode",
      title: $t("plugin.linapro-ai-core.invocation.fields.tierCode"),
      formatter: ({ cellValue }) => tierCodeLabel(String(cellValue || "")),
      minWidth: 100,
    },
    {
      field: "status",
      title: $t("plugin.linapro-ai-core.invocation.fields.status"),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.status) },
    },
    {
      field: "providerName",
      title: $t("plugin.linapro-ai-core.invocation.fields.providerName"),
      minWidth: 150,
    },
    {
      field: "modelName",
      title: $t("plugin.linapro-ai-core.invocation.fields.modelName"),
      minWidth: 180,
    },
    {
      field: "latencyMs",
      title: $t("plugin.linapro-ai-core.invocation.fields.latencyMs"),
      minWidth: 110,
    },
    {
      field: "assetSummaryJson",
      title: $t("plugin.linapro-ai-core.invocation.fields.assetSummaryJson"),
      minWidth: 180,
      showOverflow: true,
    },
    {
      field: "inputTokens",
      title: $t("plugin.linapro-ai-core.invocation.fields.inputTokens"),
      minWidth: 110,
    },
    {
      field: "outputTokens",
      title: $t("plugin.linapro-ai-core.invocation.fields.outputTokens"),
      minWidth: 120,
    },
    {
      field: "action",
      fixed: "right",
      resizable: false,
      slots: { default: "action" },
      title: $t("pages.common.actions"),
      width: 100,
    },
  ];
}
