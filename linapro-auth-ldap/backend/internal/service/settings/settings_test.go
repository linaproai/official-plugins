package settings

import "testing"

func TestValidateTLSMode(t *testing.T) {
	t.Parallel()
	if err := ValidateTLSMode("ldap.example.com", TLSModeLDAPS); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTLSMode("ldap.example.com", TLSModePlain); err == nil {
		t.Fatal("expected plain remote reject")
	}
	if err := ValidateTLSMode("localhost", TLSModePlain); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTLSMode("127.0.0.1", TLSModeStartTLS); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeTLSMode(t *testing.T) {
	t.Parallel()
	if NormalizeTLSMode("LDAPS") != TLSModeLDAPS {
		t.Fatal()
	}
	if NormalizeTLSMode("starttls") != TLSModeStartTLS {
		t.Fatal()
	}
}
