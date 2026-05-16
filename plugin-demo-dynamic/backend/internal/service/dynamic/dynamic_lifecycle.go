// This file implements lifecycle debug logging for the dynamic sample plugin.

package dynamicservice

import (
	"strconv"
	"strings"

	"lina-core/pkg/pluginbridge"
)

// RunLifecycleDebugHook logs one lifecycle callback invocation and allows the
// host operation to continue.
func (s *serviceImpl) RunLifecycleDebugHook(input *LifecycleDebugInput) error {
	if input == nil {
		input = &LifecycleDebugInput{}
	}
	fields := map[string]string{
		"pluginId":  strings.TrimSpace(input.PluginID),
		"operation": strings.TrimSpace(input.Operation),
	}
	if strings.TrimSpace(input.FromVersion) != "" {
		fields["fromVersion"] = strings.TrimSpace(input.FromVersion)
	}
	if strings.TrimSpace(input.ToVersion) != "" {
		fields["toVersion"] = strings.TrimSpace(input.ToVersion)
	}
	if input.TenantID > 0 {
		fields["tenantId"] = strconv.Itoa(input.TenantID)
	}
	if strings.TrimSpace(input.FromMode) != "" {
		fields["fromMode"] = strings.TrimSpace(input.FromMode)
	}
	if strings.TrimSpace(input.ToMode) != "" {
		fields["toMode"] = strings.TrimSpace(input.ToMode)
	}
	if input.PurgeStorageData {
		fields["purgeStorageData"] = strconv.FormatBool(input.PurgeStorageData)
	}
	return s.runtimeSvc.Log(
		int(pluginbridge.LogLevelInfo),
		"plugin-demo-dynamic lifecycle callback invoked",
		fields,
	)
}
