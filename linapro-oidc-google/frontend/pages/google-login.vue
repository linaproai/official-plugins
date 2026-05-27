<script lang="ts">
export const pluginPageMeta = {
  routePath: '/auth/google',
  title: 'Google Login',
};
</script>

<script setup lang="ts">
import { IconifyIcon } from '@vben/icons';

function handleContinue() {
  // Send the browser to the plugin's OAuth initiation endpoint. The path
  // matches the backend route registered under /api/v1/auth/* so the URL
  // works without any extra reverse-proxy configuration on /auth/*.
  window.location.href = '/api/v1/auth/google?state=google';
}
</script>

<template>
  <div style="display: flex; align-items: center; justify-content: center; min-height: 100%; padding: 24px">
    <div style="width: min(100%, 440px); padding: 32px; background: rgba(255,255,255,0.95); border: 1px solid #e5e7eb; border-radius: 16px">
      <h1 style="margin: 0 0 12px; font-size: 22px; font-weight: 700">Continue with Google</h1>
      <p style="margin: 0 0 24px; color: #4b5563; line-height: 1.7">
        Click below to start the Google login flow. The plugin will redirect
        you to Google for authorization and then bring you back to the
        workbench.
      </p>
      <a-button block size="large" type="primary" @click="handleContinue">
        <template #icon>
          <IconifyIcon icon="logos:google-icon" />
        </template>
        Continue with Google
      </a-button>
    </div>
  </div>
</template>
