// This file implements the backend summary business logic for the dynamic
// sample plugin.

package dynamicservice

import "github.com/gogf/gf/v2/util/gconv"

// backendSummaryMessage is the static intro text returned by the sample
// backend summary endpoint.
const backendSummaryMessage = "This backend example is executed through the linapro-demo-dynamic Wasm bridge runtime."

// BuildBackendSummaryPayload builds the backend summary response payload.
func (s *serviceImpl) BuildBackendSummaryPayload(input *BackendSummaryInput) *backendSummaryPayload {
	payload := &backendSummaryPayload{
		Message: backendSummaryMessage,
	}
	if input == nil {
		return payload
	}

	payload.PluginID = input.PluginID
	payload.PublicPath = input.PublicPath
	payload.Access = input.Access
	payload.Permission = input.Permission
	payload.Authenticated = input.Authenticated
	if input.HasIdentity {
		payload.Username = gconv.PtrString(input.Username)
		payload.IsSuperAdmin = gconv.PtrBool(input.IsSuperAdmin)
	}
	return payload
}
