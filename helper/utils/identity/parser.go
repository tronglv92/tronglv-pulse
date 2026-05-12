package identity

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

func FromToken(token string, publicKey *rsa.PublicKey, secret []byte) (Claims, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			return publicKey, nil
		default:
			return secret, nil
		}
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	jwtClaims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT token format")
	}
	return NewClaimsFromJWT(jwtClaims), nil
}

func FromString(s string) (Claims, error) {
	var reqData MapClaims
	if err := json.Unmarshal([]byte(s), &reqData); err != nil {
		return nil, err
	}
	return &reqData, nil
}

func NewClaimsFromJWT(claims jwt.MapClaims) Claims {
	return &MapClaims{
		Subject:     GetClaim[string](claims, "eid"),
		Name:        GetClaim[string](claims, "name"),
		Source:      GetClaim[string](claims, "source"),
		Kind:        GetStringWithDefault(claims, "kind", UserKind),
		Roles:       GetStringSlice(claims, "roles"),
		Permissions: GetStringSlice(claims, "permissions"),
		Attributes:  GetClaimMapString(claims, "attributes"),
		SingleMode:  GetClaim[bool](claims, "single_session"),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        GetClaim[string](claims, "jti"),
			Issuer:    GetClaim[string](claims, "iss"),
			Audience:  GetStringSlice(claims, "aud"),
			ExpiresAt: GetTimeFromClaim(claims, "exp"),
		},
	}
}

func GetStringWithDefault(claims jwt.MapClaims, claimName, defaultVal string) string {
	v := GetClaim[string](claims, claimName)
	if v == "" {
		return defaultVal
	}
	return v
}

func GetClaim[T any](claims jwt.MapClaims, claimName string) T {
	if value, ok := claims[claimName]; ok {
		if casted, ok := value.(T); ok {
			return casted
		}
	}
	var zero T
	return zero
}

func GetStringSlice(claims jwt.MapClaims, claimName string) []string {
	v, ok := claims[claimName]
	if !ok || v == nil {
		return nil
	}

	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func GetClaimMapString(claims jwt.MapClaims, claimName string) map[string]string {
	if value, ok := claims[claimName]; ok {
		if m, ok := value.(map[string]string); ok {
			return m
		} else if m, ok := value.(map[string]any); ok {
			res := make(map[string]string, len(m))
			for k, v := range m {
				if str, ok := v.(string); ok {
					res[k] = str
				}
			}
			return res
		}
	}
	return nil
}

func GetTimeFromClaim(claims jwt.MapClaims, name string) *jwt.NumericDate {
	val, ok := claims[name]
	if !ok {
		return nil
	}

	switch v := val.(type) {
	case float64:
		return jwt.NewNumericDate(time.Unix(int64(v), 0))
	case int64:
		return jwt.NewNumericDate(time.Unix(v, 0))
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return jwt.NewNumericDate(time.Unix(i, 0))
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return jwt.NewNumericDate(time.Unix(i, 0))
		}
	}
	return nil
}
