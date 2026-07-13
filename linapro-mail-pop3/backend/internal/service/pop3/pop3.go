// Package pop3 implements inbound POP3 transport SPI for linapro-mail-pop3.
// Probe authenticates when credentials are present so admin "receive test"
// can validate login capability; Fetch remains staged until a full client lands.
package pop3

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

// PluginID is the stable POP3 transport plugin identifier.
const PluginID = "linapro-mail-pop3"

// Transport implements spi.InboundTransport.
type Transport struct{}

var _ spi.InboundTransport = (*Transport)(nil)

// Factory constructs the POP3 inbound transport.
func Factory(_ context.Context, _ string) (spi.InboundTransport, error) {
	return &Transport{}, nil
}

// Fetch retrieves messages. Full POP3 client is staged after probe path.
func (t *Transport) Fetch(_ context.Context, endpoint mailcap.ConnectionEndpoint, _ int) (*mailcap.FetchResult, error) {
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	return nil, gerror.New("pop3: message fetch client is not implemented yet; connection probe is available")
}

// Probe validates POP3 reachability and, when credentials are present, USER/PASS login.
func (t *Transport) Probe(ctx context.Context, endpoint mailcap.ConnectionEndpoint) error {
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	conn, err := dialPOP3(ctx, endpoint)
	if err != nil {
		return err
	}
	// Close the latest connection (plain or upgraded TLS) on return.
	defer func() { _ = conn.Close() }()

	reader := bufio.NewReader(conn)
	if err = setConnDeadline(conn, 15*time.Second); err != nil {
		return err
	}
	greeting, err := readPOP3Line(reader)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToUpper(greeting), "+OK") {
		return gerror.Newf("pop3: unexpected greeting: %s", truncate(greeting, 200))
	}

	tlsMode := strings.TrimSpace(strings.ToLower(endpoint.TLSMode))
	if tlsMode == "starttls" {
		if err = writePOP3(conn, "STLS"); err != nil {
			return err
		}
		resp, readErr := readPOP3Line(reader)
		if readErr != nil {
			return readErr
		}
		if !strings.HasPrefix(strings.ToUpper(resp), "+OK") {
			return gerror.Newf("pop3: STLS failed: %s", truncate(resp, 200))
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
	if err = writePOP3(conn, "USER "+username); err != nil {
		return err
	}
	userResp, err := readPOP3Line(reader)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToUpper(userResp), "+OK") {
		return gerror.Newf("pop3: USER failed: %s", truncate(userResp, 200))
	}
	if err = writePOP3(conn, "PASS "+secret); err != nil {
		return err
	}
	passResp, err := readPOP3Line(reader)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToUpper(passResp), "+OK") {
		return gerror.Newf("pop3: PASS failed: %s", truncate(passResp, 200))
	}
	_ = writePOP3(conn, "QUIT")
	return nil
}

func dialPOP3(ctx context.Context, endpoint mailcap.ConnectionEndpoint) (net.Conn, error) {
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

func writePOP3(conn net.Conn, command string) error {
	_, err := conn.Write([]byte(command + "\r\n"))
	return err
}

func readPOP3Line(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
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
		return gerror.New("pop3: host is required")
	}
	if endpoint.Port <= 0 {
		return gerror.New("pop3: port is required")
	}
	return nil
}
