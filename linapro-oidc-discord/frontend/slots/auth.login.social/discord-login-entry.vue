<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// Platform social icon entry (Discord). Host auth.login.social renders a
// horizontal icon row under “其他登录方式”; not a full-width protocol button.
// Order is greater than the Google entry (10) so Google renders first.
export const pluginSlotMeta = {
  order: 20,
  pluginId: 'linapro-oidc-discord',
  slotKey: pluginSlotKeys.authLoginSocial,
};
</script>

<script setup lang="ts">
import { IconifyIcon } from '@vben/icons';

import { Button, Tooltip } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Discord. Full-page navigation keeps the OIDC handshake cookie-safe.
const loginStartPath = '/portal/linapro-oidc-discord/login';

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
    class="linapro-oidc-discord-entry"
    data-testid="linapro-oidc-discord-entry"
  >
    <Tooltip
      :title="$t('plugin.linapro-oidc-discord.login.button')"
      placement="top"
    >
      <Button
        class="mb-3"
        data-testid="linapro-oidc-discord-entry-button"
        shape="circle"
        type="text"
        @click="handleClick"
      >
        <IconifyIcon
          class="linapro-oidc-discord-entry__icon size-5"
          icon="logos:discord-icon"
        />
      </Button>
    </Tooltip>
  </div>
</template>
