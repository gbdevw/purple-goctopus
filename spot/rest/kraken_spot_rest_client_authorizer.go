package rest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Header used to provide the API key
	managedHeaderAPIKey = "API-Key"
	// Header used to provide the signature
	managedHeaderAPISign = "API-Sign"
)

// An authorizer for the a KrakenSpotRESTClient that signs the outgoing
// request to private Kraken spot REST API endpoints.
type KrakenSpotRESTClientAuthorizer struct {
	// API Key used to sign request.
	key string
	// Secret used to forge signature.
	secret []byte
	// Crypto - SHA
	sha hash.Hash
	// Crypto - HMAC
	mac hash.Hash
}

// # Description
//
// Factory for KrakenSpotRESTClientAuthorizer.
//
// # Inputs
//
//   - key: The API key used to sign requests
//   - secret: The base64 encoded secret used to sign request (use the value displayed when creating the API key).
//
// # Returns
//
//	The factory returns a fuly initialized authorizer or an error if the secret could not be base64 decoded.
func NewKrakenSpotRESTClientAuthorizer(key, secret string) (*KrakenSpotRESTClientAuthorizer, error) {
	// Base64 decode provided secret
	base64DecodedSecret, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		// return error
		return nil, fmt.Errorf("could not base64 decode provided secret for Kraken spot API: %w", err)
	}
	// Build and return the authorizer
	return &KrakenSpotRESTClientAuthorizer{
		key:    key,
		secret: base64DecodedSecret,
		sha:    sha256.New(),
		mac:    hmac.New(sha512.New, base64DecodedSecret),
	}, nil
}

// Authorize the request by using the request form data and the provided credentials.
//
// # WARNING
//
// The method expects request.Form data to be populated in order to extract the nonce and all
// other data required to forge the signature. The method will call req.ParseForm. For this to
// work, the provided request must have a body set, its http.Method equal to POST, PATCH or PUT
// and its content-type header be set to "application/x-www-form-urlencoded".
func (auth *KrakenSpotRESTClientAuthorizer) Authorize(ctx context.Context, req *http.Request) (*http.Request, error) {
	// Ensure request is not nil or panic as it must not be nil.
	if req == nil {
		panic("cannot authorize request: provided request is nil")
	}
	// Sign request
	select {
	case <-ctx.Done():
		// Shortcut if context has expired
		return nil, fmt.Errorf("failed to authorize request: %w", ctx.Err())
	default:
		// Check if signature is required
		if !strings.Contains(req.URL.Path, "/public") {
			// Parse form data
			err := req.ParseForm()
			if err != nil {
				return nil, fmt.Errorf("failed to authorize request: could not parse form data: %w", err)
			}
			// Sign request
			signature, err := auth.getKrakenSignature(req.URL.Path, req.Form)
			if err != nil {
				return nil, fmt.Errorf("failed to authorize request: %w", err)
			}
			// Set/Override Api-Key and API-Sign headers in request
			req.Header[managedHeaderAPIKey] = []string{auth.key}
			req.Header[managedHeaderAPISign] = []string{signature}
		}
		// Return reference to the (signed) request.
		return req, nil
	}
}

// # Description
//
// Forge the signature for a Kraken spot REST API request.
//
// # Inputs
//
//   - path: The URI path of the request.
//   - payload: The form body data which includes a "nonce" and an optional "otp" values.
//
// # Returns
//
// The request signature or an error if any.
func (auth *KrakenSpotRESTClientAuthorizer) getKrakenSignature(path string, payload url.Values) (string, error) {
	// Defer reset
	defer auth.sha.Reset()
	defer auth.mac.Reset()
	// SHA256(nonce + POST data)
	_, err := auth.sha.Write([]byte(payload.Get("nonce") + payload.Encode()))
	if err != nil {
		return "", fmt.Errorf("signature failed: could not produce SHA256(nonce + POST data): %w", err)
	}
	shasum := auth.sha.Sum(nil)
	// HMAC-SHA512 of (URI path + SHA256(nonce + POST data)) and base64 decoded secret API key
	_, err = auth.mac.Write(append([]byte(path), shasum...))
	if err != nil {
		return "", fmt.Errorf("signature failed: could not produce HMAC-SHA512(URI path + SHA256(nonce + POST data)): %w", err)
	}
	macsum := auth.mac.Sum(nil)
	// Base64 encode signature to include in header
	return base64.StdEncoding.EncodeToString(macsum), nil
}
