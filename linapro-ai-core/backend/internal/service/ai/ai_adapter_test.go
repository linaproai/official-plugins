// This file verifies provider protocol adapters using fake HTTP servers.

package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/ai/aitext"
)

type providerBaseURLKey struct{}

// TestOpenAIAdapterNormalizesURLAndMapsUsage verifies OpenAI-compatible path
// normalization, reasoning effort mapping, auth headers, and usage parsing.
func TestOpenAIAdapterNormalizesURLAndMapsUsage(t *testing.T) {
	server := testOpenAIServer(t)
	svc := New(nil, nil, server.Client()).(*serviceImpl)
	result, err := svc.callOpenAI(context.Background(), &resolvedTierBinding{
		ModelName:       "unit-openai",
		OpenaiBaseUrl:   server.URL + "/v1",
		ApiKeySecretRef: "unit-secret",
	}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, string(aitext.ThinkingEffortHigh))
	if err != nil {
		t.Fatalf("call openai adapter: %v", err)
	}
	if result.Text != "provider ok" || result.Usage.InputTokens != 11 || result.Usage.OutputTokens != 7 {
		t.Fatalf("unexpected OpenAI adapter result: %#v", result)
	}
}

// TestAnthropicAdapterMapsThinkingEffort verifies Anthropic-compatible message
// conversion and controlled thinking budget mapping.
func TestAnthropicAdapterMapsThinkingEffort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		thinking, ok := payload["thinking"].(map[string]any)
		if !ok || int(thinking["budget_tokens"].(float64)) != 32768 {
			t.Fatalf("unexpected thinking payload: %#v", payload["thinking"])
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"content":[{"type":"text","text":"anthropic ok"}],"usage":{"input_tokens":5,"output_tokens":3}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	result, err := svc.callAnthropic(context.Background(), &resolvedTierBinding{
		ModelName:        "unit-anthropic",
		AnthropicBaseUrl: server.URL,
		ApiKeySecretRef:  "unit-secret",
		SupportedEfforts: "max",
		SupportsThinking: enabledYes,
		MaxOutputTokens:  128,
		ProviderName:     "Anthropic",
		CapabilityType:   CapabilityTypeText,
		DefaultEffort:    string(aitext.ThinkingEffortMax),
	}, []aitext.Message{{Role: aitext.MessageRoleSystem, Content: "sys"}}, 128, nil, string(aitext.ThinkingEffortMax))
	if err != nil {
		t.Fatalf("call anthropic adapter: %v", err)
	}
	if result.Text != "anthropic ok" || result.Usage.InputTokens != 5 || result.Usage.OutputTokens != 3 {
		t.Fatalf("unexpected Anthropic adapter result: %#v", result)
	}
}

// TestAdapterErrorsAreRedacted verifies provider errors never expose secret markers.
func TestAdapterErrorsAreRedacted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad authorization sk-secret-token with full prompt body", http.StatusUnauthorized)
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	_, err := svc.callOpenAI(context.Background(), &resolvedTierBinding{
		ModelName:       "unit-openai",
		OpenaiBaseUrl:   server.URL,
		ApiKeySecretRef: "sk-secret-token",
	}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, "")
	if !bizerr.Is(err, CodeProviderHTTPError) {
		t.Fatalf("expected structured provider HTTP error, got %v", err)
	}
	for _, forbidden := range []string{"sk-secret-token", "full prompt body"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("expected redacted provider error, got %v", err)
		}
	}
	if err == nil {
		t.Fatalf("expected redacted provider error, got %v", err)
	}
}

// TestOpenAIAdapterRejectsUnsupportedExtendedEffort verifies protocol-specific
// effort rejection happens before the external HTTP call.
func TestOpenAIAdapterRejectsUnsupportedExtendedEffort(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	_, err := svc.callOpenAI(context.Background(), &resolvedTierBinding{
		ModelName:     "unit-openai",
		OpenaiBaseUrl: server.URL,
	}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, string(aitext.ThinkingEffortMax))
	if !bizerr.Is(err, CodeThinkingEffortUnsupported) || called {
		t.Fatalf("expected unsupported effort before HTTP call, err=%v called=%v", err, called)
	}
}

func testOpenAIServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" || r.URL.Path == "/models" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"data":[{"id":"flow-model"},{"id":"remote-model"}]}`)); err != nil {
				t.Fatalf("write model list response: %v", err)
			}
			return
		}
		if r.URL.Path != "/v1/chat/completions" && r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if !strings.Contains(r.Header.Get("Authorization"), "unit-secret") {
			t.Fatalf("missing bearer auth header")
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["reasoning_effort"] == string(aitext.ThinkingEffortHigh) && payload["model"] != "unit-openai" {
			t.Fatalf("unexpected OpenAI payload: %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"provider ok"}}],"usage":{"prompt_tokens":11,"completion_tokens":7}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func testProviderBaseURL(ctx context.Context) string {
	if value, ok := ctx.Value(providerBaseURLKey{}).(string); ok {
		return value
	}
	return "http://127.0.0.1:1/v1"
}
