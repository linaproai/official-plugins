// This file registers built-in Jobs declarations for the dynamic sample plugin
// through the governed Jobs host service.

package dynamicservice

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// Jobs heartbeat declaration constants define the built-in Jobs contract
// exported by the dynamic sample plugin.
const (
	jobHeartbeatName        = "heartbeat"
	jobHeartbeatDisplayName = "Dynamic Plugin Heartbeat"
	jobHeartbeatDesc        = "Runs the dynamic plugin built-in job through the Wasm bridge and accumulates heartbeat executions."
	jobHeartbeatPattern     = "# */10 * * * *"
	jobHeartbeatPath        = "/job-heartbeat"
	jobHeartbeatTimeout     = 30
)

// RegisterJobs declares built-in Jobs against the provided dynamic plugin
// startup facade.
func RegisterJobs(plugin pluginbridge.Declarations) error {
	if plugin == nil || plugin.Jobs() == nil {
		return gerror.New("dynamic plugin Jobs declaration facade is required")
	}
	return plugin.Jobs().Register(&protocol.JobContract{
		Name:           jobHeartbeatName,
		DisplayName:    jobHeartbeatDisplayName,
		Description:    jobHeartbeatDesc,
		Pattern:        jobHeartbeatPattern,
		Timezone:       protocol.DefaultJobContractTimezone,
		Scope:          protocol.JobScopeAllNode,
		Concurrency:    protocol.JobConcurrencySingleton,
		MaxConcurrency: 1,
		TimeoutSeconds: jobHeartbeatTimeout,
		RequestType:    "JobHeartbeatReq",
		InternalPath:   jobHeartbeatPath,
	})
}

// RegisterJobs publishes all built-in Jobs declarations for host-side
// discovery.
func (s *serviceImpl) RegisterJobs(plugin pluginbridge.Declarations) error {
	return RegisterJobs(plugin)
}
