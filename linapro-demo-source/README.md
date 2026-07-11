# linapro-demo-source

`linapro-demo-source` is the source-plugin sample for `LinaPro`. It demonstrates an in-repo plugin that is discovered by the host, installed explicitly from plugin management, and mounted into the default workspace.

## Directory Layout

```text
linapro-demo-source/
  plugin.yaml
  plugin_embed.go
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/
    pages/
  manifest/
    sql/
    sql/uninstall/
```

## What This Sample Demonstrates

- install SQL under `manifest/sql/` creates the plugin-owned table `plugin_linapro_demo_source_record`
- mock SQL under `manifest/sql/mock-data/` provides optional demo records for local sample data
- uninstall SQL under `manifest/sql/uninstall/` drops the plugin-owned table when the user confirms storage purge
- the sample page in `frontend/pages/sidebar-entry.vue` performs CRUD against the plugin-owned table and supports attachment upload/download
- plugin-owned attachment objects are stored through `pluginhost.Services.Storage()` under plugin logical paths such as `demo-record-files/...`
- disabling the plugin hides menus and routes but keeps table data and stored files
- uninstalling the plugin opens a confirmation dialog that lets the user choose whether to purge plugin-owned table data and stored files
- lifecycle callbacks log `BeforeInstall`, `AfterInstall`, `BeforeUpgrade`, `Upgrade`, `AfterUpgrade`, `BeforeDisable`, `AfterDisable`, `BeforeUninstall`, `AfterUninstall`, tenant lifecycle callbacks, and install-mode callbacks so developers can observe the source-plugin lifecycle flow

## Manifest Scope

`plugin.yaml` keeps the plugin metadata, menu declarations, and button permissions. Pages and SQL assets still follow directory conventions instead of being duplicated in metadata.

The sample declares `distribution: managed`, so it remains an ordinary plugin that is explicitly installed, enabled, upgraded, disabled, and uninstalled through plugin management. Use `distribution: builtin` only for registered source plugins that are project components and must be installed, enabled, and safely upgraded by host startup.

`plugin.yaml` does not declare source-plugin HTTP routes. Workspace navigation still comes from `menus`, while backend routes are registered by plugin code.

## Backend Integration

- implement backend entry points under `backend/`
- keep service logic under `backend/internal/service/`
- keep plugin-local ORM codegen output under `backend/internal/dao` and `backend/internal/model/{do,entity}` when the plugin accesses database tables
- register plugin APIs under `registrar.Routes().APIPrefix()`, which resolves to `/x/linapro-demo-source`; the sample appends `/api/v1` as its own route convention
- define `g.Meta path` as plugin-local business paths only (for example `/ping`, `/summary`, `/records`); do not re-embed `/plugins/linapro-demo-source/` because the host prefix already scopes the plugin
- keep public pages, portals, static routes, or plugin-owned fallback handlers on non-reserved paths instead of putting them under `/x`
- register install, upgrade, disable, uninstall, tenant, and install-mode lifecycle callbacks through the source-plugin registration entry used by the host build
- inject `Storage()` explicitly from `registrar.Services()` and use it for attachment save, download, replacement, deletion, and optional uninstall purge
- keep plugin-owned cleanup logic in the plugin service so uninstall can optionally purge `Storage()` objects before uninstall SQL drops the table

## Front-end Integration

- front-end pages are discovered from the plugin's `frontend/` directory according to repository conventions
- the sample page keeps the existing summary card and adds a `VXE` grid plus modal form for record CRUD
- the uninstall choice is surfaced by the host plugin-management page, not by the plugin page itself

## Public Assets

Source plugins may declare public static asset directories in `plugin.yaml` `public_assets`. Declared files are served from `/x-assets/{plugin-id}/{version}/...`, but this sample keeps its workbench page under the normal `frontend/pages/` discovery flow and does not require host-served public assets.

Do not use `/plugin-assets`; that legacy path is not supported.

## SQL Conventions

- install SQL lives under `manifest/sql/`
- uninstall SQL lives under `manifest/sql/uninstall/`
- install SQL should be idempotent so reinstall after a non-purge uninstall can preserve data
- uninstall SQL should be paired with plugin cleanup hooks when `Storage()` objects must be removed together with table data

## Attachment Storage Boundary

| Scenario | Sample behavior |
| --- | --- |
| Save or replace an attachment | The backend stores the object through `storagecap.Service.Put` and persists only the logical path, original file name, and record metadata. |
| Download an attachment | The backend reads the logical path through `storagecap.Service.Get` and streams the returned reader; it does not expose or open a host filesystem path. |
| Delete a record or remove an attachment | The backend deletes the stored object through `storagecap.Service.Delete`. |
| Purge on uninstall | When the host uninstall policy requests storage purge, the lifecycle callback receives plugin-scoped services, binds the recorded tenant scope, deletes recorded attachment paths, and then performs a bounded prefix cleanup for remaining `demo-record-files/` objects. |

## Review Checklist

- metadata stays minimal and accurate
- host wiring remains explicit
- pages follow directory conventions
- plugin-owned SQL is kept separate from host SQL
- disable keeps plugin-owned data intact
- uninstall supports both retain-data and purge-data paths
