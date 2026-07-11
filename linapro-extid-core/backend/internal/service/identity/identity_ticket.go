// identity_ticket.go implements short-lived verified-identity tickets used for
// bind and login orchestration. Tickets are process-local; source plugins share
// the host process so a multi-node shared store is a later hardening step.

package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-extid-core/backend/cap/extidcap"
)

const (
	ticketTTL             = 5 * time.Minute
	ticketIDBytes         = 16
	ticketIDPrefix        = "extid_"
	defaultAssuranceLevel = "code_exchange"
)

// ticketRecord is one in-memory verified-identity ticket.
type ticketRecord struct {
	Identity  extidcap.VerifiedIdentity
	ExpiresAt time.Time
	Consumed  bool
}

// ticketStore is a process-local single-use ticket registry.
type ticketStore struct {
	mu    sync.Mutex
	items map[string]*ticketRecord
}

// newTicketStore creates an empty ticket store.
func newTicketStore() *ticketStore {
	return &ticketStore{items: make(map[string]*ticketRecord)}
}

// Issue stores one verified identity and returns a ticket id.
func (s *ticketStore) Issue(_ context.Context, identity extidcap.VerifiedIdentity) (*extidcap.TicketIssueResult, error) {
	if s == nil {
		return nil, bizerr.NewCode(CodeTicketUnavailable)
	}
	identity.Provider = strings.TrimSpace(identity.Provider)
	identity.Subject = strings.TrimSpace(identity.Subject)
	if identity.Provider == "" || identity.Subject == "" {
		return nil, bizerr.NewCode(CodeIdentityInvalid)
	}
	now := time.Now()
	if identity.IssuedAt.IsZero() {
		identity.IssuedAt = now
	}
	expiresAt := now.Add(ticketTTL)
	if !identity.ExpiresAt.IsZero() && identity.ExpiresAt.Before(expiresAt) {
		expiresAt = identity.ExpiresAt
	}
	identity.ExpiresAt = expiresAt
	if strings.TrimSpace(identity.AssuranceLevel) == "" {
		identity.AssuranceLevel = defaultAssuranceLevel
	}
	id, err := newTicketID()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeTicketUnavailable)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(now)
	s.items[id] = &ticketRecord{Identity: identity, ExpiresAt: expiresAt}
	return &extidcap.TicketIssueResult{TicketID: id, ExpiresAt: expiresAt}, nil
}

// Peek returns a ticket without consuming it.
func (s *ticketStore) Peek(_ context.Context, ticketID string) (*extidcap.VerifiedIdentity, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, bizerr.NewCode(CodeTicketInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.items[ticketID]
	if !ok || rec == nil || rec.Consumed || time.Now().After(rec.ExpiresAt) {
		return nil, bizerr.NewCode(CodeTicketInvalid)
	}
	identity := rec.Identity
	return &identity, nil
}

// Consume returns and invalidates a ticket.
func (s *ticketStore) Consume(_ context.Context, ticketID string) (*extidcap.VerifiedIdentity, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, bizerr.NewCode(CodeTicketInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.items[ticketID]
	if !ok || rec == nil || rec.Consumed || time.Now().After(rec.ExpiresAt) {
		return nil, bizerr.NewCode(CodeTicketInvalid)
	}
	rec.Consumed = true
	delete(s.items, ticketID)
	identity := rec.Identity
	return &identity, nil
}

// Invalidate drops a ticket if present.
func (s *ticketStore) Invalidate(_ context.Context, ticketID string) error {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return bizerr.NewCode(CodeTicketInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, ticketID)
	return nil
}

func (s *ticketStore) purgeExpiredLocked(now time.Time) {
	for id, rec := range s.items {
		if rec == nil || rec.Consumed || now.After(rec.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func newTicketID() (string, error) {
	buf := make([]byte, ticketIDBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return ticketIDPrefix + hex.EncodeToString(buf), nil
}
