// Package pluginstorageaws is the source-plugin embedding entry for linapro-storage-aws.
package pluginstorageaws

import "embed"

// EmbeddedFiles contains the plugin manifest, frontend pages, and resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
