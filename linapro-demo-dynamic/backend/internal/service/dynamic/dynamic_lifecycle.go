// This file implements lifecycle debug logging for the dynamic sample plugin.

package dynamicservice

import "lina-core/pkg/plugin/pluginbridge/protocol"

// RunLifecycleDebugHook logs one lifecycle callback invocation and allows the
// host operation to continue.
func (s *serviceImpl) RunLifecycleDebugHook(_ *LifecycleDebugInput) error {
	return s.runtimeSvc.Log(
		int(protocol.LogLevelInfo),
		"linapro-demo-dynamic lifecycle callback invoked",
		nil,
	)
}
