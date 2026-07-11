<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the Google OIDC entry rendered
// in the host's Vben-style third-party icon row. The host slot-registry only
// mounts the button when the linapro-oidc-google plugin is installed and enabled.
export const pluginSlotMeta = {
  order: 10,
  pluginId: 'linapro-oidc-google',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import { SvgGoogleIcon } from '@vben/icons';

import { Button, Tooltip } from 'ant-design-vue';

// loginStartPath is the plugin's browser-facing route that redirects the
// user to Google. Because it is a full-page navigation (not an XHR call),
// window.location.assign is used so cookies set by the plugin route are
// respected by the browser during the OIDC handshake.
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
        class="linapro-oidc-google-entry__button mb-3"
        data-testid="linapro-oidc-google-entry-button"
        shape="circle"
        size="large"
        type="text"
        @click="handleClick"
      >
        <SvgGoogleIcon class="size-5" />
      </Button>
    </Tooltip>
  </div>
</template>

<style scoped>
/* Match Vben ThirdPartyLogin icon button footprint (rounded, fixed hit area). */
.linapro-oidc-google-entry__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  padding: 0;
}
</style>
