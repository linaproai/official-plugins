// Package pluginstorageobs is the source-plugin embedding entry for linapro-storage-obs.
package pluginstorageobs

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
