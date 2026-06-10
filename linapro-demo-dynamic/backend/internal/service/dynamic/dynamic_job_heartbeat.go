// This file implements the declared Jobs heartbeat business logic for the
// dynamic sample plugin.

package dynamicservice

import "lina-core/pkg/plugin/pluginbridge/protocol"

const (
	jobHeartbeatStateKey = "job_heartbeat_count"
)

// jobHeartbeatPayload summarizes one successful Jobs heartbeat execution.
type jobHeartbeatPayload struct {
	// Count is the accumulated execution count persisted in plugin-scoped runtime state.
	Count int `json:"count"`
	// Message describes the job execution result.
	Message string `json:"message"`
}

// BuildJobHeartbeatPayload executes the declared Jobs heartbeat task and
// returns a lightweight execution summary.
func (s *serviceImpl) BuildJobHeartbeatPayload() (*jobHeartbeatPayload, error) {
	count, found, err := s.runtimeSvc.StateGetInt(jobHeartbeatStateKey)
	if err != nil {
		return nil, err
	}
	if !found {
		count = 0
	}
	count++
	if err = s.runtimeSvc.StateSetInt(jobHeartbeatStateKey, count); err != nil {
		return nil, err
	}
	if err = s.runtimeSvc.Log(
		int(protocol.LogLevelInfo),
		"declared Jobs heartbeat executed",
		nil,
	); err != nil {
		return nil, err
	}
	return &jobHeartbeatPayload{
		Count:   count,
		Message: "Dynamic plugin Jobs heartbeat executed successfully.",
	}, nil
}
