// Package handoff implements the process-local, single-use SPA delivery store
// for host-minted external-login outcomes. Token minting stays on the host;
// this package only holds the short-lived payload so protocol plugins never put
// JWTs into redirect URLs and the host does not need a dedicated public exchange
// HTTP surface.
//
// Runtime boundary: the store is process-local and shared only within the
// current host process (source plugins share that process). Multi-node
// cluster sharing is a later hardening step; the same limitation applies to
// verified-identity tickets in the identity service.
package handoff

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

const (
	loginHandoffTTL      = 2 * time.Minute
	loginHandoffIDBytes  = 16
	loginHandoffIDPrefix = "elh_"
)

// Ensure Service implements the public handoff contract used by protocol plugins.
var _ extidcap.HandoffService = (*Service)(nil)

// Service is the process-local one-time login handoff store.
type Service struct {
	mu    sync.Mutex
	items map[string]*loginHandoffRecord
}

type loginHandoffRecord struct {
	Payload   extidcap.LoginHandoffPayload
	ExpiresAt time.Time
	Consumed  bool
}

// New creates an empty handoff store.
func New() *Service {
	return &Service{items: make(map[string]*loginHandoffRecord)}
}

// Create stores one host login outcome and returns a single-use code.
func (s *Service) Create(payload extidcap.LoginHandoffPayload) (string, error) {
	if s == nil {
		return "", bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	if strings.TrimSpace(payload.AccessToken) == "" && strings.TrimSpace(payload.PreToken) == "" {
		return "", bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	id, err := newLoginHandoffID()
	if err != nil {
		return "", err
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(now)
	s.items[id] = &loginHandoffRecord{
		Payload:   payload,
		ExpiresAt: now.Add(loginHandoffTTL),
	}
	return id, nil
}

// CreateFromHost maps a host extlogin.LoginOutput into a handoff.
func (s *Service) CreateFromHost(out *extlogin.LoginOutput) (string, error) {
	if s == nil || out == nil {
		return "", bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	return s.Create(extidcap.LoginHandoffPayload{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		PreToken:         out.PreToken,
		TenantCandidates: out.TenantCandidates,
	})
}

// Exchange consumes one handoff code and returns the stored payload.
func (s *Service) Exchange(code string) (*extidcap.LoginHandoffPayload, error) {
	if s == nil {
		return nil, bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.items[code]
	if !ok || rec == nil || rec.Consumed || now.After(rec.ExpiresAt) {
		return nil, bizerr.NewCode(extidcap.CodeLoginHandoffInvalid)
	}
	rec.Consumed = true
	delete(s.items, code)
	payload := rec.Payload
	return &payload, nil
}

func (s *Service) purgeLocked(now time.Time) {
	for id, rec := range s.items {
		if rec == nil || rec.Consumed || now.After(rec.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func newLoginHandoffID() (string, error) {
	buf := make([]byte, loginHandoffIDBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return loginHandoffIDPrefix + hex.EncodeToString(buf), nil
}
