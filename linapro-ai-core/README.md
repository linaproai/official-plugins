# LinaPro AI Core Plugin

`apps/lina-plugins/linapro-ai-core` is the official source plugin for `AI` providers, provider models, capability tiers, invocation logs, and the plugin-owned `AI` domain capability contract.

The plugin is intentionally outside `apps/lina-core`: `lina-core` owns plugin kernel governance, descriptor registration, dependency checks, owner-aware dynamic routing, authorization snapshots, audit, and cache invalidation. `linapro-ai-core` owns the `AI` public contract, provider `SPI`, dynamic guest SDK, descriptor helper, implementation, and version policy.

## Plugin Metadata

| Field | Value |
| --- | --- |
| Plugin ID | `linapro-ai-core` |
| Type | `source` |
| Distribution | `managed` |
| Scope | `platform_only` |
| Install mode | `global` |
| `i18n` | Enabled for `en-US` and `zh-CN` |

## Ownership Boundaries

| Boundary | Path | Responsibility |
| --- | --- | --- |
| Public `AI` consumer contract | `backend/cap/aicap` | Typed `AI` service, sub-capability DTOs, named method constants, status projections, public errors, descriptor metadata, and version policy |
| Provider `SPI` | `backend/cap/aicap/spi` | Provider factory interfaces, typed factory environment, provider registration helper, and descriptor helper |
| Dynamic guest SDK | `backend/cap/aicap/bridge` | Owner-aware `hostServices` declaration helper, payload codec, host-call client, and dynamic error mapping |
| Implementation | `backend/internal/service/ai` | Provider routing, model and tier resolution, invocation logging, cache usage, and external provider adapters |
| Management API | `backend/api` and `backend/internal/controller` | Provider, model, tier, and invocation-log management endpoints |
| Resources | `manifest/sql` and `manifest/i18n` | Plugin-owned schema, uninstall SQL, menu labels, plugin labels, and localized errors |

`backend/cap/aicap` must not import `backend/internal/**`, expose DAO, DO, Entity, provider secrets, private route tables, private cache snapshots, or invocation-log internals.

## Source Plugin Consumption

Source plugins that consume `AI` must import the owner contract from `lina-plugin-linapro-ai-core/backend/cap/aicap` or a subpackage such as `aicap/aitext`. They must not import `lina-core/pkg/plugin/capability/aicap` as the production owner contract, and must not import this plugin's `backend/internal/**`.

Consumers must declare the owner dependency in `plugin.yaml` and keep their `go.mod` dependency aligned:

```yaml
dependencies:
  plugins:
    - id: linapro-ai-core
      version: ">=0.1.0"
```

Source callers should consume typed services such as `AI().Text()` or an injected `aitext.Service`. They must not use a weak `Invoke(method, payload)` style gateway as the normal Go contract.

## Dynamic Plugin Consumption

Dynamic plugins call `AI` through this plugin's bridge SDK and owner-aware `hostServices` declarations. The SDK only declares, encodes, and calls the generic host-call path; runtime authorization, dependency checks, owner enablement, audit, and dispatch still happen in `lina-core`.

```yaml
dependencies:
  plugins:
    - id: linapro-ai-core
      version: ">=0.1.0"

hostServices:
  - service: ai
    owner: linapro-ai-core
    version: v1
    methods:
      - text.generate
      - text.method_status.get
```

Dynamic plugins must not encode owner identity into `service`, and must not use `dependencies.capabilities` or an optional dependency to request `AI`.

## Provider and Version Policy

`linapro-ai-core` publishes the `ai.v1` descriptor for methods that currently have a real runtime path: `text.generate`, `text.method_status.get`, and `ai.methods.status.batch_get`. Multimodal DTO packages remain under `backend/cap/aicap/*` for source contracts and future provider bindings, but those methods are not published for authorization until a real provider path is wired. Provider registration uses `aicap.ProviderDescriptor` to wrap the typed text factory into a generic descriptor with a runtime invoker that dispatches through `aicap.Service`. `pluginhost` only receives the descriptor and must not provide an `AI`-specific facade such as `ProvideAIText`.

Capability family IDs use the plugin-owned form `plugin.linapro-ai-core.ai.<family>.v1` (for example `plugin.linapro-ai-core.ai.text.v1`). They identify status, audit, and contract families and are distinct from dynamic host-service identity (`owner` + `service=ai` + `version=v1` + method).

Breaking changes to the public contract require a plugin version update and an owner descriptor version change. The plugin version and descriptor version are the compatibility boundary used by source dependency checks, dynamic manifest validation, authorization snapshots, and upgrade previews.

## Governance Notes

Every owner capability call must preserve caller plugin ID, actor, tenant, execution source, method authorization, resource declarations when present, and audit context. Public responses must not expose provider keys, internal routing configuration, raw provider request or response bodies, or another plugin's invocation logs.

The plugin has `i18n.enabled: true`, so user-visible labels, menu text, API documentation source text, owner authorization display names, and localized errors are maintained in `manifest/i18n/<locale>` and `manifest/i18n/<locale>/apidoc` when API documentation resources are added.

`AI` provider, model, tier, and invocation-log tables belong to this plugin's `manifest/sql`. `lina-core` must not add AI provider business tables when this capability evolves.

## Testing

Backend package tests belong near the changed `backend/cap/aicap` and `backend/internal/service/ai` packages. Plugin-owned E2E tests belong under:

```text
apps/lina-plugins/linapro-ai-core/hack/tests/e2e/
apps/lina-plugins/linapro-ai-core/hack/tests/pages/
apps/lina-plugins/linapro-ai-core/hack/tests/support/
```

Run plugin-scoped E2E coverage through the host runner:

```bash
pnpm -C hack/tests test:module -- plugin:linapro-ai-core
```
