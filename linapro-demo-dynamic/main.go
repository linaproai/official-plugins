//lint:file-ignore U1000 Dynamic WASM entry points are exported by symbol name and called by the host runtime.
package main

import (
	dynamicbackend "lina-plugin-linapro-demo-dynamic/backend"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// guestRuntime owns the wasm guest bridge buffers and request execution helpers.
var guestRuntime = bridgeplugin.NewGuestRuntime(dynamicbackend.HandleRequest)

//go:wasmexport lina_dynamic_route_alloc
func linaDynamicRouteAlloc(size uint32) uint32 {
	return guestRuntime.Alloc(size)
}

//go:wasmexport lina_dynamic_route_execute
func linaDynamicRouteExecute(size uint32) uint64 {
	responsePointer, responseLength, err := guestRuntime.Execute(size)
	if err != nil {
		fallback, _ := protocol.EncodeResponseEnvelope(protocol.NewInternalErrorResponse(err.Error()))
		responsePointer, responseLength, _ = guestRuntime.ExposeResponseBuffer(fallback)
	}
	return uint64(responsePointer)<<32 | uint64(responseLength)
}

//go:wasmexport lina_host_call_alloc
func linaHostCallAlloc(size uint32) uint32 {
	return guestRuntime.HostCallAlloc(size)
}

// main keeps the wasm module linkable without running any host-side process boot logic.
func main() {}
