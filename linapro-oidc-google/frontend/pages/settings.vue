<script lang="ts">
// pluginPageMeta binds this page to the workbench menu route declared in
// plugin.yaml (path: linapro-oidc-google-settings) so the dynamic-page shell
// can resolve and mount it.
export const pluginPageMeta = {
  routePath: 'linapro-oidc-google-settings',
  title: 'Google',
};
</script>

<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import { computed, onMounted, reactive, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputPassword,
  message,
  Switch,
  Tooltip,
} from 'ant-design-vue';

import { pluginApiPath, requestClient } from '#/api/request';
import { $t } from '#/locales';

const pluginID = 'linapro-oidc-google';
const formRef = ref<FormInstance>();
const labelCol = { style: { width: '180px' } };
const wrapperCol = { style: { maxWidth: '720px' } };

/** requiredRule builds the host-standard required message with a red-star label. */
function requiredRule(label: string): Rule[] {
  return [{ required: true, message: $t('ui.formRules.required', [label]) }];
}

interface SettingsItem {
  clientId: string;
  clientSecretMasked: string;
  clientSecretConfigured: boolean;
  redirectUrl: string;
  enableBackendRedirect: boolean;
  defaultBackendRedirect: string;
  backendRedirects: string;
  allowAutoProvision: boolean;
  enableOneTap: boolean;
}

interface RedirectRule {
  key: string;
  url: string;
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
    message.success($t('plugin.linapro-oidc-google.settings.copied'));
  } catch {
    message.error($t('plugin.linapro-oidc-google.settings.copyFailed'));
  }
}

const formState = reactive({
  clientId: '',
  clientSecret: '',
  redirectUrl: '',
  enableBackendRedirect: false,
  defaultBackendRedirect: '',
  rules: [] as RedirectRule[],
  allowAutoProvision: false,
  enableOneTap: false,
});

/** oneTapStateKey lets the admin pick which SSO rule the embed targets. */
const oneTapStateKey = ref('');

/**
 * oneTapEmbedCode renders the copy-paste GSI snippet for third-party pages.
 * The login_uri points at the plugin's public One Tap endpoint; an optional
 * state key routes the outcome to the matching SSO delivery rule.
 */
const oneTapEmbedCode = computed(() => {
  const stateQuery = oneTapStateKey.value.trim()
    ? `?state=${encodeURIComponent(oneTapStateKey.value.trim())}`
    : '';
  const loginUri = `${window.location.origin}/portal/${pluginID}/onetap${stateQuery}`;
  return [
    '<script src="https://accounts.google.com/gsi/client" async><' + '/script>',
    '<div id="g_id_onload"',
    `     data-client_id="${formState.clientId || 'YOUR_GOOGLE_CLIENT_ID'}"`,
    `     data-login_uri="${loginUri}"`,
    '     data-auto_select="true"',
    '     data-itp_support="true"></div>',
  ].join('\n');
});

/** copyOneTapEmbed copies the embeddable snippet to the clipboard. */
async function copyOneTapEmbed() {
  try {
    await navigator.clipboard.writeText(oneTapEmbedCode.value);
    message.success($t('plugin.linapro-oidc-google.settings.copied'));
  } catch {
    message.error($t('plugin.linapro-oidc-google.settings.copyFailed'));
  }
}

function settingsApi() {
  return pluginApiPath(pluginID, 'settings');
}

/**
 * parseRules converts the persisted JSON rules object into editable rows.
 * Malformed payloads degrade to an empty list so the page never crashes.
 */
function parseRules(raw: string): RedirectRule[] {
  if (!raw) {
    return [];
  }
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return Object.entries(parsed).map(([key, url]) => ({
        key,
        url: String(url),
      }));
    }
  } catch {
    // fall through to empty rules
  }
  return [];
}

/** serializeRules converts editable rows back into the persisted JSON object. */
function serializeRules(rules: RedirectRule[]): string {
  const dict: Record<string, string> = {};
  for (const rule of rules) {
    const key = rule.key.trim();
    const url = rule.url.trim();
    if (key && url) {
      dict[key] = url;
    }
  }
  return Object.keys(dict).length > 0 ? JSON.stringify(dict) : '';
}

function applySettings(settings: null | SettingsItem) {
  formState.clientId = settings?.clientId ?? '';
  formState.clientSecret = settings?.clientSecretMasked ?? '';
  formState.redirectUrl = settings?.redirectUrl ?? '';
  formState.enableBackendRedirect = settings?.enableBackendRedirect ?? false;
  formState.defaultBackendRedirect = settings?.defaultBackendRedirect ?? '';
  formState.rules = parseRules(settings?.backendRedirects ?? '');
  formState.allowAutoProvision = settings?.allowAutoProvision ?? false;
  formState.enableOneTap = settings?.enableOneTap ?? false;
  secretConfigured.value = settings?.clientSecretConfigured ?? false;
}

function addRule() {
  formState.rules.push({ key: '', url: '' });
}

function removeRule(index: number) {
  formState.rules.splice(index, 1);
}

/**
 * copyLoginUrl copies the SSO login entry URL carrying the rule's state key so
 * the third-party system can link directly into this provider's login flow.
 */
async function copyLoginUrl(rule: RedirectRule) {
  const key = rule.key.trim();
  if (!key) {
    return;
  }
  const loginUrl = `${window.location.origin}/portal/${pluginID}/login?state=${encodeURIComponent(key)}`;
  try {
    await navigator.clipboard.writeText(loginUrl);
    message.success($t('plugin.linapro-oidc-google.settings.copied'));
  } catch {
    message.error($t('plugin.linapro-oidc-google.settings.copyFailed'));
  }
}

async function loadSettings() {
  loading.value = true;
  try {
    const res = await requestClient.get<{ settings: SettingsItem }>(
      settingsApi(),
    );
    applySettings(res?.settings ?? null);
  } catch {
    message.error($t('plugin.linapro-oidc-google.settings.loadFailed'));
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
    const res = await requestClient.put<{ settings: SettingsItem }>(
      settingsApi(),
      {
        clientId: formState.clientId,
        clientSecret: formState.clientSecret,
        // The callback URL is fixed by the plugin route and derived by the
        // backend from the live request host; persist empty to keep derivation.
        redirectUrl: '',
        enableBackendRedirect: formState.enableBackendRedirect,
        defaultBackendRedirect: formState.defaultBackendRedirect,
        backendRedirects: serializeRules(formState.rules),
        allowAutoProvision: formState.allowAutoProvision,
        enableOneTap: formState.enableOneTap,
      },
    );
    applySettings(res?.settings ?? null);
    message.success($t('plugin.linapro-oidc-google.settings.saveSuccess'));
  } catch {
    message.error($t('plugin.linapro-oidc-google.settings.saveFailed'));
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
        Wrap Alert + Form with space-y so intro tip never collides with the first
        field. Ant Design Alert resets its own margin, so mb-* on Alert is ignored.
      -->
      <div class="flex flex-col gap-4">
        <Alert
          show-icon
          type="info"
          :message="$t('plugin.linapro-oidc-google.settings.description')"
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
          :label="$t('plugin.linapro-oidc-google.settings.clientIdLabel')"
          :rules="
            requiredRule(
              $t('plugin.linapro-oidc-google.settings.clientIdLabel'),
            )
          "
          :tooltip="$t('plugin.linapro-oidc-google.settings.clientIdHelp')"
          name="clientId"
          required
        >
          <Input
            v-model:value="formState.clientId"
            :placeholder="
              $t('plugin.linapro-oidc-google.settings.clientIdPlaceholder')
            "
            allow-clear
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-google.settings.clientSecretLabel')"
          :required="!secretConfigured"
          :rules="
            secretConfigured
              ? []
              : requiredRule(
                  $t('plugin.linapro-oidc-google.settings.clientSecretLabel'),
                )
          "
          :tooltip="$t('plugin.linapro-oidc-google.settings.clientSecretHelp')"
          name="clientSecret"
        >
          <InputPassword
            v-model:value="formState.clientSecret"
            :placeholder="
              secretConfigured
                ? $t(
                    'plugin.linapro-oidc-google.settings.clientSecretKeepPlaceholder',
                  )
                : $t(
                    'plugin.linapro-oidc-google.settings.clientSecretPlaceholder',
                  )
            "
          />
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-google.settings.redirectUrlLabel')"
          :tooltip="$t('plugin.linapro-oidc-google.settings.redirectUrlHelp')"
          name="redirectUrl"
        >
          <div class="flex items-center gap-2">
            <Input :value="displayCallbackUrl" class="flex-1" readonly />
            <Button @click="copyCallbackUrl">
              {{ $t('plugin.linapro-oidc-google.settings.copyCallbackUrl') }}
            </Button>
          </div>
          <p class="text-muted-foreground mt-1 text-xs">
            {{ $t('plugin.linapro-oidc-google.settings.redirectUrlHint') }}
          </p>
        </Form.Item>
        <Form.Item
          :label="
            $t('plugin.linapro-oidc-google.settings.defaultRedirectLabel')
          "
          :tooltip="
            $t('plugin.linapro-oidc-google.settings.defaultRedirectHelp')
          "
          name="defaultBackendRedirect"
        >
          <Input
            v-model:value="formState.defaultBackendRedirect"
            :placeholder="
              $t(
                'plugin.linapro-oidc-google.settings.defaultRedirectPlaceholder',
              )
            "
            allow-clear
          />
          <p class="text-muted-foreground mt-1 text-xs">
            {{ $t('plugin.linapro-oidc-google.settings.defaultRedirectHint') }}
          </p>
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-google.settings.autoProvisionLabel')"
          :tooltip="$t('plugin.linapro-oidc-google.settings.autoProvisionHelp')"
          name="allowAutoProvision"
        >
          <div class="flex items-center gap-3">
            <Switch v-model:checked="formState.allowAutoProvision" />
            <span class="text-muted-foreground text-sm">
              {{ $t('plugin.linapro-oidc-google.settings.autoProvisionHint') }}
            </span>
          </div>
        </Form.Item>
        <Form.Item
          :label="$t('plugin.linapro-oidc-google.settings.enableSsoLabel')"
          :tooltip="$t('plugin.linapro-oidc-google.settings.enableSsoHelp')"
          name="enableBackendRedirect"
        >
          <div class="flex items-center gap-3">
            <Switch v-model:checked="formState.enableBackendRedirect" />
            <span class="text-muted-foreground text-sm">
              {{ $t('plugin.linapro-oidc-google.settings.enableSsoHint') }}
            </span>
          </div>
        </Form.Item>
        <template v-if="formState.enableBackendRedirect">
          <div
            class="mb-3 flex items-center justify-between"
            :style="{ marginLeft: '180px', maxWidth: '720px' }"
          >
            <span class="inline-flex items-center gap-1 font-medium">
              {{ $t('plugin.linapro-oidc-google.settings.rulesTitle') }}
              <Tooltip
                :title="$t('plugin.linapro-oidc-google.settings.rulesHelp')"
              >
                <IconifyIcon
                  class="text-muted-foreground text-xs"
                  icon="ant-design:question-circle-outlined"
                />
              </Tooltip>
            </span>
            <Button size="small" @click="addRule">
              {{ $t('plugin.linapro-oidc-google.settings.addRule') }}
            </Button>
          </div>
          <div
            v-for="(rule, index) in formState.rules"
            :key="index"
            class="mb-2"
          >
            <Form.Item
              :label="$t('plugin.linapro-oidc-google.settings.ruleKeyLabel')"
              :tooltip="$t('plugin.linapro-oidc-google.settings.ruleKeyHelp')"
            >
              <Input
                v-model:value="rule.key"
                :placeholder="
                  $t('plugin.linapro-oidc-google.settings.ruleKeyPlaceholder')
                "
              />
            </Form.Item>
            <Form.Item
              :label="$t('plugin.linapro-oidc-google.settings.ruleUrlLabel')"
              :tooltip="$t('plugin.linapro-oidc-google.settings.ruleUrlHelp')"
            >
              <div class="flex flex-wrap items-center gap-2">
                <Input
                  v-model:value="rule.url"
                  class="min-w-0 flex-1"
                  :placeholder="
                    $t('plugin.linapro-oidc-google.settings.ruleUrlPlaceholder')
                  "
                />
                <Button size="small" @click="copyLoginUrl(rule)">
                  {{ $t('plugin.linapro-oidc-google.settings.copyLoginUrl') }}
                </Button>
                <Button danger size="small" @click="removeRule(index)">
                  {{ $t('plugin.linapro-oidc-google.settings.deleteRule') }}
                </Button>
              </div>
            </Form.Item>
          </div>
        </template>
        <Form.Item
          :label="$t('plugin.linapro-oidc-google.settings.oneTapLabel')"
          :tooltip="$t('plugin.linapro-oidc-google.settings.oneTapHelp')"
          name="enableOneTap"
        >
          <div class="flex items-center gap-3">
            <Switch v-model:checked="formState.enableOneTap" />
            <span class="text-muted-foreground text-sm">
              {{ $t('plugin.linapro-oidc-google.settings.oneTapHint') }}
            </span>
          </div>
        </Form.Item>
        <template v-if="formState.enableOneTap">
          <Form.Item
            :label="
              $t('plugin.linapro-oidc-google.settings.oneTapStateKeyLabel')
            "
            :tooltip="
              $t('plugin.linapro-oidc-google.settings.oneTapStateKeyHelp')
            "
          >
            <Input
              v-model:value="oneTapStateKey"
              :placeholder="
                $t(
                  'plugin.linapro-oidc-google.settings.oneTapStateKeyPlaceholder',
                )
              "
              allow-clear
            />
          </Form.Item>
          <Form.Item
            :label="$t('plugin.linapro-oidc-google.settings.oneTapEmbedTitle')"
            :tooltip="$t('plugin.linapro-oidc-google.settings.oneTapEmbedHelp')"
          >
            <div class="mb-2 flex justify-end">
              <Button size="small" @click="copyOneTapEmbed">
                {{ $t('plugin.linapro-oidc-google.settings.copyEmbed') }}
              </Button>
            </div>
            <pre
              class="bg-muted mb-2 overflow-x-auto rounded p-3 text-xs"
            ><code>{{ oneTapEmbedCode }}</code></pre>
            <p class="text-muted-foreground text-xs">
              {{ $t('plugin.linapro-oidc-google.settings.oneTapEmbedHint') }}
            </p>
          </Form.Item>
        </template>
        <Form.Item class="mt-4" label=" ">
          <Button :loading="saving" type="primary" @click="saveSettings">
            {{ $t('plugin.linapro-oidc-google.settings.save') }}
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
