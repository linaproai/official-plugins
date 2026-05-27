<script lang="ts">
export const pluginPageMeta = {
  routePath: 'linapro-oidc-discord-settings',
  title: 'Discord 登录配置',
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
  DISCORD_OAUTH_CONSOLE_URL,
  DISCORD_OAUTH_LOGIN_ENTRY_PATH,
  getDiscordOauthSettingsApiPath,
} from '../constants';

interface RedirectRuleForm {
  redirectUrl: string;
  state: string;
}

interface DiscordOAuthSettings {
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  enableBackendRedirect: boolean;
  defaultBackendRedirect: string;
  backendRedirects: string;
  enabled: boolean;
}

const saving = ref(false);
// Initial SSO rule list is empty by design: an empty state key plus the
// "登录成功后默认跳转" path together describe the normal post-login flow
// for the workbench, so adding a seeded rule on initialization would make
// the default login flow accidentally trigger SSO token delivery.
const redirectRules = ref<RedirectRuleForm[]>([]);
const form = ref<DiscordOAuthSettings>({
  clientId: '',
  clientSecret: '',
  redirectUri: '',
  enableBackendRedirect: false,
  defaultBackendRedirect: '/dashboard',
  backendRedirects: '',
  enabled: false,
});

// computedRedirectUri reflects the actual callback URL the deployment is
// reachable at right now. The settings page treats it as read-only because
// the backend always uses the registered callback path under the current
// origin; admin can only copy it for registering in Discord Developer Portal.
const computedRedirectUri = computed(() => {
  if (typeof window === 'undefined') {
    return `${DISCORD_OAUTH_LOGIN_ENTRY_PATH}/callback`;
  }
  return new URL(
    `${DISCORD_OAUTH_LOGIN_ENTRY_PATH}/callback`,
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
  const base = `${origin}${DISCORD_OAUTH_LOGIN_ENTRY_PATH}`;
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
    message.error('复制失败');
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
    const res = await requestClient.get<DiscordOAuthSettings>(
      getDiscordOauthSettingsApiPath(),
    );
    form.value = {
      clientId: res.clientId ?? '',
      clientSecret: '',
      redirectUri: computedRedirectUri.value,
      enableBackendRedirect: res.enableBackendRedirect ?? false,
      defaultBackendRedirect: res.defaultBackendRedirect ?? '/dashboard',
      backendRedirects: res.backendRedirects ?? '',
      enabled: res.enabled ?? false,
    };
    redirectRules.value = parseRedirectRules(res.backendRedirects ?? '');
  } catch {
    message.error('加载配置失败');
  }
}

async function saveSettings() {
  saving.value = true;
  try {
    await requestClient.put(getDiscordOauthSettingsApiPath(), {
      clientId: form.value.clientId,
      clientSecret: form.value.clientSecret,
      redirectUri: computedRedirectUri.value,
      enableBackendRedirect: form.value.enableBackendRedirect,
      defaultBackendRedirect: form.value.defaultBackendRedirect,
      backendRedirects: serializeRedirectRules(redirectRules.value),
      enabled: form.value.enabled,
    });
    message.success('配置已保存');
  } catch {
    message.error('保存配置失败');
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
          <a :href="DISCORD_OAUTH_CONSOLE_URL" target="_blank" rel="noopener noreferrer">{{ $t('plugins.oidc.discord.consoleName') }}</a>
          {{ $t('plugins.oidc.settings.stepCreateCredentialsSuffix') }}
        </li>
        <li>{{ $t('plugins.oidc.settings.stepWebApp') }}</li>
        <li>{{ $t('plugins.oidc.settings.stepFillCredentials') }}</li>
        <li>{{ $t('plugins.oidc.settings.stepRegisterRedirect', { section: $t('plugins.oidc.discord.registerSection') }) }}</li>
        <li>{{ $t('plugins.oidc.settings.stepEnabled', { label: $t('plugins.oidc.discord.buttonLabel') }) }}</li>
      </ol>
    </ACard>

    <ACard title="Discord 登录配置">
      <AForm class="space-y-5">
        <AFormItem>
          <div class="mb-2 font-medium">Client ID</div>
          <AInput v-model:value="form.clientId" placeholder="请输入 Discord Application Client ID" />
          <div class="mt-1 text-xs text-gray-500">从 Discord Developer Portal 获取的应用 Client ID</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">Client Secret</div>
          <AInputPassword v-model:value="form.clientSecret" placeholder="请输入 Discord Application Client Secret" />
          <div class="mt-1 text-xs text-gray-500">留空表示不修改当前密钥；如需更新请重新输入完整 Client Secret</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">Redirect URI</div>
          <div class="flex items-center gap-2">
            <AInput :value="computedRedirectUri" readonly class="flex-1" />
            <AButton @click="copyToClipboard(computedRedirectUri, 'Redirect URI 已复制')">复制</AButton>
          </div>
          <div class="mt-1 text-xs text-gray-500">根据当前站点自动生成，不可编辑。请把这个完整地址注册到 Discord Developer Portal 的「Redirects」</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 font-medium">登录成功后默认跳转</div>
          <AInput v-model:value="form.defaultBackendRedirect" placeholder="/dashboard" />
          <div class="mt-1 text-xs text-gray-500">登录成功且未命中下方 SSO 规则时，SPA 内部跳转到这个落地页（默认 /dashboard）</div>
        </AFormItem>

        <AFormItem>
          <div class="flex items-center gap-3">
            <span class="font-medium">启用 SSO Token 投递</span>
            <ASwitch v-model:checked="form.enableBackendRedirect" checked-children="启用" un-checked-children="禁用" />
          </div>
          <div class="mt-1 text-xs text-gray-500">启用后，当 OAuth 入口附带的 state 命中下方规则，会直接把 accessToken/refreshToken 作为 query 投递到规则配置的外部 URL，不再进入 SPA。state 为空或未命中时走"登录成功后默认跳转"</div>
        </AFormItem>

        <AFormItem>
          <div class="mb-2 flex items-center justify-between">
            <span class="font-medium">SSO Token 投递规则（state → 外部 URL）</span>
          </div>
          <div class="space-y-3 rounded-lg border border-dashed border-gray-300 p-4">
            <div v-if="redirectRules.length === 0" class="text-xs text-gray-500">
              当前未配置 SSO 规则。默认登录会走"登录成功后默认跳转"，不会触发 SSO 投递。
            </div>
            <div
              v-for="(rule, index) in redirectRules"
              :key="index"
              class="space-y-2"
            >
              <div class="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
                <AInput v-model:value="rule.state" placeholder="state，例如 partner-a" />
                <AInput v-model:value="rule.redirectUrl" placeholder="外部 URL，例如 https://app.example.com/sso/receive" />
                <AButton danger type="text" @click="removeRedirectRule(index)">删除</AButton>
              </div>
              <div class="flex items-center gap-2 text-xs text-gray-500">
                <span>对应登录入口：</span>
                <code class="flex-1 truncate rounded bg-gray-100 px-2 py-1">{{ buildStateLoginURL(rule.state) }}</code>
                <AButton size="small" :disabled="!rule.state.trim()" @click="copyToClipboard(buildStateLoginURL(rule.state), 'state 登录入口 URL 已复制')">复制</AButton>
              </div>
            </div>
            <AButton block type="dashed" @click="addRedirectRule">新增 SSO 规则</AButton>
          </div>
          <div class="mt-1 text-xs text-gray-500">页面自动把多行规则组装成 JSON 存到后端。state 未命中时走"登录成功后默认跳转"</div>
        </AFormItem>

        <AFormItem>
          <div class="flex items-center gap-3">
          <span class="font-medium">启用 Discord 登录</span>
          <ASwitch v-model:checked="form.enabled" checked-children="启用" un-checked-children="禁用" />
          </div>
        </AFormItem>

        <AFormItem>
          <AButton type="primary" :loading="saving" @click="saveSettings">保存配置</AButton>
        </AFormItem>
      </AForm>
    </ACard>
  </div>
</template>
