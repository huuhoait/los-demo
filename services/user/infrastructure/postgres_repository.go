package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/user/domain"
)

type PostgresUserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPostgresUserRepository(db *sqlx.DB, logger *zap.Logger) domain.UserRepository {
	return &PostgresUserRepository{
		db:     db,
		logger: logger,
	}
}

// User CRUD operations

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, phone, email_verified, phone_verified, status, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :phone, :email_verified, :phone_verified, :status, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("user_id", user.ID))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created successfully", zap.String("user_id", user.ID))
	return nil
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, email, password_hash, phone, email_verified, phone_verified, status, created_at, updated_at
		FROM users 
		WHERE id = $1 AND status != 'deleted'`

	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user not found")
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, email, password_hash, phone, email_verified, phone_verified, status, created_at, updated_at
		FROM users 
		WHERE email = $1 AND status != 'deleted'`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user not found")
		}
		r.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE users 
		SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)
	args = append(args, userID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user not found")
	}

	r.logger.Info("User updated successfully", zap.String("user_id", userID))
	return nil
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET status = 'deleted', updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user not found")
	}

	r.logger.Info("User deleted successfully", zap.String("user_id", userID))
	return nil
}

// User profile operations

func (r *PostgresUserRepository) CreateProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (id, user_id, first_name, last_name, date_of_birth, ssn_encrypted, phone, address, employment_info, financial_info, created_at, updated_at)
		VALUES (:id, :user_id, :first_name, :last_name, :date_of_birth, :ssn_encrypted, :phone, :address, :employment_info, :financial_info, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, profile)
	if err != nil {
		r.logger.Error("Failed to create profile", zap.Error(err), zap.String("user_id", profile.UserID))
		return fmt.Errorf("failed to create profile: %w", err)
	}

	r.logger.Info("Profile created successfully", zap.String("user_id", profile.UserID))
	return nil
}

func (r *PostgresUserRepository) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	var profile domain.UserProfile
	query := `
		SELECT id, user_id, first_name, last_name, date_of_birth, ssn_encrypted, phone, address, employment_info, financial_info, created_at, updated_at
		FROM user_profiles 
		WHERE user_id = $1`

	err := r.db.GetContext(ctx, &profile, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("profile not found")
		}
		r.logger.Error("Failed to get profile", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return &profile, nil
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE user_profiles 
		SET %s
		WHERE user_id = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)
	args = append(args, userID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update profile", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("profile not found")
	}

	r.logger.Info("Profile updated successfully", zap.String("user_id", userID))
	return nil
}

// User search and listing

func (r *PostgresUserRepository) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	var users []*domain.User
	query := `
		SELECT id, email, password_hash, phone, email_verified, phone_verified, status, created_at, updated_at
		FROM users 
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	r.logger.Info("Users listed successfully", zap.Int("count", len(users)))
	return users, nil
}

func (r *PostgresUserRepository) SearchUsers(ctx context.Context, criteria map[string]interface{}) ([]*domain.User, error) {
	// Build dynamic search query
	whereParts := []string{"status != 'deleted'"}
	args := make([]interface{}, 0)
	argIndex := 1

	for field, value := range criteria {
		switch field {
		case "email":
			whereParts = append(whereParts, fmt.Sprintf("email ILIKE $%d", argIndex))
			args = append(args, "%"+value.(string)+"%")
			argIndex++
		case "phone":
			whereParts = append(whereParts, fmt.Sprintf("phone = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			whereParts = append(whereParts, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "email_verified":
			whereParts = append(whereParts, fmt.Sprintf("email_verified = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "phone_verified":
			whereParts = append(whereParts, fmt.Sprintf("phone_verified = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	query := fmt.Sprintf(`
		SELECT id, email, password_hash, phone, email_verified, phone_verified, status, created_at, updated_at
		FROM users 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT 100`,
		strings.Join(whereParts, " AND "),
	)

	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		r.logger.Error("Failed to search users", zap.Error(err))
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	r.logger.Info("User search completed", zap.Int("count", len(users)))
	return users, nil
}

// KYC Repository implementation

type PostgresKYCRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPostgresKYCRepository(db *sqlx.DB, logger *zap.Logger) domain.KYCRepository {
	return &PostgresKYCRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresKYCRepository) CreateKYCVerification(ctx context.Context, verification *domain.KYCVerification) error {
	query := `
		INSERT INTO kyc_verifications (id, user_id, verification_type, provider, status, provider_reference, verification_data, created_at, updated_at)
		VALUES (:id, :user_id, :verification_type, :provider, :status, :provider_reference, :verification_data, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, verification)
	if err != nil {
		r.logger.Error("Failed to create KYC verification", zap.Error(err), zap.String("user_id", verification.UserID))
		return fmt.Errorf("failed to create KYC verification: %w", err)
	}

	r.logger.Info("KYC verification created successfully", zap.String("user_id", verification.UserID))
	return nil
}

func (r *PostgresKYCRepository) GetKYCVerification(ctx context.Context, userID, verificationType string) (*domain.KYCVerification, error) {
	var verification domain.KYCVerification
	query := `
		SELECT id, user_id, verification_type, provider, status, provider_reference, verification_data, created_at, updated_at
		FROM kyc_verifications 
		WHERE user_id = $1 AND verification_type = $2
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.db.GetContext(ctx, &verification, query, userID, verificationType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("KYC verification not found")
		}
		r.logger.Error("Failed to get KYC verification", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get KYC verification: %w", err)
	}

	return &verification, nil
}

func (r *PostgresKYCRepository) UpdateKYCVerification(ctx context.Context, verificationID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE kyc_verifications 
		SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)
	args = append(args, verificationID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update KYC verification", zap.Error(err), zap.String("verification_id", verificationID))
		return fmt.Errorf("failed to update KYC verification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("KYC verification not found")
	}

	r.logger.Info("KYC verification updated successfully", zap.String("verification_id", verificationID))
	return nil
}

func (r *PostgresKYCRepository) ListKYCVerifications(ctx context.Context, userID string) ([]*domain.KYCVerification, error) {
	var verifications []*domain.KYCVerification
	query := `
		SELECT id, user_id, verification_type, provider, status, provider_reference, verification_data, created_at, updated_at
		FROM kyc_verifications 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &verifications, query, userID)
	if err != nil {
		r.logger.Error("Failed to list KYC verifications", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to list KYC verifications: %w", err)
	}

	r.logger.Info("KYC verifications listed successfully", zap.String("user_id", userID), zap.Int("count", len(verifications)))
	return verifications, nil
}

func (r *PostgresKYCRepository) GetKYCStatus(ctx context.Context, userID string) (map[string]domain.KYCStatus, error) {
	var records []struct {
		VerificationType string           `db:"verification_type"`
		Status           domain.KYCStatus `db:"status"`
	}

	query := `
		SELECT DISTINCT ON (verification_type) verification_type, status
		FROM kyc_verifications 
		WHERE user_id = $1
		ORDER BY verification_type, created_at DESC`

	err := r.db.SelectContext(ctx, &records, query, userID)
	if err != nil {
		r.logger.Error("Failed to get KYC status", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get KYC status: %w", err)
	}

	status := make(map[string]domain.KYCStatus)
	for _, record := range records {
		status[record.VerificationType] = record.Status
	}

	r.logger.Info("KYC status retrieved successfully", zap.String("user_id", userID))
	return status, nil
}

func (r *PostgresKYCRepository) UpdateKYCStatus(ctx context.Context, userID, verificationType string, status domain.KYCStatus) error {
	query := `
		UPDATE kyc_verifications 
		SET status = $1, updated_at = NOW()
		WHERE user_id = $2 AND verification_type = $3`

	result, err := r.db.ExecContext(ctx, query, status, userID, verificationType)
	if err != nil {
		r.logger.Error("Failed to update KYC status", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update KYC status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("KYC verification not found")
	}

	r.logger.Info("KYC status updated successfully", zap.String("user_id", userID))
	return nil
}

// Document Repository implementation

type PostgresDocumentRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPostgresDocumentRepository(db *sqlx.DB, logger *zap.Logger) domain.DocumentRepository {
	return &PostgresDocumentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresDocumentRepository) CreateDocument(ctx context.Context, document *domain.Document) error {
	query := `
		INSERT INTO user_documents (id, user_id, document_type, file_path, file_size, mime_type, encryption_key, upload_ip, created_at)
		VALUES (:id, :user_id, :document_type, :file_path, :file_size, :mime_type, :encryption_key, :upload_ip, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, document)
	if err != nil {
		r.logger.Error("Failed to create document", zap.Error(err), zap.String("user_id", document.UserID))
		return fmt.Errorf("failed to create document: %w", err)
	}

	r.logger.Info("Document created successfully", zap.String("document_id", document.ID))
	return nil
}

func (r *PostgresDocumentRepository) GetDocument(ctx context.Context, documentID string) (*domain.Document, error) {
	var document domain.Document
	query := `
		SELECT id, user_id, document_type, file_path, file_size, mime_type, encryption_key, upload_ip, created_at
		FROM user_documents 
		WHERE id = $1`

	err := r.db.GetContext(ctx, &document, query, documentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("document not found")
		}
		r.logger.Error("Failed to get document", zap.Error(err), zap.String("document_id", documentID))
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &document, nil
}

func (r *PostgresDocumentRepository) GetDocumentsByUserID(ctx context.Context, userID string) ([]*domain.Document, error) {
	var documents []*domain.Document
	query := `
		SELECT id, user_id, document_type, file_path, file_size, mime_type, encryption_key, upload_ip, created_at
		FROM user_documents 
		WHERE user_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &documents, query, userID)
	if err != nil {
		r.logger.Error("Failed to get documents by user ID", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	r.logger.Info("Documents retrieved successfully", zap.String("user_id", userID), zap.Int("count", len(documents)))
	return documents, nil
}

func (r *PostgresDocumentRepository) GetDocumentsByType(ctx context.Context, userID, documentType string) ([]*domain.Document, error) {
	var documents []*domain.Document
	query := `
		SELECT id, user_id, document_type, file_path, file_size, mime_type, encryption_key, upload_ip, created_at
		FROM user_documents 
		WHERE user_id = $1 AND document_type = $2
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &documents, query, userID, documentType)
	if err != nil {
		r.logger.Error("Failed to get documents by type", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return documents, nil
}

func (r *PostgresDocumentRepository) UpdateDocument(ctx context.Context, documentID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE user_documents 
		SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "),
		argIndex,
	)
	args = append(args, documentID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update document", zap.Error(err), zap.String("document_id", documentID))
		return fmt.Errorf("failed to update document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("document not found")
	}

	r.logger.Info("Document updated successfully", zap.String("document_id", documentID))
	return nil
}

func (r *PostgresDocumentRepository) DeleteDocument(ctx context.Context, documentID string) error {
	query := `DELETE FROM user_documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, documentID)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.Error(err), zap.String("document_id", documentID))
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("document not found")
	}

	r.logger.Info("Document deleted successfully", zap.String("document_id", documentID))
	return nil
}
