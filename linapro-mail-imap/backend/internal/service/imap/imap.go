// Package imap implements inbound IMAP transport SPI for linapro-mail-imap.
// Probe authenticates when credentials are present so admin "receive test"
// can validate login capability; Fetch remains staged until a full client lands.
package imap

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
	"lina-plugin-linapro-mail-core/backend/cap/mailcap/spi"
)

// PluginID is the stable IMAP transport plugin identifier.
const PluginID = "linapro-mail-imap"

// Transport implements spi.InboundTransport.
type Transport struct{}

var _ spi.InboundTransport = (*Transport)(nil)

// Factory constructs the IMAP inbound transport.
func Factory(_ context.Context, _ string) (spi.InboundTransport, error) {
	return &Transport{}, nil
}

// Fetch retrieves messages. Full IMAP protocol parsing is staged; Probe is
// available first while Fetch returns a structured not-ready error until the
// protocol client lands.
func (t *Transport) Fetch(_ context.Context, endpoint mailcap.ConnectionEndpoint, _ int) (*mailcap.FetchResult, error) {
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	return nil, gerror.New("imap: message fetch client is not implemented yet; connection probe is available")
}

// Probe validates IMAP reachability and, when credentials are present, LOGIN.
func (t *Transport) Probe(ctx context.Context, endpoint mailcap.ConnectionEndpoint) error {
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	conn, err := dialIMAP(ctx, endpoint)
	if err != nil {
		return err
	}
	// Close the latest connection (plain or upgraded TLS) on return.
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	if err = setConnDeadline(conn, 15*time.Second); err != nil {
		return err
	}
	greeting, err := readIMAPLine(reader)
	if err != nil {
		return err
	}
	upper := strings.ToUpper(greeting)
	if !strings.HasPrefix(upper, "* OK") && !strings.HasPrefix(upper, "* PREAUTH") {
		return gerror.Newf("imap: unexpected greeting: %s", truncate(greeting, 200))
	}

	tlsMode := strings.TrimSpace(strings.ToLower(endpoint.TLSMode))
	if tlsMode == "starttls" {
		if err = writeIMAP(conn, "a000 STARTTLS"); err != nil {
			return err
		}
		resp, readErr := readIMAPTagged(reader, "a000")
		if readErr != nil {
			return readErr
		}
		if !strings.HasPrefix(strings.ToUpper(resp), "A000 OK") {
			return gerror.Newf("imap: STARTTLS failed: %s", truncate(resp, 200))
		}
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: endpoint.Host,
			MinVersion: tls.VersionTLS12,
		})
		if err = tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		conn = tlsConn
		reader = bufio.NewReader(conn)
		if err = setConnDeadline(conn, 15*time.Second); err != nil {
			return err
		}
	}

	username := strings.TrimSpace(endpoint.Username)
	secret := endpoint.Secret
	if username == "" || strings.TrimSpace(secret) == "" {
		// Reachability-only when credentials are absent.
		return nil
	}
	if err = writeIMAP(conn, fmt.Sprintf("a001 LOGIN %s %s", quoteIMAP(username), quoteIMAP(secret))); err != nil {
		return err
	}
	loginResp, err := readIMAPTagged(reader, "a001")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToUpper(loginResp), "A001 OK") {
		return gerror.Newf("imap: login failed: %s", truncate(loginResp, 200))
	}
	_ = writeIMAP(conn, "a002 LOGOUT")
	return nil
}

func dialIMAP(ctx context.Context, endpoint mailcap.ConnectionEndpoint) (net.Conn, error) {
	addr := net.JoinHostPort(endpoint.Host, fmt.Sprintf("%d", endpoint.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	tlsMode := strings.TrimSpace(strings.ToLower(endpoint.TLSMode))
	if tlsMode == "tls" {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config: &tls.Config{
				ServerName: endpoint.Host,
				MinVersion: tls.VersionTLS12,
			},
		}
		return tlsDialer.DialContext(ctx, "tcp", addr)
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

func writeIMAP(conn net.Conn, command string) error {
	_, err := conn.Write([]byte(command + "\r\n"))
	return err
}

func readIMAPLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// readIMAPTagged drains untagged lines until the expected tag response arrives.
func readIMAPTagged(reader *bufio.Reader, tag string) (string, error) {
	prefix := strings.ToUpper(tag) + " "
	for {
		line, err := readIMAPLine(reader)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(strings.ToUpper(line), prefix) {
			return line, nil
		}
	}
}

func quoteIMAP(value string) string {
	// IMAP quoted-string: escape backslash and double-quote.
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func setConnDeadline(conn net.Conn, d time.Duration) error {
	return conn.SetDeadline(time.Now().Add(d))
}

func truncate(value string, limit int) string {
	if limit <= 0 || len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func validateEndpoint(endpoint mailcap.ConnectionEndpoint) error {
	if strings.TrimSpace(endpoint.Host) == "" {
		return gerror.New("imap: host is required")
	}
	if endpoint.Port <= 0 {
		return gerror.New("imap: port is required")
	}
	return nil
}
