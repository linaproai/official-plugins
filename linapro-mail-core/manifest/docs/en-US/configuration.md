# Configuration

This plugin does not expose extra runtime settings in the current release. After it is enabled, use the entry points declared in `plugin.yaml`.

## Notes

- Transport plugins share the mail capability published by `linapro-mail-core`.
- Connection secrets should be stored through Mail settings and left blank when unchanged.

## Entry Points

| Name | Path |
| --- | --- |
| Mail | `linapro-mail-core-settings` |
