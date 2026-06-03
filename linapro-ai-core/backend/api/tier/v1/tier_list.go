// This file declares the list-tier request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for listing AI capability tiers.
type ListReq struct {
	g.Meta           `path:"/ai/tiers" method:"get" tags:"AI Tiers" summary:"List AI capability tiers" dc:"Return the fixed basic, standard, and advanced AI tiers for one capability method with their primary binding projection and latest test summary." permission:"ai:tier:list"`
	CapabilityType   string `json:"capabilityType" d:"text" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod string `json:"capabilityMethod" d:"generate" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
}

// ListRes defines the response for listing AI capability tiers.
type ListRes struct {
	List []*TierItem `json:"list" dc:"AI capability tier list" eg:"[]"`
}
