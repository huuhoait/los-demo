package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/loan-api/domain"
)

// LoanRepository implements domain.LoanRepository interface
type LoanRepository struct {
	db     *Connection
	logger *zap.Logger
}

// NewLoanRepository creates a new loan repository
func NewLoanRepository(db *Connection, logger *zap.Logger) *LoanRepository {
	return &LoanRepository{
		db:     db,
		logger: logger,
	}
}

// CreateApplication creates a new loan application
func (r *LoanRepository) CreateApplication(ctx context.Context, app *domain.LoanApplication) error {
	logger := r.logger.With(
		zap.String("operation", "create_application"),
		zap.String("application_id", app.ID),
		zap.String("user_id", app.UserID),
	)

	query := `
		INSERT INTO loan_applications (
			id, user_id, application_number, loan_amount, loan_purpose, requested_term_months,
			annual_income, monthly_income, employment_status, monthly_debt_payments,
			current_state, status, risk_score, workflow_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err := r.db.Exec(ctx, query,
		app.ID, app.UserID, app.ApplicationNumber, app.LoanAmount, app.LoanPurpose, app.RequestedTerm,
		app.AnnualIncome, app.MonthlyIncome, app.EmploymentStatus, app.MonthlyDebt,
		app.CurrentState, app.Status, app.RiskScore, app.WorkflowID,
		time.Now().UTC(), time.Now().UTC(),
	)

	if err != nil {
		logger.Error("Failed to create application", zap.Error(err))
		return fmt.Errorf("failed to create application: %w", err)
	}

	logger.Info("Application created successfully", zap.String("application_id", app.ID))
	return nil
}

// GetApplicationByID retrieves a loan application by ID
func (r *LoanRepository) GetApplicationByID(ctx context.Context, id string) (*domain.LoanApplication, error) {
	logger := r.logger.With(
		zap.String("operation", "get_application_by_id"),
		zap.String("application_id", id),
	)

	query := `
		SELECT 
			id, user_id, application_number, loan_amount, loan_purpose, requested_term_months,
			annual_income, monthly_income, employment_status, monthly_debt_payments,
			current_state, status, risk_score, workflow_id, created_at, updated_at
		FROM loan_applications WHERE id = $1`

	var app domain.LoanApplication
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.UserID, &app.ApplicationNumber, &app.LoanAmount, &app.LoanPurpose, &app.RequestedTerm,
		&app.AnnualIncome, &app.MonthlyIncome, &app.EmploymentStatus, &app.MonthlyDebt,
		&app.CurrentState, &app.Status, &app.RiskScore, &app.WorkflowID,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Application not found", zap.String("application_id", id))
			return nil, fmt.Errorf("application not found: %s", id)
		}
		logger.Error("Failed to get application by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	app.CreatedAt = createdAt
	app.UpdatedAt = updatedAt

	logger.Info("Application retrieved successfully", zap.String("application_id", id))
	return &app, nil
}

// GetApplicationsByUserID retrieves all applications for a user
func (r *LoanRepository) GetApplicationsByUserID(ctx context.Context, userID string) ([]*domain.LoanApplication, error) {
	logger := r.logger.With(
		zap.String("operation", "get_applications_by_user_id"),
		zap.String("user_id", userID),
	)

	query := `
		SELECT 
			id, user_id, application_number, loan_amount, loan_purpose, requested_term_months,
			annual_income, monthly_income, employment_status, monthly_debt_payments,
			current_state, status, risk_score, workflow_id, created_at, updated_at
		FROM loan_applications WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to query applications by user ID", zap.Error(err))
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var applications []*domain.LoanApplication
	for rows.Next() {
		var app domain.LoanApplication
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&app.ID, &app.UserID, &app.ApplicationNumber, &app.LoanAmount, &app.LoanPurpose, &app.RequestedTerm,
			&app.AnnualIncome, &app.MonthlyIncome, &app.EmploymentStatus, &app.MonthlyDebt,
			&app.CurrentState, &app.Status, &app.RiskScore, &app.WorkflowID,
			&createdAt, &updatedAt,
		)

		if err != nil {
			logger.Error("Failed to scan application row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}

		app.CreatedAt = createdAt
		app.UpdatedAt = updatedAt
		applications = append(applications, &app)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over application rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	logger.Info("Applications retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(applications)))
	return applications, nil
}

// UpdateApplication updates an existing loan application
func (r *LoanRepository) UpdateApplication(ctx context.Context, app *domain.LoanApplication) error {
	logger := r.logger.With(
		zap.String("operation", "update_application"),
		zap.String("application_id", app.ID),
	)

	query := `
		UPDATE loan_applications SET 
			loan_amount = $1, loan_purpose = $2, requested_term_months = $3,
			annual_income = $4, monthly_income = $5, employment_status = $6, monthly_debt_payments = $7,
			current_state = $8, status = $9, risk_score = $10, workflow_id = $11, updated_at = $12
		WHERE id = $13`

	result, err := r.db.Exec(ctx, query,
		app.LoanAmount, app.LoanPurpose, app.RequestedTerm,
		app.AnnualIncome, app.MonthlyIncome, app.EmploymentStatus, app.MonthlyDebt,
		app.CurrentState, app.Status, app.RiskScore, app.WorkflowID,
		time.Now().UTC(), app.ID,
	)

	if err != nil {
		logger.Error("Failed to update application", zap.Error(err))
		return fmt.Errorf("failed to update application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No application found to update", zap.String("application_id", app.ID))
		return fmt.Errorf("application not found: %s", app.ID)
	}

	logger.Info("Application updated successfully", zap.String("application_id", app.ID))
	return nil
}

// DeleteApplication deletes a loan application by ID
func (r *LoanRepository) DeleteApplication(ctx context.Context, id string) error {
	logger := r.logger.With(
		zap.String("operation", "delete_application"),
		zap.String("application_id", id),
	)

	query := `DELETE FROM loan_applications WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete application", zap.Error(err))
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No application found to delete", zap.String("application_id", id))
		return fmt.Errorf("application not found: %s", id)
	}

	logger.Info("Application deleted successfully", zap.String("application_id", id))
	return nil
}

// CreateOffer creates a new loan offer
func (r *LoanRepository) CreateOffer(ctx context.Context, offer *domain.LoanOffer) error {
	logger := r.logger.With(
		zap.String("operation", "create_offer"),
		zap.String("offer_id", offer.ID),
		zap.String("application_id", offer.ApplicationID),
	)

	query := `
		INSERT INTO loan_offers (
			id, application_id, offer_amount, interest_rate, term_months,
			monthly_payment, total_interest, apr, expires_at, status, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err := r.db.Exec(ctx, query,
		offer.ID, offer.ApplicationID, offer.OfferAmount, offer.InterestRate, offer.TermMonths,
		offer.MonthlyPayment, offer.TotalInterest, offer.APR, offer.ExpiresAt, offer.Status,
		time.Now().UTC(),
	)

	if err != nil {
		logger.Error("Failed to create offer", zap.Error(err))
		return fmt.Errorf("failed to create offer: %w", err)
	}

	logger.Info("Offer created successfully", zap.String("offer_id", offer.ID))
	return nil
}

// GetOfferByApplicationID retrieves a loan offer by application ID
func (r *LoanRepository) GetOfferByApplicationID(ctx context.Context, applicationID string) (*domain.LoanOffer, error) {
	logger := r.logger.With(
		zap.String("operation", "get_offer_by_application_id"),
		zap.String("application_id", applicationID),
	)

	query := `
		SELECT 
			id, application_id, offer_amount, interest_rate, term_months,
			monthly_payment, total_interest, apr, expires_at, status, created_at, updated_at
		FROM loan_offers WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`

	var offer domain.LoanOffer
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, query, applicationID).Scan(
		&offer.ID, &offer.ApplicationID, &offer.OfferAmount, &offer.InterestRate, &offer.TermMonths,
		&offer.MonthlyPayment, &offer.TotalInterest, &offer.APR, &offer.ExpiresAt, &offer.Status,
		&createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Offer not found", zap.String("application_id", applicationID))
			return nil, fmt.Errorf("offer not found: %s", applicationID)
		}
		logger.Error("Failed to get offer by application ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	offer.CreatedAt = createdAt

	logger.Info("Offer retrieved successfully", zap.String("offer_id", offer.ID))
	return &offer, nil
}

// UpdateOffer updates an existing loan offer
func (r *LoanRepository) UpdateOffer(ctx context.Context, offer *domain.LoanOffer) error {
	logger := r.logger.With(
		zap.String("operation", "update_offer"),
		zap.String("offer_id", offer.ID),
	)

	query := `
		UPDATE loan_offers SET 
			offer_amount = $1, interest_rate = $2, term_months = $3,
			monthly_payment = $4, total_interest = $5, apr = $6, expires_at = $7, status = $8, updated_at = $9
		WHERE id = $10`

	result, err := r.db.Exec(ctx, query,
		offer.OfferAmount, offer.InterestRate, offer.TermMonths,
		offer.MonthlyPayment, offer.TotalInterest, offer.APR, offer.ExpiresAt, offer.Status,
		time.Now().UTC(), offer.ID,
	)

	if err != nil {
		logger.Error("Failed to update offer", zap.Error(err))
		return fmt.Errorf("failed to update offer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("No offer found to update", zap.String("offer_id", offer.ID))
		return fmt.Errorf("offer not found: %s", offer.ID)
	}

	logger.Info("Offer updated successfully", zap.String("offer_id", offer.ID))
	return nil
}

// CreateStateTransition creates a new state transition record
func (r *LoanRepository) CreateStateTransition(ctx context.Context, transition *domain.StateTransition) error {
	logger := r.logger.With(
		zap.String("operation", "create_state_transition"),
		zap.String("transition_id", transition.ID),
		zap.String("application_id", transition.ApplicationID),
	)

	query := `
		INSERT INTO state_transitions (
			id, application_id, from_state, to_state, transition_reason, triggered_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)`

	fromState := "initiated"
	if transition.FromState != nil {
		fromState = string(*transition.FromState)
	}

	triggeredBy := "system"
	if transition.UserID != nil {
		triggeredBy = *transition.UserID
	}

	_, err := r.db.Exec(ctx, query,
		transition.ID, transition.ApplicationID, fromState, transition.ToState,
		transition.TransitionReason, triggeredBy, time.Now().UTC(),
	)

	if err != nil {
		logger.Error("Failed to create state transition", zap.Error(err))
		return fmt.Errorf("failed to create state transition: %w", err)
	}

	logger.Info("State transition created successfully", zap.String("transition_id", transition.ID))
	return nil
}

// GetStateTransitions retrieves all state transitions for an application
func (r *LoanRepository) GetStateTransitions(ctx context.Context, applicationID string) ([]*domain.StateTransition, error) {
	logger := r.logger.With(
		zap.String("operation", "get_state_transitions"),
		zap.String("application_id", applicationID),
	)

	query := `
		SELECT 
			id, application_id, from_state, to_state, transition_reason, triggered_by, created_at
		FROM state_transitions WHERE application_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, applicationID)
	if err != nil {
		logger.Error("Failed to query state transitions", zap.Error(err))
		return nil, fmt.Errorf("failed to query state transitions: %w", err)
	}
	defer rows.Close()

	var transitions []*domain.StateTransition
	for rows.Next() {
		var transition domain.StateTransition
		var fromState, triggeredBy string
		var createdAt time.Time

		err := rows.Scan(
			&transition.ID, &transition.ApplicationID, &fromState, &transition.ToState,
			&transition.TransitionReason, &triggeredBy, &createdAt,
		)

		if err != nil {
			logger.Error("Failed to scan state transition row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan state transition: %w", err)
		}

		// Convert string to ApplicationState pointer
		if fromState != "" {
			state := domain.ApplicationState(fromState)
			transition.FromState = &state
		}

		if triggeredBy != "" && triggeredBy != "system" {
			transition.UserID = &triggeredBy
		}

		transition.CreatedAt = createdAt
		transitions = append(transitions, &transition)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over state transition rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	logger.Info("State transitions retrieved successfully",
		zap.String("application_id", applicationID),
		zap.Int("count", len(transitions)))
	return transitions, nil
}

// SaveWorkflowExecution saves a workflow execution record
func (r *LoanRepository) SaveWorkflowExecution(ctx context.Context, execution *domain.WorkflowExecution) error {
	logger := r.logger.With(
		zap.String("operation", "save_workflow_execution"),
		zap.String("execution_id", execution.ID),
		zap.String("workflow_id", execution.WorkflowID),
	)

	query := `
		INSERT INTO workflow_executions (
			id, workflow_id, application_id, status, start_time, end_time, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			end_time = EXCLUDED.end_time,
			updated_at = EXCLUDED.updated_at`

	var endTime *time.Time
	if execution.EndTime != nil {
		endTime = execution.EndTime
	}

	_, err := r.db.Exec(ctx, query,
		execution.ID, execution.WorkflowID, execution.ApplicationID, execution.Status,
		execution.StartTime, endTime, time.Now().UTC(), time.Now().UTC(),
	)

	if err != nil {
		logger.Error("Failed to save workflow execution", zap.Error(err))
		return fmt.Errorf("failed to save workflow execution: %w", err)
	}

	logger.Info("Workflow execution saved successfully", zap.String("execution_id", execution.ID))
	return nil
}

// GetWorkflowExecutionByApplicationID retrieves a workflow execution by application ID
func (r *LoanRepository) GetWorkflowExecutionByApplicationID(ctx context.Context, applicationID string) (*domain.WorkflowExecution, error) {
	logger := r.logger.With(
		zap.String("operation", "get_workflow_execution_by_application_id"),
		zap.String("application_id", applicationID),
	)

	query := `
		SELECT 
			id, workflow_id, application_id, status, start_time, end_time, created_at, updated_at
		FROM workflow_executions WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`

	var execution domain.WorkflowExecution
	var startTime, createdAt, updatedAt time.Time
	var endTime *time.Time

	err := r.db.QueryRow(ctx, query, applicationID).Scan(
		&execution.ID, &execution.WorkflowID, &execution.ApplicationID, &execution.Status,
		&startTime, &endTime, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("Workflow execution not found", zap.String("application_id", applicationID))
			return nil, fmt.Errorf("workflow execution not found: %s", applicationID)
		}
		logger.Error("Failed to get workflow execution by application ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get workflow execution: %w", err)
	}

	execution.StartTime = startTime
	execution.EndTime = endTime
	execution.CreatedAt = createdAt

	logger.Info("Workflow execution retrieved successfully", zap.String("execution_id", execution.ID))
	return &execution, nil
}
