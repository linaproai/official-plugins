// Package v1 declares platform single-account mail settings DTOs.
package v1

// SettingsItem is the masked projection for the admin settings page.
// Secrets are never returned in plaintext; empty password on save keeps the previous value.
type SettingsItem struct {
	// Name is an internal account label (may be derived from fromAddress/smtpUsername).
	Name string `json:"name" dc:"Internal account label; may be derived when not supplied by the admin form"`
	// FromAddress is the default From address for outbound mail (defaults to smtpUsername when empty on save).
	FromAddress string `json:"fromAddress" dc:"Default From address for outbound mail"`
	// SMTP outbound connection fields.
	SmtpHost               string `json:"smtpHost" dc:"SMTP server host"`
	SmtpPort               int    `json:"smtpPort" dc:"SMTP server port"`
	SmtpUsername           string `json:"smtpUsername" dc:"Mailbox login account shared by SMTP and inbound"`
	SmtpPasswordConfigured bool   `json:"smtpPasswordConfigured" dc:"Whether SMTP password is stored"`
	SmtpTlsMode            string `json:"smtpTlsMode" dc:"SMTP TLS mode: disable, starttls, tls"`
	// InboundKind is empty/none, imap, or pop3.
	InboundKind               string `json:"inboundKind" dc:"Inbound protocol: none, imap, pop3"`
	InboundHost               string `json:"inboundHost" dc:"Inbound server host"`
	InboundPort               int    `json:"inboundPort" dc:"Inbound server port"`
	InboundUsername           string `json:"inboundUsername" dc:"Inbound username"`
	InboundPasswordConfigured bool   `json:"inboundPasswordConfigured" dc:"Whether inbound password is stored"`
	InboundTlsMode            string `json:"inboundTlsMode" dc:"Inbound TLS mode: disable, starttls, tls"`
	// Configured reports whether a platform account already exists.
	Configured bool `json:"configured" dc:"Whether a platform mail account is configured"`
}
