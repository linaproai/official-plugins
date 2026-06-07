# linapro-demo-dynamic

`linapro-demo-dynamic` is the dynamic WASM plugin sample for LinaPro. It demonstrates the smallest end-to-end path for a governed runtime plugin.

## What It Demonstrates

- one menu entry rendered inside the default management workspace
- one standalone static page that does not depend on the host UI framework
- demo backend routes executed through the dynamic plugin bridge
- governed access to `runtime`, `storage`, `network`, `data`, `config`, `manifest`, `hostConfig`, `org`, and `tenant` host services through `pkg/plugin/pluginbridge/guest`
- source-compatible `Before*` precondition handlers and `After*` notification handlers are auto-discovered from backend controller methods, then write runtime debug logs for lifecycle flow inspection

## Directory Layout

```text
linapro-demo-dynamic/
  main.go
  plugin_embed.go
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## Build

Build all dynamic plugin artifacts:

```bash
make wasm
```

Build only this sample:

```bash
make wasm p=linapro-demo-dynamic
```

The runtime artifact is written to `temp/output/linapro-demo-dynamic.wasm`.

## Backend Contract

The sample exposes governed routes under `/x/linapro-demo-dynamic/api/v1` and keeps business logic inside `backend/internal/service/`. The host requires only `/x/{pluginId}`; this sample keeps `/api/v1/...` as its own route group by declaring it in `backend/plugin.go` through `RegisterRoutes`.

API DTO files under `backend/api/` stay resource-local and do not own route group prefixes. To add another group, create a separate API package such as `backend/api/dynamic/v2` or `backend/api/dynamic/interface/m1`, keep DTO paths local to that package, and add another `RegisterRoutes` binding such as `registrar.Group("/api/v2", "dynamic/v2")` or `registrar.Group("/interface/m1", "dynamic/interface/m1")`. The host will publish them as `/x/linapro-demo-dynamic/api/v2/...` or `/x/linapro-demo-dynamic/interface/m1/...`.

## Public Assets

The sample declares `public_assets` in `plugin.yaml`:

```yaml
public_assets:
  - source: frontend/pages
    mount: /
    index: index.html
```

Only files matching that declaration are served from `/x-assets/linapro-demo-dynamic/v0.1.0/...`. When the mount directory itself is requested, `index` selects the default file and falls back to `index.html` when omitted. The management workspace menu continues to use `system/plugin/dynamic-page` and passes the `/x-assets/.../mount.js` URL as the hosted resource; it does not use `/x-assets/...` as the workbench route itself.

## Host Services

The sample requests the following host services in `plugin.yaml`:

- `runtime`
- `storage`
- `network`
- `data`
- `config`
- `manifest`
- `hostConfig`
- `org`
- `tenant`

These declarations are reviewed and authorized by the host during plugin lifecycle operations.

Guest business host-service clients are imported from `lina-core/pkg/plugin/pluginbridge/guest`. The same package is also used by the sample's bridge files for protocol envelopes, route dispatch, lifecycle contracts, cron contracts, and response helpers.

The `manifest` host service example authorizes only `config/profile.yaml` and `config/config.yaml`. The `/api/v1/manifest-demo` route reads those two packaged files through `manifest.get` and the embedded page displays the returned profile and config preview, so the sample shows the full declaration-to-use flow. Runtime-effective configuration is still read through the dedicated `config` host service; SQL and i18n lifecycle resources are not included in this manifest host-service authorization example.

## Lifecycle Logging

The dynamic sample implements `BeforeInstall`, `AfterInstall`, `BeforeUpgrade`, `AfterUpgrade`, `BeforeDisable`, `AfterDisable`, `BeforeUninstall`, `AfterUninstall`, `BeforeTenantDisable`, `AfterTenantDisable`, `BeforeTenantDelete`, `AfterTenantDelete`, `BeforeInstallModeChange`, and `AfterInstallModeChange` controller methods. `linactl wasm` auto-discovers those methods and embeds lifecycle contracts in the WASM artifact. Each handler returns `ok=true` and writes a runtime log entry with the operation and available transition fields.

## Review Checklist

- `plugin.yaml` keeps metadata and host-service declarations clear.
- frontend assets match the declared plugin access mode.
- the built WASM artifact is reproducible from the source tree.
- backend logic stays inside service components instead of controllers.
