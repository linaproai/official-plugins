package v1

import "github.com/gogf/gf/v2/frame/g"

// SaveSettingsReq persists the single platform mail account and its connections.
// Empty passwords keep previously stored secrets.
// Empty fromAddress defaults to smtpUsername. Name is optional and may be derived server-side.
type SaveSettingsReq struct {
	g.Meta          `path:"/mail/settings" method:"put" tags:"Mail Settings" summary:"Save platform mail settings" dc:"Upsert the single platform mail account with SMTP and optional IMAP/POP3 inbound. Empty passwords keep previous values. Empty fromAddress defaults to smtpUsername." permission:"linapro-mail-core:settings:update"`
	Name            string `json:"name" v:"max-length:128" dc:"Optional internal account label; omitted values are derived from fromAddress or smtpUsername"`
	FromAddress     string `json:"fromAddress" v:"max-length:255" dc:"Default From address; empty defaults to smtpUsername"`
	SmtpHost        string `json:"smtpHost" v:"required|max-length:255" dc:"SMTP server host"`
	SmtpPort        int    `json:"smtpPort" v:"required|min:1|max:65535" dc:"SMTP server port"`
	SmtpUsername    string `json:"smtpUsername" v:"required|max-length:255" dc:"Mailbox login account shared by SMTP and inbound"`
	SmtpPassword    string `json:"smtpPassword" v:"max-length:512" dc:"SMTP password; empty keeps previous"`
	SmtpTlsMode     string `json:"smtpTlsMode" v:"max-length:32" dc:"SMTP TLS mode: disable, starttls, tls"`
	InboundKind     string `json:"inboundKind" v:"max-length:16" dc:"Inbound protocol: none, imap, pop3"`
	InboundHost     string `json:"inboundHost" v:"max-length:255" dc:"Inbound server host"`
	InboundPort     int    `json:"inboundPort" v:"min:0|max:65535" dc:"Inbound server port"`
	InboundUsername string `json:"inboundUsername" v:"max-length:255" dc:"Inbound username; empty defaults to smtpUsername"`
	InboundPassword string `json:"inboundPassword" v:"max-length:512" dc:"Inbound password; empty keeps previous"`
	InboundTlsMode  string `json:"inboundTlsMode" v:"max-length:32" dc:"Inbound TLS mode: disable, starttls, tls"`
}

// SaveSettingsRes returns the masked projection after save.
type SaveSettingsRes struct {
	g.Meta   `mime:"application/json"`
	Settings *SettingsItem `json:"settings" dc:"Masked settings after save"`
}
