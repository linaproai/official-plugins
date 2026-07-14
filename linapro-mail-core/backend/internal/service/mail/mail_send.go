// This file implements mailcap.Service send, probe, and fetch via transport SPI.

package mail

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
	"lina-plugin-linapro-mail-core/backend/internal/model/entity"
)

// Send delivers one message through the account outbound connection and SMTP SPI.
func (s *serviceImpl) Send(ctx context.Context, in mailcap.SendInput) (*mailcap.SendResult, error) {
	account, err := s.ResolveAccount(ctx, in.AccountID)
	if err != nil {
		return nil, err
	}
	if account.OutboundConnectionId <= 0 {
		return nil, bizerr.NewCode(mailcap.CodeMailAccountOutboundNotConfigured)
	}
	conn, err := s.GetConnection(ctx, account.OutboundConnectionId)
	if err != nil {
		return nil, err
	}
	if conn.Status != StatusEnabled {
		return nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", conn.Kind))
	}
	endpoint := connectionEndpoint(conn)
	_, transport, err := spi.ResolveOutbound(ctx, mailcap.Kind(conn.Kind), s.enablement())
	if err != nil {
		return nil, err
	}
	message := in.Message
	if strings.TrimSpace(message.From) == "" {
		message.From = account.FromAddress
	}
	return transport.Send(ctx, endpoint, message)
}

// ProbeConnection tests one connection through its transport SPI.
func (s *serviceImpl) ProbeConnection(ctx context.Context, connectionID int64) error {
	conn, err := s.GetConnection(ctx, connectionID)
	if err != nil {
		return err
	}
	endpoint := connectionEndpoint(conn)
	kind := mailcap.Kind(conn.Kind)
	switch kind {
	case mailcap.KindSMTP:
		_, transport, resolveErr := spi.ResolveOutbound(ctx, kind, s.enablement())
		if resolveErr != nil {
			return resolveErr
		}
		return transport.Probe(ctx, endpoint)
	case mailcap.KindIMAP, mailcap.KindPOP3:
		_, transport, resolveErr := spi.ResolveInbound(ctx, kind, s.enablement())
		if resolveErr != nil {
			return resolveErr
		}
		return transport.Probe(ctx, endpoint)
	default:
		return bizerr.NewCode(CodeConnectionKindInvalid)
	}
}

// Fetch pulls inbound messages for one account.
func (s *serviceImpl) Fetch(ctx context.Context, in mailcap.FetchInput) (*mailcap.FetchResult, error) {
	account, err := s.ResolveAccount(ctx, in.AccountID)
	if err != nil {
		return nil, err
	}
	if account.InboundConnectionId <= 0 {
		return nil, bizerr.NewCode(mailcap.CodeMailAccountInboundNotConfigured)
	}
	conn, err := s.GetConnection(ctx, account.InboundConnectionId)
	if err != nil {
		return nil, err
	}
	kind := mailcap.Kind(conn.Kind)
	if kind != mailcap.KindIMAP && kind != mailcap.KindPOP3 {
		return nil, bizerr.NewCode(mailcap.CodeMailInboundUnavailable)
	}
	_, transport, err := spi.ResolveInbound(ctx, kind, s.enablement())
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	return transport.Fetch(ctx, connectionEndpoint(conn), limit)
}

func connectionEndpoint(conn *entity.Connection) mailcap.ConnectionEndpoint {
	return mailcap.ConnectionEndpoint{
		ConnectionID: conn.Id,
		Kind:         mailcap.Kind(conn.Kind),
		Host:         conn.Host,
		Port:         conn.Port,
		Username:     conn.Username,
		// Secret material is resolved later by transport settings using SecretRef.
		// Endpoint only carries the reference-safe fields for protocol dialing.
		Secret:    conn.SecretRef,
		TLSMode:   conn.TlsMode,
		AuthMode:  conn.AuthMode,
		ExtraJSON: conn.ExtraJson,
	}
}

type stateEnablement struct {
	state plugincap.StateService
}

func (e stateEnablement) IsProviderEnabled(ctx context.Context, pluginID string) bool {
	if e.state == nil {
		return false
	}
	enabled, err := e.state.IsProviderEnabled(ctx, plugincap.PluginID(pluginID))
	if err != nil {
		return false
	}
	return enabled
}

func (s *serviceImpl) enablement() spi.EnablementReader {
	if s == nil || s.pluginState == nil {
		return nil
	}
	return stateEnablement{state: s.pluginState}
}
