<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-storage-oss-settings',
  title: 'Alibaba Cloud OSS',
};
</script>

<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import { h, onMounted, reactive, ref } from 'vue';

import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputPassword,
  Modal,
  Switch,
  message,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-storage-oss';
const formRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };
const loading = ref(false);
const saving = ref(false);
const testing = ref(false);
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


/** Show failure detail in a modal so operators can read the cause (aligned with mail settings). */
function showErrorModal(title: string, detail: string) {
  const text = (detail || '').trim() || title;
  Modal.error({
    title,
    width: 560,
    centered: true,
    content: h(
      'pre',
      {
        'data-testid': 'storage-error-modal-detail',
        style: {
          margin: 0,
          maxHeight: '320px',
          overflow: 'auto',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          fontSize: '12px',
          lineHeight: 1.5,
        },
      },
      text,
    ),
  });
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
      const detail = (data?.message || '').trim() || t('testFailed');
      showErrorModal(t('testFailed'), detail);
    }
  } catch (error: any) {
    const detail =
      error?.response?.data?.message || error?.message || t('testFailed');
    showErrorModal(t('testFailed'), String(detail));
  } finally {
    testing.value = false;
  }
}


onMounted(loadSettings);
</script>

<template>
  <div class="p-4">
    <Card :loading="loading">
      <div class="flex flex-col gap-4">
        <Alert show-icon type="info" :message="t('description')" />
        <Form
          ref="formRef"
          :colon="false"
          :label-col="labelCol"
          :model="formState"
          :wrapper-col="wrapperCol"
          class="storage-settings-form"
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
            :rules="requiredRule(t('regionLabel'))"
            :tooltip="t('regionHelp')"
            name="region"
            required
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

<style scoped>
/*
 * Align raw ant-design Form labels with host useVbenForm conventions:
 * medium (semi-bold) weight and no trailing colon (handled via :colon="false").
 */
.storage-settings-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
}
</style>
