package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// keySet caches OIDC JSON web keys.
type keySet struct {
	issuer  string
	client  *http.Client
	mu      sync.RWMutex
	jwksURI string
	keys    map[string]any
	expires time.Time
}

// discoveryDocument contains OIDC discovery metadata.
type discoveryDocument struct {
	JWKSURI string `json:"jwks_uri"`
}

// jwkSet contains JSON web keys.
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// jwk contains one JSON web key.
type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
	Curve     string `json:"crv"`
	X         string `json:"x"`
	Y         string `json:"y"`
}

// newKeySet creates a key cache.
func newKeySet(issuer string) *keySet {
	return &keySet{issuer: strings.TrimRight(issuer, "/"), client: http.DefaultClient, keys: map[string]any{}}
}

// Key returns the verification key for token.
func (set *keySet) Key(ctx context.Context, token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, fmt.Errorf("token kid is required")
	}
	if key, ok := set.cached(kid); ok {
		return key, nil
	}
	if err := set.refresh(ctx); err != nil {
		return nil, err
	}
	if key, ok := set.cached(kid); ok {
		return key, nil
	}
	return nil, fmt.Errorf("jwks key %s not found", kid)
}

// cached returns a cached key when fresh.
func (set *keySet) cached(kid string) (any, bool) {
	set.mu.RLock()
	defer set.mu.RUnlock()
	if time.Now().UTC().After(set.expires) {
		return nil, false
	}
	key, ok := set.keys[kid]
	return key, ok
}

// refresh refreshes OIDC discovery and JWKS.
func (set *keySet) refresh(ctx context.Context) error {
	set.mu.Lock()
	defer set.mu.Unlock()
	if time.Now().UTC().Before(set.expires) {
		return nil
	}
	jwksURI, err := set.discovery(ctx)
	if err != nil {
		return err
	}
	keys, err := set.fetchKeys(ctx, jwksURI)
	if err != nil {
		return err
	}
	set.jwksURI = jwksURI
	set.keys = keys
	set.expires = time.Now().UTC().Add(10 * time.Minute)
	return nil
}

// discovery returns the provider JWKS URI.
func (set *keySet) discovery(ctx context.Context) (string, error) {
	if set.jwksURI != "" {
		return set.jwksURI, nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, set.issuer+"/.well-known/openid-configuration", nil)
	if err != nil {
		return "", err
	}
	response, err := set.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("oidc discovery failed with status %d", response.StatusCode)
	}
	var discovery discoveryDocument
	if err := json.NewDecoder(response.Body).Decode(&discovery); err != nil {
		return "", err
	}
	if strings.TrimSpace(discovery.JWKSURI) == "" {
		return "", fmt.Errorf("oidc discovery jwks_uri is required")
	}
	return discovery.JWKSURI, nil
}

// fetchKeys returns public verification keys by kid.
func (set *keySet) fetchKeys(ctx context.Context, uri string) (map[string]any, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	response, err := set.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("jwks fetch failed with status %d", response.StatusCode)
	}
	var payload jwkSet
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	keys := map[string]any{}
	for _, key := range payload.Keys {
		if parsed, ok := key.rsaPublicKey(); ok {
			keys[key.KeyID] = parsed
		}
		if parsed, ok := key.ecdsaPublicKey(); ok {
			keys[key.KeyID] = parsed
		}
	}
	return keys, nil
}

// rsaPublicKey returns an RSA public key from a JWK.
func (key jwk) rsaPublicKey() (*rsa.PublicKey, bool) {
	if key.KeyID == "" || key.KeyType != "RSA" || key.Modulus == "" || key.Exponent == "" {
		return nil, false
	}
	modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil {
		return nil, false
	}
	exponent, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil {
		return nil, false
	}
	value := 0
	for _, b := range exponent {
		value = value<<8 + int(b)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(modulus), E: value}, true
}

// ecdsaPublicKey returns an ECDSA public key from a JWK.
func (key jwk) ecdsaPublicKey() (*ecdsa.PublicKey, bool) {
	curve, ok := namedCurve(key.Curve)
	if key.KeyID == "" || key.KeyType != "EC" || key.X == "" || key.Y == "" || !ok {
		return nil, false
	}
	x, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, false
	}
	y, err := base64.RawURLEncoding.DecodeString(key.Y)
	if err != nil {
		return nil, false
	}
	xValue := new(big.Int).SetBytes(x)
	yValue := new(big.Int).SetBytes(y)
	if !curve.IsOnCurve(xValue, yValue) {
		return nil, false
	}
	return &ecdsa.PublicKey{Curve: curve, X: xValue, Y: yValue}, true
}

// namedCurve returns the elliptic curve for a JWK curve name.
func namedCurve(name string) (elliptic.Curve, bool) {
	switch name {
	case "P-256":
		return elliptic.P256(), true
	case "P-384":
		return elliptic.P384(), true
	case "P-521":
		return elliptic.P521(), true
	default:
		return nil, false
	}
}
