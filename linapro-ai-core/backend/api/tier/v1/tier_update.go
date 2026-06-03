// This file declares the update-tier request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateReq defines the request for updating one AI capability tier.
type UpdateReq struct {
	g.Meta           `path:"/ai/tiers/{code}" method:"put" tags:"AI Tiers" summary:"Update AI capability tier" dc:"Update one AI capability tier within a capability method, including enabled state, default thinking effort, and primary provider-model binding." permission:"ai:tier:update"`
	CapabilityType   string `json:"capabilityType" d:"text" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod string `json:"capabilityMethod" d:"generate" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	Code             string `json:"code" v:"required|in:basic,standard,advanced" dc:"Tier code: basic, standard, advanced" eg:"basic"`
	ProviderId       int64  `json:"providerId" dc:"Provider ID for the primary binding; 0 keeps the existing binding" eg:"1"`
	ModelId          int64  `json:"modelId" dc:"Model ID for the primary binding; 0 keeps the existing binding" eg:"1"`
	DefaultEffort    string `json:"defaultEffort" dc:"Default thinking effort: low, medium, high, xhigh, max, or empty" eg:"low"`
	Enabled          int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateRes defines the response for updating one AI capability tier.
type UpdateRes struct{}
