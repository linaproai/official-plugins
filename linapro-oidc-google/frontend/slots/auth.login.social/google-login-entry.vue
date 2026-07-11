<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// Platform social icon entry (Google). Host auth.login.social renders a
// horizontal icon row under “其他登录方式”; not a full-width protocol button.
export const pluginSlotMeta = {
  order: 10,
  pluginId: 'linapro-oidc-google',
  slotKey: pluginSlotKeys.authLoginSocial,
};
</script>

<script setup lang="ts">
import { SvgGoogleIcon } from '@vben/icons';

import { Button, Tooltip } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Google. Full-page navigation keeps the OIDC handshake cookie-safe.
const loginStartPath = '/portal/linapro-oidc-google/login';

function handleClick() {
  // Echo the live SPA login path so configuration errors and OIDC callbacks
  // return to history-mode or hash-mode pages without a hard-coded router form.
  const returnTo = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  const target = new URL(loginStartPath, window.location.origin);
  target.searchParams.set('returnTo', returnTo);
  window.location.assign(target.toString());
}
</script>

<template>
  <div
    class="linapro-oidc-google-entry"
    data-testid="linapro-oidc-google-entry"
  >
    <Tooltip
      :title="$t('plugin.linapro-oidc-google.login.button')"
      placement="top"
    >
      <Button
        class="mb-3"
        data-testid="linapro-oidc-google-entry-button"
        shape="circle"
        type="text"
        @click="handleClick"
      >
        <SvgGoogleIcon class="size-5" />
      </Button>
    </Tooltip>
  </div>
</template>
