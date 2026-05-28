package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents JWT token claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Email            string   `json:"email,omitempty"`
	PreferredUsername string  `json:"preferred_username,omitempty"`
	Roles            []string `json:"roles,omitempty"`
}

// Validator handles JWT token validation
type Validator struct {
	keycloakURL   string
	realm         string
	jwksURL       string
	verifySig     bool
	audience      string
	cachedKeys    map[string]interface{}
	keysUpdatedAt time.Time
}

// NewValidator creates a new JWT validator
func NewValidator(keycloakURL, realm, jwksURL string, verifySig bool, audience string) *Validator {
	return &Validator{
		keycloakURL: keycloakURL,
		realm:       realm,
		jwksURL:     jwksURL,
		verifySig:   verifySig,
		audience:    audience,
		cachedKeys:  make(map[string]interface{}),
	}
}

// DecodeUnverified decodes JWT token without signature verification
// This is used to extract claims quickly without calling Keycloak
func (v *Validator) DecodeUnverified(tokenString string) (map[string]interface{}, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}

	return result, nil
}

// ValidateWithSignature validates JWT token with signature verification
// Uses cached JWKS keys to avoid calling Keycloak on every request
func (v *Validator) ValidateWithSignature(tokenString string) (*TokenClaims, error) {
	// Try to validate with cached keys first
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		kid := token.Header["kid"].(string)
		
		if key, ok := v.cachedKeys[kid]; ok {
			return key, nil
		}
		
		// Fetch JWKS if key not found (simplified - in production use proper JWKS fetching)
		return nil, fmt.Errorf("key not found: %s", kid)
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate audience if configured
	if v.audience != "" {
		found := false
		for _, aud := range claims.Audience {
			if aud == v.audience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return claims, nil
}

// ExtractUserInfo extracts user information from JWT claims
func ExtractUserInfo(claims map[string]interface{}) (userID, username, email string, roles []string) {
	userID = getStringClaim(claims, "sub")
	username = getStringClaim(claims, "preferred_username")
	email = getStringClaim(claims, "email")
	roles = getStringSliceClaim(claims, "realm_access", "roles")
	
	if len(roles) == 0 {
		roles = getStringSliceClaim(claims, "roles")
	}

	return
}

func getStringClaim(claims map[string]interface{}, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

func getStringSliceClaim(claims map[string]interface{}, keys ...string) []string {
	var current interface{} = claims
	
	for i, key := range keys {
		if i == len(keys)-1 {
			if v, ok := current.(map[string]interface{})[key].([]interface{}); ok {
				result := make([]string, len(v))
				for j, item := range v {
					if s, ok := item.(string); ok {
						result[j] = s
					}
				}
				return result
			}
			break
		}
		
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			break
		}
	}
	
	return []string{}
}

// ClaimsToJSON converts claims to JSON string for storage
func ClaimsToJSON(claims map[string]interface{}) (string, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
