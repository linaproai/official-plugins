// This file declares capability method default parameter DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MethodDefaultParamItem is one capability method default parameter projection.
type MethodDefaultParamItem struct {
	Id                int64  `json:"id" dc:"Method default parameter ID" eg:"1"`
	CapabilityType    string `json:"capabilityType" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod  string `json:"capabilityMethod" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	DefaultParamsJson string `json:"defaultParamsJson" dc:"Method-specific default params JSON without provider secrets" eg:"{}"`
	Enabled           int    `json:"enabled" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
	CreatedAt         int64  `json:"createdAt" dc:"Creation time, Unix timestamp in milliseconds" eg:"1717200000000"`
	UpdatedAt         int64  `json:"updatedAt" dc:"Update time, Unix timestamp in milliseconds" eg:"1717200000000"`
}

// ListMethodDefaultsReq defines the request for listing method defaults.
type ListMethodDefaultsReq struct {
	g.Meta `path:"/ai/method-defaults" method:"get" tags:"AI Method Defaults" summary:"List AI method defaults" dc:"List method-specific default parameter projections for all governed AI capability methods." permission:"ai:tier:list"`
}

// ListMethodDefaultsRes defines the response for listing method defaults.
type ListMethodDefaultsRes struct {
	List []*MethodDefaultParamItem `json:"list" dc:"Capability method default parameter list" eg:"[]"`
}

// UpdateMethodDefaultReq defines the request for updating one method default.
type UpdateMethodDefaultReq struct {
	g.Meta            `path:"/ai/method-defaults/{capabilityType}/{capabilityMethod}" method:"put" tags:"AI Method Defaults" summary:"Update AI method default" dc:"Update method-specific default parameters for one governed AI capability method and invalidate capability method resolution cache after successful persistence." permission:"ai:tier:update"`
	CapabilityType    string `json:"capabilityType" v:"required" dc:"Capability type such as text, image, embedding, audio, vision, document, safety, or video" eg:"image"`
	CapabilityMethod  string `json:"capabilityMethod" v:"required" dc:"Capability method within the type, such as generate, create, transcribe, analyze, moderate, or operation.get" eg:"generate"`
	DefaultParamsJson string `json:"defaultParamsJson" dc:"Method-specific default params JSON without provider secrets" eg:"{}"`
	Enabled           int    `json:"enabled" d:"1" dc:"Enabled flag: 0=disabled 1=enabled" eg:"1"`
}

// UpdateMethodDefaultRes defines the response for updating one method default.
type UpdateMethodDefaultRes struct{}
