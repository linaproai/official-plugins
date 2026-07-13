// Package backend wires linapro-mail-smtp into the host plugin registry and
// mail-core outbound SPI.
package backend

import (
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	mailsmtp "lina-plugin-linapro-mail-smtp"
	smtpsvc "lina-plugin-linapro-mail-smtp/backend/internal/service/smtp"
)

const pluginID = smtpsvc.PluginID

func init() {
	if err := spi.RegisterOutbound(pluginID, mailcap.KindSMTP, smtpsvc.Factory); err != nil {
		panic(err)
	}
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(mailsmtp.EmbeddedFiles)
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}
