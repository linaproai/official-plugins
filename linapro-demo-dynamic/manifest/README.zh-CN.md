# Manifest Resources

该目录用于存放`linapro-demo-dynamic`的安装期资源和可选`mock`资源。

- `config/config.example.yaml`：记录本地开发和运维审查可参考的运行期配置键。
- `config/config.yaml`：提供当前动态插件产物打包的默认运行期配置值。
- `config/profile.yaml`：提供一个用于`manifest.get`宿主服务演示的小型打包 profile。
- `sql/001-linapro-demo-dynamic-records.sql`：安装时创建插件自有示例表。
- `sql/mock-data/001-linapro-demo-dynamic-mock-data.sql`：提供本地样例数据使用的可选演示记录。
- `i18n/`：存放插件自有运行时翻译资源。

动态插件会在`plugin.yaml`中声明`manifest`宿主服务，并仅授权读取`config/profile.yaml`和`config/config.yaml`。`manifest.get`只读取这些打包文件的原始资源，不替代`Plugins().Config()`、SQL 或 i18n 生命周期管线。
