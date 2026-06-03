// This file declares model capability management DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ModelCapabilityItem is the explicit model capability method projection.
type ModelCapabilityItem struct {
	Id                int64    `json:"id" dc:"Model capability ID" eg:"1"`
	ModelId           int64    `json:"modelId" dc:"Owning model ID" eg:"1"`
	EndpointId        int64    `json:"endpointId" dc:"Preferred provider endpoint ID; 0 means model default" eg:"1"`
	CapabilityType    string   `json:"capabilityType" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod  string   `json:"capabilityMethod" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	InputModalities   []string `json:"inputModalities" dc:"Input modalities accepted by this model method" eg:"text,image"`
	OutputModalities  []string `json:"outputModalities" dc:"Output modalities produced by this model method" eg:"image"`
	MaxInputTokens    int      `json:"maxInputTokens" dc:"Maximum input tokens; 0 means unspecified" eg:"128000"`
	MaxOutputTokens   int      `json:"maxOutputTokens" dc:"Maximum output tokens; 0 means unspecified" eg:"8192"`
	MaxInputAssets    int      `json:"maxInputAssets" dc:"Maximum number of input asset references; 0 means unspecified" eg:"4"`
	MaxOutputAssets   int      `json:"maxOutputAssets" dc:"Maximum number of output asset references; 0 means unspecified" eg:"1"`
	MaxAssetBytes     int64    `json:"maxAssetBytes" dc:"Maximum bytes per referenced asset; 0 means unspecified" eg:"10485760"`
	SupportsStreaming int      `json:"supportsStreaming" dc:"Streaming support flag: 0=no 1=yes" eg:"0"`
	SupportsOperation int      `json:"supportsOperation" dc:"Provider operation support flag: 0=no 1=yes" eg:"1"`
	SupportsThinking  int      `json:"supportsThinking" dc:"Thinking effort support flag for this model method: 0=no 1=yes" eg:"1"`
	SupportedEfforts  []string `json:"supportedEfforts" dc:"Supported thinking efforts for this model method: low, medium, high, xhigh, max" eg:"low,medium,high"`
	DefaultParamsJson string   `json:"defaultParamsJson" dc:"Method-specific default params JSON without provider secrets" eg:"{}"`
	Enabled           int      `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	CreatedAt         int64    `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt         int64    `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ModelCapabilityInput is one explicit model capability method save item.
type ModelCapabilityInput struct {
	EndpointId        int64    `json:"endpointId" dc:"Preferred provider endpoint ID; 0 means model default" eg:"1"`
	CapabilityType    string   `json:"capabilityType" v:"required" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod  string   `json:"capabilityMethod" v:"required" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	InputModalities   []string `json:"inputModalities" dc:"Input modalities accepted by this model method" eg:"text,image"`
	OutputModalities  []string `json:"outputModalities" dc:"Output modalities produced by this model method" eg:"image"`
	MaxInputTokens    int      `json:"maxInputTokens" dc:"Maximum input tokens; 0 means unspecified" eg:"128000"`
	MaxOutputTokens   int      `json:"maxOutputTokens" dc:"Maximum output tokens; 0 means unspecified" eg:"8192"`
	MaxInputAssets    int      `json:"maxInputAssets" dc:"Maximum number of input asset references; 0 means unspecified" eg:"4"`
	MaxOutputAssets   int      `json:"maxOutputAssets" dc:"Maximum number of output asset references; 0 means unspecified" eg:"1"`
	MaxAssetBytes     int64    `json:"maxAssetBytes" dc:"Maximum bytes per referenced asset; 0 means unspecified" eg:"10485760"`
	SupportsStreaming int      `json:"supportsStreaming" dc:"Streaming support flag: 0=no 1=yes" eg:"0"`
	SupportsOperation int      `json:"supportsOperation" dc:"Provider operation support flag: 0=no 1=yes" eg:"1"`
	SupportsThinking  int      `json:"supportsThinking" dc:"Thinking effort support flag for this model method: 0=no 1=yes" eg:"1"`
	SupportedEfforts  []string `json:"supportedEfforts" dc:"Supported thinking efforts for this model method: low, medium, high, xhigh, max" eg:"low,medium,high"`
	DefaultParamsJson string   `json:"defaultParamsJson" dc:"Method-specific default params JSON without provider secrets" eg:"{}"`
	Enabled           int      `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// ListCapabilitiesReq defines the request for listing model capability methods.
type ListCapabilitiesReq struct {
	g.Meta `path:"/ai/models/{id}/capabilities" method:"get" tags:"AI Model Capabilities" summary:"List model capabilities" dc:"List explicit capability method declarations for one AI model, including modality, asset, token, thinking effort, streaming, and provider operation limits." permission:"ai:provider:list"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
}

// ListCapabilitiesRes defines the response for listing model capability methods.
type ListCapabilitiesRes struct {
	List []*ModelCapabilityItem `json:"list" dc:"Model capability method list" eg:"[]"`
}

// UpsertCapabilitiesReq defines the request for replacing model capability methods.
type UpsertCapabilitiesReq struct {
	g.Meta `path:"/ai/models/{id}/capabilities" method:"put" tags:"AI Model Capabilities" summary:"Upsert model capabilities" dc:"Replace or insert explicit capability method declarations for one AI model. Endpoint references must belong to the same provider as the model." permission:"ai:provider:update"`
	Id     int64                  `json:"id" v:"required|min:1" dc:"Model ID" eg:"1"`
	Items  []ModelCapabilityInput `json:"items" v:"required" dc:"Capability method save items" eg:"[]"`
}

// UpsertCapabilitiesRes defines the response for replacing model capability methods.
type UpsertCapabilitiesRes struct{}
