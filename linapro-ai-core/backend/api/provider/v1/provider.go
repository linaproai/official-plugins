// This file declares shared AI provider and model API DTO projections.

package v1

// ProviderItem is the provider projection returned by list and detail APIs.
type ProviderItem struct {
	Id                int64  `json:"id" dc:"Provider ID" eg:"1"`
	Name              string `json:"name" dc:"Provider display name" eg:"OpenAI"`
	WebsiteUrl        string `json:"websiteUrl" dc:"Provider website URL" eg:"https://openai.com"`
	Remark            string `json:"remark" dc:"Provider remark" eg:"Production text models"`
	OpenaiBaseUrl     string `json:"openaiBaseUrl" dc:"OpenAI-compatible base URL" eg:"https://api.openai.com/v1"`
	AnthropicBaseUrl  string `json:"anthropicBaseUrl" dc:"Anthropic-compatible base URL" eg:"https://api.anthropic.com/v1"`
	ApiKeySecretRef   string `json:"apiKeySecretRef" dc:"Masked API key secret reference; plaintext API keys are never returned" eg:"sk-***abcd"`
	Enabled           int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	ModelCount        int    `json:"modelCount" dc:"Number of models under this provider" eg:"3"`
	EnabledModelCount int    `json:"enabledModelCount" dc:"Number of enabled models under this provider" eg:"2"`
	CreatedAt         int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt         int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ModelItem is the AI model projection returned by provider model APIs.
type ModelItem struct {
	Id               int64    `json:"id" dc:"Model ID" eg:"1"`
	ProviderId       int64    `json:"providerId" dc:"Owning provider ID" eg:"1"`
	CapabilityType   string   `json:"capabilityType" dc:"Capability type; first release supports text" eg:"text"`
	ModelName        string   `json:"modelName" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol         string   `json:"protocol" dc:"Provider protocol: openai or anthropic" eg:"openai"`
	Source           string   `json:"source" dc:"Model source: manual or api" eg:"manual"`
	SupportsThinking int      `json:"supportsThinking" dc:"Thinking effort support flag: 0=no 1=yes" eg:"1"`
	SupportedEfforts []string `json:"supportedEfforts" dc:"Supported thinking efforts: low, medium, high, xhigh, max" eg:"low,medium,high"`
	MaxInputTokens   int      `json:"maxInputTokens" dc:"Maximum input tokens; 0 means unspecified" eg:"128000"`
	MaxOutputTokens  int      `json:"maxOutputTokens" dc:"Maximum output tokens; 0 means unspecified" eg:"8192"`
	Enabled          int      `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	CreatedAt        int64    `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt        int64    `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}
