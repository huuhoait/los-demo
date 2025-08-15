package infrastructure

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/auth/domain"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// JWTTokenManager implements token management using JWT
type JWTTokenManager struct {
	signingKey     []byte
	issuer         string
	accessTokenTTL time.Duration
	cache          domain.CacheService
	logger         *zap.Logger
	localizer      *i18n.Localizer
}

// NewJWTTokenManager creates a new JWT token manager
func NewJWTTokenManager(
	signingKey []byte,
	issuer string,
	accessTokenTTL time.Duration,
	cache domain.CacheService,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *JWTTokenManager {
	return &JWTTokenManager{
		signingKey:     signingKey,
		issuer:         issuer,
		accessTokenTTL: accessTokenTTL,
		cache:          cache,
		logger:         logger,
		localizer:      localizer,
	}
}

// GenerateAccessToken creates a new JWT access token
func (j *JWTTokenManager) GenerateAccessToken(ctx context.Context, user *domain.User, sessionID string) (string, time.Time, error) {
	logger := j.logger.With(
		zap.String("operation", "generate_access_token"),
		zap.String("user_id", user.ID),
		zap.String("session_id", sessionID),
	)

	now := time.Now()
	expiresAt := now.Add(j.accessTokenTTL)

	claims := &domain.JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   user.ID,
			Audience:  []string{"los-api"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.signingKey)
	if err != nil {
		logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", time.Time{}, domain.NewAuthError(domain.AUTH_019,
			j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.token_generation_failed"}),
			"Failed to generate access token")
	}

	logger.Debug("Access token generated successfully",
		zap.String("token_id", claims.ID),
		zap.Time("expires_at", expiresAt))

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a new refresh token
func (j *JWTTokenManager) GenerateRefreshToken(ctx context.Context) (string, error) {
	// Generate cryptographically secure random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		j.logger.Error("Failed to generate random bytes for refresh token", zap.Error(err))
		return "", domain.NewAuthError(domain.AUTH_019,
			j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.token_generation_failed"}),
			"Failed to generate refresh token")
	}

	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

// ValidateAccessToken validates and parses a JWT access token
func (j *JWTTokenManager) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.JWTClaims, error) {
	logger := j.logger.With(
		zap.String("operation", "validate_access_token"),
	)

	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.signingKey, nil
	})

	if err != nil {
		logger.Warn("Failed to parse JWT token", zap.Error(err))

		// Check specific error types - simplified for JWT v5
		if err.Error() == "token is expired" {
			return nil, domain.NewAuthError(domain.AUTH_005,
				j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.token_expired"}),
				"Access token has expired")
		}
		if err.Error() == "token is not valid yet" {
			return nil, domain.NewAuthError(domain.AUTH_004,
				j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.token_not_valid_yet"}),
				"Access token is not valid yet")
		}

		return nil, domain.NewAuthError(domain.AUTH_004,
			j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_token"}),
			"Invalid access token")
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		logger.Warn("Invalid JWT claims")
		return nil, domain.NewAuthError(domain.AUTH_004,
			j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_token"}),
			"Invalid access token claims")
	}

	// Additional validation
	if claims.UserID == "" {
		logger.Warn("Missing user ID in token claims")
		return nil, domain.NewAuthError(domain.AUTH_004,
			j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.invalid_token"}),
			"Invalid token claims")
	}

	logger.Debug("Access token validated successfully",
		zap.String("user_id", claims.UserID),
		zap.String("token_id", claims.ID))

	return claims, nil
}

// RevokeToken adds a token to the revocation list
func (j *JWTTokenManager) RevokeToken(ctx context.Context, token string) error {
	logger := j.logger.With(
		zap.String("operation", "revoke_token"),
	)

	// Parse token to get expiration time
	claims, err := j.ValidateAccessToken(ctx, token)
	if err != nil {
		// Token is already invalid, consider it revoked
		return nil
	}

	// Store token ID in revocation list until expiration
	revokeKey := "revoked_token:" + claims.ID
	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)

	if ttl > 0 {
		if err := j.cache.Set(ctx, revokeKey, true, ttl); err != nil {
			logger.Error("Failed to store revoked token", zap.Error(err))
			return domain.NewAuthError(domain.AUTH_018,
				j.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "auth.revocation_failed"}),
				"Failed to revoke token")
		}
	}

	logger.Info("Token revoked successfully", zap.String("token_id", claims.ID))
	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (j *JWTTokenManager) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	// Parse token to get ID
	claims, err := j.ValidateAccessToken(ctx, token)
	if err != nil {
		// If token is invalid, consider it revoked
		return true, nil
	}

	revokeKey := "revoked_token:" + claims.ID
	exists, err := j.cache.Exists(ctx, revokeKey)
	if err != nil {
		j.logger.Error("Failed to check token revocation status", zap.Error(err))
		// If cache is unavailable, assume token is not revoked to avoid blocking valid requests
		return false, nil
	}

	return exists, nil
}

// Helper function to generate unique token ID
func generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
