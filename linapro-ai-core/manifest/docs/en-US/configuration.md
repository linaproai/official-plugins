# Configuration

This plugin does not expose extra runtime settings in the current release. After it is enabled, use the entry points declared in `plugin.yaml`.

## Notes

- Configure providers, endpoints, model identities, and capability tiers from the `AI Hub` pages.
- Store provider secrets only through the plugin settings workflow; masked secrets should be left blank when unchanged.

## Entry Points

| Name | Path |
| --- | --- |
| Providers | `/ai/providers` |
| Models | `/ai/models` |
| AI Tiers | `/ai/tiers` |
| Request Logs | `/ai/invocations` |
