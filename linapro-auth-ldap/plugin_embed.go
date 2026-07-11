// Package pluginauthldap is the source-plugin embedding entry for linapro-auth-ldap.
package pluginauthldap

import "embed"

// EmbeddedFiles contains plugin.yaml, frontend, and manifest resources.
//
//go:embed plugin.yaml frontend manifest
var EmbeddedFiles embed.FS
