// This file embeds the plugin manifest, frontend pages, and manifest resources.

package mailpop3

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and manifest resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
