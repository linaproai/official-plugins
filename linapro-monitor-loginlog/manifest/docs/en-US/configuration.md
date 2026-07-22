# Configuration

This plugin does not expose extra runtime settings in the current release. After it is enabled, use the entry points declared in `plugin.yaml`.

## Notes

- Most monitoring plugins use built-in collection or host event paths.
- Retention and cleanup behavior is governed by the plugin implementation and declared operations.

## Entry Points

| Name | Path |
| --- | --- |
| Login History | `/monitor/loginlog` |
