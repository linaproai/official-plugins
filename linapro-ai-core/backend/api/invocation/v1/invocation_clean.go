// This file declares the invocation-log cleanup request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CleanReq defines the request for clearing masked AI invocation logs.
type CleanReq struct {
	g.Meta    `path:"/ai/invocations/clean" method:"delete" tags:"Smart Center/Invocation Logs" summary:"Clear AI invocation logs" dc:"Clear masked AI invocation logs within a specified creation time range, or clear all logs when no range is provided." permission:"ai:invocation:clear"`
	StartedAt int64 `json:"startedAt" dc:"Optional cleanup start time, Unix timestamp in milliseconds. Empty means no lower bound." eg:"1717200000000"`
	EndedAt   int64 `json:"endedAt" dc:"Optional cleanup end time, Unix timestamp in milliseconds. Empty means no upper bound." eg:"1717286400000"`
}

// CleanRes defines the invocation-log cleanup response.
type CleanRes struct {
	Deleted int `json:"deleted" dc:"Number of invocation logs actually deleted" eg:"500"`
}
