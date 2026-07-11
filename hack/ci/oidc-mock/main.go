// Command oidc-mock is a minimal OIDC provider used by official-plugins CI.
// It supports discovery, authorize (auto-login), token (auth code + PKCE S256),
// and JWKS so linapro-oidc-generic can run a real code+PKCE+id_token login path.
//
// oidc-mock 是官方插件 CI 使用的最小 OIDC 提供方。
// 支持 discovery、authorize（自动登录）、token（授权码 + PKCE S256）和 JWKS，
// 以便 linapro-oidc-generic 走真实的 code+PKCE+id_token 登录路径。
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const kid = "ci-oidc-mock-1"

type authCodeRecord struct {
	ClientID            string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
	Subject             string
	Email               string
	Name                string
	ExpiresAt           time.Time
}

type server struct {
	issuer       string
	clientID     string
	clientSecret string
	privateKey   *rsa.PrivateKey
	mu           sync.Mutex
	codes        map[string]authCodeRecord
}

func main() {
	listen := flag.String("listen", "127.0.0.1:18080", "listen address")
	issuer := flag.String("issuer", "", "issuer URL (default http://<listen>)")
	clientID := flag.String("client-id", "linapro-ci", "allowed OAuth client_id")
	clientSecret := flag.String("client-secret", "linapro-ci-secret", "allowed OAuth client_secret")
	flag.Parse()

	if strings.TrimSpace(*issuer) == "" {
		*issuer = "http://" + *listen
	}
	*issuer = strings.TrimRight(strings.TrimSpace(*issuer), "/")

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate RSA key: %v", err)
	}

	s := &server{
		issuer:       *issuer,
		clientID:     *clientID,
		clientSecret: *clientSecret,
		privateKey:   key,
		codes:        map[string]authCodeRecord{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleDiscovery)
	mux.HandleFunc("/jwks.json", s.handleJWKS)
	mux.HandleFunc("/authorize", s.handleAuthorize)
	mux.HandleFunc("/token", s.handleToken)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen %s: %v", *listen, err)
	}
	log.Printf("oidc-mock listening on %s issuer=%s client_id=%s", *listen, *issuer, *clientID)
	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func (s *server) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":                                s.issuer,
		"authorization_endpoint":                s.issuer + "/authorize",
		"token_endpoint":                        s.issuer + "/token",
		"jwks_uri":                              s.issuer + "/jwks.json",
		"userinfo_endpoint":                     s.issuer + "/userinfo",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "email", "profile"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"code_challenge_methods_supported":      []string{"S256"},
	})
}

func (s *server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	n := base64.RawURLEncoding.EncodeToString(s.privateKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.privateKey.E)).Bytes())
	writeJSON(w, http.StatusOK, map[string]any{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": kid,
				"n":   n,
				"e":   e,
			},
		},
	})
}

func (s *server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := strings.TrimSpace(q.Get("client_id"))
	redirectURI := strings.TrimSpace(q.Get("redirect_uri"))
	responseType := strings.TrimSpace(q.Get("response_type"))
	state := strings.TrimSpace(q.Get("state"))
	nonce := strings.TrimSpace(q.Get("nonce"))
	challenge := strings.TrimSpace(q.Get("code_challenge"))
	method := strings.TrimSpace(q.Get("code_challenge_method"))
	scope := strings.TrimSpace(q.Get("scope"))

	if clientID != s.clientID {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}
	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if redirectURI == "" || state == "" {
		http.Error(w, "redirect_uri and state are required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(scope, "openid") {
		http.Error(w, "scope must include openid", http.StatusBadRequest)
		return
	}
	if challenge == "" || !strings.EqualFold(method, "S256") {
		http.Error(w, "PKCE S256 is required", http.StatusBadRequest)
		return
	}

	// Auto-login test subject. Optional overrides via query for future cases.
	// 自动登录测试主体。可通过 query 覆盖以便后续扩展用例。
	subject := firstNonEmpty(q.Get("login_hint"), "ci-oidc-user")
	email := firstNonEmpty(q.Get("email"), "ci-oidc-user@example.com")
	name := firstNonEmpty(q.Get("name"), "CI OIDC User")

	code, err := randomToken(24)
	if err != nil {
		http.Error(w, "entropy failure", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.codes[code] = authCodeRecord{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       challenge,
		CodeChallengeMethod: method,
		Nonce:               nonce,
		Subject:             subject,
		Email:               email,
		Name:                name,
		ExpiresAt:           time.Now().Add(5 * time.Minute),
	}
	s.mu.Unlock()

	target, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	query := target.Query()
	query.Set("code", code)
	query.Set("state", state)
	target.RawQuery = query.Encode()
	http.Redirect(w, r, target.String(), http.StatusFound)
}

func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	clientID, clientSecret := s.clientCredentials(r)
	if clientID != s.clientID || clientSecret != s.clientSecret {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":             "invalid_client",
			"error_description": "client authentication failed",
		})
		return
	}
	if r.Form.Get("grant_type") != "authorization_code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}

	code := strings.TrimSpace(r.Form.Get("code"))
	redirectURI := strings.TrimSpace(r.Form.Get("redirect_uri"))
	verifier := strings.TrimSpace(r.Form.Get("code_verifier"))

	s.mu.Lock()
	rec, ok := s.codes[code]
	if ok {
		delete(s.codes, code)
	}
	s.mu.Unlock()
	if !ok || time.Now().After(rec.ExpiresAt) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if rec.ClientID != clientID || rec.RedirectURI != redirectURI {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if !pkceS256Valid(verifier, rec.CodeChallenge) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "pkce verification failed",
		})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":            s.issuer,
		"sub":            rec.Subject,
		"aud":            clientID,
		"exp":            now.Add(10 * time.Minute).Unix(),
		"iat":            now.Unix(),
		"nonce":          rec.Nonce,
		"email":          rec.Email,
		"email_verified": true,
		"name":           rec.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	idToken, err := token.SignedString(s.privateKey)
	if err != nil {
		http.Error(w, "sign id_token failed", http.StatusInternalServerError)
		return
	}
	accessToken, err := randomToken(16)
	if err != nil {
		http.Error(w, "entropy failure", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   600,
		"id_token":     idToken,
		"scope":        "openid email profile",
	})
}

func (s *server) clientCredentials(r *http.Request) (string, string) {
	clientID := strings.TrimSpace(r.Form.Get("client_id"))
	clientSecret := strings.TrimSpace(r.Form.Get("client_secret"))
	if clientID != "" {
		return clientID, clientSecret
	}
	user, pass, ok := r.BasicAuth()
	if ok {
		return user, pass
	}
	return "", ""
}

func pkceS256Valid(verifier, challenge string) bool {
	if verifier == "" || challenge == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(sum[:])
	return expected == challenge
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Fprintf(os.Stderr, "write json: %v\n", err)
	}
}
