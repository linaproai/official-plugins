// Package backend wires linapro-mail-imap into the host plugin registry and
// mail-core inbound SPI.
package backend

import (
	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	mailimap "lina-plugin-linapro-mail-imap"
	imapsvc "lina-plugin-linapro-mail-imap/backend/internal/service/imap"
)

const pluginID = imapsvc.PluginID

func init() {
	if err := spi.RegisterInbound(pluginID, mailcap.KindIMAP, imapsvc.Factory); err != nil {
		panic(err)
	}
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(mailimap.EmbeddedFiles)
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(err)
	}
}
