# 认证 CI 夹具

`official-plugins` GitHub Actions 用于真实 LDAP / OIDC 登录覆盖的轻量后端。

| 路径 | 作用 |
|------|------|
| `ldap-mock/` | 最小 LDAP 目录（对种子用户`alice`的 simple bind + search） |
| `oidc-mock/` | 最小 OIDC 提供方（discovery、authorize 自动登录、token + PKCE S256、JWKS） |
| `docker-compose.auth.yml` | 可选真实 OpenLDAP，供本地试验（CI 不依赖） |

集成测试与插件代码同目录，并使用`//go:build integration`：

- `linapro-auth-ldap/backend/internal/service/ldapauth/ldapauth_integration_test.go`
- `linapro-oidc-generic/backend/internal/service/oauth/oauth_integration_test.go`

宿主会话签发通过`extlogin.Service`打桩；目录绑定与 OIDC 协议流量为真实网络 I/O。

## 本地运行

```bash
# 在 apps/lina-plugins（或 monorepo 中 plugins 位于 apps/lina-plugins）
export GOWORK=off
go run ./hack/ci/ldap-mock -listen 127.0.0.1:1389 &
go run ./hack/ci/oidc-mock -listen 127.0.0.1:18080 &

# 然后在 monorepo 布局且 go.work 覆盖 plugins + lina-core 时：
go test ./linapro-auth-ldap/backend/internal/service/ldapauth/ -tags=integration -count=1 -v
go test ./linapro-oidc-generic/backend/internal/service/oauth/ -tags=integration -count=1 -v
```
