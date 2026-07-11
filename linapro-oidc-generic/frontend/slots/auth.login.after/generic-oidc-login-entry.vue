<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

// pluginSlotMeta identifies this component as the generic OIDC entry.
export const pluginSlotMeta = {
  order: 30,
  pluginId: 'linapro-oidc-generic',
  slotKey: pluginSlotKeys.authLoginAfter,
};
</script>

<script setup lang="ts">
import { VbenButton } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

const loginStartPath = '/portal/linapro-oidc-generic/login';

function handleClick() {
  const returnTo = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  const target = new URL(loginStartPath, window.location.origin);
  target.searchParams.set('returnTo', returnTo);
  window.location.assign(target.toString());
}
</script>

<template>
  <div
    class="linapro-oidc-generic-entry w-full"
    data-testid="linapro-oidc-generic-entry"
  >
    <!--
      Match host AuthenticationLogin primary button metrics (h-9 / text-sm)
      via VbenButton defaults; outline keeps primary CTA emphasis.
    -->
    <VbenButton
      class="linapro-oidc-generic-entry__button w-full"
      data-testid="linapro-oidc-generic-entry-button"
      type="button"
      variant="outline"
      @click="handleClick"
    >
      <span class="inline-flex min-w-0 items-center justify-center gap-2">
        <IconifyIcon class="size-4 shrink-0" icon="mdi:shield-key-outline" />
        <span class="truncate">
          {{ $t('plugin.linapro-oidc-generic.login.button') }}
        </span>
      </span>
    </VbenButton>
  </div>
</template>
