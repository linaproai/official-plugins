// Package bridge provides the linapro-ai-core dynamic-plugin guest SDK for
// owner-aware AI host services. It owns AI request codecs, manifest declaration
// helpers, and typed guest clients while all provider implementations and
// business storage remain inside the linapro-ai-core backend internals.
package bridge

import (
	"strings"

	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-plugin-linapro-ai-core/backend/cap/aicap/spi"
)

// AllMethods returns every AI method currently published by the owner
// descriptor.
func AllMethods() []string {
	descriptors := spi.MethodDescriptors()
	methods := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if method := strings.TrimSpace(descriptor.Method); method != "" {
			methods = append(methods, method)
		}
	}
	return methods
}

// TextMethods returns the text AI methods normally needed by a dynamic plugin
// that only generates text and reads text method status.
func TextMethods() []string {
	return []string{
		spi.MethodTextGenerate,
		spi.MethodTextStatusGet,
	}
}

// StatusMethods returns cross-capability AI status methods for dynamic plugins
// that need method-level degradation checks.
func StatusMethods() []string {
	return []string{
		spi.MethodTextStatusGet,
		spi.MethodStatusesBatchGet,
	}
}

// HostServiceSpec returns one owner-aware ai.v1 host-service declaration.
// Passing no methods declares every method currently published by the owner.
func HostServiceSpec(methods ...string) *protocol.HostServiceSpec {
	return &protocol.HostServiceSpec{
		Owner:   spi.OwnerPluginID,
		Service: spi.ServiceAI,
		Version: spi.VersionV1,
		Methods: normalizeMethods(methods),
	}
}

// normalizeMethods trims declaration methods and falls back to the full owner
// method set when the caller passes no explicit methods.
func normalizeMethods(methods []string) []string {
	if len(methods) == 0 {
		return AllMethods()
	}
	normalized := make([]string, 0, len(methods))
	for _, method := range methods {
		if value := strings.TrimSpace(method); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}
