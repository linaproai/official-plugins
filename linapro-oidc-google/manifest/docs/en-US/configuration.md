# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `allow_auto_provision` | Google OIDC - Allow Auto Provision | Whether unknown directory users may be auto-provisioned. Allowed values: true, false. |
| `backend_redirects` | Google OIDC - Backend Redirects | Allowed backend redirect targets, typically one entry per line or a delimited list. |
| `client_id` | Google OIDC - Client ID | OAuth/OIDC client identifier issued by the identity provider. |
| `client_secret` | Google OIDC - Client Secret | OAuth/OIDC client secret. Leave blank or masked to keep the existing value. |
| `default_backend_redirect` | Google OIDC - Default Backend Redirect | Default post-login backend redirect target when backend redirect is enabled. |
| `enable_backend_redirect` | Google OIDC - Enable Backend Redirect | Whether backend redirect allow-list enforcement is enabled. Allowed values: true, false. |
| `enable_one_tap` | Google OIDC - Enable One Tap | Whether Google One Tap is enabled on the login page. Allowed values: true, false. |
| `redirect_url` | Google OIDC - Redirect URL | OAuth/OIDC callback URL registered with the identity provider. |

## Notes

- Provider secrets are sensitive and should be stored through the plugin settings page only.
- Leave masked secret fields blank when keeping the existing value.

## Entry Points

| Name | Path |
| --- | --- |
| Google | `linapro-oidc-google-settings` |
