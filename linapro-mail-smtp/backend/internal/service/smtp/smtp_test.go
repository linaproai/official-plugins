// This file tests SMTP transport message assembly and endpoint validation.

package smtp

import (
	"strings"
	"testing"

	"lina-plugin-linapro-mail-core/backend/cap/mailcap"
)

func TestValidateEndpoint(t *testing.T) {
	if err := validateEndpoint(mailcap.ConnectionEndpoint{}); err == nil {
		t.Fatal("expected host required")
	}
	if err := validateEndpoint(mailcap.ConnectionEndpoint{Host: "smtp.example.com", Port: 587}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildMessage(t *testing.T) {
	body := string(buildMessage("a@example.com", mailcap.MailMessage{
		To:       []string{"b@example.com"},
		Subject:  "hello",
		TextBody: "world",
	}))
	if !strings.Contains(body, "Subject: hello") || !strings.Contains(body, "world") {
		t.Fatalf("unexpected body: %s", body)
	}
}
