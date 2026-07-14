// This file declares the platform settings receive-test API DTO.
// It probes inbound IMAP/POP3 using the current admin form values without persisting.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ReceiveTestMailReq probes inbound receive capability using current form values.
// It does not persist settings. Empty passwords reuse previously stored secrets.
// SMTP fields are accepted for form reuse but ignored for the inbound probe itself.
type ReceiveTestMailReq struct {
	g.Meta          `path:"/mail/settings/receive-test" method:"post" tags:"Mail Settings" summary:"Test platform mail receive" dc:"Probe IMAP or POP3 receive capability through the inbound settings currently entered on the admin form (not necessarily the last saved default account). Empty passwords reuse stored secrets. Requires inboundKind imap or pop3." permission:"linapro-mail-core:settings:view"`
	Name            string `json:"name" v:"max-length:128" dc:"Optional internal account label" eg:"platform"`
	FromAddress     string `json:"fromAddress" v:"max-length:255" dc:"From address ignored for receive-test; accepted for form reuse" eg:"noreply@example.com"`
	SmtpHost        string `json:"smtpHost" v:"max-length:255" dc:"SMTP host ignored for receive-test; accepted for form reuse" eg:"smtp.example.com"`
	SmtpPort        int    `json:"smtpPort" v:"min:0|max:65535" dc:"SMTP port ignored for receive-test" eg:"587"`
	SmtpUsername    string `json:"smtpUsername" v:"max-length:255" dc:"Mailbox login account shared by SMTP and inbound" eg:"noreply@example.com"`
	SmtpPassword    string `json:"smtpPassword" v:"max-length:512" dc:"SMTP password; empty reuses stored; used as inbound fallback when inbound password is empty" eg:""`
	SmtpTlsMode     string `json:"smtpTlsMode" v:"max-length:32" dc:"SMTP TLS mode ignored for receive-test" eg:"starttls"`
	InboundKind     string `json:"inboundKind" v:"required|max-length:16" dc:"Inbound protocol: imap or pop3 (none is rejected)" eg:"imap"`
	InboundHost     string `json:"inboundHost" v:"required|max-length:255" dc:"Inbound server host" eg:"imap.example.com"`
	InboundPort     int    `json:"inboundPort" v:"required|min:1|max:65535" dc:"Inbound server port" eg:"993"`
	InboundUsername string `json:"inboundUsername" v:"max-length:255" dc:"Inbound username; empty defaults to smtpUsername" eg:""`
	InboundPassword string `json:"inboundPassword" v:"max-length:512" dc:"Inbound password; empty reuses stored or smtp password" eg:""`
	InboundTlsMode  string `json:"inboundTlsMode" v:"max-length:32" dc:"Inbound TLS mode: disable, starttls, tls" eg:"tls"`
}

// ReceiveTestMailRes reports whether the inbound receive probe completed.
type ReceiveTestMailRes struct {
	g.Meta  `mime:"application/json"`
	OK      bool   `json:"ok" dc:"Whether the inbound receive probe succeeded" eg:"true"`
	Message string `json:"message" dc:"Human-readable receive probe result or failure detail" eg:"ok"`
}
