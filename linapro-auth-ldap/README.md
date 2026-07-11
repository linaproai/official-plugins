# linapro-auth-ldap

Managed source plugin for LDAP/Active Directory sign-in via a login-page form.
Depends on `linapro-extlogin-core`.

**Provider:** `ldap:default` · **Auto-provision:** default off · **TLS:** LDAPS/StartTLS (plain only on localhost)

## Install

1. Enable `linapro-extlogin-core`
2. Enable `linapro-auth-ldap`
3. Configure host/TLS/search or DN template
4. Use **Continue with LDAP** on the login page

## Security

- Password used only for bind; never stored or logged
- Unified auth failure message
- Handoff-only session delivery to SPA
