package models

import "github.com/golang-jwt/jwt/v5"

// Claims represents the JWT claims structure
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginRequest represents the incoming login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the response containing the JWT token
type LoginResponse struct {
	Token string `json:"token"`
}
