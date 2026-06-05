// This file declares shared AI tier API DTO projections.

package v1

// TierBindingItem is the primary provider-model binding projection for a tier.
type TierBindingItem struct {
	ProviderId   int64  `json:"providerId" dc:"Provider ID bound to the tier" eg:"1"`
	ProviderName string `json:"providerName" dc:"Provider display name" eg:"OpenAI"`
	ModelId      int64  `json:"modelId" dc:"Model ID bound to the tier" eg:"1"`
	ModelName    string `json:"modelName" dc:"Model name bound to the tier" eg:"gpt-4.1-mini"`
	Protocol     string `json:"protocol" dc:"Provider protocol: openai, anthropic, voyage, openai-compatible, or anthropic-compatible" eg:"openai"`
	Enabled      int    `json:"enabled" dc:"Binding enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// TierItem is the AI capability tier projection.
type TierItem struct {
	Id                   int64            `json:"id" dc:"Tier ID" eg:"1"`
	CapabilityType       string           `json:"capabilityType" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod     string           `json:"capabilityMethod" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	Code                 string           `json:"code" dc:"Tier code: basic, standard, advanced" eg:"basic"`
	DisplayName          string           `json:"displayName" dc:"Tier display name" eg:"Basic"`
	Description          string           `json:"description" dc:"Tier description" eg:"Low-cost image generation tier"`
	DefaultEffort        string           `json:"defaultEffort" dc:"Default thinking effort: low, medium, high, xhigh, max, or empty" eg:"low"`
	Enabled              int              `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	SortOrder            int              `json:"sortOrder" dc:"Stable tier sort order" eg:"1"`
	Binding              *TierBindingItem `json:"binding" dc:"Primary provider-model binding projection" eg:"{}"`
	LastTestStatus       string           `json:"lastTestStatus" dc:"Last tier test status: success or failed" eg:"success"`
	LastTestLatencyMs    int              `json:"lastTestLatencyMs" dc:"Last tier test latency in milliseconds" eg:"300"`
	LastTestErrorSummary string           `json:"lastTestErrorSummary" dc:"Masked last test error summary" eg:""`
	LastTestAt           int64            `json:"lastTestAt" dc:"Last tier test time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt            int64            `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// TextMessage carries one plain-text test message.
type TextMessage struct {
	Role    string `json:"role" dc:"Message role: system, user, assistant" eg:"user"`
	Content string `json:"content" dc:"Plain text message content" eg:"Say hello"`
}
