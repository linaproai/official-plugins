<script lang="ts">
// pluginPageMeta binds this page to the workbench menu route declared in
// plugin.yaml (path: linapro-oidc-generic-settings) so the dynamic-page shell
// can resolve and mount it.
export const pluginPageMeta = {
  routePath: 'linapro-oidc-generic-settings',
  title: 'OIDC Settings',
};
</script>

<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import { computed, onMounted, reactive, ref } from 'vue';

import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputPassword,
  message,
  Switch,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-oidc-generic';
const formRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };

/** requiredRule builds the host-standard required message with a red-star label. */
function requiredRule(label: string): Rule[] {
  return [{ required: true, message: $t('ui.formRules.required', [label]) }];
}

interface SettingsItem {
  connectionKey: string;
  displayName: string;
  issuer: string;
  clientId: string;
  clientSecretMasked: string;
  clientSecretConfigured: boolean;
  redirectUrl: string;
  scopes: string;
  allowAutoProvision: boolean;
  defaultBackendRedirect: string;
}

const loading = ref(false);
const saving = ref(false);
const secretConfigured = ref(false);

/**
 * displayCallbackUrl is the fixed plugin callback URL derived from the current
 * origin. It is display-only: the backend derives the same URL from the live
 * request host, so admins only need to copy it into the IdP console.
 */
const displayCallbackUrl = computed(
  () => `${window.location.origin}/portal/${pluginID}/callback`,
);

/** copyCallbackUrl copies the callback URL for pasting into the IdP console. */
async function copyCallbackUrl() {
  try {
    await navigator.clipboard.writeText(displayCallbackUrl.value);
    message.success($t('plugin.linapro-oidc-generic.settings.copied'));
  } catch {
    message.error($t('plugin.linapro-oidc-generic.settings.copyFailed'));
  }
}

const formState = reactive({
  displayName: '',
  issuer: '',
  clientId: '',
  clientSecret: '',
  redirectUrl: '',
  scopes: 'openid email profile',
  allowAutoProvision: false,
  defaultBackendRedirect: '',
  connectionKey: 'default',
});

function settingsApi() {
  return pluginApiPath(pluginID, 'settings');
}

async function loadSettings() {
  loading.value = true;
  try {
    const data = await requestClient.get<{ settings: SettingsItem }>(
      settingsApi(),
    );
    const settings = data?.settings;
    formState.connectionKey = settings?.connectionKey || 'default';
    formState.displayName = settings?.displayName || '';
    formState.issuer = settings?.issuer || '';
    formState.clientId = settings?.clientId || '';
    formState.redirectUrl = settings?.redirectUrl || '';
    formState.scopes = settings?.scopes || 'openid email profile';
    formState.allowAutoProvision = !!settings?.allowAutoProvision;
    formState.defaultBackendRedirect = settings?.defaultBackendRedirect || '';
    secretConfigured.value = !!settings?.clientSecretConfigured;
    formState.clientSecret = '';
  } catch {
    message.error($t('plugin.linapro-oidc-generic.settings.loadFailed'));
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
    await requestClient.put(settingsApi(), {
      displayName: formState.displayName,
      issuer: formState.issuer,
      clientId: formState.clientId,
      clientSecret: formState.clientSecret,
      redirectUrl: formState.redirectUrl,
      scopes: formState.scopes,
      allowAutoProvision: formState.allowAutoProvision,
      defaultBackendRedirect: formState.defaultBackendRedirect,
    });
    message.success($t('plugin.linapro-oidc-generic.settings.saveSuccess'));
    await loadSettings();
  } catch {
    message.error($t('plugin.linapro-oidc-generic.settings.saveFailed'));
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
          :message="$t('plugin.linapro-oidc-generic.settings.description')"
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
          :label="$t('plugin.linapro-oidc-generic.settings.connectionKeyLabel')"
          :tooltip="
            $t('plugin.linapro-oidc-generic.settings.connectionKeyHelp')
          "
          name="connectionKey"
        >
          <Input :value="formState.connectionKey" disabled />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.displayNameLabel')"
          :tooltip="$t('plugin.linapro-oidc-generic.settings.displayNameHelp')"
          name="displayName"
        >
          <Input
            v-model:value="formState.displayName"
            :placeholder="
              $t('plugin.linapro-oidc-generic.settings.displayNamePlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.issuerLabel')"
          :rules="
            requiredRule($t('plugin.linapro-oidc-generic.settings.issuerLabel'))
          "
          :tooltip="$t('plugin.linapro-oidc-generic.settings.issuerHelp')"
          name="issuer"
          required
        >
          <Input
            v-model:value="formState.issuer"
            :placeholder="
              $t('plugin.linapro-oidc-generic.settings.issuerPlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.clientIdLabel')"
          :rules="
            requiredRule(
              $t('plugin.linapro-oidc-generic.settings.clientIdLabel'),
            )
          "
          :tooltip="$t('plugin.linapro-oidc-generic.settings.clientIdHelp')"
          name="clientId"
          required
        >
          <Input
            v-model:value="formState.clientId"
            :placeholder="
              $t('plugin.linapro-oidc-generic.settings.clientIdPlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.clientSecretLabel')"
          :required="!secretConfigured"
          :rules="
            secretConfigured
              ? []
              : requiredRule(
                  $t('plugin.linapro-oidc-generic.settings.clientSecretLabel'),
                )
          "
          :tooltip="$t('plugin.linapro-oidc-generic.settings.clientSecretHelp')"
          name="clientSecret"
        >
          <InputPassword
            v-model:value="formState.clientSecret"
            :placeholder="
              secretConfigured
                ? $t(
                    'plugin.linapro-oidc-generic.settings.clientSecretKeepPlaceholder',
                  )
                : $t(
                    'plugin.linapro-oidc-generic.settings.clientSecretPlaceholder',
                  )
            "
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.redirectUrlLabel')"
          :tooltip="$t('plugin.linapro-oidc-generic.settings.redirectUrlHelp')"
          name="redirectUrl"
        >
          <div class="flex items-center gap-2">
            <Input :value="displayCallbackUrl" class="flex-1" readonly />
            <Button @click="copyCallbackUrl">
              {{ $t('plugin.linapro-oidc-generic.settings.copyCallbackUrl') }}
            </Button>
          </div>
          <p class="text-muted-foreground mt-1 text-xs">
            {{ $t('plugin.linapro-oidc-generic.settings.redirectUrlHint') }}
          </p>
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.scopesLabel')"
          :tooltip="$t('plugin.linapro-oidc-generic.settings.scopesHelp')"
          name="scopes"
        >
          <Input
            v-model:value="formState.scopes"
            :placeholder="
              $t('plugin.linapro-oidc-generic.settings.scopesPlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-generic.settings.defaultRedirectLabel')"
          :tooltip="
            $t('plugin.linapro-oidc-generic.settings.defaultRedirectHelp')
          "
          name="defaultBackendRedirect"
        >
          <Input
            v-model:value="formState.defaultBackendRedirect"
            :placeholder="
              $t(
                'plugin.linapro-oidc-generic.settings.defaultRedirectPlaceholder',
              )
            "
            allow-clear
          />
          <p class="text-muted-foreground mt-1 text-xs">
            {{ $t('plugin.linapro-oidc-generic.settings.defaultRedirectHint') }}
          </p>
        </Form.Item>
        <Form.Item
          :label="
            $t('plugin.linapro-oidc-generic.settings.autoProvisionLabel')
          "
          :tooltip="
            $t('plugin.linapro-oidc-generic.settings.autoProvisionHelp')
          "
          name="allowAutoProvision"
        >
          <div class="flex items-center gap-3">
            <Switch v-model:checked="formState.allowAutoProvision" />
            <span class="text-muted-foreground text-sm">
              {{ $t('plugin.linapro-oidc-generic.settings.autoProvisionHint') }}
            </span>
          </div>
        </Form.Item>
        <Form.Item class="mt-4" label=" ">
          <Button :loading="saving" type="primary" @click="saveSettings">
            {{ $t('plugin.linapro-oidc-generic.settings.save') }}
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
