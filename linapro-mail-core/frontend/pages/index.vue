<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-mail-core-settings',
  title: 'Mail',
};
</script>

<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import { computed, h, onMounted, reactive, ref, watch } from 'vue';

import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  InputPassword,
  Modal,
  Select,
  message,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-mail-core';
const formRef = ref<FormInstance>();
const testMailFormRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };
const loading = ref(false);
const saving = ref(false);
const testing = ref(false);
const sendingTestMail = ref(false);
const receivingTestMail = ref(false);
const testMailOpen = ref(false);
/** Shared account password already stored (SMTP/inbound use the same credentials). */
const passwordConfigured = ref(false);

const formState = reactive({
  // Mailbox login account (required); shared by SMTP and optional IMAP/POP3.
  username: '',
  password: '',
  // Optional From; empty means default to username on save.
  fromAddress: '',
  smtpHost: '',
  smtpPort: 587,
  smtpTlsMode: 'starttls',
  inboundKind: 'none',
  inboundHost: '',
  inboundPort: 993,
  inboundTlsMode: 'tls',
});

const testMailState = reactive({
  to: '',
  subject: '',
  body: '',
});

const inboundEnabled = computed(() => formState.inboundKind !== 'none');

function settingsApi() {
  return pluginApiPath(pluginID, 'mail/settings');
}

function t(key: string) {
  return $t(`plugin.${pluginID}.settings.${key}`);
}

function requiredRule(label: string): Rule[] {
  return [{ required: true, message: $t('ui.formRules.required', [label]) }];
}

function tlsOptions() {
  return [
    { label: t('tlsDisable'), value: 'disable' },
    { label: t('tlsStarttls'), value: 'starttls' },
    { label: t('tlsTls'), value: 'tls' },
  ];
}

function inboundKindOptions() {
  return [
    { label: t('inboundNone'), value: 'none' },
    { label: 'IMAP', value: 'imap' },
    { label: 'POP3', value: 'pop3' },
  ];
}

function applySettings(s: Record<string, any> | undefined | null) {
  const data = s ?? {};
  // Prefer outbound credentials; fall back to inbound if only inbound was set historically.
  const account = data.smtpUsername || data.inboundUsername || '';
  formState.username = account;
  formState.password = '';
  // When stored From equals account, leave the field empty so "optional default" is clear.
  const from = data.fromAddress || '';
  formState.fromAddress = from && from !== account ? from : '';
  formState.smtpHost = data.smtpHost || '';
  formState.smtpPort = data.smtpPort || 587;
  formState.smtpTlsMode = data.smtpTlsMode || 'starttls';
  formState.inboundKind = data.inboundKind || 'none';
  formState.inboundHost = data.inboundHost || '';
  formState.inboundPort = data.inboundPort || defaultInboundPort(formState.inboundKind);
  formState.inboundTlsMode = data.inboundTlsMode || 'tls';
  passwordConfigured.value = !!(
    data.smtpPasswordConfigured || data.inboundPasswordConfigured
  );
}

function defaultInboundPort(kind: string) {
  if (kind === 'pop3') {
    return 995;
  }
  if (kind === 'imap') {
    return 993;
  }
  return 993;
}

/**
 * Apply protocol-specific defaults when the inbound kind changes.
 * Always refresh port/TLS for imap↔pop3 so switching protocols is visible.
 */
function applyInboundKindDefaults(kind: string) {
  if (kind === 'imap') {
    formState.inboundPort = 993;
    formState.inboundTlsMode = 'tls';
    return;
  }
  if (kind === 'pop3') {
    formState.inboundPort = 995;
    formState.inboundTlsMode = 'tls';
  }
}

function inboundHostPlaceholder() {
  if (formState.inboundKind === 'pop3') {
    return t('inboundHostPlaceholderPop3');
  }
  return t('inboundHostPlaceholder');
}

watch(
  () => formState.inboundKind,
  (kind, previous) => {
    if (kind === previous) {
      return;
    }
    applyInboundKindDefaults(String(kind || 'none'));
  },
);

/**
 * Resolve Select change payload to a kind string.
 * Ant Design Vue may emit a raw value or an option object as the first argument.
 */
function resolveInboundKind(value: unknown): string {
  if (typeof value === 'string' || typeof value === 'number') {
    return String(value);
  }
  if (value && typeof value === 'object') {
    const option = value as { value?: string | number };
    if (option.value !== undefined && option.value !== null) {
      return String(option.value);
    }
  }
  return String(formState.inboundKind || 'none');
}

/** Select @change: always re-apply protocol defaults from the selected kind. */
function onInboundKindChange(value: unknown) {
  const kind = resolveInboundKind(value);
  formState.inboundKind = kind;
  applyInboundKindDefaults(kind);
}

async function loadSettings() {
  loading.value = true;
  try {
    const data = await requestClient.get<{ settings: Record<string, any> }>(
      settingsApi(),
    );
    applySettings(data?.settings);
  } catch {
    message.error(t('loadFailed'));
  } finally {
    loading.value = false;
  }
}

/** Build API payload: one shared username/password for SMTP and inbound. */
function payload() {
  const username = formState.username;
  const password = formState.password;
  return {
    fromAddress: formState.fromAddress,
    smtpHost: formState.smtpHost,
    smtpPort: formState.smtpPort,
    smtpUsername: username,
    smtpPassword: password,
    smtpTlsMode: formState.smtpTlsMode,
    inboundKind: formState.inboundKind,
    inboundHost: formState.inboundHost,
    inboundPort: formState.inboundPort,
    inboundUsername: username,
    inboundPassword: password,
    inboundTlsMode: formState.inboundTlsMode,
  };
}

/** Show failure detail in a modal so operators can read and copy the cause. */
function showErrorModal(title: string, detail: string) {
  const text = (detail || '').trim() || title;
  Modal.error({
    title,
    width: 560,
    centered: true,
    content: h(
      'pre',
      {
        'data-testid': 'mail-error-modal-detail',
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
      payload(),
    );
    applySettings(data?.settings);
    message.success(t('saveSuccess'));
  } catch {
    message.error(t('saveFailed'));
  } finally {
    saving.value = false;
  }
}

async function testConnection() {
  try {
    await formRef.value?.validate([
      'username',
      'smtpHost',
      'smtpPort',
    ]);
  } catch {
    return;
  }
  testing.value = true;
  try {
    const data = await requestClient.post<{ ok: boolean; message: string }>(
      pluginApiPath(pluginID, 'mail/settings/test'),
      payload(),
    );
    if (data?.ok) {
      message.success(t('testSuccess'));
    } else {
      const detail = (data?.message || '').trim() || t('testFailed');
      showErrorModal(t('testErrorTitle'), detail);
    }
  } catch (error: any) {
    const detail =
      error?.response?.data?.message || error?.message || t('testFailed');
    showErrorModal(t('testErrorTitle'), String(detail));
  } finally {
    testing.value = false;
  }
}

async function openSendTestMail() {
  try {
    await formRef.value?.validate([
      'username',
      'smtpHost',
      'smtpPort',
    ]);
  } catch {
    return;
  }
  testMailState.to = '';
  testMailState.subject = '';
  testMailState.body = t('sendTestBodyDefault');
  testMailOpen.value = true;
}

function closeSendTestMail() {
  testMailOpen.value = false;
}

async function testReceive() {
  if (!inboundEnabled.value) {
    showErrorModal(t('receiveTestErrorTitle'), t('receiveTestInboundRequired'));
    return;
  }
  try {
    await formRef.value?.validate([
      'username',
      'inboundHost',
      'inboundPort',
    ]);
  } catch {
    return;
  }
  receivingTestMail.value = true;
  try {
    const data = await requestClient.post<{ ok: boolean; message: string }>(
      pluginApiPath(pluginID, 'mail/settings/receive-test'),
      payload(),
    );
    if (data?.ok) {
      message.success(t('receiveTestSuccess'));
      return;
    }
    const detail = (data?.message || '').trim() || t('receiveTestFailed');
    showErrorModal(t('receiveTestErrorTitle'), detail);
  } catch (error: any) {
    const detail =
      error?.response?.data?.message || error?.message || t('receiveTestFailed');
    showErrorModal(t('receiveTestErrorTitle'), String(detail));
  } finally {
    receivingTestMail.value = false;
  }
}

async function submitSendTestMail() {
  try {
    await testMailFormRef.value?.validate();
  } catch {
    // Keep the dialog open when client validation fails.
    return Promise.reject(new Error('validation'));
  }
  sendingTestMail.value = true;
  try {
    const data = await requestClient.post<{ ok: boolean; message: string }>(
      pluginApiPath(pluginID, 'mail/settings/send-test'),
      {
        ...payload(),
        to: testMailState.to,
        subject: testMailState.subject,
        body: testMailState.body,
      },
    );
    if (data?.ok) {
      message.success(t('sendTestSuccess'));
      testMailOpen.value = false;
      return;
    }
    const detail = (data?.message || '').trim() || t('sendTestFailed');
    showErrorModal(t('sendTestErrorTitle'), detail);
    return Promise.reject(new Error(detail));
  } catch (error: any) {
    // Re-throw validation rejections without an extra error modal.
    if (error?.message === 'validation') {
      return Promise.reject(error);
    }
    const detail =
      error?.response?.data?.message || error?.message || t('sendTestFailed');
    showErrorModal(t('sendTestErrorTitle'), String(detail));
    return Promise.reject(error instanceof Error ? error : new Error(String(detail)));
  } finally {
    sendingTestMail.value = false;
  }
}

onMounted(loadSettings);
</script>

<template>
  <div class="p-4" data-testid="mail-settings-page">
    <Card :loading="loading">
      <div class="flex flex-col gap-4">
        <Alert
          data-testid="mail-settings-tip"
          show-icon
          type="info"
          :message="t('description')"
        />
        <Form
          ref="formRef"
          :colon="false"
          :label-col="labelCol"
          :model="formState"
          :wrapper-col="wrapperCol"
          class="mail-settings-form"
          layout="horizontal"
        >
          <Form.Item
            :label="t('usernameLabel')"
            :rules="requiredRule(t('usernameLabel'))"
            :tooltip="t('usernameHelp')"
            name="username"
            required
          >
            <Input
              v-model:value="formState.username"
              :placeholder="t('usernamePlaceholder')"
              allow-clear
              data-testid="mail-settings-username"
            />
          </Form.Item>
          <Form.Item
            :label="t('passwordLabel')"
            :required="!passwordConfigured"
            :rules="passwordConfigured ? [] : requiredRule(t('passwordLabel'))"
            :tooltip="t('passwordHelp')"
            name="password"
          >
            <InputPassword
              v-model:value="formState.password"
              :placeholder="
                passwordConfigured
                  ? t('passwordKeepPlaceholder')
                  : t('passwordPlaceholder')
              "
              data-testid="mail-settings-password"
            />
          </Form.Item>
          <Form.Item
            :label="t('fromAddressLabel')"
            :tooltip="t('fromAddressHelp')"
            name="fromAddress"
          >
            <Input
              v-model:value="formState.fromAddress"
              :placeholder="t('fromAddressPlaceholder')"
              allow-clear
              data-testid="mail-settings-from"
            />
          </Form.Item>

          <div class="mb-2 text-sm font-medium opacity-80">
            {{ t('smtpSection') }}
          </div>
          <Form.Item
            :label="t('smtpHostLabel')"
            :rules="requiredRule(t('smtpHostLabel'))"
            name="smtpHost"
            required
          >
            <Input
              v-model:value="formState.smtpHost"
              :placeholder="t('smtpHostPlaceholder')"
              allow-clear
              data-testid="mail-settings-smtp-host"
            />
          </Form.Item>
          <Form.Item
            :label="t('smtpPortLabel')"
            :rules="requiredRule(t('smtpPortLabel'))"
            name="smtpPort"
            required
          >
            <InputNumber
              v-model:value="formState.smtpPort"
              :min="1"
              :max="65535"
              class="w-full"
              data-testid="mail-settings-smtp-port"
            />
          </Form.Item>
          <Form.Item
            :label="t('smtpTlsLabel')"
            :tooltip="t('smtpTlsHelp')"
            name="smtpTlsMode"
          >
            <Select
              v-model:value="formState.smtpTlsMode"
              :options="tlsOptions()"
              data-testid="mail-settings-smtp-tls"
            />
          </Form.Item>

          <div class="mb-2 text-sm font-medium opacity-80">
            {{ t('inboundSection') }}
          </div>
          <Form.Item
            :label="t('inboundKindLabel')"
            :tooltip="t('inboundKindHelp')"
            name="inboundKind"
          >
            <Select
              v-model:value="formState.inboundKind"
              :options="inboundKindOptions()"
              data-testid="mail-settings-inbound-kind"
              @change="onInboundKindChange"
            />
          </Form.Item>
          <template v-if="inboundEnabled">
            <Form.Item
              :label="t('inboundHostLabel')"
              :rules="requiredRule(t('inboundHostLabel'))"
              name="inboundHost"
              required
            >
              <Input
                v-model:value="formState.inboundHost"
                :placeholder="inboundHostPlaceholder()"
                allow-clear
                data-testid="mail-settings-inbound-host"
              />
            </Form.Item>
            <Form.Item
              :label="t('inboundPortLabel')"
              :rules="requiredRule(t('inboundPortLabel'))"
              name="inboundPort"
              required
            >
              <InputNumber
                v-model:value="formState.inboundPort"
                :min="1"
                :max="65535"
                class="w-full"
                data-testid="mail-settings-inbound-port"
              />
            </Form.Item>
            <Form.Item
              :label="t('inboundTlsLabel')"
              :tooltip="t('inboundTlsHelp')"
              name="inboundTlsMode"
            >
              <Select
                v-model:value="formState.inboundTlsMode"
                :options="tlsOptions()"
                data-testid="mail-settings-inbound-tls"
              />
            </Form.Item>
          </template>

          <Form.Item :wrapper-col="{ offset: 0 }">
            <div class="flex flex-wrap gap-3">
              <Button
                :loading="testing"
                data-testid="mail-settings-test"
                @click="testConnection"
              >
                {{ t('testConnection') }}
              </Button>
              <Button
                data-testid="mail-settings-send-test"
                @click="openSendTestMail"
              >
                {{ t('sendTestMail') }}
              </Button>
              <Button
                :loading="receivingTestMail"
                data-testid="mail-settings-receive-test"
                @click="testReceive"
              >
                {{ t('receiveTestMail') }}
              </Button>
              <Button
                type="primary"
                :loading="saving"
                data-testid="mail-settings-save"
                @click="saveSettings"
              >
                {{ t('save') }}
              </Button>
            </div>
          </Form.Item>
        </Form>
      </div>
    </Card>

    <Modal
      v-model:open="testMailOpen"
      :confirm-loading="sendingTestMail"
      :ok-text="t('sendTestSubmit')"
      :cancel-text="t('sendTestCancel')"
      :title="t('sendTestTitle')"
      :width="560"
      destroy-on-close
      data-testid="mail-send-test-modal"
      @ok="submitSendTestMail"
      @cancel="closeSendTestMail"
    >
      <div
        class="flex flex-col"
        data-testid="mail-send-test-body-layout"
        style="gap: 16px"
      >
        <Alert
          data-testid="mail-send-test-help"
          show-icon
          type="info"
          :message="t('sendTestHelp')"
          style="margin-bottom: 0"
        />
        <Form
          ref="testMailFormRef"
          :colon="false"
          :label-col="{ style: { width: '96px' } }"
          :model="testMailState"
          class="mail-settings-form"
          layout="horizontal"
          style="margin-top: 0"
        >
          <Form.Item
            :label="t('sendTestToLabel')"
            :rules="requiredRule(t('sendTestToLabel'))"
            name="to"
            required
          >
            <Input
              v-model:value="testMailState.to"
              :placeholder="t('sendTestToPlaceholder')"
              allow-clear
              data-testid="mail-send-test-to"
            />
          </Form.Item>
          <Form.Item
            :label="t('sendTestSubjectLabel')"
            name="subject"
          >
            <Input
              v-model:value="testMailState.subject"
              :placeholder="t('sendTestSubjectPlaceholder')"
              allow-clear
              data-testid="mail-send-test-subject"
            />
          </Form.Item>
          <Form.Item
            :label="t('sendTestBodyLabel')"
            :rules="requiredRule(t('sendTestBodyLabel'))"
            name="body"
            required
          >
            <Input.TextArea
              v-model:value="testMailState.body"
              :placeholder="t('sendTestBodyPlaceholder')"
              :rows="5"
              allow-clear
              data-testid="mail-send-test-body"
            />
          </Form.Item>
        </Form>
      </div>
    </Modal>
  </div>
</template>

<style scoped>
/*
 * Align raw ant-design Form labels with host useVbenForm conventions:
 * medium (semi-bold) weight and no trailing colon (handled via :colon="false").
 */
.mail-settings-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
}
</style>
