//go:build !wasip1

// This file keeps host-side Go tests and tooling able to execute the dynamic
// demo backend through the generic reflected dispatcher. The Wasm runtime build
// does not compile this file; linactl injects a generated zero-reflection
// dispatcher for wasip1 builds instead.

package backend

import (
	"lina-plugin-linapro-demo-dynamic/backend/internal/controller/dynamic"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// guestRouteDispatcher is the host-build reflected bridge dispatcher for
// ordinary Go tests. Runtime Wasm builds use the generated dispatcher.
var guestRouteDispatcher = bridgeplugin.MustNewGuestControllerRouteDispatcher(dynamic.New())

// HandleRequest dispatches bridge requests to the matching dynamic controller
// method using the build-time RequestType contract.
func HandleRequest(
	request *protocol.BridgeRequestEnvelopeV1,
) (*protocol.BridgeResponseEnvelopeV1, error) {
	return guestRouteDispatcher.HandleRequest(request)
}
