# LinaPro Plugins

`apps/lina-plugins/` is the first-party plugin workspace for LinaPro.

LinaPro positions `apps/lina-core` as the stable full-stack framework host. The host keeps universal framework capabilities, governance, and plugin extension surfaces, while business modules, operational pages, demo capabilities, and optional domain features live here as plugins. This keeps the host reusable and avoids binding core contracts to a specific management workspace page.

The workspace currently contains source plugins compiled into the host, plus one dynamic `WASM` plugin example for runtime delivery.

## Plugin Inventory

| Plugin | Type | Scope | Install mode | Capability |
|--------|------|-------|--------------|------------|
| `linapro-tenant-core` | `source` | `platform_only` | `global` | Tenant entities, member relationships, tenant resolution, and tenant lifecycle governance |
| `linapro-org-core` | `source` | `tenant_aware` | `global` | Department management and post management |
| `linapro-content-notice` | `source` | `tenant_aware` | `tenant_scoped` | Notice announcement management |
| `linapro-monitor-online` | `source` | `tenant_aware` | `tenant_scoped` | Online user query and forced logout governance |
| `linapro-monitor-server` | `source` | `platform_only` | `global` | Server monitor collection, cleanup, and query |
| `linapro-monitor-operlog` | `source` | `tenant_aware` | `tenant_scoped` | Operation log persistence and governance pages |
| `linapro-monitor-loginlog` | `source` | `tenant_aware` | `tenant_scoped` | Login log persistence and governance pages |
| `linapro-ops-demo-guard` | `source` | `tenant_aware` | `global` | Demo-environment read-only protection and global write-operation interception |
| `linapro-oidc-google` | `source` | `platform_only` | `global` | Google / Gmail `OAuth2` login entry and `OIDC` provider metadata |
| `linapro-oidc-discord` | `source` | `platform_only` | `global` | Discord `OAuth2` login entry and `OIDC` provider metadata |
| `linapro-demo-source` | `source` | `tenant_aware` | `tenant_scoped` | Source plugin example for menu pages, public routes, and protected routes |
| `linapro-demo-dynamic` | `dynamic` | `tenant_aware` | `tenant_scoped` | Dynamic `WASM` plugin example for embedded menu pages, plugin-owned `SQL` table `CRUD`, and standalone static pages |

The root `go.mod` and `lina-plugins.go` wire the source plugins that are compiled with the host. `linapro-demo-dynamic` is intentionally not wired as a source plugin; it is the runtime plugin reference used by the `WASM` build and lifecycle flow.

## Workspace Files

| Path | Purpose |
|------|---------|
| `go.mod` | Local Go workspace module for source plugin compilation checks |
| `lina-plugins.go` | Explicit source plugin import registry for host compilation |
| `Makefile` | Plugin workspace build and code generation entry; `make wasm p=<plugin-id>`, `make ctrl p=<plugin-id>`, and `make dao p=<plugin-id>` delegate to `linactl` |
| `package.json` | Frontend workspace metadata for source plugin packages |
| `<plugin-id>/plugin.yaml` | Plugin manifest, metadata, menus, install mode, `i18n`, assets, dependencies, and host service declarations |
| `<plugin-id>/Makefile` | Plugin-local code generation wrapper that includes the shared root `hack/makefiles/plugin.codegen.mk` target fragment |
| `<plugin-id>/README.md` | English plugin-level guide |
| `<plugin-id>/README.zh-CN.md` | Chinese plugin-level guide |

## Repository Mounting

The main `linapro` repository mounts this workspace at `apps/lina-plugins` as a Git submodule:

```bash
git submodule update --init --recursive
```

The configured SSH remote is:

```text
git@github.com:linaproai/official-plugins.git
```

## Plugin Directory Contract

Each plugin directory is owned by the plugin. Lifecycle resources, frontend pages, backend code, `SQL` assets, `i18n` resources, and tests should stay inside `apps/lina-plugins/<plugin-id>/`.

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/                  API DTOs, route contracts, and metadata
    internal/
      controller/         Plugin request handling and response projection
      service/            Plugin business orchestration and domain logic
      dao/                Plugin-local generated DAO objects when database access exists
      model/do/           Plugin-local generated DO objects when database access exists
      model/entity/       Plugin-local generated entity objects when database access exists
    hack/config.yaml      Plugin-local GoFrame code generation config when DAO generation exists
    plugin.go             Backend registration, route registration, lifecycle entry, or dynamic bridge entry
  frontend/pages/         Plugin-owned pages or public static assets
  manifest/
    sql/                  Install SQL assets
    sql/mock-data/        Optional mock or demo SQL assets
    sql/uninstall/        Optional uninstall SQL assets
    i18n/<locale>/        Plugin i18n resources
  hack/tests/             Optional plugin-owned E2E tests, page objects, and helpers
  go.mod                  Plugin-local Go module
  Makefile                Plugin-local code generation wrapper
  plugin.yaml             Plugin manifest
  plugin_embed.go         Embedded asset registration entry
  README.md               English guide
  README.zh-CN.md         Chinese guide
```

Plugin root `Makefile` files are thin wrappers around the repository-level
`hack/makefiles/plugin.codegen.mk` fragment. The shared fragment derives the
target backend from the including plugin directory, so plugin `Makefile` files
must not hard-code `apps/lina-plugins/<plugin-id>/backend`. Run `make ctrl` or
`make dao` inside a plugin directory to use that plugin's
`backend/hack/config.yaml`, or run `make ctrl p=<plugin-id>` and
`make dao p=<plugin-id>` from `apps/lina-plugins/`.

`backend/internal/service/` is the only valid location for plugin business services. Do not create `backend/service/`. Dynamic plugins keep the same `backend/api/`, `backend/plugin.go`, `backend/internal/controller/`, and `backend/internal/service/` shape; their bridge files only adapt `WASM` and `pluginbridge` protocols. Guest business capability clients must come from `lina-core/pkg/plugin/capability/guest`, not from the `pluginbridge` root package.

## Source Plugins

Source plugins are compiled with `apps/lina-core` through explicit registration. They are suitable for first-party framework capabilities that should be delivered with the host build but still remain outside the host core domain.

Source plugin development rules:

1. Create or update `apps/lina-plugins/<plugin-id>/`.
2. Keep plugin metadata, menus, page mounts, lifecycle resources, `SQL` assets, and `i18n` assets in `plugin.yaml` and `manifest/`.
3. Keep backend implementation under `backend/`, with business logic in `backend/internal/service/`.
4. Keep frontend pages under `frontend/pages/` or declare public asset directories through `plugin.yaml` `public_assets`.
5. Register source plugins explicitly in `apps/lina-plugins/lina-plugins.go` when they must be compiled into the host.

## Dynamic Plugins

Dynamic plugins are delivered as runtime-managed `WASM` artifacts. Use `linapro-demo-dynamic/` as the reference for upload, install, enable, disable, uninstall, `hostServices`, public static assets, and plugin-owned data access through governed host services.

Build all dynamic plugins, or one plugin with `p=<plugin-id>`:

```bash
make -C apps/lina-plugins wasm
make -C apps/lina-plugins wasm p=linapro-demo-dynamic
```

Dynamic plugins must declare `type: dynamic` in `plugin.yaml`, keep `main.go` and `go.mod` as the guest build entry, use `hostServices` to describe runtime capabilities and resource boundaries, and import runtime, storage, data, cache, config, notify, cron, and related business host-service clients from `lina-core/pkg/plugin/capability/guest`.

## Host and Plugin Boundary

The host owns stable framework surfaces and top-level catalogs such as `dashboard`, `platform`, `org`, `content`, `monitor`, `setting`, `scheduler`, `extension`, and `developer`. Plugins choose their own mount points through `plugin.yaml` `parent_key`. When the referenced parent already exists, such as a stable host catalog, the host resolves it during sync. The host also owns a small set of on-demand catalogs, such as `auth-provider`; the host owns their name, icon, and sort, materializes them when a plugin first mounts under them, and removes them once the last child plugin is uninstalled. A `parent_key` that names a parent the host neither has nor owns is still rejected to avoid orphaned menu trees.

Plugin-owned tables, menus, pages, hooks, cron jobs, public assets, and lifecycle resources stay in the plugin directory. Host code should depend on stable plugin service seams and published packages rather than hard-coding plugin-specific page structures or menu composition details.

## Routes and Public Assets

Source plugin `HTTP` routes are registered by plugin backend code. Do not declare public routes, portal routes, workspace `API` routes, or route groups in `plugin.yaml`.

Use `registrar.Routes().APIPrefix()` for source plugin APIs. The returned prefix is `/x/{plugin-id}`. Segments after that prefix are plugin-defined route content, such as `/api/v1`, `/api/v2`, or `/interface/m1`.

`plugin.yaml` `menus` is the source of truth for management workspace navigation and permissions. Registering an `HTTP` route does not create menus, permission nodes, `OpenAPI` entries, or workspace route metadata.

Plugins may declare public static assets through `plugin.yaml` `public_assets`. The host serves declared assets from `/x-assets/{plugin-id}/{version}/...` and treats each declaration as the plugin author's publication boundary.

## Testing

Plugin-owned Playwright coverage belongs under:

```text
apps/lina-plugins/<plugin-id>/hack/tests/e2e/
apps/lina-plugins/<plugin-id>/hack/tests/pages/
apps/lina-plugins/<plugin-id>/hack/tests/support/
```

Run a single plugin test scope through the host test runner:

```bash
pnpm -C hack/tests test:module -- plugin:<plugin-id>
```

Use plugin-local `api_contract_test.go` and Go package tests for backend contract checks where applicable.

## Version Upgrades

When an installed source plugin bumps `plugin.yaml` `version`, discovery does not silently replace the effective host version.

- The current effective version remains pinned in `sys_plugin.version` and `release_id`.
- The higher discovered source version is stored as a prepared `sys_plugin_release`.
- The host must process the source plugin through the supported plugin workspace update flow before startup is allowed to continue.
- If the update is skipped, startup fails fast and reports the plugins that still require attention.

## References

- `apps/lina-plugins/linapro-demo-source/README.md`
- `apps/lina-plugins/linapro-demo-dynamic/README.md`
