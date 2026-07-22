# linapro-auth-ldap

Managed source plugin that provides **general-purpose enterprise LDAP / Active Directory directory login**.

English | [简体中文](README.zh-CN.md) 

Depends on `linapro-extlogin-core`. The same capability can power login entries and other product flows.

**Provider:** `ldap:default` · **Auto-provision:** default off · **TLS:** LDAPS/StartTLS (plain only on localhost)

## Install

1. Enable `linapro-extlogin-core`
2. Enable `linapro-auth-ldap`
3. Configure host/TLS/search or DN template
4. Use the LDAP login entry (for example **Continue with LDAP** on the host login page)

## Security

- Password used only for bind; never stored or logged
- Unified auth failure message
- Handoff-only session delivery to SPA
