// extidcap_handoff.go publishes the SPA handoff DTO and a thin process-bound
// facade. The store implementation lives in backend/internal/service/handoff and
// is bound by the owner plugin at startup. Protocol plugins call the package-level
// Create helpers; HTTP exchange should inject HandoffService via constructor.

package extidcap

import (
	"sync"

	"lina-core/pkg/plugin/capability/authcap/extlogin"
)

// LoginHandoffPayload is the single-use SPA exchange body.
type LoginHandoffPayload struct {
	AccessToken      string
	RefreshToken     string
	PreToken         string
	TenantCandidates []extlogin.TenantCandidate
}

// HandoffService is the owner-plugin runtime for one-time login handoff codes.
// Implementations must be process-shared for create and exchange paths.
type HandoffService interface {
	// Create stores one host login outcome and returns a single-use code.
	Create(payload LoginHandoffPayload) (string, error)
	// CreateFromHost maps a host extlogin.LoginOutput into a handoff.
	CreateFromHost(out *extlogin.LoginOutput) (string, error)
	// Exchange consumes one handoff code and returns the stored payload.
	Exchange(code string) (*LoginHandoffPayload, error)
}

var (
	handoffMu    sync.RWMutex
	boundHandoff HandoffService
)

// BindHandoffService wires the owner plugin handoff implementation for package
// facades used by protocol plugins. Passing nil clears the binding.
func BindHandoffService(svc HandoffService) {
	handoffMu.Lock()
	defer handoffMu.Unlock()
	boundHandoff = svc
}

// BoundHandoffService returns the currently bound handoff implementation.
func BoundHandoffService() HandoffService {
	handoffMu.RLock()
	defer handoffMu.RUnlock()
	return boundHandoff
}

// CreateLoginHandoff stores one host login outcome and returns a single-use code.
// Requires BindHandoffService from the owner plugin.
func CreateLoginHandoff(payload LoginHandoffPayload) (string, error) {
	svc := BoundHandoffService()
	if svc == nil {
		return "", errHandoffInvalid()
	}
	return svc.Create(payload)
}

// CreateLoginHandoffFromHost maps a host extlogin.LoginOutput into a handoff.
// Requires BindHandoffService from the owner plugin.
func CreateLoginHandoffFromHost(out *extlogin.LoginOutput) (string, error) {
	svc := BoundHandoffService()
	if svc == nil {
		return "", errHandoffInvalid()
	}
	return svc.CreateFromHost(out)
}

// ExchangeLoginHandoff consumes one handoff code and returns the stored payload.
// Prefer injecting HandoffService into controllers; this facade remains for tests
// and thin adapters. Requires BindHandoffService from the owner plugin.
func ExchangeLoginHandoff(code string) (*LoginHandoffPayload, error) {
	svc := BoundHandoffService()
	if svc == nil {
		return nil, errHandoffInvalid()
	}
	return svc.Exchange(code)
}
