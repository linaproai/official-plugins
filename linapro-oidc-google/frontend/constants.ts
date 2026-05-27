// Package frontend constants centralizes Google OIDC plugin route and API
// path fragments so the settings page does not duplicate hard-coded strings.
// Keep this file tiny and plugin-local: it should only contain stable paths
// that are shared by the Google frontend entrypoints.

import { pluginApiPath } from '#/api/request';

export function getGoogleOidcSettingsApiPath() {
  return pluginApiPath(
    'linapro-oidc-google',
    'plugin/linapro-oidc-google/settings',
  );
}

export const GOOGLE_OIDC_AUTHORIZATION_URL =
  'https://accounts.google.com/o/oauth2/v2/auth';

// Google OAuth2 callback path served by the plugin under the host's
// /api/v1/* namespace so admin guidance and Google Cloud Console
// redirect URIs match the actual backend route.
export const GOOGLE_OIDC_CALLBACK_URL = '/api/v1/auth/google/callback';

// Google OAuth2 login entry path that the settings page exposes for
// administrators who need to build per-state SSO entry URLs (one URL per
// configured state key under SSO Token 投递规则).
export const GOOGLE_OIDC_LOGIN_ENTRY_PATH = '/api/v1/auth/google';

// External console URL operators visit to create OAuth2 credentials. The
// settings page links to this URL from the instructions card.
export const GOOGLE_OIDC_CONSOLE_URL = 'https://console.cloud.google.com';
