// Package backend wires the dynamic demo plugin bridge handlers and build-time
// route group declarations into the Lina dynamic plugin runtime.
package backend

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// dynamicAPIV1GroupPrefix is the demo plugin-owned API route group prefix.
const dynamicAPIV1GroupPrefix = "/api/v1"

// RegisterRoutes declares dynamic plugin route groups for the WASM builder.
// The function is a build-time convention entrypoint: the builder parses this
// method by name, extracts registrar.Group calls, and uses those bindings when
// it combines API DTO path metadata into dynamic route contracts. It is not a
// runtime HTTP handler and is not called by the Lina host through normal Go
// references.
//
// The registrar.Group(dynamicAPIV1GroupPrefix, "dynamic/v1") call means:
// dynamicAPIV1GroupPrefix declares the plugin-owned route group prefix
// "/api/v1", while "dynamic/v1" points to the backend/api-relative DTO package
// backend/api/dynamic/v1. DTO paths in that package, such as
// path:"/backend-summary", are therefore built as "/api/v1/backend-summary";
// the host later exposes the final public path under "/x/{pluginId}".
//
// RegisterRoutes 声明动态插件在 WASM 构建阶段使用的路由分组。该方法是构建期
// 约定入口：构建器会按方法名解析这里的 registrar.Group 调用，并把解析到的分组
// 绑定用于组合 API DTO 中的 path 元数据，生成动态路由契约。它不是运行时 HTTP
// 处理器，也不会被宿主通过普通 Go 引用直接调用。
//
// registrar.Group(dynamicAPIV1GroupPrefix, "dynamic/v1") 的含义是：
// dynamicAPIV1GroupPrefix 声明插件自有路由分组前缀 "/api/v1"；"dynamic/v1"
// 指向相对于 backend/api 的 DTO 包 backend/api/dynamic/v1。该包中 DTO 声明的
// path，例如 path:"/backend-summary"，会在构建期组合成
// "/api/v1/backend-summary"；宿主运行时再把它暴露到最终公开路径
// "/x/{pluginId}" 下。
func RegisterRoutes(registrar pluginbridge.DynamicRouteRegistrar) error {
	if registrar == nil {
		return gerror.New("linapro-demo-dynamic route registrar is required")
	}
	// Bind backend/api/dynamic/v1 DTO routes to the plugin-owned /api/v1 group.
	// 将 backend/api/dynamic/v1 下的 DTO 路由绑定到插件自有的 /api/v1 分组。
	return registrar.Group(dynamicAPIV1GroupPrefix, "dynamic/v1")
}
