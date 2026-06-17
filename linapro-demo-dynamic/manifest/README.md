# Manifest Resources

This directory stores install-time and optional mock resources for `linapro-demo-dynamic`.

- `config/config.example.yaml`: documents configurable runtime keys for local development and operator review.
- `config/config.yaml`: provides packaged default runtime config values for this dynamic plugin artifact.
- `config/profile.yaml`: provides a small manifest-only profile used by the manifest host-service demos to verify `manifest.get`.
- `sql/001-linapro-demo-dynamic-records.sql`: creates the plugin-owned demo table during install.
- `sql/mock-data/001-linapro-demo-dynamic-mock-data.sql`: provides optional local demo records.
- `i18n/`: stores plugin-owned runtime translation resources.

The dynamic plugin declares a `manifest` host service in `plugin.yaml` for `config/profile.yaml` and `config/config.yaml` only. `manifest.get` reads these packaged files as raw resources and does not replace `Plugins().Config()`, SQL, or i18n lifecycle pipelines.
