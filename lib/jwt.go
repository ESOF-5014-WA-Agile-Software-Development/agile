package lib

import "github.com/dgrijalva/jwt-go"

type JWTClaims struct {
	UserId    uint     `json:"id,omitempty"`
	UserName  string   `json:"name,omitempty"`
	UserEmail string   `json:"email,omitempty"`
	UserRole  string   `json:"user_role,omitempty"`
	MetaMasks []string `json:"meta_masks,omitempty"`

	jwt.StandardClaims
}
