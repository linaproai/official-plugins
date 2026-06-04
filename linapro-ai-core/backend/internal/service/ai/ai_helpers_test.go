// This file verifies Smart Center helper behavior that does not require
// database fixtures.

package ai

import "testing"

func TestMaskSecretRefPreservesRecognizablePrefixAndSuffix(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "openai key", value: "sk-1234567890", want: "sk-**********90"},
		{name: "non prefixed key", value: "plain-secret", want: "plain-**********et"},
		{name: "already masked", value: "sk-**********90", want: "sk-**********90"},
		{name: "short secret", value: "ab", want: "**"},
		{name: "empty secret", value: " ", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := maskSecretRef(test.value); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestCapabilityMethodDefaultsAndCacheKey(t *testing.T) {
	if got := normalizeCapabilityMethod(""); got != CapabilityMethodGenerate {
		t.Fatalf("expected default method %q, got %q", CapabilityMethodGenerate, got)
	}

	generateKey := tierCacheKey(CapabilityTypeText, CapabilityMethodGenerate, TierCodeBasic)
	otherKey := tierCacheKey(CapabilityTypeText, "summarize", TierCodeBasic)
	if generateKey == otherKey {
		t.Fatalf("expected method-specific cache keys, got %q", generateKey)
	}
}

func TestProviderRequestModelNameStripsPlatformSuffix(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		want      string
	}{
		{name: "single suffix", modelName: "mimo-v2.5-pro[1m]", want: "mimo-v2.5-pro"},
		{name: "multiple suffixes", modelName: "mimo-v2.5-pro[1m][coding]", want: "mimo-v2.5-pro"},
		{name: "trims spaces", modelName: "  claude-sonnet[agent]  ", want: "claude-sonnet"},
		{name: "middle bracket unchanged", modelName: "model[official]-2026", want: "model[official]-2026"},
		{name: "empty suffix unchanged", modelName: "model[]", want: "model[]"},
		{name: "malformed suffix unchanged", modelName: "model[a[b]]", want: "model[a[b]]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := providerRequestModelName(test.modelName); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}
