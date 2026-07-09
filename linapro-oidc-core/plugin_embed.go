// Package pluginoidccore embeds the linapro-oidc-core plugin manifest and
// lifecycle SQL assets for source-plugin registration. The plugin ships no
// frontend pages: it is a capability-only plugin providing the external
// identity storage and provider engine.
package pluginoidccore

import "embed"

// EmbeddedFiles contains the plugin manifest and convention-based SQL and i18n assets.
//
//go:embed plugin.yaml manifest
var EmbeddedFiles embed.FS
