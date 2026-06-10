// Package backend wires the dynamic demo plugin bridge handlers and
// declaration-time plugin capabilities into the Lina dynamic plugin runtime.
package backend

import (
	dynamicservice "lina-plugin-linapro-demo-dynamic/backend/internal/service/dynamic"

	"github.com/gogf/gf/v2/errors/gerror"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
)

// dynamicAPIV1GroupPrefix is the demo plugin-owned API route group prefix.
const dynamicAPIV1GroupPrefix = "/api/v1"

// RegisterPlugin declares dynamic plugin startup capabilities. The WASM
// builder parses this function by name, extracts plugin.Routes().Group calls,
// and uses those bindings when it combines API DTO path metadata into dynamic
// route contracts. Host-driven Jobs discovery can execute the same function
// with a runtime facade whose Routes facade is a no-op and whose Jobs facade
// submits governed jobs.register declarations.
//
// The plugin.Routes().Group(dynamicAPIV1GroupPrefix, "dynamic/v1") call means:
// dynamicAPIV1GroupPrefix declares the plugin-owned route group prefix
// "/api/v1", while "dynamic/v1" points to the backend/api-relative DTO package
// backend/api/dynamic/v1. DTO paths in that package, such as
// path:"/backend-summary", are therefore built as "/api/v1/backend-summary";
// the host later exposes the final public path under "/x/{pluginId}".
//
// RegisterPlugin 声明动态插件启动期能力。构建器会按方法名解析这里的
// plugin.Routes().Group 调用，并把解析到的分组绑定用于组合 API DTO 中的 path
// 元数据，生成动态路由契约。宿主 Jobs 发现期可以用运行时 facade 执行同一个函数；
// 此时 Routes facade 为 no-op，Jobs facade 会提交受治理的 jobs.register 声明。
//
// plugin.Routes().Group(dynamicAPIV1GroupPrefix, "dynamic/v1") 的含义是：
// dynamicAPIV1GroupPrefix 声明插件自有路由分组前缀 "/api/v1"；"dynamic/v1"
// 指向相对于 backend/api 的 DTO 包 backend/api/dynamic/v1。该包中 DTO 声明的
// path，例如 path:"/backend-summary"，会在构建期组合成
// "/api/v1/backend-summary"；宿主运行时再把它暴露到最终公开路径
// "/x/{pluginId}" 下。
func RegisterPlugin(plugin bridgeplugin.Declarations) error {
	if plugin == nil {
		return gerror.New("linapro-demo-dynamic dynamic plugin declaration facade is required")
	}
	// Bind backend/api/dynamic/v1 DTO routes to the plugin-owned /api/v1 group.
	// 将 backend/api/dynamic/v1 下的 DTO 路由绑定到插件自有的 /api/v1 分组。
	if err := plugin.Routes().Group(dynamicAPIV1GroupPrefix, "dynamic/v1"); err != nil {
		return err
	}
	return dynamicservice.RegisterJobs(plugin)
}
