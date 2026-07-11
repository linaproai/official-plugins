<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the Discord OIDC entry rendered
// in the host's Vben-style third-party icon row. The host slot-registry only
// mounts the button when the linapro-oidc-discord plugin is installed and enabled.
// The order is greater than the Google entry (10) so Google renders first.
export const pluginSlotMeta = {
  order: 20,
  pluginId: 'linapro-oidc-discord',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import { IconifyIcon } from '@vben/icons';

import { Button, Tooltip } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Discord. Because it is a full-page navigation (not an XHR call),
// window.location.assign is used so cookies set by the plugin route are
// respected by the browser during the OIDC handshake.
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
        class="linapro-oidc-discord-entry__button mb-3"
        data-testid="linapro-oidc-discord-entry-button"
        shape="circle"
        size="large"
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

<style scoped>
/* Match Vben ThirdPartyLogin icon button footprint (rounded, fixed hit area). */
.linapro-oidc-discord-entry__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
}
</style>
