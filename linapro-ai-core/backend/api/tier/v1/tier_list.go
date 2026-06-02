// This file declares the list-tier request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListReq defines the request for listing AI capability tiers.
type ListReq struct {
	g.Meta         `path:"/ai/tiers" method:"get" tags:"AI Tiers" summary:"List AI capability tiers" dc:"Return the fixed basic, standard, and advanced text AI tiers with their primary binding projection and latest test summary." permission:"ai:tier:list"`
	CapabilityType string `json:"capabilityType" d:"text" dc:"Capability type; first release supports text" eg:"text"`
}

// ListRes defines the response for listing AI capability tiers.
type ListRes struct {
	List []*TierItem `json:"list" dc:"AI capability tier list" eg:"[]"`
}
