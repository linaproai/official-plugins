<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the "Continue with Discord"
// entry rendered below the login form. The host slot-registry only mounts
// the button when the linapro-oidc-discord plugin is installed and enabled.
// The order is greater than the Google entry (10) so Google renders first.
export const pluginSlotMeta = {
  order: 20,
  pluginId: 'linapro-oidc-discord',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import { IconifyIcon } from '@vben/icons';

import { Button } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Discord. Because it is a full-page navigation (not an XHR call),
// window.location.assign is used so cookies set by the plugin route are
// respected by the browser during the OIDC handshake.
const loginStartPath = '/portal/linapro-oidc-discord/login';

function handleClick() {
  window.location.assign(loginStartPath);
}
</script>

<template>
  <div
    class="linapro-oidc-discord-entry"
    data-testid="linapro-oidc-discord-entry"
  >
    <Button
      block
      class="linapro-oidc-discord-entry__button"
      data-testid="linapro-oidc-discord-entry-button"
      size="large"
      @click="handleClick"
    >
      <template #icon>
        <IconifyIcon
          class="linapro-oidc-discord-entry__icon"
          icon="logos:discord-icon"
        />
      </template>
      {{ $t('plugin.linapro-oidc-discord.login.button') }}
    </Button>
  </div>
</template>

<style scoped>
.linapro-oidc-discord-entry {
  display: flex;
  justify-content: center;
  width: 100%;
}

.linapro-oidc-discord-entry__button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.linapro-oidc-discord-entry__icon {
  width: 18px;
  height: 18px;
}
</style>
