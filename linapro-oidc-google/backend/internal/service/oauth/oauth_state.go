// oauth_state.go implements the self-contained HMAC-signed OAuth state used
// by the Google OIDC login flow. The state token embeds the optional business
// state key, a random nonce, and an absolute expiry, and is signed with a key
// derived from the client secret. Because the token is self-validating, the
// flow does not depend on browser cookies surviving the cross-site round trip
// through Google, which removes an entire class of SameSite/proxy cookie
// failures.

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

// stateTTL bounds the OAuth state lifetime to mitigate replay attacks.
const stateTTL = 10 * time.Minute

// stateNonceByteSize is the number of random bytes bound to one authorize
// round trip.
const stateNonceByteSize = 16

// stateMACKeyPrefix scopes the HMAC key to this plugin so a signature can
// never be replayed against another plugin's flow.
const stateMACKeyPrefix = "linapro-oidc-google::"

// StatePayload holds the decoded contents of one OAuth state token.
type StatePayload struct {
	// StateKey carries the optional business state key used by the SSO
	// delivery rules.
	StateKey string `json:"stateKey"`
	// Nonce is a random value bound to one authorization round trip.
	Nonce string `json:"nonce"`
	// ExpiresAt is the absolute unix-second deadline beyond which the state
	// is rejected.
	ExpiresAt int64 `json:"expiresAt"`
}

// StateCodec encodes and validates self-contained OAuth state tokens. The
// interface keeps the orchestration testable with a deterministic codec.
type StateCodec interface {
	// Encode produces one signed state token embedding the business state key.
	Encode(ctx context.Context, stateKey string, clientSecret string) (string, error)
	// Decode verifies the signature and expiry and returns the payload.
	Decode(ctx context.Context, state string, clientSecret string) (StatePayload, error)
}

// hmacStateCodec is the production HMAC-SHA256 state codec.
type hmacStateCodec struct{}

// NewHMACStateCodec returns the production self-contained state codec.
func NewHMACStateCodec() StateCodec {
	return &hmacStateCodec{}
}

// Encode serializes and signs one state payload.
func (c *hmacStateCodec) Encode(_ context.Context, stateKey string, clientSecret string) (string, error) {
	nonceBuffer := make([]byte, stateNonceByteSize)
	if _, err := rand.Read(nonceBuffer); err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: state entropy read failed"), CodeStateGenerateFailed)
	}
	payload := StatePayload{
		StateKey:  strings.TrimSpace(stateKey),
		Nonce:     base64.RawURLEncoding.EncodeToString(nonceBuffer),
		ExpiresAt: time.Now().Add(stateTTL).Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", bizerr.WrapCode(gerror.Wrap(err, "oidc google: encode state payload failed"), CodeStateGenerateFailed)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	return encoded + "." + computeStateMAC(encoded, clientSecret), nil
}

// Decode verifies the HMAC signature and expiry of one state token.
func (c *hmacStateCodec) Decode(_ context.Context, state string, clientSecret string) (StatePayload, error) {
	parts := strings.SplitN(strings.TrimSpace(state), ".", 2)
	if len(parts) != 2 {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc google: state token is malformed"), CodeCallbackStateMismatch)
	}
	expected := computeStateMAC(parts[0], clientSecret)
	if !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc google: state signature mismatch"), CodeCallbackStateMismatch)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return StatePayload{}, bizerr.WrapCode(gerror.Wrap(err, "oidc google: decode state payload failed"), CodeCallbackStateMismatch)
	}
	var payload StatePayload
	if err = json.Unmarshal(raw, &payload); err != nil {
		return StatePayload{}, bizerr.WrapCode(gerror.Wrap(err, "oidc google: parse state payload failed"), CodeCallbackStateMismatch)
	}
	if payload.ExpiresAt > 0 && time.Now().Unix() > payload.ExpiresAt {
		return StatePayload{}, bizerr.WrapCode(gerror.New("oidc google: state has expired"), CodeCallbackStateMismatch)
	}
	return payload, nil
}

// computeStateMAC derives one HMAC-SHA256 signature scoped to this plugin.
func computeStateMAC(encodedPayload string, clientSecret string) string {
	mac := hmac.New(sha256.New, []byte(stateMACKeyPrefix+clientSecret))
	mac.Write([]byte(encodedPayload))
	return hex.EncodeToString(mac.Sum(nil))
}
