// This file verifies mail transport SPI registration, resolve rules, and
// global enable conflict helpers.

package spi

import (
	"context"
	"testing"

	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

type staticEnablement map[string]bool

func (e staticEnablement) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return e[pluginID]
}

func TestResolveOutboundUniqueAndConflict(t *testing.T) {
	t.Cleanup(ResetRegistryForTest)
	ctx := context.Background()
	if err := RegisterOutbound("linapro-mail-smtp", mailcap.KindSMTP, func(context.Context, string) (OutboundTransport, error) {
		return stubOutbound{}, nil
	}); err != nil {
		t.Fatalf("register smtp: %v", err)
	}
	if err := RegisterOutbound("linapro-mail-smtp-alt", mailcap.KindSMTP, func(context.Context, string) (OutboundTransport, error) {
		return stubOutbound{}, nil
	}); err != nil {
		t.Fatalf("register smtp alt: %v", err)
	}

	pluginID, transport, err := ResolveOutbound(ctx, mailcap.KindSMTP, staticEnablement{
		"linapro-mail-smtp": true,
	})
	if err != nil || pluginID != "linapro-mail-smtp" || transport == nil {
		t.Fatalf("unique resolve failed: id=%s err=%v transport=%v", pluginID, err, transport)
	}

	_, _, err = ResolveOutbound(ctx, mailcap.KindSMTP, staticEnablement{
		"linapro-mail-smtp":     true,
		"linapro-mail-smtp-alt": true,
	})
	if err == nil {
		t.Fatal("expected conflict when two smtp providers enabled")
	}
}

func TestHasEnabledPeerSameKind(t *testing.T) {
	t.Cleanup(ResetRegistryForTest)
	ctx := context.Background()
	if err := RegisterOutbound("linapro-mail-smtp", mailcap.KindSMTP, func(context.Context, string) (OutboundTransport, error) {
		return stubOutbound{}, nil
	}); err != nil {
		t.Fatalf("register smtp: %v", err)
	}
	if err := RegisterOutbound("linapro-mail-smtp-alt", mailcap.KindSMTP, func(context.Context, string) (OutboundTransport, error) {
		return stubOutbound{}, nil
	}); err != nil {
		t.Fatalf("register smtp alt: %v", err)
	}
	conflict, kind, peers := HasEnabledPeerSameKind(ctx, "linapro-mail-smtp-alt", staticEnablement{
		"linapro-mail-smtp": true,
	})
	if !conflict || kind != mailcap.KindSMTP || len(peers) != 1 || peers[0] != "linapro-mail-smtp" {
		t.Fatalf("unexpected conflict result: conflict=%v kind=%s peers=%v", conflict, kind, peers)
	}
}

type stubOutbound struct{}

func (stubOutbound) Send(context.Context, mailcap.ConnectionEndpoint, mailcap.MailMessage) (*mailcap.SendResult, error) {
	return &mailcap.SendResult{}, nil
}

func (stubOutbound) Probe(context.Context, mailcap.ConnectionEndpoint) error {
	return nil
}
