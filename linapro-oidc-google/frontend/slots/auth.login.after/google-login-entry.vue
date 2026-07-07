<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the "Continue with Google"
// entry rendered below the login form. The host slot-registry only mounts
// the button when the linapro-oidc-google plugin is installed and enabled.
export const pluginSlotMeta = {
  order: 10,
  pluginId: 'linapro-oidc-google',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import { IconifyIcon } from '@vben/icons';

import { Button } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Google. Because it is a full-page navigation (not an XHR call),
// window.location.assign is used so cookies set by the plugin route are
// respected by the browser during the OIDC handshake.
const loginStartPath = '/portal/linapro-oidc-google/login';

function handleClick() {
  window.location.assign(loginStartPath);
}
</script>

<template>
  <div class="linapro-oidc-google-entry" data-testid="linapro-oidc-google-entry">
    <Button
      block
      class="linapro-oidc-google-entry__button"
      data-testid="linapro-oidc-google-entry-button"
      size="large"
      @click="handleClick"
    >
      <template #icon>
        <IconifyIcon
          class="linapro-oidc-google-entry__icon"
          icon="logos:google-icon"
        />
      </template>
      {{ $t('plugin.linapro-oidc-google.login.button') }}
    </Button>
  </div>
</template>

<style scoped>
.linapro-oidc-google-entry {
  display: flex;
  justify-content: center;
  width: 100%;
}

.linapro-oidc-google-entry__button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.linapro-oidc-google-entry__icon {
  width: 18px;
  height: 18px;
}
</style>
