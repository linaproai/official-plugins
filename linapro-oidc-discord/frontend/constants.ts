// Package frontend constants centralizes Discord OAuth2 plugin route and API
// path fragments so the settings page does not duplicate hard-coded strings.
// Keep this file tiny and plugin-local: it should only contain stable paths
// that are shared by the Discord frontend entrypoints.

import { pluginApiPath } from '#/api/request';

export function getDiscordOauthSettingsApiPath() {
  return pluginApiPath(
    'linapro-oidc-discord',
    'plugin/linapro-oidc-discord/settings',
  );
}

export const DISCORD_OAUTH_AUTHORIZATION_URL =
  'https://discord.com/oauth2/authorize';

// Discord OAuth2 callback path served by the plugin under the host's
// /api/v1/* namespace so admin guidance and Discord Developer Portal
// redirect URIs match the actual backend route.
export const DISCORD_OAUTH_CALLBACK_URL = '/api/v1/auth/discord/callback';

// Discord OAuth2 login entry path that the settings page exposes for
// administrators who need to build per-state SSO entry URLs (one URL per
// configured state key under SSO Token 投递规则).
export const DISCORD_OAUTH_LOGIN_ENTRY_PATH = '/api/v1/auth/discord';

// External console URL operators visit to create OAuth2 applications. The
// settings page links to this URL from the instructions card.
export const DISCORD_OAUTH_CONSOLE_URL = 'https://discord.com/developers/applications';
