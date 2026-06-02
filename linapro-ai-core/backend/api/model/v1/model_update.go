// This file declares model update and delete DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating an AI model.
type UpdateReq struct {
	g.Meta           `path:"/ai/models/{id}" method:"put" tags:"AI Models" summary:"Update AI model" dc:"Update one AI model after verifying it remains attached to its provider and still declares supported text capability metadata." permission:"ai:provider:update"`
	Id               int64    `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
	CapabilityType   string   `json:"capabilityType" d:"text" dc:"Capability type; first release supports text" eg:"text"`
	ModelName        string   `json:"modelName" v:"required|max-length:128" dc:"Provider model name" eg:"gpt-4.1-mini"`
	Protocol         string   `json:"protocol" v:"required|in:openai,anthropic" dc:"Provider protocol: openai or anthropic" eg:"openai"`
	SupportsThinking int      `json:"supportsThinking" dc:"Thinking effort support flag: 0=no 1=yes" eg:"1"`
	SupportedEfforts []string `json:"supportedEfforts" dc:"Supported thinking efforts: low, medium, high, xhigh, max" eg:"low,medium,high"`
	MaxInputTokens   int      `json:"maxInputTokens" dc:"Maximum input tokens; 0 means unspecified" eg:"128000"`
	MaxOutputTokens  int      `json:"maxOutputTokens" dc:"Maximum output tokens; 0 means unspecified" eg:"8192"`
	Enabled          int      `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateRes defines the response for updating an AI model.
type UpdateRes struct{}

// DeleteReq defines the request for deleting an AI model.
type DeleteReq struct {
	g.Meta `path:"/ai/models/{id}" method:"delete" tags:"AI Models" summary:"Delete AI model" dc:"Delete one AI model after verifying no AI capability tier binding references it." permission:"ai:provider:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
}

// DeleteRes defines the response for deleting an AI model.
type DeleteRes struct{}
