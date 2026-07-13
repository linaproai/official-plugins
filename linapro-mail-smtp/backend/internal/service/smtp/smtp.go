// Package smtp implements the outbound SMTP transport SPI for linapro-mail-smtp.
package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
)

// PluginID is the stable SMTP transport plugin identifier.
const PluginID = "linapro-mail-smtp"

// Transport implements spi.OutboundTransport.
type Transport struct{}

// Ensure Transport implements OutboundTransport.
var _ spi.OutboundTransport = (*Transport)(nil)

// Factory constructs the SMTP outbound transport.
func Factory(_ context.Context, _ string) (spi.OutboundTransport, error) {
	return &Transport{}, nil
}

// Send delivers one message over SMTP.
func (t *Transport) Send(ctx context.Context, endpoint mailcap.ConnectionEndpoint, message mailcap.MailMessage) (*mailcap.SendResult, error) {
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	if len(message.To) == 0 {
		return nil, gerror.New("smtp: at least one recipient is required")
	}
	addr := net.JoinHostPort(endpoint.Host, fmt.Sprintf("%d", endpoint.Port))
	from := strings.TrimSpace(message.From)
	if from == "" {
		from = endpoint.Username
	}
	body := buildMessage(from, message)
	auth := smtp.PlainAuth("", endpoint.Username, endpoint.Secret, endpoint.Host)
	recipients := append([]string{}, message.To...)
	recipients = append(recipients, message.Cc...)
	recipients = append(recipients, message.Bcc...)

	switch strings.TrimSpace(endpoint.TLSMode) {
	case "tls":
		tlsConfig := &tls.Config{ServerName: endpoint.Host, MinVersion: tls.VersionTLS12}
		tlsDialer := &tls.Dialer{
			NetDialer: &net.Dialer{Timeout: 15 * time.Second},
			Config:    tlsConfig,
		}
		conn, err := tlsDialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, endpoint.Host)
		if err != nil {
			return nil, err
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return nil, err
		}
		if err = client.Mail(from); err != nil {
			return nil, err
		}
		for _, rcpt := range recipients {
			if err = client.Rcpt(rcpt); err != nil {
				return nil, err
			}
		}
		writer, err := client.Data()
		if err != nil {
			return nil, err
		}
		if _, err = writer.Write(body); err != nil {
			_ = writer.Close()
			return nil, err
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		if err = client.Quit(); err != nil {
			return nil, err
		}
	default:
		// starttls / disable: use net/smtp SendMail with plain dial; STARTTLS is
		// negotiated by the library when the server advertises it.
		if err := smtp.SendMail(addr, auth, from, recipients, body); err != nil {
			return nil, err
		}
	}
	return &mailcap.SendResult{MessageID: ""}, nil
}

// Probe validates SMTP connectivity with a short dial/TLS attempt.
func (t *Transport) Probe(ctx context.Context, endpoint mailcap.ConnectionEndpoint) error {
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	addr := net.JoinHostPort(endpoint.Host, fmt.Sprintf("%d", endpoint.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if strings.TrimSpace(endpoint.TLSMode) == "tls" {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: endpoint.Host, MinVersion: tls.VersionTLS12})
		if err = tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		_ = tlsConn.Close()
	}
	return nil
}

func validateEndpoint(endpoint mailcap.ConnectionEndpoint) error {
	if strings.TrimSpace(endpoint.Host) == "" {
		return gerror.New("smtp: host is required")
	}
	if endpoint.Port <= 0 {
		return gerror.New("smtp: port is required")
	}
	return nil
}

func buildMessage(from string, message mailcap.MailMessage) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(message.To, ", ") + "\r\n")
	if len(message.Cc) > 0 {
		b.WriteString("Cc: " + strings.Join(message.Cc, ", ") + "\r\n")
	}
	b.WriteString("Subject: " + message.Subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	if strings.TrimSpace(message.HTMLBody) != "" {
		b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		b.WriteString(message.HTMLBody)
	} else {
		b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		b.WriteString(message.TextBody)
	}
	return []byte(b.String())
}
