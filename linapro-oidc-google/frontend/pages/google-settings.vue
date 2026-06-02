<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-oidc-google-settings',
  title: 'plugin:linapro-oidc-google:settings',
};
</script>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import {
  Button as AButton,
  Card as ACard,
  Form as AForm,
  FormItem as AFormItem,
  Input as AInput,
  InputPassword as AInputPassword,
  Switch as ASwitch,
  message,
} from 'ant-design-vue';

import { $t } from '#/locales';
import { requestClient } from '#/api/request';
import {
  GOOGLE_OIDC_CONSOLE_URL,
  GOOGLE_OIDC_LOGIN_ENTRY_PATH,
  getGoogleOidcSettingsApiPath,
} from '../constants';

interface RedirectRuleForm {
  redirectUrl: string;
  state: string;
}

interface GoogleOidcSettings {
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  enableBackendRedirect: boolean;
  defaultBackendRedirect: string;
  backendRedirects: string;
}

const saving = ref(false);
// Initial SSO rule list is empty by design: an empty state key plus the
// "登录成功后默认跳转" path together describe the normal post-login flow
// for the workbench, so adding a seeded rule on initialization would make
// the default login flow accidentally trigger SSO token delivery.
const redirectRules = ref<RedirectRuleForm[]>([]);
const form = ref<GoogleOidcSettings>({
  clientId: '',
  clientSecret: '',
  redirectUri: '',
  enableBackendRedirect: false,
  defaultBackendRedirect: '/dashboard/analytics',
  backendRedirects: '',
});

// computedRedirectUri reflects the actual callback URL the deployment is
// reachable at right now. The settings page treats it as read-only because
// the backend always uses the registered callback path under the current
// origin; admin can only copy it for registering in Google Cloud Console.
const computedRedirectUri = computed(() => {
  if (typeof window === 'undefined') {
    return `${GOOGLE_OIDC_LOGIN_ENTRY_PATH}/callback`;
  }
  return new URL(
    `${GOOGLE_OIDC_LOGIN_ENTRY_PATH}/callback`,
    window.location.origin,
  ).toString();
});

const loginEntryOrigin = computed(() => {
  if (typeof window === 'undefined') {
    return '';
  }
  return window.location.origin;
});

function buildStateLoginURL(state: string): string {
  const origin = loginEntryOrigin.value;
  const base = `${origin}${GOOGLE_OIDC_LOGIN_ENTRY_PATH}`;
  const trimmed = state.trim();
  if (!trimmed) {
    return base;
  }
  return `${base}?state=${encodeURIComponent(trimmed)}`;
}

async function copyToClipboard(text: string, successMessage: string) {
  if (!text) {
    return;
  }
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const helper = document.createElement('textarea');
      helper.value = text;
      helper.style.position = 'fixed';
      helper.style.opacity = '0';
      document.body.append(helper);
      helper.select();
      document.execCommand('copy');
      helper.remove();
    }
    message.success(successMessage);
  } catch {
    message.error($t('plugins.oidc.settings.copyFailed'));
  }
}

function parseRedirectRules(value: string): RedirectRuleForm[] {
  if (!value.trim()) {
    return [];
  }
  try {
    const parsed = JSON.parse(value) as Record<string, unknown>;
    return Object.entries(parsed)
      .map(([state, redirectUrl]) => ({
        redirectUrl: typeof redirectUrl === 'string' ? redirectUrl.trim() : '',
        state: state.trim(),
      }))
      .filter((item) => item.state && item.redirectUrl);
  } catch {
    return [];
  }
}

function serializeRedirectRules(rules: RedirectRuleForm[]) {
  const payload: Record<string, string> = {};
  for (const rule of rules) {
    const state = rule.state.trim();
    const redirectUrl = rule.redirectUrl.trim();
    if (!state || !redirectUrl) {
      continue;
    }
    payload[state] = redirectUrl;
  }
  return JSON.stringify(payload);
}

function addRedirectRule() {
  redirectRules.value.push({ state: '', redirectUrl: '' });
}

function removeRedirectRule(index: number) {
  redirectRules.value.splice(index, 1);
}

async function loadSettings() {
  try {
    const res = await requestClient.get<GoogleOidcSettings>(
      getGoogleOidcSettingsApiPath(),
    );
    form.value = {
      clientId: res.clientId ?? '',
      clientSecret: '',
      redirectUri: computedRedirectUri.value,
      enableBackendRedirect: res.enableBackendRedirect ?? false,
      defaultBackendRedirect: res.defaultBackendRedirect ?? '/dashboard/analytics',
      backendRedirects: res.backendRedirects ?? '',
    };
    redirectRules.value = parseRedirectRules(res.backendRedirects ?? '');
  } catch {
    message.error($t('plugins.oidc.settings.loadFailed'));
  }
}

async function saveSettings() {
  saving.value = true;
  try {
    await requestClient.put(getGoogleOidcSettingsApiPath(), {
      clientId: form.value.clientId,
      clientSecret: form.value.clientSecret,
      redirectUri: computedRedirectUri.value,
      enableBackendRedirect: form.value.enableBackendRedirect,
      defaultBackendRedirect: form.value.defaultBackendRedirect,
      backendRedirects: serializeRedirectRules(redirectRules.value),
    });
    message.success($t('plugins.oidc.settings.saveSuccess'));
  } catch {
    message.error($t('plugins.oidc.settings.saveFailed'));
  } finally {
    saving.value = false;
  }
}

onMounted(loadSettings);
</script>

<template>
  <div style="max-width: 720px; margin: 0 auto; padding: 24px">
    <ACard :title="$t('plugins.oidc.settings.instructionsCard')" style="margin-bottom: 16px">
      <ol style="padding-left: 20px; line-height: 2">
        <li>
          {{ $t('plugins.oidc.settings.stepCreateCredentialsPrefix') }}
          <a :href="GOOGLE_OIDC_CONSOLE_URL" target="_blank" rel="noopener noreferrer">{{ $t('plugins.oidc.google.consoleName') }}</a>
          {{ $t('plugins.oidc.settings.stepCreateCredentialsSuffix') }}
        </li>
        <li>{{ $t('plugins.oidc.settings.stepWebApp') }}</li>
        <li>{{ $t('plugins.oidc.settings.stepFillCredentials') }}</li>
        <li>{{ $t('plugins.oidc.settings.stepRegisterRedirect', { section: $t('plugins.oidc.google.registerSection') }) }}</li>
        <li>{{ $t('plugins.oidc.settings.stepEnabled', { label: $t('plugins.oidc.google.buttonLabel') }) }}</li>
      </ol>
    </ACard>

    <ACard :title="$t('plugins.oidc.settings.cardTitle', { provider: $t('plugins.oidc.google.displayName') })">
      <AForm class="space-y-5">
        <AFormItem>
          <div class="mb-2 font-medium">{{ $t('plugins.oidc.settings.clientIdLabel') }}</div>
          <AInput v-model:value="form.clientId" :placeholder="$t('plugins.oidc.settings.clientIdPlaceholder', { provider: $t('plugins.oidc.google.displayName') })" />
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.clientIdHelp', { console: $t('plugins.oidc.google.consoleName') }) }}</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">{{ $t('plugins.oidc.settings.clientSecretLabel') }}</div>
          <AInputPassword v-model:value="form.clientSecret" :placeholder="$t('plugins.oidc.settings.clientSecretPlaceholder', { provider: $t('plugins.oidc.google.displayName') })" />
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.clientSecretHelp') }}</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">{{ $t('plugins.oidc.settings.redirectUriLabel') }}</div>
          <div class="flex items-center gap-2">
            <AInput :value="computedRedirectUri" readonly class="flex-1" />
            <AButton @click="copyToClipboard(computedRedirectUri, $t('plugins.oidc.settings.redirectUriCopied'))">{{ $t('plugins.oidc.settings.copy') }}</AButton>
          </div>
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.redirectUriHelp', { section: $t('plugins.oidc.google.registerSection') }) }}</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">{{ $t('plugins.oidc.settings.defaultRedirectLabel') }}</div>
          <AInput v-model:value="form.defaultBackendRedirect" placeholder="/dashboard/analytics" />
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.defaultRedirectHelp') }}</div>
        </AFormItem>

        <AFormItem>
          <div class="flex items-center gap-3">
            <span class="font-medium">{{ $t('plugins.oidc.settings.enableSsoLabel') }}</span>
            <ASwitch v-model:checked="form.enableBackendRedirect" :checked-children="$t('plugins.oidc.settings.enabledText')" :un-checked-children="$t('plugins.oidc.settings.disabledText')" />
          </div>
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.enableSsoHelp') }}</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 flex items-center justify-between">
            <span class="font-medium">{{ $t('plugins.oidc.settings.rulesTitle') }}</span>
          </div>
          <div class="space-y-3 rounded-lg border border-dashed border-gray-300 p-4">
            <div v-if="redirectRules.length === 0" class="text-xs text-gray-500">
              {{ $t('plugins.oidc.settings.rulesEmpty') }}
            </div>
            <div
              v-for="(rule, index) in redirectRules"
              :key="index"
              class="space-y-2"
            >
              <div class="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
                <AInput v-model:value="rule.state" :placeholder="$t('plugins.oidc.settings.statePlaceholder')" />
                <AInput v-model:value="rule.redirectUrl" :placeholder="$t('plugins.oidc.settings.receiverPlaceholder')" />
                <AButton danger type="text" @click="removeRedirectRule(index)">{{ $t('plugins.oidc.settings.delete') }}</AButton>
              </div>
              <div class="flex items-center gap-2 text-xs text-gray-500">
                <span>{{ $t('plugins.oidc.settings.entryLabel') }}</span>
                <code class="flex-1 truncate rounded bg-gray-100 px-2 py-1">{{ buildStateLoginURL(rule.state) }}</code>
                <AButton size="small" :disabled="!rule.state.trim()" @click="copyToClipboard(buildStateLoginURL(rule.state), $t('plugins.oidc.settings.stateEntryCopied'))">{{ $t('plugins.oidc.settings.copy') }}</AButton>
              </div>
            </div>
            <AButton block type="dashed" @click="addRedirectRule">{{ $t('plugins.oidc.settings.addRule') }}</AButton>
          </div>
          <div class="mt-1 text-xs text-gray-500">{{ $t('plugins.oidc.settings.rulesHelp') }}</div>
        </AFormItem>

        <AFormItem>
          <AButton type="primary" :loading="saving" @click="saveSettings">{{ $t('plugins.oidc.settings.save') }}</AButton>
        </AFormItem>
      </AForm>
    </ACard>
  </div>
</template>
