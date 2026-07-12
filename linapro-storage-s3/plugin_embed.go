// Package pluginstorages3 is the source-plugin embedding entry for linapro-storage-s3.
package pluginstorages3

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
