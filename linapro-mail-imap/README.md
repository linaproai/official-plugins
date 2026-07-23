# Mail - IMAP

`linapro-mail-imap` retrieves email from IMAP mailboxes for `linapro-mail-core`.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- `kind=imap` inbound transport SPI registration
- IMAP-specific mailbox fetching logic

It does not own Connection/Account tables; those belong to `linapro-mail-core`.

## Host Boundary

`linapro-mail-core` owns the Connection/Account persistence, the public `mailcap` contract, and transport SPI resolution. This plugin only implements the IMAP transport and depends on `linapro-mail-core`.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-mail-imap` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Install mode | `global` |
