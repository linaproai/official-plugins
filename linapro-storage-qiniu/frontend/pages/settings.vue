<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-storage-qiniu-settings',
  title: 'Qiniu Kodo',
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
  Switch,
  message,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-storage-qiniu';
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
  accessKeyID: '',
  secretAccessKey: '',
  region: '',
  bucket: '',
  endpoint: '',
  pathPrefix: '',
  forcePathStyle: false,
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
    formState.accessKeyID = s.accessKeyID || '';
    formState.region = s.region || '';
    formState.bucket = s.bucket || '';
    formState.endpoint = s.endpoint || '';
    formState.pathPrefix = s.pathPrefix || '';
    formState.forcePathStyle = !!s.forcePathStyle;
    secretConfigured.value = !!s.secretAccessKeyConfigured;
    formState.secretAccessKey = '';
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
        accessKeyID: formState.accessKeyID,
        secretAccessKey: formState.secretAccessKey,
        region: formState.region,
        bucket: formState.bucket,
        endpoint: formState.endpoint,
        pathPrefix: formState.pathPrefix,
        forcePathStyle: formState.forcePathStyle,
      },
    );
    const s = data?.settings ?? {};
    secretConfigured.value = !!s.secretAccessKeyConfigured;
    formState.secretAccessKey = '';
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
        accessKeyID: formState.accessKeyID,
        secretAccessKey: formState.secretAccessKey,
        region: formState.region,
        bucket: formState.bucket,
        endpoint: formState.endpoint,
        pathPrefix: formState.pathPrefix,
        forcePathStyle: formState.forcePathStyle,
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
            :label="t('accessKeyIdLabel')"
            :rules="requiredRule(t('accessKeyIdLabel'))"
            :tooltip="t('accessKeyIdHelp')"
            name="accessKeyID"
            required
          >
            <Input
              v-model:value="formState.accessKeyID"
              :placeholder="t('accessKeyIdPlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="t('secretAccessKeyLabel')"
            :required="!secretConfigured"
            :rules="secretConfigured ? [] : requiredRule(t('secretAccessKeyLabel'))"
            :tooltip="t('secretAccessKeyHelp')"
            name="secretAccessKey"
          >
            <InputPassword
              v-model:value="formState.secretAccessKey"
              :placeholder="
                secretConfigured
                  ? t('secretAccessKeyKeepPlaceholder')
                  : t('secretAccessKeyPlaceholder')
              "
            />
          </Form.Item>
          <Form.Item
            :label="t('regionLabel')"
            :tooltip="t('regionHelp')"
            name="region"
          >
            <Input
              v-model:value="formState.region"
              :placeholder="t('regionPlaceholder')"
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="t('bucketLabel')"
            :rules="requiredRule(t('bucketLabel'))"
            :tooltip="t('bucketHelp')"
            name="bucket"
            required
          >
            <Input
              v-model:value="formState.bucket"
              :placeholder="t('bucketPlaceholder')"
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
          <Form.Item
            v-if="false"
            :label="t('forcePathStyleLabel')"
            :tooltip="t('forcePathStyleHelp')"
            name="forcePathStyle"
          >
            <Switch v-model:checked="formState.forcePathStyle" />
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
