// This file translates dynamic demo business errors into bridge responses
// via the shared pluginbridge ErrorClassifier composition.

package dynamic

import (
	dynamicservice "lina-plugin-linapro-demo-dynamic/backend/internal/service/dynamic"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// dynamicErrorClassifier maps dynamic sample business errors to normalized
// bridge responses. BindJSON sentinels are handled by pluginbridge itself.
var dynamicErrorClassifier = bridgeplugin.NewErrorClassifier(
	bridgeplugin.NewErrorCase(dynamicservice.IsDemoRecordInvalidInput, protocol.NewBadRequestResponse),
	bridgeplugin.NewErrorCase(dynamicservice.IsDemoRecordNotFound, protocol.NewNotFoundResponse),
)

// wrapDynamicError converts one dynamic sample business error into a
// prebuilt bridge response error so typed guest controllers can return it
// through the standard error channel.
func wrapDynamicError(err error) error {
	if err == nil {
		return nil
	}
	return bridgeplugin.NewResponseError(dynamicErrorClassifier.Classify(err))
}
