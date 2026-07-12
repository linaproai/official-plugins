<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-storage-azure-settings',
  title: 'Azure Blob Storage',
};
</script>

<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import { onMounted, reactive, ref } from 'vue';

import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputPassword,
  message,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-storage-azure';
const formRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };
const loading = ref(false);
const saving = ref(false);
const testing = ref(false);
const testErrorDetail = ref<string | null>(null);
const secretConfigured = ref(false);

function requiredRule(label: string): Rule[] {
  return [{ required: true, message: $t('ui.formRules.required', [label]) }];
}

const formState = reactive({
  accountName: '',
  accountKey: '',
  container: '',
  endpoint: '',
  pathPrefix: '',
});

function settingsApi() {
  return pluginApiPath(pluginID, 'settings');
}

function t(key: string) {
  return $t(`plugin.${pluginID}.settings.${key}`);
}

async function loadSettings() {
  loading.value = true;
  try {
    const data = await requestClient.get<{ settings: Record<string, any> }>(
      settingsApi(),
    );
    const s = data?.settings ?? {};
    formState.accountName = s.accountName || '';
    formState.container = s.container || '';
    formState.endpoint = s.endpoint || '';
    formState.pathPrefix = s.pathPrefix || '';
    secretConfigured.value = !!s.accountKeyConfigured;
    formState.accountKey = '';
  } catch {
    message.error(t('loadFailed'));
  } finally {
    loading.value = false;
  }
}

async function saveSettings() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  saving.value = true;
  try {
    const data = await requestClient.put<{ settings: Record<string, any> }>(
      settingsApi(),
      {
        accountName: formState.accountName,
        accountKey: formState.accountKey,
        container: formState.container,
        endpoint: formState.endpoint,
        pathPrefix: formState.pathPrefix,
      },
    );
    const s = data?.settings ?? {};
    secretConfigured.value = !!s.accountKeyConfigured;
    formState.accountKey = '';
    message.success(t('saveSuccess'));
  } catch {
    message.error(t('saveFailed'));
  } finally {
    saving.value = false;
  }
}

async function testConnection() {
  testing.value = true;
  testErrorDetail.value = null;
  try {
    const data = await requestClient.post<{ ok: boolean; message: string }>(
      pluginApiPath(pluginID, 'settings/test-connection'),
      {
        accountName: formState.accountName,
        accountKey: formState.accountKey,
        container: formState.container,
        endpoint: formState.endpoint,
        pathPrefix: formState.pathPrefix,
      },
    );
    if (data?.ok) {
      message.success(t('testSuccess'));
    } else {
      const detail = (data?.message || '').trim();
      testErrorDetail.value = detail || t('testFailed');
      message.error(t('testFailed'));
    }
  } catch {
    testErrorDetail.value = t('testFailed');
    message.error(t('testFailed'));
  } finally {
    testing.value = false;
  }
}

async function copyTestErrorDetail() {
  const text = testErrorDetail.value;
  if (!text) {
    return;
  }
  try {
    await navigator.clipboard.writeText(text);
    message.success(t('copied'));
  } catch {
    message.error(t('copyFailed'));
  }
}

function clearTestErrorDetail() {
  testErrorDetail.value = null;
}

onMounted(loadSettings);
</script>

<template>
  <div class="p-4">
    <Card :loading="loading">
      <div class="flex flex-col gap-4">
        <Alert show-icon type="info" :message="t('description')" />
        <Alert
          v-if="testErrorDetail"
          data-testid="storage-test-result-alert"
          class="storage-test-result-alert [&_.ant-alert-message]:mb-1 [&_.ant-alert-message]:text-sm [&_.ant-alert-message]:font-normal"
          show-icon
          closable
          type="error"
          @close="clearTestErrorDetail"
        >
          <template #message>
            <span
              class="text-sm font-normal leading-normal"
              data-testid="storage-test-error-title"
            >{{ t('testFailed') }}</span>
          </template>
          <template #description>
            <pre
              class="m-0 max-h-48 overflow-auto whitespace-pre-wrap break-words text-xs leading-relaxed"
              data-testid="storage-test-error-detail"
            >{{ testErrorDetail }}</pre>
          </template>
          <template #action>
            <Button size="small" @click="copyTestErrorDetail">
              {{ t('copyDetail') }}
            </Button>
          </template>
        </Alert>
        <Form
          ref="formRef"
          :colon="false"
          :label-col="labelCol"
          :model="formState"
          :wrapper-col="wrapperCol"
          layout="horizontal"
        >
          <Form.Item
            :label="t('accountNameLabel')"
            :rules="requiredRule(t('accountNameLabel'))"
            :tooltip="t('accountNameHelp')"
            name="accountName"
            required
          >
            <Input
              v-model:value="formState.accountName"
              :placeholder="t('accountNamePlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="t('accountKeyLabel')"
            :required="!secretConfigured"
            :rules="secretConfigured ? [] : requiredRule(t('accountKeyLabel'))"
            :tooltip="t('accountKeyHelp')"
            name="accountKey"
          >
            <InputPassword
              v-model:value="formState.accountKey"
              :placeholder="
                secretConfigured
                  ? t('accountKeyKeepPlaceholder')
                  : t('accountKeyPlaceholder')
              "
            />
          </Form.Item>
          <Form.Item
            :label="t('containerLabel')"
            :rules="requiredRule(t('containerLabel'))"
            :tooltip="t('containerHelp')"
            name="container"
            required
          >
            <Input
              v-model:value="formState.container"
              :placeholder="t('containerPlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="t('endpointLabel')"
            :tooltip="t('endpointHelp')"
            name="endpoint"
          >
            <Input
              v-model:value="formState.endpoint"
              :placeholder="t('endpointPlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="t('pathPrefixLabel')"
            :tooltip="t('pathPrefixHelp')"
            name="pathPrefix"
          >
            <Input
              v-model:value="formState.pathPrefix"
              :placeholder="t('pathPrefixPlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item :wrapper-col="{ offset: 0 }">
            <div class="flex gap-3">
              <Button :loading="testing" @click="testConnection">
                {{ t('testConnection') }}
              </Button>
              <Button type="primary" :loading="saving" @click="saveSettings">
                {{ t('save') }}
              </Button>
            </div>
          </Form.Item>
        </Form>
      </div>
    </Card>
  </div>
</template>
