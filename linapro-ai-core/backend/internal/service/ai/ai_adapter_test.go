// This file verifies provider protocol adapters using fake HTTP servers.

package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

type providerBaseURLKey struct{}

// TestOpenAIAdapterNormalizesURLAndMapsUsage verifies OpenAI-compatible path
// normalization, reasoning effort mapping, auth headers, and usage parsing.
func TestOpenAIAdapterNormalizesURLAndMapsUsage(t *testing.T) {
	server := testOpenAIServer(t)
	svc := New(nil, nil, server.Client()).(*serviceImpl)
	result, err := svc.callOpenAI(context.Background(), &resolvedTierBinding{
		ModelName:         "unit-openai",
		EndpointBaseUrl:   server.URL + "/v1",
		EndpointSecretRef: "unit-secret",
	}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, string(aitext.ThinkingEffortHigh))
	if err != nil {
		t.Fatalf("call openai adapter: %v", err)
	}
	if result.Text != "provider ok" || result.Usage.InputTokens != 11 || result.Usage.OutputTokens != 7 {
		t.Fatalf("unexpected OpenAI adapter result: %#v", result)
	}
}

// TestProviderAdaptersStripPlatformModelSuffix verifies tool-routing suffixes
// remain platform-side and are not sent in provider protocol payloads.
func TestProviderAdaptersStripPlatformModelSuffix(t *testing.T) {
	t.Run("openai", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode OpenAI request: %v", err)
			}
			if payload["model"] != "mimo-v2.5-pro" {
				t.Fatalf("expected stripped OpenAI model, got %#v", payload["model"])
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)); err != nil {
				t.Fatalf("write OpenAI response: %v", err)
			}
		}))
		defer server.Close()

		svc := New(nil, nil, server.Client()).(*serviceImpl)
		result, err := svc.callOpenAI(context.Background(), &resolvedTierBinding{
			ModelName:       "mimo-v2.5-pro[1m]",
			EndpointBaseUrl: server.URL,
		}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, "")
		if err != nil {
			t.Fatalf("call OpenAI adapter: %v", err)
		}
		if result.Text != "ok" {
			t.Fatalf("unexpected OpenAI result: %#v", result)
		}
	})

	t.Run("anthropic", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode Anthropic request: %v", err)
			}
			if payload["model"] != "claude-sonnet" {
				t.Fatalf("expected stripped Anthropic model, got %#v", payload["model"])
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`)); err != nil {
				t.Fatalf("write Anthropic response: %v", err)
			}
		}))
		defer server.Close()

		svc := New(nil, nil, server.Client()).(*serviceImpl)
		result, err := svc.callAnthropic(context.Background(), &resolvedTierBinding{
			ModelName:       "claude-sonnet[codex][fast]",
			EndpointBaseUrl: server.URL,
		}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 32, nil, "")
		if err != nil {
			t.Fatalf("call Anthropic adapter: %v", err)
		}
		if result.Text != "ok" {
			t.Fatalf("unexpected Anthropic result: %#v", result)
		}
	})
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
		ModelName:         "unit-anthropic",
		EndpointBaseUrl:   server.URL,
		EndpointSecretRef: "unit-secret",
		ProviderName:      "Anthropic",
		CapabilityType:    CapabilityTypeText,
		DefaultEffort:     string(aitext.ThinkingEffortMax),
	}, []aitext.Message{{Role: aitext.MessageRoleSystem, Content: "sys"}}, 128, nil, string(aitext.ThinkingEffortMax))
	if err != nil {
		t.Fatalf("call anthropic adapter: %v", err)
	}
	if result.Text != "anthropic ok" || result.Usage.InputTokens != 5 || result.Usage.OutputTokens != 3 {
		t.Fatalf("unexpected Anthropic adapter result: %#v", result)
	}
}

// TestAnthropicAdapterRetriesVersionedURLAndCachesBase verifies 404 fallback
// and process-local URL cache reuse for Anthropic-compatible generation.
func TestAnthropicAdapterRetriesVersionedURLAndCachesBase(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		switch r.URL.Path {
		case "/anthropic/messages":
			http.NotFound(w, r)
		case "/anthropic/v1/messages":
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"content":[{"type":"text","text":"anthropic ok"}],"usage":{"input_tokens":5,"output_tokens":3}}`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	binding := &resolvedTierBinding{
		ModelName:         "unit-anthropic",
		Protocol:          ProtocolAnthropicCompatible,
		EndpointBaseUrl:   server.URL + "/anthropic",
		EndpointSecretRef: "unit-secret",
	}
	for i := 0; i < 2; i++ {
		result, err := svc.callAnthropic(context.Background(), binding, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 128, nil, "")
		if err != nil {
			t.Fatalf("call anthropic adapter on attempt %d: %v", i+1, err)
		}
		if result.Text != "anthropic ok" {
			t.Fatalf("unexpected Anthropic adapter result: %#v", result)
		}
	}

	mu.Lock()
	got := append([]string(nil), paths...)
	mu.Unlock()
	want := []string{"/anthropic/messages", "/anthropic/v1/messages", "/anthropic/v1/messages"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected paths: got %v want %v", got, want)
	}
}

// TestOpenAIModelListRetriesVersionedURLAndCachesBase verifies the same 404
// fallback and cache path for OpenAI-compatible model synchronization.
func TestOpenAIModelListRetriesVersionedURLAndCachesBase(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		switch r.URL.Path {
		case "/proxy/models":
			http.NotFound(w, r)
		case "/proxy/v1/models":
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"data":[{"id":"flow-model"},{"id":"remote-model"}]}`)); err != nil {
				t.Fatalf("write model list response: %v", err)
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	endpoint := &entity.ProviderEndpoint{
		Protocol:  ProtocolOpenAICompatible,
		BaseUrl:   server.URL + "/proxy",
		SecretRef: "unit-secret",
	}
	for i := 0; i < 2; i++ {
		models, err := svc.listOpenAIModels(context.Background(), endpoint)
		if err != nil {
			t.Fatalf("list OpenAI models on attempt %d: %v", i+1, err)
		}
		names := make([]string, 0, len(models))
		for _, model := range models {
			names = append(names, model.Name)
			if len(model.Capabilities) != 0 {
				t.Fatalf("OpenAI /models must not imply capabilities, got %#v", model.Capabilities)
			}
		}
		if strings.Join(names, ",") != "flow-model,remote-model" {
			t.Fatalf("unexpected model names: %v", names)
		}
	}

	mu.Lock()
	got := append([]string(nil), paths...)
	mu.Unlock()
	want := []string{"/proxy/models", "/proxy/v1/models", "/proxy/v1/models"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected paths: got %v want %v", got, want)
	}
}

// TestAnthropicAdapterDoesNotDuplicateVersionSuffix verifies already-versioned
// base URLs are requested directly and never retried with /v1/v1.
func TestAnthropicAdapterDoesNotDuplicateVersionSuffix(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		http.NotFound(w, r)
	}))
	defer server.Close()

	svc := New(nil, nil, server.Client()).(*serviceImpl)
	_, err := svc.callAnthropic(context.Background(), &resolvedTierBinding{
		ModelName:       "unit-anthropic",
		Protocol:        ProtocolAnthropicCompatible,
		EndpointBaseUrl: server.URL + "/anthropic/v1",
	}, []aitext.Message{{Role: aitext.MessageRoleUser, Content: "hello"}}, 128, nil, "")
	if !bizerr.Is(err, CodeProviderHTTPError) {
		t.Fatalf("expected structured provider HTTP error, got %v", err)
	}

	mu.Lock()
	got := append([]string(nil), paths...)
	mu.Unlock()
	want := []string{"/anthropic/v1/messages"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected paths: got %v want %v", got, want)
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
		ModelName:         "unit-openai",
		EndpointBaseUrl:   server.URL,
		EndpointSecretRef: "sk-secret-token",
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
		ModelName:       "unit-openai",
		EndpointBaseUrl: server.URL,
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
