package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/huuhoait/los-demo/services/auth/domain"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

// AuthService implements the authentication use cases
type AuthService struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenManager domain.TokenManager
	cache        domain.CacheService
	auditLogger  domain.AuditLogger
	logger       *zap.Logger
	localizer    *i18n.Localizer // Use shared i18n Localizer

	// Configuration
	maxLoginAttempts int
	lockoutDuration  time.Duration
	sessionDuration  time.Duration
	cleanupInterval  time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenManager domain.TokenManager,
	cache domain.CacheService,
	auditLogger domain.AuditLogger,
	logger *zap.Logger,
	localizer *customI18n.Localizer, // Use custom i18n Localizer
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		tokenManager:     tokenManager,
		cache:            cache,
		auditLogger:      auditLogger,
		logger:           logger,
		localizer:        localizer,
		maxLoginAttempts: 5,
		lockoutDuration:  time.Minute * 15,
		sessionDuration:  time.Hour * 24 * 30, // 30 days
		cleanupInterval:  time.Hour * 24,      // Daily cleanup
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, email, password string, ipAddress, userAgent string) (*domain.TokenResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "login"),
		zap.String("email", email),
		zap.String("ip_address", ipAddress),
	)

	// Check rate limiting
	if err := s.CheckRateLimit(ctx, ipAddress); err != nil {
		logger.Warn("Rate limit exceeded for login attempt")
		return nil, err
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		s.logFailedLogin(ctx, "", email, ipAddress, userAgent, domain.AUTH_001)
		return nil, domain.NewAuthError(domain.AUTH_001,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.invalid_credentials", nil),
			"Invalid email or password provided")
	}

	// Check if account is locked
	if err := s.checkAccountLockout(ctx, user.ID); err != nil {
		logger.Warn("Account locked", zap.String("user_id", user.ID))
		s.logFailedLogin(ctx, user.ID, email, ipAddress, userAgent, domain.AUTH_002)
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.Warn("Invalid password", zap.String("user_id", user.ID))
		s.incrementFailedAttempts(ctx, user.ID)
		s.logFailedLogin(ctx, user.ID, email, ipAddress, userAgent, domain.AUTH_001)
		return nil, domain.NewAuthError(domain.AUTH_001,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.invalid_credentials", nil),
			"Invalid email or password provided")
	}

	// Check account status
	if user.Status != "active" {
		logger.Warn("Account not active", zap.String("user_id", user.ID), zap.String("status", user.Status))
		s.logFailedLogin(ctx, user.ID, email, ipAddress, userAgent, domain.AUTH_003)
		return nil, domain.NewAuthError(domain.AUTH_003,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.account_disabled", nil),
			"User account is disabled")
	}

	// Clear failed attempts on successful authentication
	s.clearFailedAttempts(ctx, user.ID)

	// Create session
	session, err := s.CreateSession(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.session_creation_failed", nil),
			"Failed to create user session")
	}

	// Generate tokens
	accessToken, expiresAt, err := s.tokenManager.GenerateAccessToken(ctx, user, session.ID)
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_019,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.token_generation_failed", nil),
			"Failed to generate access token")
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		logger.Warn("Failed to update last login", zap.Error(err))
	}

	// Log successful login
	s.logSuccessfulLogin(ctx, user.ID, session.ID, ipAddress, userAgent)

	logger.Info("User logged in successfully", zap.String("user_id", user.ID))

	return &domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, ipAddress, userAgent string) (*domain.TokenResponse, error) {
	logger := s.logger.With(
		zap.String("operation", "refresh_token"),
		zap.String("ip_address", ipAddress),
	)

	// Validate refresh token
	session, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Warn("Invalid refresh token")
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", session.UserID))
		return nil, domain.NewAuthError(domain.AUTH_016,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.user_not_found", nil),
			"User not found")
	}

	// Check account status
	if user.Status != "active" {
		logger.Warn("Account not active", zap.String("user_id", user.ID))
		return nil, domain.NewAuthError(domain.AUTH_003,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.account_disabled", nil),
			"User account is disabled")
	}

	// Generate new refresh token
	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(ctx)
	if err != nil {
		logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_019,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.token_generation_failed", nil),
			"Failed to generate refresh token")
	}

	// Update session
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(s.sessionDuration)
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Error("Failed to update session", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.session_update_failed", nil),
			"Failed to update session")
	}

	// Generate new access token
	accessToken, expiresAt, err := s.tokenManager.GenerateAccessToken(ctx, user, session.ID)
	if err != nil {
		logger.Error("Failed to generate access token", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_019,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.token_generation_failed", nil),
			"Failed to generate access token")
	}

	// Log token refresh
	s.auditLogger.LogAuthEvent(ctx, &domain.AuthEvent{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		EventType: "refresh",
		SessionID: session.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Timestamp: time.Now(),
	})

	logger.Info("Token refreshed successfully", zap.String("user_id", user.ID))

	return &domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// Logout invalidates the user session
func (s *AuthService) Logout(ctx context.Context, userID, sessionID string) error {
	logger := s.logger.With(
		zap.String("operation", "logout"),
		zap.String("user_id", userID),
		zap.String("session_id", sessionID),
	)

	// Invalidate session
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		logger.Error("Failed to delete session", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.logout_failed", nil),
			"Failed to logout user")
	}

	// Log logout
	s.auditLogger.LogAuthEvent(ctx, &domain.AuthEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventType: "logout",
		SessionID: sessionID,
		Success:   true,
		Timestamp: time.Now(),
	})

	logger.Info("User logged out successfully")
	return nil
}

// LogoutAll invalidates all user sessions
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	logger := s.logger.With(
		zap.String("operation", "logout_all"),
		zap.String("user_id", userID),
	)

	// Get all user sessions
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user sessions", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.logout_failed", nil),
			"Failed to logout user")
	}

	// Delete all sessions
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		logger.Error("Failed to delete user sessions", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.logout_failed", nil),
			"Failed to logout user")
	}

	// Log logout all
	for _, session := range sessions {
		s.auditLogger.LogAuthEvent(ctx, &domain.AuthEvent{
			ID:        uuid.New().String(),
			UserID:    userID,
			EventType: "logout_all",
			SessionID: session.ID,
			Success:   true,
			Timestamp: time.Now(),
		})
	}

	logger.Info("All user sessions logged out successfully", zap.Int("sessions_count", len(sessions)))
	return nil
}

// ValidateAccessToken validates and parses an access token
func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (*domain.AuthContext, error) {
	claims, err := s.tokenManager.ValidateAccessToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if token is revoked
	revoked, err := s.tokenManager.IsTokenRevoked(ctx, token)
	if err != nil {
		s.logger.Error("Failed to check token revocation", zap.Error(err))
	}
	if revoked {
		return nil, domain.NewAuthError(domain.AUTH_006,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.token_revoked", nil),
			"Token has been revoked")
	}

	// Verify session exists
	session, err := s.sessionRepo.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, domain.NewAuthError(domain.AUTH_008,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.session_not_found", nil),
			"Session not found")
	}

	// Check session expiration
	if session.ExpiresAt.Before(time.Now()) {
		return nil, domain.NewAuthError(domain.AUTH_009,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.session_expired", nil),
			"Session has expired")
	}

	return &domain.AuthContext{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		SessionID: claims.SessionID,
	}, nil
}

// ValidateRefreshToken validates a refresh token
func (s *AuthService) ValidateRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByRefreshToken(ctx, token)
	if err != nil {
		return nil, domain.NewAuthError(domain.AUTH_007,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.invalid_refresh_token", nil),
			"Invalid refresh token")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, domain.NewAuthError(domain.AUTH_009,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.session_expired", nil),
			"Session has expired")
	}

	return session, nil
}

// Helper methods

func (s *AuthService) checkAccountLockout(ctx context.Context, userID string) error {
	lockoutKey := "lockout:" + userID
	exists, err := s.cache.Exists(ctx, lockoutKey)
	if err != nil {
		s.logger.Error("Failed to check lockout status", zap.Error(err))
	}
	if exists {
		return domain.NewAuthError(domain.AUTH_002,
			s.localizer.Localize(customI18n.GetLanguageFromContext(ctx), "auth.account_locked", nil),
			"Account is temporarily locked due to too many failed login attempts")
	}
	return nil
}

func (s *AuthService) incrementFailedAttempts(ctx context.Context, userID string) {
	attemptsKey := "failed_attempts:" + userID
	attempts, err := s.cache.Increment(ctx, attemptsKey)
	if err != nil {
		s.logger.Error("Failed to increment failed attempts", zap.Error(err))
		return
	}

	// Set expiration for first attempt
	if attempts == 1 {
		s.cache.SetExpiration(ctx, attemptsKey, s.lockoutDuration)
	}

	// Lock account if max attempts reached
	if attempts >= int64(s.maxLoginAttempts) {
		lockoutKey := "lockout:" + userID
		s.cache.Set(ctx, lockoutKey, true, s.lockoutDuration)
		s.logger.Warn("Account locked due to failed attempts",
			zap.String("user_id", userID),
			zap.Int64("attempts", attempts))
	}
}

func (s *AuthService) clearFailedAttempts(ctx context.Context, userID string) {
	attemptsKey := "failed_attempts:" + userID
	s.cache.Delete(ctx, attemptsKey)
}

func (s *AuthService) logSuccessfulLogin(ctx context.Context, userID, sessionID, ipAddress, userAgent string) {
	s.auditLogger.LogAuthEvent(ctx, &domain.AuthEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventType: "login",
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Timestamp: time.Now(),
	})
}

func (s *AuthService) logFailedLogin(ctx context.Context, userID, email, ipAddress, userAgent, errorCode string) {
	s.auditLogger.LogAuthEvent(ctx, &domain.AuthEvent{
		ID:           uuid.New().String(),
		UserID:       userID,
		EventType:    "failed_login",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      false,
		ErrorCode:    errorCode,
		ErrorMessage: "Login attempt failed",
		Metadata:     map[string]interface{}{"email": email},
		Timestamp:    time.Now(),
	})
}
