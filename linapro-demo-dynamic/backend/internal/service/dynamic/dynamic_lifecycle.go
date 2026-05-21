// This file implements lifecycle debug logging for the dynamic sample plugin.

package dynamicservice

import "lina-core/pkg/pluginbridge"

// RunLifecycleDebugHook logs one lifecycle callback invocation and allows the
// host operation to continue.
func (s *serviceImpl) RunLifecycleDebugHook(input *LifecycleDebugInput) error {
	if input == nil {
		input = &LifecycleDebugInput{}
	}
	return s.runtimeSvc.Log(
		int(pluginbridge.LogLevelInfo),
		"linapro-demo-dynamic lifecycle callback invoked",
		nil,
	)
}
