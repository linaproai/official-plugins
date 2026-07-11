<script lang="ts">
// pluginPageMeta binds this page to the workbench menu route declared in
// plugin.yaml (path: linapro-auth-ldap-settings) so the dynamic-page shell
// can resolve and mount it.
export const pluginPageMeta = {
  routePath: 'linapro-auth-ldap-settings',
  title: 'LDAP Settings',
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
  Select,
  Switch,
  message,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-auth-ldap';
const formRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };
const loading = ref(false);
const saving = ref(false);
const secretConfigured = ref(false);

/** requiredRule builds the host-standard required message with a red-star label. */
function requiredRule(label: string): Rule[] {
  return [{ required: true, message: $t('ui.formRules.required', [label]) }];
}

const formState = reactive({
  connectionKey: 'default',
  displayName: '',
  host: '',
  port: '636',
  tlsMode: 'ldaps',
  bindDn: '',
  bindPassword: '',
  baseDn: '',
  userFilter: '(uid={username})',
  userDnTemplate: '',
  subjectAttr: 'entryUUID',
  emailAttr: 'mail',
  displayNameAttr: 'cn',
  allowAutoProvision: false,
});

const tlsOptions = [
  { value: 'ldaps', label: 'LDAPS' },
  { value: 'starttls', label: 'StartTLS' },
  { value: 'plain', label: 'Plain (localhost only)' },
];

function settingsApi() {
  return pluginApiPath(pluginID, 'settings');
}

async function loadSettings() {
  loading.value = true;
  try {
    const data = await requestClient.get<{ settings: Record<string, any> }>(
      settingsApi(),
    );
    const s = data?.settings ?? {};
    formState.connectionKey = s.connectionKey || 'default';
    formState.displayName = s.displayName || '';
    formState.host = s.host || '';
    formState.port = s.port || '636';
    formState.tlsMode = s.tlsMode || 'ldaps';
    formState.bindDn = s.bindDn || '';
    formState.baseDn = s.baseDn || '';
    formState.userFilter = s.userFilter || '(uid={username})';
    formState.userDnTemplate = s.userDnTemplate || '';
    formState.subjectAttr = s.subjectAttr || 'entryUUID';
    formState.emailAttr = s.emailAttr || 'mail';
    formState.displayNameAttr = s.displayNameAttr || 'cn';
    formState.allowAutoProvision = !!s.allowAutoProvision;
    secretConfigured.value = !!s.bindPasswordConfigured;
    formState.bindPassword = '';
  } catch {
    message.error($t('plugin.linapro-auth-ldap.settings.loadFailed'));
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
    await requestClient.put(settingsApi(), { ...formState });
    message.success($t('plugin.linapro-auth-ldap.settings.saveSuccess'));
    await loadSettings();
  } catch {
    message.error($t('plugin.linapro-auth-ldap.settings.saveFailed'));
  } finally {
    saving.value = false;
  }
}

onMounted(loadSettings);
</script>

<template>
  <div class="p-4">
    <!-- Menu already names the page; omit Card title to avoid duplicate heading. -->
    <Card :loading="loading">
      <!--
        Wrap Alert + Form with gap so intro tip never collides with the first
        field. Ant Design Alert resets its own margin, so mb-* on Alert is ignored.
      -->
      <div class="flex flex-col gap-4">
        <Alert
          show-icon
          type="info"
          :message="$t('plugin.linapro-auth-ldap.settings.description')"
        />
        <Form
          ref="formRef"
          :colon="false"
          :label-col="labelCol"
          :model="formState"
          :wrapper-col="wrapperCol"
          class="auth-settings-form"
          layout="horizontal"
        >
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.displayNameLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.displayNameHelp')"
          name="displayName"
        >
          <Input v-model:value="formState.displayName" allow-clear />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.hostLabel')"
          :rules="
            requiredRule($t('plugin.linapro-auth-ldap.settings.hostLabel'))
          "
          :tooltip="$t('plugin.linapro-auth-ldap.settings.hostHelp')"
          name="host"
          required
        >
          <Input
            v-model:value="formState.host"
            placeholder="ldap.example.com"
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.portLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.portHelp')"
          name="port"
        >
          <Input v-model:value="formState.port" allow-clear />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.tlsModeLabel')"
          :rules="
            requiredRule($t('plugin.linapro-auth-ldap.settings.tlsModeLabel'))
          "
          :tooltip="$t('plugin.linapro-auth-ldap.settings.tlsModeHelp')"
          name="tlsMode"
          required
        >
          <Select v-model:value="formState.tlsMode" :options="tlsOptions" />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.bindDnLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.bindDnHelp')"
          name="bindDn"
        >
          <Input v-model:value="formState.bindDn" allow-clear />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.bindPasswordLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.bindPasswordHelp')"
          name="bindPassword"
        >
          <InputPassword
            v-model:value="formState.bindPassword"
            :placeholder="
              secretConfigured
                ? $t(
                    'plugin.linapro-auth-ldap.settings.bindPasswordKeepPlaceholder',
                  )
                : $t(
                    'plugin.linapro-auth-ldap.settings.bindPasswordPlaceholder',
                  )
            "
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.baseDnLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.baseDnHelp')"
          name="baseDn"
        >
          <Input
            v-model:value="formState.baseDn"
            placeholder="ou=people,dc=example,dc=com"
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.userFilterLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.userFilterHelp')"
          name="userFilter"
        >
          <Input
            v-model:value="formState.userFilter"
            placeholder="(uid={username})"
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.userDnTemplateLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.userDnTemplateHelp')"
          name="userDnTemplate"
        >
          <Input
            v-model:value="formState.userDnTemplate"
            :placeholder="
              $t('plugin.linapro-auth-ldap.settings.userDnTemplatePlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.subjectAttrLabel')"
          :rules="
            requiredRule(
              $t('plugin.linapro-auth-ldap.settings.subjectAttrLabel'),
            )
          "
          :tooltip="$t('plugin.linapro-auth-ldap.settings.subjectAttrHelp')"
          name="subjectAttr"
          required
        >
          <Input
            v-model:value="formState.subjectAttr"
            placeholder="entryUUID"
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.emailAttrLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.emailAttrHelp')"
          name="emailAttr"
        >
          <Input v-model:value="formState.emailAttr" allow-clear />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.displayNameAttrLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.displayNameAttrHelp')"
          name="displayNameAttr"
        >
          <Input v-model:value="formState.displayNameAttr" allow-clear />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-auth-ldap.settings.autoProvisionLabel')"
          :tooltip="$t('plugin.linapro-auth-ldap.settings.autoProvisionHelp')"
          name="allowAutoProvision"
        >
          <div class="flex items-center gap-3">
            <Switch v-model:checked="formState.allowAutoProvision" />
            <span class="text-muted-foreground text-sm">
              {{ $t('plugin.linapro-auth-ldap.settings.autoProvisionHint') }}
            </span>
          </div>
        </Form.Item>
        <Form.Item class="mt-4" label=" ">
          <Button :loading="saving" type="primary" @click="saveSettings">
            {{ $t('plugin.linapro-auth-ldap.settings.save') }}
          </Button>
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
.auth-settings-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
}
</style>
