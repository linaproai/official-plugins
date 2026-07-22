# Configuration

## Settings Fields

| Field | Name | Description |
| --- | --- | --- |
| `allow_auto_provision` | LDAP Login - Allow Auto Provision | Whether unknown directory users may be auto-provisioned. Allowed values: true, false. |
| `base_dn` | LDAP Login - Base DN | Base distinguished name used when searching for users. |
| `bind_dn` | LDAP Login - Bind DN | Service account distinguished name used to bind to the directory. |
| `bind_password` | LDAP Login - Bind Password | Service account password. Leave blank or masked to keep the existing value. |
| `connection_key` | LDAP Login - Connection Key | Stable connection identifier used by the host external-login seam. |
| `display_name` | LDAP Login - Display Name | Login entry label shown to end users. |
| `display_name_attr` | LDAP Login - Display Name Attribute | Directory attribute used for the user display name. |
| `email_attr` | LDAP Login - Email Attribute | Directory attribute used for the user email address. |
| `host` | LDAP Login - Host | Directory server hostname or IP address. |
| `port` | LDAP Login - Port | Directory server TCP port. |
| `subject_attr` | LDAP Login - Subject Attribute | Directory attribute used as the external subject identifier. |
| `tls_mode` | LDAP Login - TLS Mode | LDAP TLS mode for directory connections. |
| `user_dn_template` | LDAP Login - User DN Template | Optional DN template used to construct the user DN directly. |
| `user_filter` | LDAP Login - User Filter | LDAP filter template used to locate a user entry. |

## Notes

- Provider secrets are sensitive and should be stored through the plugin settings page only.
- Leave masked secret fields blank when keeping the existing value.

## Entry Points

| Name | Path |
| --- | --- |
| LDAP | `linapro-auth-ldap-settings` |
