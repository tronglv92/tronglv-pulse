package security

import (
	"context"
	"crypto/rsa"
	"pulse/helper/utils/httpc"
	"pulse/helper/utils/toolkit/cryptox"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/collection"
)

const JWKSKey = "security:jwks"

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWKSProvider interface {
	GetSecret() []byte
	GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error)
	Refresh(ctx context.Context) error
}

type jwksProvider struct {
	url         string
	secret      []byte
	httpClient  httpc.Service
	cacheClient *collection.Cache
	keys        map[string]*rsa.PublicKey
}

func NewJWKSProvider(url string, secret []byte, cacheClient *collection.Cache, httpClient httpc.Service) JWKSProvider {
	return &jwksProvider{
		url:         url,
		secret:      secret,
		httpClient:  httpClient,
		cacheClient: cacheClient,
		keys:        make(map[string]*rsa.PublicKey),
	}
}

func (p *jwksProvider) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	resp, err := p.httpClient.Get(ctx, p.url)
	if err != nil {
		return nil, err
	}

	var jwks JWKS
	if err = json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := cryptox.ParseRSAKey(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	return keys, nil
}

func (p *jwksProvider) GetSecret() []byte {
	return p.secret
}

func (p *jwksProvider) GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	jwks, err := p.cacheClient.Take(JWKSKey, func() (any, error) {
		return p.fetch(ctx)
	})
	if err != nil {
		return nil, err
	}

	if items, ok := jwks.(map[string]*rsa.PublicKey); ok {
		key, ok := items[kid]
		if !ok {
			return nil, fmt.Errorf("kid not found after refresh")
		}
		return key, nil
	}
	return nil, fmt.Errorf("jwks key not found after refresh")
}

func (p *jwksProvider) Refresh(ctx context.Context) error {
	_, err := p.cacheClient.Take(JWKSKey, func() (any, error) {
		return p.fetch(ctx)
	})
	return err
}
