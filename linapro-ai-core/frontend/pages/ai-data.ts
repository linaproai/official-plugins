import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Tier } from './ai-client';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';

function statusTag(value: number | string) {
  const enabled = Number(value) === 1 || value === 'success';
  const failed = value === 'failed';
  const color = failed ? 'error' : enabled ? 'success' : 'default';
  const label =
    value === 'success'
      ? $t('plugin.linapro-ai-core.common.success')
      : value === 'failed'
        ? $t('plugin.linapro-ai-core.common.failed')
        : Number(value) === 1
          ? $t('plugin.linapro-ai-core.common.enabled')
          : $t('plugin.linapro-ai-core.common.disabled');
  return h(Tag, { color }, () => label);
}

export function buildEnabledOptions() {
  return [
    { label: $t('plugin.linapro-ai-core.common.enabled'), value: 1 },
    { label: $t('plugin.linapro-ai-core.common.disabled'), value: 0 },
  ];
}

export const protocolOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Anthropic', value: 'anthropic' },
];

export function buildEffortOptions() {
  return [
    { label: $t('plugin.linapro-ai-core.effort.empty'), value: '' },
    { label: 'low', value: 'low' },
    { label: 'medium', value: 'medium' },
    { label: 'high', value: 'high' },
    { label: 'xhigh', value: 'xhigh' },
    { label: 'max', value: 'max' },
  ];
}

export function tierDisplayName(tier: Pick<Tier, 'code' | 'displayName'> | undefined) {
  const code = tier?.code?.trim();
  if (!code) {
    return tier?.displayName || '';
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
      component: 'Input',
      fieldName: 'keyword',
      label: $t('plugin.linapro-ai-core.provider.fields.keyword'),
    },
    {
      component: 'Select',
      fieldName: 'enabled',
      label: $t('pages.common.status'),
      componentProps: { options: buildEnabledOptions() },
    },
  ];
}

export function buildProviderColumns(): VxeGridProps['columns'] {
  return [
    {
      field: 'name',
      title: $t('plugin.linapro-ai-core.provider.fields.name'),
      minWidth: 160,
    },
    {
      field: 'enabled',
      title: $t('pages.common.status'),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.enabled) },
    },
    {
      field: 'modelCount',
      title: $t('plugin.linapro-ai-core.provider.fields.modelCount'),
      minWidth: 120,
    },
    {
      field: 'enabledModelCount',
      title: $t('plugin.linapro-ai-core.provider.fields.enabledModelCount'),
      minWidth: 140,
    },
    {
      field: 'openaiBaseUrl',
      title: $t('plugin.linapro-ai-core.provider.fields.openaiBaseUrl'),
      minWidth: 220,
    },
    {
      field: 'anthropicBaseUrl',
      title: $t('plugin.linapro-ai-core.provider.fields.anthropicBaseUrl'),
      minWidth: 220,
    },
    {
      field: 'updatedAt',
      title: $t('pages.common.updatedAt'),
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: 'action',
      fixed: 'right',
      resizable: false,
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      width: 180,
    },
  ];
}

export function buildProviderFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('plugin.linapro-ai-core.provider.fields.name'),
      rules: 'required',
    },
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
      component: 'Input',
      fieldName: 'websiteUrl',
      label: $t('plugin.linapro-ai-core.provider.fields.websiteUrl'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'Input',
      fieldName: 'openaiBaseUrl',
      label: $t('plugin.linapro-ai-core.provider.fields.openaiBaseUrl'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'Input',
      fieldName: 'anthropicBaseUrl',
      label: $t('plugin.linapro-ai-core.provider.fields.anthropicBaseUrl'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'InputPassword',
      fieldName: 'apiKeySecretRef',
      label: $t('plugin.linapro-ai-core.provider.fields.apiKeySecretRef'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('pages.common.remark'),
      formItemClass: 'col-span-2',
      componentProps: { rows: 3 },
    },
  ];
}

export function buildModelFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'modelName',
      label: $t('plugin.linapro-ai-core.model.fields.modelName'),
      rules: 'required',
    },
    {
      component: 'Select',
      fieldName: 'protocol',
      label: $t('plugin.linapro-ai-core.model.fields.protocol'),
      rules: 'selectRequired',
      componentProps: { options: protocolOptions },
    },
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
      component: 'RadioGroup',
      fieldName: 'supportsThinking',
      label: $t('plugin.linapro-ai-core.model.fields.supportsThinking'),
      defaultValue: 0,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: buildEnabledOptions(),
      },
    },
    {
      component: 'Select',
      fieldName: 'supportedEfforts',
      label: $t('plugin.linapro-ai-core.model.fields.supportedEfforts'),
      formItemClass: 'col-span-2',
      componentProps: {
        mode: 'multiple',
        options: buildEffortOptions().filter((item) => item.value),
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'maxInputTokens',
      label: $t('plugin.linapro-ai-core.model.fields.maxInputTokens'),
    },
    {
      component: 'InputNumber',
      fieldName: 'maxOutputTokens',
      label: $t('plugin.linapro-ai-core.model.fields.maxOutputTokens'),
    },
  ];
}

export function buildTierColumns(): VxeGridProps['columns'] {
  return [
    {
      field: 'displayName',
      title: $t('plugin.linapro-ai-core.tier.fields.displayName'),
      formatter: ({ row }) => tierDisplayName(row),
      minWidth: 130,
    },
    {
      field: 'enabled',
      title: $t('pages.common.status'),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.enabled) },
    },
    {
      field: 'binding.providerName',
      title: $t('plugin.linapro-ai-core.tier.fields.provider'),
      minWidth: 150,
    },
    {
      field: 'binding.modelName',
      title: $t('plugin.linapro-ai-core.tier.fields.model'),
      minWidth: 180,
    },
    {
      field: 'binding.protocol',
      title: $t('plugin.linapro-ai-core.model.fields.protocol'),
      minWidth: 100,
    },
    {
      field: 'defaultEffort',
      title: $t('plugin.linapro-ai-core.tier.fields.defaultEffort'),
      minWidth: 120,
    },
    {
      field: 'lastTestStatus',
      title: $t('plugin.linapro-ai-core.tier.fields.lastTestStatus'),
      minWidth: 120,
      slots: { default: ({ row }) => row.lastTestStatus ? statusTag(row.lastTestStatus) : '-' },
    },
    {
      field: 'updatedAt',
      title: $t('pages.common.updatedAt'),
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: 'action',
      fixed: 'right',
      resizable: false,
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      width: 190,
    },
  ];
}

export function buildInvocationQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'purpose',
      label: $t('plugin.linapro-ai-core.invocation.fields.purpose'),
    },
    {
      component: 'Select',
      fieldName: 'tierCode',
      label: $t('plugin.linapro-ai-core.invocation.fields.tierCode'),
      componentProps: {
        options: ['basic', 'standard', 'advanced'].map((value) => ({
          label: tierCodeLabel(value),
          value,
        })),
      },
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('plugin.linapro-ai-core.invocation.fields.status'),
      componentProps: {
        options: [
          { label: $t('plugin.linapro-ai-core.common.success'), value: 'success' },
          { label: $t('plugin.linapro-ai-core.common.failed'), value: 'failed' },
        ],
      },
    },
    {
      component: 'Input',
      fieldName: 'sourcePluginId',
      label: $t('plugin.linapro-ai-core.invocation.fields.sourcePluginId'),
    },
  ];
}

export function buildInvocationColumns(): VxeGridProps['columns'] {
  return [
    {
      field: 'createdAt',
      title: $t('pages.common.createdAt'),
      formatter: ({ cellValue }) => formatTimestamp(cellValue),
      minWidth: 180,
    },
    {
      field: 'purpose',
      title: $t('plugin.linapro-ai-core.invocation.fields.purpose'),
      minWidth: 180,
    },
    {
      field: 'tierCode',
      title: $t('plugin.linapro-ai-core.invocation.fields.tierCode'),
      formatter: ({ cellValue }) => tierCodeLabel(String(cellValue || '')),
      minWidth: 100,
    },
    {
      field: 'status',
      title: $t('plugin.linapro-ai-core.invocation.fields.status'),
      minWidth: 100,
      slots: { default: ({ row }) => statusTag(row.status) },
    },
    {
      field: 'providerName',
      title: $t('plugin.linapro-ai-core.invocation.fields.providerName'),
      minWidth: 150,
    },
    {
      field: 'modelName',
      title: $t('plugin.linapro-ai-core.invocation.fields.modelName'),
      minWidth: 180,
    },
    {
      field: 'latencyMs',
      title: $t('plugin.linapro-ai-core.invocation.fields.latencyMs'),
      minWidth: 110,
    },
    {
      field: 'inputTokens',
      title: $t('plugin.linapro-ai-core.invocation.fields.inputTokens'),
      minWidth: 110,
    },
    {
      field: 'outputTokens',
      title: $t('plugin.linapro-ai-core.invocation.fields.outputTokens'),
      minWidth: 120,
    },
    {
      field: 'action',
      fixed: 'right',
      resizable: false,
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      width: 100,
    },
  ];
}
