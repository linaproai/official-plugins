module lina-plugins

go 1.25.0

require (
	lina-plugin-linapro-content-notice v0.0.0
	lina-plugin-linapro-demo-source v0.0.0
	lina-plugin-linapro-monitor-loginlog v0.0.0
	lina-plugin-linapro-monitor-online v0.0.0
	lina-plugin-linapro-monitor-operlog v0.0.0
	lina-plugin-linapro-monitor-server v0.0.0
	lina-plugin-linapro-oidc-discord v0.0.0
	lina-plugin-linapro-oidc-google v0.0.0
	lina-plugin-linapro-ops-demo-guard v0.0.0
	lina-plugin-linapro-org-core v0.0.0
	lina-plugin-linapro-tenant-core v0.0.0
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/emirpasic/gods/v2 v2.0.0-alpha // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogf/gf/v2 v2.10.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grokify/html-strip-tags-go v0.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mssola/useragent v1.0.0 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.0.9 // indirect
	github.com/olekukonko/tablewriter v1.1.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shirou/gopsutil/v4 v4.26.3 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/excelize/v2 v2.10.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lina-core v0.0.0 // indirect
)

replace lina-core => ../lina-core

replace lina-plugin-linapro-content-notice => ./linapro-content-notice

replace lina-plugin-linapro-ops-demo-guard => ./linapro-ops-demo-guard

replace lina-plugin-linapro-demo-source => ./linapro-demo-source

replace lina-plugin-linapro-monitor-loginlog => ./linapro-monitor-loginlog

replace lina-plugin-linapro-monitor-online => ./linapro-monitor-online

replace lina-plugin-linapro-monitor-operlog => ./linapro-monitor-operlog

replace lina-plugin-linapro-monitor-server => ./linapro-monitor-server

replace lina-plugin-linapro-tenant-core => ./linapro-tenant-core

replace lina-plugin-linapro-org-core => ./linapro-org-core

replace lina-plugin-linapro-oidc-google => ./linapro-oidc-google

replace lina-plugin-linapro-oidc-discord => ./linapro-oidc-discord
