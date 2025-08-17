package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/auth/domain"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB, logger *zap.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	logger := r.logger.With(
		zap.String("operation", "get_user_by_id"),
		zap.String("user_id", id),
	)

	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("User not found")
			return nil, domain.NewAuthError(domain.AUTH_016, "User not found", "No user exists with the provided ID")
		}
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to retrieve user")
	}

	logger.Debug("User retrieved successfully")
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	logger := r.logger.With(
		zap.String("operation", "get_user_by_email"),
		zap.String("email", email),
	)

	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at
		FROM users 
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("User not found")
			return nil, domain.NewAuthError(domain.AUTH_016, "User not found", "No user exists with the provided email")
		}
		logger.Error("Failed to get user by email", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to retrieve user")
	}

	logger.Debug("User retrieved successfully")
	return &user, nil
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	logger := r.logger.With(
		zap.String("operation", "create_user"),
		zap.String("email", user.Email),
	)

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Role, user.Status, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to create user")
	}

	logger.Info("User created successfully", zap.String("user_id", user.ID))
	return nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	logger := r.logger.With(
		zap.String("operation", "update_user"),
		zap.String("user_id", user.ID),
	)

	query := `
		UPDATE users 
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, 
		    role = $6, status = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Role, user.Status, user.UpdatedAt)

	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update user")
	}

	if rowsAffected == 0 {
		logger.Debug("No rows affected - user not found")
		return domain.NewAuthError(domain.AUTH_016, "User not found", "No user exists with the provided ID")
	}

	logger.Info("User updated successfully")
	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	logger := r.logger.With(
		zap.String("operation", "update_last_login"),
		zap.String("user_id", userID),
	)

	query := `
		UPDATE users 
		SET last_login_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, userID, now)

	if err != nil {
		logger.Error("Failed to update last login", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update last login")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update last login")
	}

	if rowsAffected == 0 {
		logger.Debug("No rows affected - user not found")
		return domain.NewAuthError(domain.AUTH_016, "User not found", "No user exists with the provided ID")
	}

	logger.Debug("Last login updated successfully")
	return nil
}

// PostgresSessionRepository implements SessionRepository using PostgreSQL
type PostgresSessionRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPostgresSessionRepository creates a new PostgreSQL session repository
func NewPostgresSessionRepository(db *sqlx.DB, logger *zap.Logger) *PostgresSessionRepository {
	return &PostgresSessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new session
func (r *PostgresSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	logger := r.logger.With(
		zap.String("operation", "create_session"),
		zap.String("session_id", session.ID),
		zap.String("user_id", session.UserID),
	)

	query := `
		INSERT INTO user_sessions (id, user_id, refresh_token, expires_at, created_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.RefreshToken, session.ExpiresAt,
		session.CreatedAt, session.IPAddress, session.UserAgent)

	if err != nil {
		logger.Error("Failed to create session", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to create session")
	}

	logger.Debug("Session created successfully")
	return nil
}

// GetByID retrieves a session by ID
func (r *PostgresSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	logger := r.logger.With(
		zap.String("operation", "get_session_by_id"),
		zap.String("session_id", id),
	)

	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at, ip_address, user_agent
		FROM user_sessions 
		WHERE id = $1`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("Session not found")
			return nil, domain.NewAuthError(domain.AUTH_008, "Session not found", "No session exists with the provided ID")
		}
		logger.Error("Failed to get session by ID", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to retrieve session")
	}

	logger.Debug("Session retrieved successfully")
	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *PostgresSessionRepository) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	logger := r.logger.With(
		zap.String("operation", "get_session_by_refresh_token"),
	)

	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at, ip_address, user_agent
		FROM user_sessions 
		WHERE refresh_token = $1`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("Session not found for refresh token")
			return nil, domain.NewAuthError(domain.AUTH_007, "Invalid refresh token", "No session exists with the provided refresh token")
		}
		logger.Error("Failed to get session by refresh token", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to retrieve session")
	}

	logger.Debug("Session retrieved by refresh token successfully")
	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *PostgresSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	logger := r.logger.With(
		zap.String("operation", "get_sessions_by_user_id"),
		zap.String("user_id", userID),
	)

	query := `
		SELECT id, user_id, refresh_token, expires_at, created_at, ip_address, user_agent
		FROM user_sessions 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var sessions []*domain.Session
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		logger.Error("Failed to get sessions by user ID", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to retrieve sessions")
	}

	logger.Debug("Sessions retrieved successfully", zap.Int("count", len(sessions)))
	return sessions, nil
}

// Update updates an existing session
func (r *PostgresSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	logger := r.logger.With(
		zap.String("operation", "update_session"),
		zap.String("session_id", session.ID),
	)

	query := `
		UPDATE user_sessions 
		SET refresh_token = $2, expires_at = $3, ip_address = $4, user_agent = $5
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		session.ID, session.RefreshToken, session.ExpiresAt, session.IPAddress, session.UserAgent)

	if err != nil {
		logger.Error("Failed to update session", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update session")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to update session")
	}

	if rowsAffected == 0 {
		logger.Debug("No rows affected - session not found")
		return domain.NewAuthError(domain.AUTH_008, "Session not found", "No session exists with the provided ID")
	}

	logger.Debug("Session updated successfully")
	return nil
}

// Delete deletes a session by ID
func (r *PostgresSessionRepository) Delete(ctx context.Context, id string) error {
	logger := r.logger.With(
		zap.String("operation", "delete_session"),
		zap.String("session_id", id),
	)

	query := `DELETE FROM user_sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete session", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete session")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete session")
	}

	if rowsAffected == 0 {
		logger.Debug("No rows affected - session not found")
		return domain.NewAuthError(domain.AUTH_008, "Session not found", "No session exists with the provided ID")
	}

	logger.Debug("Session deleted successfully")
	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	logger := r.logger.With(
		zap.String("operation", "delete_sessions_by_user_id"),
		zap.String("user_id", userID),
	)

	query := `DELETE FROM user_sessions WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to delete user sessions", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete user sessions")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete user sessions")
	}

	logger.Debug("User sessions deleted successfully", zap.Int64("count", rowsAffected))
	return nil
}

// DeleteExpired removes expired sessions
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) error {
	logger := r.logger.With(
		zap.String("operation", "delete_expired_sessions"),
	)

	query := `DELETE FROM user_sessions WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		logger.Error("Failed to delete expired sessions", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete expired sessions")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get affected rows", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_017, "Database error", "Failed to delete expired sessions")
	}

	logger.Info("Expired sessions cleaned up", zap.Int64("count", rowsAffected))
	return nil
}
