// This file defines typed Jobs callback DTOs for the dynamic plugin sample.

package v1

// RegisterJobsReq is the typed request for built-in Jobs declaration discovery.
type RegisterJobsReq struct{}

// RegisterJobsRes is the response placeholder for Jobs declaration discovery.
type RegisterJobsRes struct{}

// JobHeartbeatReq is the typed request for executing the sample Jobs heartbeat.
type JobHeartbeatReq struct{}

// JobHeartbeatRes summarizes one sample Jobs heartbeat execution.
type JobHeartbeatRes struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}
