# wasm: Build all dynamic wasm plugins, or use p=<plugin-id> for one plugin
# wasm: 构建全部或指定 dynamic wasm 插件；可通过 `p=<plugin-id>` 定向构建
.PHONY: wasm
wasm:
	@go run ../../hack/tools/linactl wasm p="$(p)" out="$(out)"

# ctrl: Generate controller scaffolding for one plugin backend; requires p=<plugin-id>
# ctrl: 为指定插件后端生成控制器骨架；需要 p=<plugin-id>
.PHONY: ctrl
ctrl:
	@$(if $(p),go run ../../hack/tools/linactl ctrl p="$(p)",$(error p=<plugin-id> is required))

# dao: Generate DAO/DO/Entity files for one plugin backend; requires p=<plugin-id>
# dao: 为指定插件后端生成 DAO/DO/Entity 文件；需要 p=<plugin-id>
.PHONY: dao
dao:
	@$(if $(p),go run ../../hack/tools/linactl dao p="$(p)",$(error p=<plugin-id> is required))
