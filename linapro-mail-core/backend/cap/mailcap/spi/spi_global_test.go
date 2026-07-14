// This file covers global enable conflict helpers for transport kind uniqueness.

package spi

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/pluginhost"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

func TestGlobalBeforeEnableVetoSecondSMTP(t *testing.T) {
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

	// First smtp already enabled: enabling the alt must be vetoed.
	ok, reason, err := GlobalBeforeEnableVeto(ctx, pluginhost.NewSourcePluginGlobalLifecycleInput(
		"linapro-mail-smtp-alt",
		pluginhost.LifecycleHookGlobalBeforeEnable.String(),
	), staticEnablement{
		"linapro-mail-smtp": true,
	})
	if ok || reason == "" {
		t.Fatalf("expected veto, ok=%v reason=%q err=%v", ok, reason, err)
	}

	// Non-mail plugin must pass quickly.
	ok, reason, err = GlobalBeforeEnableVeto(ctx, pluginhost.NewSourcePluginGlobalLifecycleInput(
		"linapro-content-notice",
		pluginhost.LifecycleHookGlobalBeforeEnable.String(),
	), staticEnablement{"linapro-mail-smtp": true})
	if !ok || reason != "" || err != nil {
		t.Fatalf("expected non-mail pass, ok=%v reason=%q err=%v", ok, reason, err)
	}
}

func TestSMTPAndIMAPCoexist(t *testing.T) {
	t.Cleanup(ResetRegistryForTest)
	ctx := context.Background()
	if err := RegisterOutbound("linapro-mail-smtp", mailcap.KindSMTP, func(context.Context, string) (OutboundTransport, error) {
		return stubOutbound{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterInbound("linapro-mail-imap", mailcap.KindIMAP, func(context.Context, string) (InboundTransport, error) {
		return stubInbound{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	enablement := staticEnablement{
		"linapro-mail-smtp": true,
		"linapro-mail-imap": true,
	}
	if _, _, err := ResolveOutbound(ctx, mailcap.KindSMTP, enablement); err != nil {
		t.Fatalf("smtp resolve: %v", err)
	}
	if _, _, err := ResolveInbound(ctx, mailcap.KindIMAP, enablement); err != nil {
		t.Fatalf("imap resolve: %v", err)
	}
	// Enabling imap while smtp is enabled must not conflict.
	ok, _, err := GlobalBeforeEnableVeto(ctx, pluginhost.NewSourcePluginGlobalLifecycleInput(
		"linapro-mail-imap",
		pluginhost.LifecycleHookGlobalBeforeEnable.String(),
	), staticEnablement{"linapro-mail-smtp": true})
	if !ok || err != nil {
		t.Fatalf("smtp+imap should coexist: ok=%v err=%v", ok, err)
	}
}

type stubInbound struct{}

func (stubInbound) Fetch(context.Context, mailcap.ConnectionEndpoint, int) (*mailcap.FetchResult, error) {
	return &mailcap.FetchResult{Items: nil}, nil
}

func (stubInbound) Probe(context.Context, mailcap.ConnectionEndpoint) error {
	return nil
}
