// Package pluginstorageoss is the source-plugin embedding entry for linapro-storage-oss.
package pluginstorageoss

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
