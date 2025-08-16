package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"loan-api/domain"
)

// UserRepository implements domain.UserRepository interface
type UserRepository struct {
	db     *Connection
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Connection, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	logger := r.logger.With(
		zap.String("operation", "create_user"),
		zap.String("email", user.Email),
	)

	query := `
		INSERT INTO users (
			first_name, last_name, email, phone_number, date_of_birth, ssn,
			street_address, city, state, zip_code, country, residence_type, time_at_address_months,
			employer_name, job_title, time_employed_months, work_phone, work_email,
			bank_name, account_type, account_number, routing_number,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		) RETURNING id`

	var userID string
	err := r.db.QueryRow(ctx, query,
		user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.DateOfBirth, user.SSN,
		user.Address.StreetAddress, user.Address.City, user.Address.State, user.Address.ZipCode,
		user.Address.Country, user.Address.ResidenceType, user.Address.TimeAtAddress,
		user.EmploymentInfo.EmployerName, user.EmploymentInfo.JobTitle, user.EmploymentInfo.TimeEmployed,
		user.EmploymentInfo.WorkPhone, user.EmploymentInfo.WorkEmail,
		user.BankingInfo.BankName, user.BankingInfo.AccountType, user.BankingInfo.AccountNumber,
		user.BankingInfo.RoutingNumber,
		time.Now().UTC(), time.Now().UTC(),
	).Scan(&userID)

	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = userID
	logger.Info("User created successfully", zap.String("user_id", userID))
	return userID, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	logger := r.logger.With(
		zap.String("operation", "get_user_by_id"),
		zap.String("user_id", id),
	)

	query := `
		SELECT 
			id, first_name, last_name, email, phone_number, date_of_birth, ssn,
			street_address, city, state, zip_code, country, residence_type, time_at_address_months,
			employer_name, job_title, time_employed_months, work_phone, work_email,
			bank_name, account_type, account_number, routing_number,
			created_at, updated_at
		FROM users WHERE id = $1`

	var user domain.User
	var dateOfBirth time.Time
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &dateOfBirth, &user.SSN,
		&user.Address.StreetAddress, &user.Address.City, &user.Address.State, &user.Address.ZipCode,
		&user.Address.Country, &user.Address.ResidenceType, &user.Address.TimeAtAddress,
		&user.EmploymentInfo.EmployerName, &user.EmploymentInfo.JobTitle, &user.EmploymentInfo.TimeEmployed,
		&user.EmploymentInfo.WorkPhone, &user.EmploymentInfo.WorkEmail,
		&user.BankingInfo.BankName, &user.BankingInfo.AccountType, &user.BankingInfo.AccountNumber,
		&user.BankingInfo.RoutingNumber,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("User not found", zap.String("user_id", id))
			return nil, fmt.Errorf("user not found: %s", id)
		}
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.DateOfBirth = dateOfBirth
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	logger.Info("User retrieved successfully", zap.String("user_id", id))
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	logger := r.logger.With(
		zap.String("operation", "get_user_by_email"),
		zap.String("email", email),
	)

	query := `
		SELECT 
			id, first_name, last_name, email, phone_number, date_of_birth, ssn,
			street_address, city, state, zip_code, country, residence_type, time_at_address_months,
			employer_name, job_title, time_employed_months, work_phone, work_email,
			bank_name, account_type, account_number, routing_number,
			created_at, updated_at
		FROM users WHERE email = $1`

	var user domain.User
	var dateOfBirth time.Time
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &dateOfBirth, &user.SSN,
		&user.Address.StreetAddress, &user.Address.City, &user.Address.State, &user.Address.ZipCode,
		&user.Address.Country, &user.Address.ResidenceType, &user.Address.TimeAtAddress,
		&user.EmploymentInfo.EmployerName, &user.EmploymentInfo.JobTitle, &user.EmploymentInfo.TimeEmployed,
		&user.EmploymentInfo.WorkPhone, &user.EmploymentInfo.WorkEmail,
		&user.BankingInfo.BankName, &user.BankingInfo.AccountType, &user.BankingInfo.AccountNumber,
		&user.BankingInfo.RoutingNumber,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("User not found", zap.String("email", email))
			return nil, fmt.Errorf("user not found: %s", email)
		}
		logger.Error("Failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.DateOfBirth = dateOfBirth
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	logger.Info("User retrieved successfully", zap.String("email", email))
	return &user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	logger := r.logger.With(
		zap.String("operation", "update_user"),
		zap.String("user_id", user.ID),
	)

	query := `
		UPDATE users SET 
			first_name = $1, last_name = $2, email = $3, phone_number = $4, date_of_birth = $5, ssn = $6,
			street_address = $7, city = $8, state = $9, zip_code = $10, country = $11, residence_type = $12, time_at_address_months = $13,
			employer_name = $14, job_title = $15, time_employed_months = $16, work_phone = $17, work_email = $18,
			bank_name = $19, account_type = $20, account_number = $21, routing_number = $22,
			updated_at = $23
		WHERE id = $24`

	result, err := r.db.Exec(ctx, query,
		user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.DateOfBirth, user.SSN,
		user.Address.StreetAddress, user.Address.City, user.Address.State, user.Address.ZipCode,
		user.Address.Country, user.Address.ResidenceType, user.Address.TimeAtAddress,
		user.EmploymentInfo.EmployerName, user.EmploymentInfo.JobTitle, user.EmploymentInfo.TimeEmployed,
		user.EmploymentInfo.WorkPhone, user.EmploymentInfo.WorkEmail,
		user.BankingInfo.BankName, user.BankingInfo.AccountType, user.BankingInfo.AccountNumber,
		user.BankingInfo.RoutingNumber,
		time.Now().UTC(), user.ID,
	)

	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No user found to update", zap.String("user_id", user.ID))
		return fmt.Errorf("user not found: %s", user.ID)
	}

	logger.Info("User updated successfully", zap.String("user_id", user.ID))
	return nil
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	logger := r.logger.With(
		zap.String("operation", "delete_user"),
		zap.String("user_id", id),
	)

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No user found to delete", zap.String("user_id", id))
		return fmt.Errorf("user not found: %s", id)
	}

	logger.Info("User deleted successfully", zap.String("user_id", id))
	return nil
}
