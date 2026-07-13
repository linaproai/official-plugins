// Package backend wires linapro-mail-pop3 into the host plugin registry and
// mail-core inbound SPI.
package backend

import (
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	mailpop3 "lina-plugin-linapro-mail-pop3"
	pop3svc "lina-plugin-linapro-mail-pop3/backend/internal/service/pop3"
)

const pluginID = pop3svc.PluginID

func init() {
	if err := spi.RegisterInbound(pluginID, mailcap.KindPOP3, pop3svc.Factory); err != nil {
		panic(err)
	}
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(mailpop3.EmbeddedFiles)
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}
