<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the LDAP directory login entry.
export const pluginSlotMeta = {
  order: 40,
  pluginId: 'linapro-auth-ldap',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, reactive } from 'vue';

import { useVbenForm, useVbenModal, VbenButton, z } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { message } from 'ant-design-vue';

import { $t } from '#/locales';
import { useAuthStore } from '#/store';

const authStore = useAuthStore();

const formSchema = computed((): VbenFormSchema[] => {
  const usernameLabel = $t('plugin.linapro-auth-ldap.login.usernameLabel');
  const passwordLabel = $t('plugin.linapro-auth-ldap.login.passwordLabel');
  return [
    {
      component: 'VbenInput',
      componentProps: {
        autocomplete: 'username',
        'data-testid': 'linapro-auth-ldap-username',
        placeholder: usernameLabel,
      },
      fieldName: 'username',
      label: usernameLabel,
      // Host-standard per-field required copy (not a shared credentials blurb).
      rules: z.string().trim().min(1, {
        message: $t('ui.formRules.required', [usernameLabel]),
      }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        autocomplete: 'current-password',
        'data-testid': 'linapro-auth-ldap-password',
        placeholder: passwordLabel,
      },
      fieldName: 'password',
      label: passwordLabel,
      rules: z.string().min(1, {
        message: $t('ui.formRules.required', [passwordLabel]),
      }),
    },
  ];
});

const [LoginForm, formApi] = useVbenForm(
  reactive({
    commonConfig: {
      // Match host AuthenticationLogin: placeholder-only fields on auth surfaces.
      componentProps: {
        class: 'w-full',
      },
      hideLabel: true,
      hideRequiredMark: true,
    },
    schema: formSchema,
    showDefaultActions: false,
  }),
);

const [LoginModal, modalApi] = useVbenModal({
  centered: true,
  fullscreenButton: false,
  onClosed: async () => {
    await formApi.resetForm();
  },
  onConfirm: submit,
});

async function submit() {
  const { valid } = await formApi.validate();
  if (!valid) {
    return;
  }

  const values = await formApi.getValues<{
    password: string;
    username: string;
  }>();
  modalApi.lock(true);
  try {
    const res = await fetch('/portal/linapro-auth-ldap/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        password: values.password,
        username: values.username.trim(),
      }),
    });
    const body = await res.json().catch(() => ({}));
    const data = body?.data ?? body;
    const handoff = data?.handoff ?? data?.Handoff;
    if (!res.ok || !handoff) {
      const code =
        body?.code || body?.message || 'PLUGIN_AUTH_LDAP_AUTH_FAILED';
      message.error(String(code));
      return;
    }
    await modalApi.close();
    await authStore.completeExternalLoginFromHandoff(String(handoff));
  } catch {
    message.error($t('plugin.linapro-auth-ldap.login.failed'));
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <div
    class="linapro-auth-ldap-entry w-full"
    data-testid="linapro-auth-ldap-entry"
  >
    <!--
      Match host AuthenticationLogin primary button metrics (h-9 / text-sm)
      via VbenButton defaults; outline keeps primary CTA emphasis.
    -->
    <VbenButton
      class="linapro-auth-ldap-entry__button w-full"
      data-testid="linapro-auth-ldap-entry-button"
      type="button"
      variant="outline"
      @click="modalApi.open()"
    >
      <span class="inline-flex min-w-0 items-center justify-center gap-2">
        <IconifyIcon
          class="size-4 shrink-0"
          icon="mdi:folder-account-outline"
        />
        <span class="truncate">
          {{ $t('plugin.linapro-auth-ldap.login.button') }}
        </span>
      </span>
    </VbenButton>

    <LoginModal
      :cancel-text="$t('common.cancel')"
      :confirm-text="$t('common.login')"
      :title="$t('plugin.linapro-auth-ldap.login.modalTitle')"
      class="w-[440px] max-w-[calc(100vw-2rem)]"
    >
      <div @keydown.enter.prevent="submit">
        <LoginForm />
      </div>
    </LoginModal>
  </div>
</template>
