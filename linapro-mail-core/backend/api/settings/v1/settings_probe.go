package v1

import "github.com/gogf/gf/v2/frame/g"

// TestSettingsReq probes SMTP and optional inbound connectivity without persisting.
// Empty passwords use previously stored secrets when an account already exists.
// Empty fromAddress defaults to smtpUsername for consistency with save.
type TestSettingsReq struct {
	g.Meta          `path:"/mail/settings/test" method:"post" tags:"Mail Settings" summary:"Test platform mail connectivity" dc:"Probe SMTP and optional IMAP/POP3 with form values without persisting. Empty passwords reuse stored secrets." permission:"linapro-mail-core:settings:view"`
	Name            string `json:"name" v:"max-length:128" dc:"Optional internal account label"`
	FromAddress     string `json:"fromAddress" v:"max-length:255" dc:"Default From address; empty defaults to smtpUsername"`
	SmtpHost        string `json:"smtpHost" v:"required|max-length:255" dc:"SMTP server host"`
	SmtpPort        int    `json:"smtpPort" v:"required|min:1|max:65535" dc:"SMTP server port"`
	SmtpUsername    string `json:"smtpUsername" v:"required|max-length:255" dc:"Mailbox login account shared by SMTP and inbound"`
	SmtpPassword    string `json:"smtpPassword" v:"max-length:512" dc:"SMTP password; empty reuses stored"`
	SmtpTlsMode     string `json:"smtpTlsMode" v:"max-length:32" dc:"SMTP TLS mode"`
	InboundKind     string `json:"inboundKind" v:"max-length:16" dc:"Inbound protocol: none, imap, pop3"`
	InboundHost     string `json:"inboundHost" v:"max-length:255" dc:"Inbound server host"`
	InboundPort     int    `json:"inboundPort" v:"min:0|max:65535" dc:"Inbound server port"`
	InboundUsername string `json:"inboundUsername" v:"max-length:255" dc:"Inbound username; empty defaults to smtpUsername"`
	InboundPassword string `json:"inboundPassword" v:"max-length:512" dc:"Inbound password; empty reuses stored"`
	InboundTlsMode  string `json:"inboundTlsMode" v:"max-length:32" dc:"Inbound TLS mode"`
}

// TestSettingsRes reports probe outcome.
type TestSettingsRes struct {
	g.Meta  `mime:"application/json"`
	OK      bool   `json:"ok" dc:"Whether all required probes succeeded"`
	Message string `json:"message" dc:"Human-readable probe result"`
}
