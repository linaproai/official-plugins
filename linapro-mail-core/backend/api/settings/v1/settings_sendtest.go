package v1

import "github.com/gogf/gf/v2/frame/g"

// SendTestMailReq sends one diagnostic email using the current form SMTP values.
// It does not persist settings. Empty passwords reuse previously stored secrets.
// Empty fromAddress defaults to smtpUsername.
type SendTestMailReq struct {
	g.Meta          `path:"/mail/settings/send-test" method:"post" tags:"Mail Settings" summary:"Send platform test mail" dc:"Send one diagnostic email through the SMTP settings currently entered on the admin form (not necessarily the last saved default account). Empty passwords reuse stored secrets. Empty fromAddress defaults to smtpUsername." permission:"linapro-mail-core:settings:view"`
	Name            string `json:"name" v:"max-length:128" dc:"Optional internal account label" eg:"platform"`
	FromAddress     string `json:"fromAddress" v:"max-length:255" dc:"Default From address; empty defaults to smtpUsername" eg:"noreply@example.com"`
	SmtpHost        string `json:"smtpHost" v:"required|max-length:255" dc:"SMTP server host" eg:"smtp.example.com"`
	SmtpPort        int    `json:"smtpPort" v:"required|min:1|max:65535" dc:"SMTP server port" eg:"587"`
	SmtpUsername    string `json:"smtpUsername" v:"required|max-length:255" dc:"Mailbox login account shared by SMTP and inbound" eg:"noreply@example.com"`
	SmtpPassword    string `json:"smtpPassword" v:"max-length:512" dc:"SMTP password; empty reuses stored" eg:""`
	SmtpTlsMode     string `json:"smtpTlsMode" v:"max-length:32" dc:"SMTP TLS mode: disable, starttls, tls" eg:"starttls"`
	InboundKind     string `json:"inboundKind" v:"max-length:16" dc:"Inbound protocol ignored for send-test; accepted for form reuse" eg:"none"`
	InboundHost     string `json:"inboundHost" v:"max-length:255" dc:"Inbound host ignored for send-test" eg:""`
	InboundPort     int    `json:"inboundPort" v:"min:0|max:65535" dc:"Inbound port ignored for send-test" eg:"0"`
	InboundUsername string `json:"inboundUsername" v:"max-length:255" dc:"Inbound username ignored for send-test" eg:""`
	InboundPassword string `json:"inboundPassword" v:"max-length:512" dc:"Inbound password ignored for send-test" eg:""`
	InboundTlsMode  string `json:"inboundTlsMode" v:"max-length:32" dc:"Inbound TLS mode ignored for send-test" eg:""`
	// To is the recipient address for the diagnostic message.
	To string `json:"to" v:"required|max-length:255" dc:"Recipient email address for the diagnostic message" eg:"ops@example.com"`
	// Subject is the optional subject; empty uses a server default.
	Subject string `json:"subject" v:"max-length:255" dc:"Optional subject; empty uses a server default" eg:"Mail configuration test"`
	// Body is the plain-text body of the diagnostic message.
	Body string `json:"body" v:"required|max-length:4000" dc:"Plain-text body of the diagnostic message" eg:"This is a test message from LinaPro mail settings."`
}

// SendTestMailRes reports whether the diagnostic send completed.
type SendTestMailRes struct {
	g.Meta  `mime:"application/json"`
	OK      bool   `json:"ok" dc:"Whether the diagnostic mail was accepted by the SMTP transport" eg:"true"`
	Message string `json:"message" dc:"Human-readable send result or failure detail" eg:"ok"`
}
