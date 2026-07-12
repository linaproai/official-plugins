package v1

import "github.com/gogf/gf/v2/frame/g"

// TestConnectionReq probes bucket connectivity with current form or saved settings.
type TestConnectionReq struct {
	g.Meta          `path:"/settings/test-connection" method:"post" tags:"Storage / Qiniu Kodo" summary:"Test object storage connectivity" dc:"Probe the target bucket with the submitted or previously saved credentials without persisting changes." permission:"linapro-storage-qiniu:settings:view"`
	AccessKeyID     string `json:"accessKeyID" v:"max-length:256"`
	SecretAccessKey string `json:"secretAccessKey" v:"max-length:512"`
	Region          string `json:"region" v:"max-length:128"`
	Bucket          string `json:"bucket" v:"max-length:256"`
	Endpoint        string `json:"endpoint" v:"max-length:512"`
	PathPrefix      string `json:"pathPrefix" v:"max-length:256"`
}

// TestConnectionRes reports probe success.
type TestConnectionRes struct {
	OK      bool   `json:"ok" dc:"Whether the connectivity probe succeeded"`
	Message string `json:"message" dc:"Human-readable probe result"`
}
