// Package pluginoidcgeneric is the source-plugin embedding entry for the
// linapro-oidc-generic plugin. It only owns the embedded filesystem binding so
// the backend package can register plugin.yaml, frontend slot components, and
// manifest i18n resources with the host at compile time.
package pluginoidcgeneric

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend slot components, and
// manifest resources shipped with the linapro-oidc-generic source plugin.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
