// This file declares the tier-test request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TestReq defines the request for testing a saved or draft AI tier binding.
type TestReq struct {
	g.Meta           `path:"/ai/tiers/{code}/test" method:"post" tags:"AI Tiers" summary:"Test AI capability tier" dc:"Execute a lightweight provider test against a saved or draft tier binding within one capability method without persisting draft binding changes. Text methods use messages; other methods may return provider-unavailable until an adapter supports that method." permission:"ai:tier:test"`
	CapabilityType   string        `json:"capabilityType" d:"text" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"text"`
	CapabilityMethod string        `json:"capabilityMethod" d:"generate" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	Code             string        `json:"code" v:"required|in:basic,standard,advanced" dc:"Tier code: basic, standard, advanced" eg:"basic"`
	ProviderId       int64         `json:"providerId" dc:"Optional draft provider ID. When omitted, the saved tier binding is used." eg:"1"`
	ModelId          int64         `json:"modelId" dc:"Optional draft model ID. When omitted, the saved tier binding is used." eg:"1"`
	ThinkingEffort   string        `json:"thinkingEffort" dc:"Optional draft thinking effort: low, medium, high, xhigh, max, or empty" eg:"low"`
	MaxOutputTokens  int           `json:"maxOutputTokens" d:"128" dc:"Maximum output tokens for the lightweight test" eg:"128"`
	Messages         []TextMessage `json:"messages" dc:"Optional test messages. Defaults to a short health-check prompt when empty." eg:"[]"`
}

// TestRes defines the response for testing a saved or draft AI tier binding.
type TestRes struct {
	Status         string `json:"status" dc:"Test status: success or failed" eg:"success"`
	LatencyMs      int    `json:"latencyMs" dc:"Test latency in milliseconds" eg:"300"`
	ProviderName   string `json:"providerName" dc:"Actual provider display name" eg:"OpenAI"`
	ModelName      string `json:"modelName" dc:"Actual model name" eg:"gpt-4.1-mini"`
	Protocol       string `json:"protocol" dc:"Actual provider protocol" eg:"openai"`
	ThinkingEffort string `json:"thinkingEffort" dc:"Actual thinking effort applied by the model" eg:"low"`
	ErrorSummary   string `json:"errorSummary" dc:"Masked failure summary when the test fails" eg:""`
	TestedAt       int64  `json:"testedAt" dc:"Test time, Unix timestamp in milliseconds" eg:"1717200000000"`
}
