// This file implements narrow OpenAI-compatible and Anthropic-compatible HTTP
// adapters for text generation and public model-list synchronization.

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// adapterResult carries one provider text-generation result.
type adapterResult struct {
	Text           string
	Usage          aitext.Usage
	LatencyMs      int
	ThinkingEffort string
}

// remoteModel carries public model identity and provider-confirmed capabilities.
type remoteModel struct {
	Name         string
	Capabilities []remoteModelCapability
}

// remoteModelCapability carries one confirmed remote model capability.
type remoteModelCapability struct {
	CapabilityType   string
	CapabilityMethod string
	InputModalities  []string
	OutputModalities []string
}

// listRemoteModels returns public model identities from one selected provider endpoint.
func (s *serviceImpl) listRemoteModels(ctx context.Context, endpoint *entity.ProviderEndpoint) ([]remoteModel, error) {
	if endpoint == nil {
		return nil, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	switch endpoint.Protocol {
	case protocolOpenAI, protocolOpenAICompatible:
		return s.listOpenAIModels(ctx, endpoint)
	case protocolAnthropic, protocolAnthropicCompatible:
		return s.listAnthropicModels(ctx, endpoint)
	default:
		return nil, bizerr.NewCode(CodeRequestInvalid)
	}
}

// callProvider executes one text-generation request against the selected protocol.
func (s *serviceImpl) callProvider(
	ctx context.Context,
	binding *resolvedTierBinding,
	messages []aitext.Message,
	maxOutputTokens int,
	temperature *float64,
	effort string,
) (*adapterResult, error) {
	switch binding.Protocol {
	case protocolOpenAI, protocolOpenAICompatible:
		return s.callOpenAI(ctx, binding, messages, maxOutputTokens, temperature, effort)
	case protocolAnthropic, protocolAnthropicCompatible:
		return s.callAnthropic(ctx, binding, messages, maxOutputTokens, temperature, effort)
	default:
		return nil, bizerr.NewCode(CodeRequestInvalid)
	}
}

// listOpenAIModels reads OpenAI-compatible /models data.
func (s *serviceImpl) listOpenAIModels(ctx context.Context, endpointRow *entity.ProviderEndpoint) ([]remoteModel, error) {
	resp, err := s.doProviderRequest(ctx, endpointRow.Protocol, endpointRow.BaseUrl, "/models", http.MethodGet, nil, func(req *http.Request) {
		addBearerAuth(req, endpointRow.SecretRef)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readProviderHTTPError(resp)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, gerror.Wrap(err, "decode OpenAI model list failed")
	}
	models := make([]remoteModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		if name := strings.TrimSpace(item.ID); name != "" {
			models = append(models, remoteModel{Name: name})
		}
	}
	return models, nil
}

// listAnthropicModels reads Anthropic-compatible /models data.
func (s *serviceImpl) listAnthropicModels(ctx context.Context, endpointRow *entity.ProviderEndpoint) ([]remoteModel, error) {
	resp, err := s.doProviderRequest(ctx, endpointRow.Protocol, endpointRow.BaseUrl, "/models", http.MethodGet, nil, func(req *http.Request) {
		addAnthropicHeaders(req, endpointRow.SecretRef)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readProviderHTTPError(resp)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, gerror.Wrap(err, "decode Anthropic model list failed")
	}
	models := make([]remoteModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		if name := strings.TrimSpace(item.ID); name != "" {
			models = append(models, remoteModel{Name: name})
		}
	}
	return models, nil
}

// callOpenAI executes one OpenAI-compatible chat completion request.
func (s *serviceImpl) callOpenAI(
	ctx context.Context,
	binding *resolvedTierBinding,
	messages []aitext.Message,
	maxOutputTokens int,
	temperature *float64,
	effort string,
) (*adapterResult, error) {
	if effort == string(aitext.ThinkingEffortXHigh) || effort == string(aitext.ThinkingEffortMax) {
		return nil, bizerr.NewCode(CodeThinkingEffortUnsupported)
	}
	payload := map[string]any{
		"model":    providerRequestModelName(binding.ModelName),
		"messages": openAIMessages(messages),
	}
	if maxOutputTokens > 0 {
		payload["max_tokens"] = maxOutputTokens
	}
	if temperature != nil {
		payload["temperature"] = *temperature
	}
	if effort != "" {
		payload["reasoning_effort"] = effort
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := s.doProviderRequest(ctx, binding.Protocol, binding.EndpointBaseUrl, "/chat/completions", http.MethodPost, body, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
		addBearerAuth(req, binding.EndpointSecretRef)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	latencyMs := int(time.Since(start).Milliseconds())
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readProviderHTTPError(resp)
	}
	var payloadResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			InputTokens      int `json:"input_tokens"`
			OutputTokens     int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payloadResp); err != nil {
		return nil, gerror.Wrap(err, "decode OpenAI response failed")
	}
	if len(payloadResp.Choices) == 0 {
		return nil, gerror.New("OpenAI response has no choices")
	}
	text := payloadResp.Choices[0].Message.Content
	if text == "" {
		text = payloadResp.Choices[0].Text
	}
	return &adapterResult{
		Text: text,
		Usage: aitext.Usage{
			InputTokens:  firstNonZero(payloadResp.Usage.PromptTokens, payloadResp.Usage.InputTokens),
			OutputTokens: firstNonZero(payloadResp.Usage.CompletionTokens, payloadResp.Usage.OutputTokens),
		},
		LatencyMs:      latencyMs,
		ThinkingEffort: effort,
	}, nil
}

// callAnthropic executes one Anthropic-compatible messages request.
func (s *serviceImpl) callAnthropic(
	ctx context.Context,
	binding *resolvedTierBinding,
	messages []aitext.Message,
	maxOutputTokens int,
	temperature *float64,
	effort string,
) (*adapterResult, error) {
	systemPrompt, anthropicMessages := anthropicMessages(messages)
	payload := map[string]any{
		"model":      providerRequestModelName(binding.ModelName),
		"messages":   anthropicMessages,
		"max_tokens": maxOutputTokens,
	}
	if maxOutputTokens <= 0 {
		payload["max_tokens"] = 128
	}
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}
	if temperature != nil {
		payload["temperature"] = *temperature
	}
	if effort != "" {
		payload["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": anthropicThinkingBudget(effort),
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := s.doProviderRequest(ctx, binding.Protocol, binding.EndpointBaseUrl, "/messages", http.MethodPost, body, func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
		addAnthropicHeaders(req, binding.EndpointSecretRef)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	latencyMs := int(time.Since(start).Milliseconds())
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, readProviderHTTPError(resp)
	}
	var payloadResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payloadResp); err != nil {
		return nil, gerror.Wrap(err, "decode Anthropic response failed")
	}
	var builder strings.Builder
	for _, item := range payloadResp.Content {
		if item.Type == "" || item.Type == "text" {
			builder.WriteString(item.Text)
		}
	}
	text := builder.String()
	if strings.TrimSpace(text) == "" {
		return nil, gerror.New("Anthropic response has no text content")
	}
	return &adapterResult{
		Text: text,
		Usage: aitext.Usage{
			InputTokens:  payloadResp.Usage.InputTokens,
			OutputTokens: payloadResp.Usage.OutputTokens,
		},
		LatencyMs:      latencyMs,
		ThinkingEffort: effort,
	}, nil
}

// doProviderRequest executes one provider HTTP request, retrying a 404 once with
// a /v1-suffixed OpenAI or Anthropic base URL and caching the working base URL.
func (s *serviceImpl) doProviderRequest(
	ctx context.Context,
	protocol string,
	baseURL string,
	resourcePath string,
	method string,
	body []byte,
	configure func(*http.Request),
) (*http.Response, error) {
	cacheKey := providerURLCacheKey(protocol, baseURL)
	if cachedBaseURL, ok := s.cachedProviderBaseURL(cacheKey); ok {
		resp, err := s.doProviderRequestOnce(ctx, cachedBaseURL, resourcePath, method, body, configure)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusNotFound {
			return resp, nil
		}
		if err = resp.Body.Close(); err != nil {
			return nil, gerror.Wrap(err, "close cached provider 404 response failed")
		}
		s.deleteProviderBaseURLCache(cacheKey)
	}

	resp, err := s.doProviderRequestOnce(ctx, baseURL, resourcePath, method, body, configure)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusNotFound || !supportsProviderVersionFallback(protocol) {
		return resp, nil
	}

	versionedBaseURL, changed, err := versionedProviderBaseURL(baseURL, resourcePath)
	if err != nil || !changed {
		return resp, nil
	}
	if err = resp.Body.Close(); err != nil {
		return nil, gerror.Wrap(err, "close provider 404 response failed")
	}

	retryResp, err := s.doProviderRequestOnce(ctx, versionedBaseURL, resourcePath, method, body, configure)
	if err != nil {
		return nil, err
	}
	if shouldCacheProviderVersionedURL(retryResp.StatusCode) {
		s.cacheProviderBaseURL(cacheKey, versionedBaseURL)
	}
	return retryResp, nil
}

// doProviderRequestOnce builds and sends a single provider HTTP request.
func (s *serviceImpl) doProviderRequestOnce(
	ctx context.Context,
	baseURL string,
	resourcePath string,
	method string,
	body []byte,
	configure func(*http.Request),
) (*http.Response, error) {
	endpoint, err := normalizeBaseURL(baseURL, resourcePath)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	if configure != nil {
		configure(req)
	}
	return s.httpClient.Do(req)
}

// cachedProviderBaseURL returns a process-local corrected provider base URL.
func (s *serviceImpl) cachedProviderBaseURL(cacheKey string) (string, bool) {
	s.providerURLMu.RLock()
	defer s.providerURLMu.RUnlock()
	baseURL, ok := s.providerURLCache[cacheKey]
	return baseURL, ok
}

// cacheProviderBaseURL stores one provider base URL proven by a retry response.
func (s *serviceImpl) cacheProviderBaseURL(cacheKey string, baseURL string) {
	s.providerURLMu.Lock()
	defer s.providerURLMu.Unlock()
	s.providerURLCache[cacheKey] = baseURL
}

// deleteProviderBaseURLCache removes a stale process-local provider base URL.
func (s *serviceImpl) deleteProviderBaseURLCache(cacheKey string) {
	s.providerURLMu.Lock()
	defer s.providerURLMu.Unlock()
	delete(s.providerURLCache, cacheKey)
}

// providerURLCacheKey scopes corrected base URL cache entries by protocol and
// configured base URL so provider updates naturally use a new cache entry.
func providerURLCacheKey(protocol string, baseURL string) string {
	return strings.TrimSpace(protocol) + "\x00" + strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// supportsProviderVersionFallback limits /v1 retry behavior to the two
// provider protocol families that use OpenAI-compatible or Anthropic endpoints.
func supportsProviderVersionFallback(protocol string) bool {
	switch protocol {
	case protocolOpenAI, protocolOpenAICompatible, protocolAnthropic, protocolAnthropicCompatible:
		return true
	default:
		return false
	}
}

// shouldCacheProviderVersionedURL treats non-404, non-5xx retry responses as a
// recognized provider path. The caller still handles 4xx responses as failures.
func shouldCacheProviderVersionedURL(statusCode int) bool {
	return statusCode != http.StatusNotFound && statusCode < http.StatusInternalServerError
}

// versionedProviderBaseURL returns the base URL with /v1 inserted before the
// resource path unless that base portion already ends with /v1.
func versionedProviderBaseURL(base string, resourcePath string) (string, bool, error) {
	endpoint, err := normalizeBaseURL(base, resourcePath)
	if err != nil {
		return "", false, err
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", false, bizerr.NewCode(CodeProviderProtocolRequired)
	}
	resourcePath = "/" + strings.TrimLeft(resourcePath, "/")
	var (
		endpointPath   = strings.TrimRight(parsed.Path, "/")
		resourceSuffix = strings.TrimRight(resourcePath, "/")
		basePath       = strings.TrimRight(strings.TrimSuffix(endpointPath, resourceSuffix), "/")
	)
	if basePath == "/v1" || strings.HasSuffix(basePath, "/v1") {
		parsed.Path = basePath
		return parsed.String(), false, nil
	}
	if basePath == "" {
		parsed.Path = "/v1"
	} else {
		parsed.Path = basePath + "/v1"
	}
	return parsed.String(), true, nil
}

// normalizeBaseURL appends the resource path unless base already points at it.
func normalizeBaseURL(base string, resourcePath string) (string, error) {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		return "", bizerr.NewCode(CodeProviderProtocolRequired)
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", bizerr.NewCode(CodeProviderProtocolRequired)
	}
	resourcePath = "/" + strings.TrimLeft(resourcePath, "/")
	if strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), strings.TrimRight(resourcePath, "/")) {
		return parsed.String(), nil
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + resourcePath
	return parsed.String(), nil
}

// openAIMessages converts framework messages to OpenAI-compatible messages.
func openAIMessages(messages []aitext.Message) []map[string]string {
	out := make([]map[string]string, 0, len(messages))
	for _, message := range messages {
		out = append(out, map[string]string{
			"role":    string(message.Role),
			"content": message.Content,
		})
	}
	return out
}

// anthropicMessages converts framework messages to Anthropic-compatible messages.
func anthropicMessages(messages []aitext.Message) (string, []map[string]string) {
	var systemPrompt strings.Builder
	out := make([]map[string]string, 0, len(messages))
	for _, message := range messages {
		if message.Role == aitext.MessageRoleSystem {
			if systemPrompt.Len() > 0 {
				systemPrompt.WriteString("\n")
			}
			systemPrompt.WriteString(message.Content)
			continue
		}
		role := "user"
		if message.Role == aitext.MessageRoleAssistant {
			role = "assistant"
		}
		out = append(out, map[string]string{"role": role, "content": message.Content})
	}
	if len(out) == 0 {
		out = append(out, map[string]string{"role": "user", "content": "Health check"})
	}
	return systemPrompt.String(), out
}

// anthropicThinkingBudget maps platform efforts to Anthropic thinking budgets.
func anthropicThinkingBudget(effort string) int {
	switch effort {
	case string(aitext.ThinkingEffortLow):
		return 1024
	case string(aitext.ThinkingEffortMedium):
		return 4096
	case string(aitext.ThinkingEffortHigh):
		return 8192
	case string(aitext.ThinkingEffortXHigh):
		return 16384
	case string(aitext.ThinkingEffortMax):
		return 32768
	default:
		return 0
	}
}

// addBearerAuth applies an OpenAI-compatible bearer token.
func addBearerAuth(req *http.Request, secretRef string) {
	if strings.TrimSpace(secretRef) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(secretRef))
	}
}

// addAnthropicHeaders applies Anthropic-compatible authentication headers.
func addAnthropicHeaders(req *http.Request, secretRef string) {
	if strings.TrimSpace(secretRef) != "" {
		req.Header.Set("X-Api-Key", strings.TrimSpace(secretRef))
	}
	req.Header.Set("Anthropic-Version", "2023-06-01")
}

// readProviderHTTPError reports provider HTTP failures without exposing the
// provider response body, which can contain echoed prompts or upstream diagnostics.
func readProviderHTTPError(resp *http.Response) error {
	return bizerr.NewCode(CodeProviderHTTPError, bizerr.P("status", resp.StatusCode))
}

// firstNonZero returns the first non-zero value.
func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
