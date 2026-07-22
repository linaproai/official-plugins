# Mail - POP3

`linapro-mail-pop3` retrieves email from POP3 mailboxes for `linapro-mail-core`.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- `kind=pop3` inbound transport SPI registration
- POP3-specific mailbox fetching logic

It does not own Connection/Account tables; those belong to `linapro-mail-core`.

## Host Boundary

`linapro-mail-core` owns the Connection/Account persistence, the public `mailcap` contract, and transport SPI resolution. This plugin only implements the POP3 transport and depends on `linapro-mail-core`.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-mail-pop3` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Install mode | `global` |
