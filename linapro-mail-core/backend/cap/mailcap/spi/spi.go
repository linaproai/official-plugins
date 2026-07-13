// Package spi defines transport provider registration and kind singleton
// resolution for linapro-mail-core. Protocol plugins register factories here;
// conflict detection and Resolve live with the owner.
package spi

import (
	"context"
	"sort"
	"strings"
	"sync"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginhost"

	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

// OwnerPluginID is the mail owner plugin ID.
const OwnerPluginID = mailcap.OwnerPluginID

// OutboundTransport sends mail for one connection endpoint.
type OutboundTransport interface {
	// Send delivers one message using the supplied endpoint.
	Send(ctx context.Context, endpoint mailcap.ConnectionEndpoint, message mailcap.MailMessage) (*mailcap.SendResult, error)
	// Probe validates connectivity for the endpoint.
	Probe(ctx context.Context, endpoint mailcap.ConnectionEndpoint) error
}

// InboundTransport fetches mail for one connection endpoint.
type InboundTransport interface {
	// Fetch pulls a bounded set of messages.
	Fetch(ctx context.Context, endpoint mailcap.ConnectionEndpoint, limit int) (*mailcap.FetchResult, error)
	// Probe validates connectivity for the endpoint.
	Probe(ctx context.Context, endpoint mailcap.ConnectionEndpoint) error
}

// OutboundFactory constructs one outbound transport for a plugin ID.
type OutboundFactory func(ctx context.Context, pluginID string) (OutboundTransport, error)

// InboundFactory constructs one inbound transport for a plugin ID.
type InboundFactory func(ctx context.Context, pluginID string) (InboundTransport, error)

// EnablementReader reports whether a provider plugin may serve requests.
type EnablementReader interface {
	// IsProviderEnabled reports whether pluginID is currently provider-enabled.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
}

type outboundRegistration struct {
	pluginID string
	kind     mailcap.Kind
	factory  OutboundFactory
}

type inboundRegistration struct {
	pluginID string
	kind     mailcap.Kind
	factory  InboundFactory
}

var registry = struct {
	sync.RWMutex
	outbound []outboundRegistration
	inbound  []inboundRegistration
}{}

// RegisterOutbound registers one outbound transport factory for a kind.
func RegisterOutbound(pluginID string, kind mailcap.Kind, factory OutboundFactory) error {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	if factory == nil {
		return bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	registry.Lock()
	defer registry.Unlock()
	for _, item := range registry.outbound {
		if item.pluginID == pluginID && item.kind == kind {
			return bizerr.NewCode(
				mailcap.CodeMailTransportConflict,
				bizerr.P("kind", kind.String()),
				bizerr.P("providerIds", pluginID),
			)
		}
	}
	registry.outbound = append(registry.outbound, outboundRegistration{
		pluginID: pluginID,
		kind:     kind,
		factory:  factory,
	})
	return nil
}

// RegisterInbound registers one inbound transport factory for a kind.
func RegisterInbound(pluginID string, kind mailcap.Kind, factory InboundFactory) error {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	if factory == nil {
		return bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	registry.Lock()
	defer registry.Unlock()
	for _, item := range registry.inbound {
		if item.pluginID == pluginID && item.kind == kind {
			return bizerr.NewCode(
				mailcap.CodeMailTransportConflict,
				bizerr.P("kind", kind.String()),
				bizerr.P("providerIds", pluginID),
			)
		}
	}
	registry.inbound = append(registry.inbound, inboundRegistration{
		pluginID: pluginID,
		kind:     kind,
		factory:  factory,
	})
	return nil
}

// ResolveOutbound selects the unique enabled outbound transport for kind.
func ResolveOutbound(
	ctx context.Context,
	kind mailcap.Kind,
	enablement EnablementReader,
) (pluginID string, transport OutboundTransport, err error) {
	active := activeOutboundPluginIDs(ctx, kind, enablement)
	if len(active) == 0 {
		return "", nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	if len(active) > 1 {
		return "", nil, bizerr.NewCode(
			mailcap.CodeMailTransportConflict,
			bizerr.P("kind", kind.String()),
			bizerr.P("providerIds", strings.Join(active, ",")),
		)
	}
	pluginID = active[0]
	factory := outboundFactoryFor(pluginID, kind)
	if factory == nil {
		return pluginID, nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	transport, err = factory(ctx, pluginID)
	if err != nil {
		return pluginID, nil, err
	}
	if transport == nil {
		return pluginID, nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	return pluginID, transport, nil
}

// ResolveInbound selects the unique enabled inbound transport for kind.
func ResolveInbound(
	ctx context.Context,
	kind mailcap.Kind,
	enablement EnablementReader,
) (pluginID string, transport InboundTransport, err error) {
	active := activeInboundPluginIDs(ctx, kind, enablement)
	if len(active) == 0 {
		return "", nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	if len(active) > 1 {
		return "", nil, bizerr.NewCode(
			mailcap.CodeMailTransportConflict,
			bizerr.P("kind", kind.String()),
			bizerr.P("providerIds", strings.Join(active, ",")),
		)
	}
	pluginID = active[0]
	factory := inboundFactoryFor(pluginID, kind)
	if factory == nil {
		return pluginID, nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	transport, err = factory(ctx, pluginID)
	if err != nil {
		return pluginID, nil, err
	}
	if transport == nil {
		return pluginID, nil, bizerr.NewCode(mailcap.CodeMailTransportUnavailable, bizerr.P("kind", kind.String()))
	}
	return pluginID, transport, nil
}

// KindForPlugin returns the registered transport kind for pluginID when known.
func KindForPlugin(pluginID string) (mailcap.Kind, bool) {
	pluginID = strings.TrimSpace(pluginID)
	registry.RLock()
	defer registry.RUnlock()
	for _, item := range registry.outbound {
		if item.pluginID == pluginID {
			return item.kind, true
		}
	}
	for _, item := range registry.inbound {
		if item.pluginID == pluginID {
			return item.kind, true
		}
	}
	return "", false
}

// HasEnabledPeerSameKind reports whether another enabled transport already
// occupies the same kind as targetPluginID.
func HasEnabledPeerSameKind(ctx context.Context, targetPluginID string, enablement EnablementReader) (bool, mailcap.Kind, []string) {
	kind, ok := KindForPlugin(targetPluginID)
	if !ok {
		return false, "", nil
	}
	var peers []string
	if isOutboundKind(kind) {
		for _, id := range activeOutboundPluginIDs(ctx, kind, enablement) {
			if id != strings.TrimSpace(targetPluginID) {
				peers = append(peers, id)
			}
		}
	}
	if isInboundKind(kind) {
		for _, id := range activeInboundPluginIDs(ctx, kind, enablement) {
			if id != strings.TrimSpace(targetPluginID) {
				peers = append(peers, id)
			}
		}
	}
	sort.Strings(peers)
	return len(peers) > 0, kind, peers
}

// GlobalBeforeEnableVeto implements GlobalBeforeEnable for mail-core.
// enablement may be nil; when nil, only registration presence is considered and
// the handler allows enable (runtime Resolve still enforces uniqueness).
func GlobalBeforeEnableVeto(
	ctx context.Context,
	input pluginhost.SourcePluginGlobalLifecycleInput,
	enablement EnablementReader,
) (ok bool, reason string, err error) {
	if input == nil {
		return true, "", nil
	}
	target := strings.TrimSpace(input.TargetPluginID())
	if target == "" {
		return true, "", nil
	}
	if _, registered := KindForPlugin(target); !registered {
		return true, "", nil
	}
	if enablement == nil {
		return true, "", nil
	}
	conflict, kind, peers := HasEnabledPeerSameKind(ctx, target, enablement)
	if !conflict {
		return true, "", nil
	}
	return false, mailcap.CodeMailTransportConflict.MessageKey(), bizerr.NewCode(
		mailcap.CodeMailTransportConflict,
		bizerr.P("kind", kind.String()),
		bizerr.P("providerIds", strings.Join(peers, ",")),
	)
}

func isOutboundKind(kind mailcap.Kind) bool {
	return kind == mailcap.KindSMTP
}

func isInboundKind(kind mailcap.Kind) bool {
	return kind == mailcap.KindIMAP || kind == mailcap.KindPOP3
}

func activeOutboundPluginIDs(ctx context.Context, kind mailcap.Kind, enablement EnablementReader) []string {
	registry.RLock()
	defer registry.RUnlock()
	ids := make([]string, 0)
	for _, item := range registry.outbound {
		if item.kind != kind {
			continue
		}
		if enablement != nil && !enablement.IsProviderEnabled(ctx, item.pluginID) {
			continue
		}
		if enablement == nil {
			// Without enablement, treat registered providers as candidates only
			// for conflict enumeration when explicitly requested; Resolve uses
			// enablement from host. Keep empty when nil to fail closed.
			continue
		}
		ids = append(ids, item.pluginID)
	}
	sort.Strings(ids)
	return ids
}

func activeInboundPluginIDs(ctx context.Context, kind mailcap.Kind, enablement EnablementReader) []string {
	registry.RLock()
	defer registry.RUnlock()
	ids := make([]string, 0)
	for _, item := range registry.inbound {
		if item.kind != kind {
			continue
		}
		if enablement != nil && !enablement.IsProviderEnabled(ctx, item.pluginID) {
			continue
		}
		if enablement == nil {
			continue
		}
		ids = append(ids, item.pluginID)
	}
	sort.Strings(ids)
	return ids
}

func outboundFactoryFor(pluginID string, kind mailcap.Kind) OutboundFactory {
	registry.RLock()
	defer registry.RUnlock()
	for _, item := range registry.outbound {
		if item.pluginID == pluginID && item.kind == kind {
			return item.factory
		}
	}
	return nil
}

func inboundFactoryFor(pluginID string, kind mailcap.Kind) InboundFactory {
	registry.RLock()
	defer registry.RUnlock()
	for _, item := range registry.inbound {
		if item.pluginID == pluginID && item.kind == kind {
			return item.factory
		}
	}
	return nil
}

// ResetRegistryForTest clears SPI registrations. Tests only.
func ResetRegistryForTest() {
	registry.Lock()
	defer registry.Unlock()
	registry.outbound = nil
	registry.inbound = nil
}
