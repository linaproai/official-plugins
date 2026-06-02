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
	"lina-core/pkg/plugin/capability/ai/aitext"
	"lina-plugin-linapro-ai-core/backend/internal/model/entity"
)

// adapterResult carries one provider text-generation result.
type adapterResult struct {
	Text           string
	Usage          aitext.Usage
	LatencyMs      int
	ThinkingEffort string
}

// listRemoteModels returns public model names from the selected protocol.
func (s *serviceImpl) listRemoteModels(ctx context.Context, provider *entity.Provider, protocol string) ([]string, error) {
	switch protocol {
	case ProtocolOpenAI:
		return s.listOpenAIModels(ctx, provider)
	case ProtocolAnthropic:
		return s.listAnthropicModels(ctx, provider)
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
	case ProtocolOpenAI:
		return s.callOpenAI(ctx, binding, messages, maxOutputTokens, temperature, effort)
	case ProtocolAnthropic:
		return s.callAnthropic(ctx, binding, messages, maxOutputTokens, temperature, effort)
	default:
		return nil, bizerr.NewCode(CodeRequestInvalid)
	}
}

// listOpenAIModels reads OpenAI-compatible /models data.
func (s *serviceImpl) listOpenAIModels(ctx context.Context, provider *entity.Provider) ([]string, error) {
	endpoint, err := normalizeBaseURL(provider.OpenaiBaseUrl, "/models")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	addBearerAuth(req, provider.ApiKeySecretRef)
	resp, err := s.httpClient.Do(req)
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
	names := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) != "" {
			names = append(names, item.ID)
		}
	}
	return names, nil
}

// listAnthropicModels reads Anthropic-compatible /models data.
func (s *serviceImpl) listAnthropicModels(ctx context.Context, provider *entity.Provider) ([]string, error) {
	endpoint, err := normalizeBaseURL(provider.AnthropicBaseUrl, "/models")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	addAnthropicHeaders(req, provider.ApiKeySecretRef)
	resp, err := s.httpClient.Do(req)
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
	names := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) != "" {
			names = append(names, item.ID)
		}
	}
	return names, nil
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
	endpoint, err := normalizeBaseURL(binding.OpenaiBaseUrl, "/chat/completions")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"model":    binding.ModelName,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	addBearerAuth(req, binding.ApiKeySecretRef)
	start := time.Now()
	resp, err := s.httpClient.Do(req)
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
	endpoint, err := normalizeBaseURL(binding.AnthropicBaseUrl, "/messages")
	if err != nil {
		return nil, err
	}
	systemPrompt, anthropicMessages := anthropicMessages(messages)
	payload := map[string]any{
		"model":      binding.ModelName,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	addAnthropicHeaders(req, binding.ApiKeySecretRef)
	start := time.Now()
	resp, err := s.httpClient.Do(req)
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
		req.Header.Set("x-api-key", strings.TrimSpace(secretRef))
	}
	req.Header.Set("anthropic-version", "2023-06-01")
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
