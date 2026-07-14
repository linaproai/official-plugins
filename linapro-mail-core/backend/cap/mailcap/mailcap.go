// Package mailcap defines the plugin-owned mail domain public contract for
// source consumers. Implementations and persistence live under backend/internal.
package mailcap

import (
	"context"
)

// OwnerPluginID is the stable plugin ID of the mail domain owner.
const OwnerPluginID = "linapro-mail-core"

// Kind identifies one transport protocol family.
type Kind string

// Transport kind constants.
const (
	// KindSMTP identifies SMTP outbound transport.
	KindSMTP Kind = "smtp"
	// KindIMAP identifies IMAP inbound transport.
	KindIMAP Kind = "imap"
	// KindPOP3 identifies POP3 inbound transport.
	KindPOP3 Kind = "pop3"
)

// String returns the canonical kind value.
func (k Kind) String() string {
	return string(k)
}

// ConnectionEndpoint is the protocol-neutral endpoint passed to transport SPI.
type ConnectionEndpoint struct {
	// ConnectionID is the mail-core connection identifier.
	ConnectionID int64
	// Kind is the transport kind.
	Kind Kind
	// Host is the mail server host.
	Host string
	// Port is the mail server port.
	Port int
	// Username is the authentication username.
	Username string
	// Secret is the resolved secret material for the transport call.
	// Callers outside mail-core must not persist this value.
	Secret string
	// TLSMode is the TLS mode token (for example disable, starttls, tls).
	TLSMode string
	// AuthMode is the authentication mode token (for example password).
	AuthMode string
	// ExtraJSON carries protocol extension fields without secret values.
	ExtraJSON string
}

// MailMessage is one outbound message payload.
type MailMessage struct {
	// From is the optional From address; empty uses Account default.
	From string
	// To contains recipient addresses.
	To []string
	// Cc contains carbon-copy addresses.
	Cc []string
	// Bcc contains blind carbon-copy addresses.
	Bcc []string
	// Subject is the message subject.
	Subject string
	// TextBody is the plain-text body.
	TextBody string
	// HTMLBody is the optional HTML body.
	HTMLBody string
}

// SendInput describes one mail send request through mailcap.
type SendInput struct {
	// AccountID selects one Account; zero means resolve the default Account.
	AccountID int64
	// Message is the outbound message.
	Message MailMessage
}

// SendResult describes one send outcome.
type SendResult struct {
	// MessageID is an optional provider/message identifier.
	MessageID string
}

// FetchInput describes one inbound fetch request.
type FetchInput struct {
	// AccountID selects one Account with inbound binding.
	AccountID int64
	// Limit bounds how many messages to fetch.
	Limit int
}

// FetchedMessage is one inbound message projection.
type FetchedMessage struct {
	// UID is the protocol message identifier when available.
	UID string
	// Subject is the message subject.
	Subject string
	// From is the sender address projection.
	From string
	// ReceivedAt is the receive timestamp as Unix milliseconds when known.
	ReceivedAt int64
}

// FetchResult groups inbound messages.
type FetchResult struct {
	// Items contains fetched messages.
	Items []FetchedMessage
}

// Service is the plugin-visible mail domain service.
type Service interface {
	// Send delivers one message through the Account outbound connection.
	Send(ctx context.Context, in SendInput) (*SendResult, error)
	// ProbeConnection tests one Connection through its transport SPI.
	ProbeConnection(ctx context.Context, connectionID int64) error
	// Fetch pulls inbound messages for one Account.
	Fetch(ctx context.Context, in FetchInput) (*FetchResult, error)
}
