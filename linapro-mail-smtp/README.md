# Mail - SMTP

`linapro-mail-smtp` sends email through a standard SMTP server for `linapro-mail-core`.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- `kind=smtp` outbound transport SPI registration
- SMTP-specific connection probing and sending logic

It does not own Connection/Account tables; those belong to `linapro-mail-core`.

## Host Boundary

`linapro-mail-core` owns the Connection/Account persistence, the public `mailcap` contract, and transport SPI resolution. This plugin only implements the SMTP transport and depends on `linapro-mail-core`.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-mail-smtp` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Install mode | `global` |
