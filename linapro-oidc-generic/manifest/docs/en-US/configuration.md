# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `allow_auto_provision` | Generic OIDC - Allow Auto Provision | Whether unknown directory users may be auto-provisioned. Allowed values: true, false. |
| `client_id` | Generic OIDC - Client ID | OAuth/OIDC client identifier issued by the identity provider. |
| `client_secret` | Generic OIDC - Client Secret | OAuth/OIDC client secret. Leave blank or masked to keep the existing value. |
| `connection_key` | Generic OIDC - Connection Key | Stable connection identifier used by the host external-login seam. |
| `default_backend_redirect` | Generic OIDC - Default Backend Redirect | Default post-login backend redirect target when backend redirect is enabled. |
| `display_name` | Generic OIDC - Display Name | Login entry label shown to end users. |
| `issuer` | Generic OIDC - Issuer | OIDC issuer URL used for discovery and token validation. |
| `redirect_url` | Generic OIDC - Redirect URL | OAuth/OIDC callback URL registered with the identity provider. |
| `scopes` | Generic OIDC - Scopes | Space-separated OAuth/OIDC scopes requested during authorization. |

## Notes

- Provider secrets are sensitive and should be stored through the plugin settings page only.
- Leave masked secret fields blank when keeping the existing value.

## Entry Points

| Name | Path |
| --- | --- |
| OIDC | `linapro-oidc-generic-settings` |
