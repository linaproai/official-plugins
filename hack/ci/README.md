# Auth CI Fixtures

Lightweight backends used by `official-plugins` GitHub Actions for real LDAP and OIDC login coverage.

| Path | Role |
|------|------|
| `ldap-mock/` | Minimal LDAP directory (simple bind + search for seed user `alice`) |
| `oidc-mock/` | Minimal OIDC provider (discovery, authorize auto-login, token + PKCE S256, JWKS) |
| `docker-compose.auth.yml` | Optional real OpenLDAP for local experiments (CI does not require it) |

Integration tests live next to plugin code with `//go:build integration`:

- `linapro-auth-ldap/backend/internal/service/ldapauth/ldapauth_integration_test.go`
- `linapro-oidc-generic/backend/internal/service/oauth/oauth_integration_test.go`

Host session minting is stubbed through `extlogin.Service`; directory bind and OIDC protocol traffic are real network I/O.

## Local run

```bash
# from apps/lina-plugins (or monorepo with plugins at apps/lina-plugins)
export GOWORK=off
go run ./hack/ci/ldap-mock -listen 127.0.0.1:1389 &
go run ./hack/ci/oidc-mock -listen 127.0.0.1:18080 &

# then, with monorepo layout and go.work covering plugins + lina-core:
go test ./linapro-auth-ldap/backend/internal/service/ldapauth/ -tags=integration -count=1 -v
go test ./linapro-oidc-generic/backend/internal/service/oauth/ -tags=integration -count=1 -v
```
