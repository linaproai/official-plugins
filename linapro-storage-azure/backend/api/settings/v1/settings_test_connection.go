// POST /settings/test-connection for Azure Blob connectivity probe.
package v1

import "github.com/gogf/gf/v2/frame/g"

// TestConnectionReq probes container connectivity with current form or saved settings.
type TestConnectionReq struct {
	g.Meta      `path:"/settings/test-connection" method:"post" tags:"Storage / Azure Blob" summary:"Test object storage connectivity" dc:"Probe the target container with the submitted or previously saved credentials without persisting changes." permission:"linapro-storage-azure:settings:view"`
	AccountName string `json:"accountName" v:"max-length:256"`
	AccountKey  string `json:"accountKey" v:"max-length:512"`
	Container   string `json:"container" v:"max-length:256"`
	Endpoint    string `json:"endpoint" v:"max-length:512"`
	PathPrefix  string `json:"pathPrefix" v:"max-length:256"`
}

// TestConnectionRes reports probe success.
type TestConnectionRes struct {
	OK      bool   `json:"ok" dc:"Whether the connectivity probe succeeded"`
	Message string `json:"message" dc:"Human-readable probe result"`
}
