// Package pluginoidcdiscord is the source-plugin embedding entry for the
// linapro-oidc-discord reference plugin. It only owns the embedded filesystem
// binding so the backend package can register plugin.yaml, frontend slot
// components, and manifest i18n resources with the host at compile time.
package pluginoidcdiscord

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend slot components, and
// manifest resources shipped with the linapro-oidc-discord source plugin.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
