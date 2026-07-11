// oauth_state.go implements HMAC-signed OAuth state with PKCE and nonce fields.

package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

const (
	stateTTL             = 10 * time.Minute
	stateNonceByteSize   = 16
	stateMACKeyPrefix    = "linapro-oidc-generic::"
	pkceVerifierByteSize = 32
)

// StatePayload holds decoded OAuth state contents including PKCE verifier.
type StatePayload struct {
	StateKey     string `json:"stateKey"`
	ReturnTo     string `json:"returnTo,omitempty"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"codeVerifier"`
	OIDCNonce    string `json:"oidcNonce"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// StateCodec encodes and validates self-contained OAuth state tokens.
type StateCodec interface {
	Encode(ctx context.Context, stateKey string, returnTo string, clientSecret string, codeVerifier string, oidcNonce string) (string, error)
	Decode(ctx context.Context, state string, clientSecret string) (StatePayload, error)
}

type hmacStateCodec struct{}

// NewHMACStateCodec returns the production self-contained state codec.
func NewHMACStateCodec() StateCodec {
	return &hmacStateCodec{}
}

func (c *hmacStateCodec) Encode(_ context.Context, stateKey string, returnTo string, clientSecret string, codeVerifier string, oidcNonce string) (string, error) {
	nonceBuffer := make([]byte, stateNonceByteSize)
	if _, err := rand.Read(nonceBuffer); err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc generic: state entropy read failed"), CodeStateGenerateFailed)
	}
	payload := StatePayload{
		StateKey:     strings.TrimSpace(stateKey),
		ReturnTo:     sanitizeReturnTo(returnTo),
		Nonce:        base64.RawURLEncoding.EncodeToString(nonceBuffer),
		CodeVerifier: strings.TrimSpace(codeVerifier),
		OIDCNonce:    strings.TrimSpace(oidcNonce),
		ExpiresAt:    time.Now().Add(stateTTL).Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc generic: encode state payload failed"), CodeStateGenerateFailed)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	return encoded + "." + computeStateMAC(encoded, clientSecret), nil
}

func (c *hmacStateCodec) Decode(_ context.Context, state string, clientSecret string) (StatePayload, error) {
	parts := strings.SplitN(strings.TrimSpace(state), ".", 2)
	if len(parts) != 2 {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc generic: state token is malformed"), CodeCallbackStateMismatch)
	}
	expected := computeStateMAC(parts[0], clientSecret)
	if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc generic: state signature mismatch"), CodeCallbackStateMismatch)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return StatePayload{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: decode state payload failed"), CodeCallbackStateMismatch)
	}
	var payload StatePayload
	if err = json.Unmarshal(raw, &payload); err != nil {
		return StatePayload{}, bizerr.WrapCode(gerror.Wrap(err, "oidc generic: parse state payload failed"), CodeCallbackStateMismatch)
	}
	if payload.ExpiresAt > 0 && time.Now().Unix() > payload.ExpiresAt {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc generic: state has expired"), CodeCallbackStateMismatch)
	}
	return payload, nil
}

func computeStateMAC(encodedPayload string, clientSecret string) string {
	mac := hmac.New(sha256.New, []byte(stateMACKeyPrefix+clientSecret))
	mac.Write([]byte(encodedPayload))
	return hex.EncodeToString(mac.Sum(nil))
}

func newPKCEVerifier() (string, error) {
	buf := make([]byte, pkceVerifierByteSize)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func pkceChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newOIDCNonce() (string, error) {
	buf := make([]byte, stateNonceByteSize)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
