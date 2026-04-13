package auth

import "github.com/golang-jwt/jwt/v5"

// Claims represents the JWT claims issued by the Zentrale's OIDC issuer.
// Shared between zentrale (mints tokens) and workspace (validates tokens).
type Claims struct {
	jwt.RegisteredClaims
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	IsAdmin     bool   `json:"is_admin,omitempty"`
}
