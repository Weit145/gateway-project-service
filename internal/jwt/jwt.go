package jwt

import (
	"crypto/rsa"
	"fmt"
	"strings"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type accessClaims struct {
	Type string `json:"typ"`
	gojwt.RegisteredClaims
}

type JWTService struct {
	publicKey *rsa.PublicKey
}

func NewJWTService(publicKey *rsa.PublicKey) *JWTService {
	return &JWTService{
		publicKey: publicKey,
	}
}

func (j *JWTService) CheckToken(authHeader string) (uuid.UUID, uuid.UUID, error) {
	if j.publicKey == nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("JWT public key is not configured: %w", errs.ErrInternal)
	}

	authHeader = strings.TrimSpace(authHeader)
	tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || strings.TrimSpace(tokenString) == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("missing bearer token: %w", errs.ErrUnauthorized)
	}

	claims := &accessClaims{}
	token, err := gojwt.ParseWithClaims(
		strings.TrimSpace(tokenString),
		claims,
		func(token *gojwt.Token) (interface{}, error) {
			return j.publicKey, nil
		},
		gojwt.WithValidMethods([]string{gojwt.SigningMethodRS256.Alg()}),
		gojwt.WithIssuer("auth-service"),
		gojwt.WithAudience("gateway"),
		gojwt.WithExpirationRequired(),
		gojwt.WithIssuedAt(),
	)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid bearer token: %v: %w", err, errs.ErrUnauthorized)
	}
	if !token.Valid {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid bearer token: %w", errs.ErrUnauthorized)
	}

	if claims.Type != "access" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token type %q: %w", claims.Type, errs.ErrUnauthorized)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token subject: %v: %w", err, errs.ErrUnauthorized)
	}

	jwtId, err := uuid.Parse(claims.ID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token ID: %v: %w", err, errs.ErrUnauthorized)
	}

	return userID, jwtId, nil
}
