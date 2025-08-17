package application

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/loan-api/infrastructure/workflow"
	"github.com/huuhoait/los-demo/services/loan-api/pkg/i18n"
)

// UserRepository interface for user data persistence
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (string, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
}

// LoanRepository interface for data persistence
type LoanRepository interface {
	CreateApplication(ctx context.Context, app *domain.LoanApplication) error
	GetApplicationByID(ctx context.Context, id string) (*domain.LoanApplication, error)
	GetApplicationsByUserID(ctx context.Context, userID string) ([]*domain.LoanApplication, error)
	UpdateApplication(ctx context.Context, app *domain.LoanApplication) error
	DeleteApplication(ctx context.Context, id string) error

	CreateOffer(ctx context.Context, offer *domain.LoanOffer) error
	GetOfferByApplicationID(ctx context.Context, applicationID string) (*domain.LoanOffer, error)
	UpdateOffer(ctx context.Context, offer *domain.LoanOffer) error

	CreateStateTransition(ctx context.Context, transition *domain.StateTransition) error
	GetStateTransitions(ctx context.Context, applicationID string) ([]*domain.StateTransition, error)

	SaveWorkflowExecution(ctx context.Context, execution *domain.WorkflowExecution) error
	GetWorkflowExecutionByApplicationID(ctx context.Context, applicationID string) (*domain.WorkflowExecution, error)
}

// LoanService handles loan business logic
type LoanService struct {
	userRepo             UserRepository
	repo                 LoanRepository
	workflowOrchestrator *workflow.LoanWorkflowOrchestrator
	logger               *zap.Logger
	localizer            *i18n.Localizer
}

// NewLoanService creates a new loan service
func NewLoanService(userRepo UserRepository, repo LoanRepository, workflowOrchestrator *workflow.LoanWorkflowOrchestrator, logger *zap.Logger, localizer *i18n.Localizer) *LoanService {
	return &LoanService{
		userRepo:             userRepo,
		repo:                 repo,
		workflowOrchestrator: workflowOrchestrator,
		logger:               logger,
		localizer:            localizer,
	}
}

// generateApplicationNumber generates a unique application number
func (s *LoanService) generateApplicationNumber() string {
	// Generate application number with format: LOAN-YYYYMMDD-HHMMSS-XXXX
	// Where XXXX is a random 4-digit number for uniqueness
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	timeStr := now.Format("150405")
	randomNum := rand.Intn(9999)

	return fmt.Sprintf("LOAN-%s-%s-%04d", dateStr, timeStr, randomNum)
}

// CreateApplication creates a new loan application with user creation
func (s *LoanService) CreateApplication(ctx context.Context, req *domain.CreateApplicationRequest) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("operation", "create_application"),
	)

	// Validate request
	validation := req.Validate()
	if !validation.Valid {
		logger.Warn("Application validation failed", zap.Any("errors", validation.Errors))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_018,
			Message:     "Application validation failed",
			Description: fmt.Sprintf("Validation errors: %v", validation.Errors),
			HTTPStatus:  400,
		}
	}

	// Check if user already exists by email
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.User.Email)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		logger.Error("Failed to check existing user", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	var userID string
	if existingUser != nil {
		// User exists, use existing user ID
		userID = existingUser.ID
		logger.Info("Using existing user", zap.String("user_id", userID))
	} else {
		// Create new user
		user := req.User
		user.ID = uuid.New().String()
		user.CreatedAt = time.Now().UTC()
		user.UpdatedAt = time.Now().UTC()

		userID, err = s.userRepo.CreateUser(ctx, &user)
		if err != nil {
			logger.Error("Failed to create user", zap.Error(err))
			return nil, &domain.LoanError{
				Code:        domain.LOAN_023,
				Message:     "Failed to create user",
				Description: err.Error(),
				HTTPStatus:  500,
			}
		}
		logger.Info("User created successfully", zap.String("user_id", userID))
	}

	// Create loan application
	application := &domain.LoanApplication{
		ID:                uuid.New().String(),
		UserID:            userID,
		ApplicationNumber: s.generateApplicationNumber(),
		LoanAmount:        req.LoanAmount,
		LoanPurpose:       req.LoanPurpose,
		AnnualIncome:      req.AnnualIncome,
		MonthlyIncome:     req.MonthlyIncome,
		MonthlyDebt:       req.MonthlyDebt,
		RequestedTerm:     req.RequestedTerm,
		EmploymentStatus:  req.EmploymentStatus,
		CurrentState:      domain.StateInitiated,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Save application to database
	if err := s.repo.CreateApplication(ctx, application); err != nil {
		logger.Error("Failed to create application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Failed to create application",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Create initial state transition
	transition := &domain.StateTransition{
		ID:               uuid.New().String(),
		ApplicationID:    application.ID,
		FromState:        nil,
		ToState:          domain.StateInitiated,
		TransitionReason: "Application created",
		Automated:        false,
		UserID:           &userID,
		Metadata:         map[string]interface{}{"source": "api"},
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.repo.CreateStateTransition(ctx, transition); err != nil {
		logger.Warn("Failed to create state transition", zap.Error(err))
		// Don't fail the entire operation for this
	}

	// Start initial workflow for the application
	if s.workflowOrchestrator != nil {
		logger.Info("Starting initial workflow for application",
			zap.String("application_id", application.ID))

		workflowExecution, err := s.workflowOrchestrator.StartLoanProcessingWorkflow(ctx, application)
		if err != nil {
			logger.Error("Failed to start workflow", zap.Error(err))
			// Don't fail the application creation if workflow fails
		} else {
			// Update application with workflow ID
			application.WorkflowID = &workflowExecution.WorkflowID
			if err := s.repo.UpdateApplication(ctx, application); err != nil {
				logger.Error("Failed to update application with workflow ID", zap.Error(err))
			}

			// Save workflow execution
			workflowRecord := &domain.WorkflowExecution{
				ID:            uuid.New().String(),
				WorkflowID:    workflowExecution.WorkflowID,
				ApplicationID: application.ID,
				Status:        workflowExecution.Status,
				Input:         workflowExecution.Input,
				StartTime:     workflowExecution.StartTime,
				CreatedAt:     time.Now().UTC(),
			}
			if err := s.repo.SaveWorkflowExecution(ctx, workflowRecord); err != nil {
				logger.Error("Failed to save workflow execution", zap.Error(err))
			}

			logger.Info("Workflow started successfully",
				zap.String("application_id", application.ID),
				zap.String("workflow_id", workflowExecution.WorkflowID))
		}
	}

	logger.Info("Application created successfully",
		zap.String("application_id", application.ID),
		zap.String("user_id", userID),
		zap.String("state", string(application.CurrentState)),
	)

	return application, nil
}

// GetApplication retrieves a loan application by ID
func (s *LoanService) GetApplication(ctx context.Context, id string) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", id),
		zap.String("operation", "get_application"),
	)

	application, err := s.repo.GetApplicationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn("Application not found")
			return nil, &domain.LoanError{
				Code:        domain.LOAN_010,
				Message:     "Application not found",
				Description: fmt.Sprintf("No application found with ID: %s", id),
				HTTPStatus:  404,
			}
		}
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	return application, nil
}

// GetApplicationsByUser retrieves all applications for a user
func (s *LoanService) GetApplicationsByUser(ctx context.Context, userID string) ([]*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("user_id", userID),
		zap.String("operation", "get_applications_by_user"),
	)

	applications, err := s.repo.GetApplicationsByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get applications", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	return applications, nil
}

// UpdateApplication updates an existing loan application
func (s *LoanService) UpdateApplication(ctx context.Context, id string, req *domain.UpdateApplicationRequest) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", id),
		zap.String("operation", "update_application"),
	)

	// Get existing application
	application, err := s.repo.GetApplicationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn("Application not found")
			return nil, &domain.LoanError{
				Code:        domain.LOAN_010,
				Message:     "Application not found",
				Description: fmt.Sprintf("No application found with ID: %s", id),
				HTTPStatus:  404,
			}
		}
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Update fields if provided
	if req.LoanAmount != nil && *req.LoanAmount > 0 {
		application.LoanAmount = *req.LoanAmount
	}
	if req.LoanPurpose != nil {
		application.LoanPurpose = *req.LoanPurpose
	}
	if req.AnnualIncome != nil && *req.AnnualIncome > 0 {
		application.AnnualIncome = *req.AnnualIncome
	}
	if req.MonthlyIncome != nil && *req.MonthlyIncome > 0 {
		application.MonthlyIncome = *req.MonthlyIncome
	}
	if req.MonthlyDebt != nil && *req.MonthlyDebt > 0 {
		application.MonthlyDebt = *req.MonthlyDebt
	}
	if req.RequestedTerm != nil && *req.RequestedTerm > 0 {
		application.RequestedTerm = *req.RequestedTerm
	}

	application.UpdatedAt = time.Now().UTC()

	// Save updated application
	if err := s.repo.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to update application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Failed to update application",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Application updated successfully")
	return application, nil
}

// SubmitApplication submits an application for processing
func (s *LoanService) SubmitApplication(ctx context.Context, id string) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", id),
		zap.String("operation", "submit_application"),
	)

	// Get existing application
	application, err := s.repo.GetApplicationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warn("Application not found")
			return nil, &domain.LoanError{
				Code:        domain.LOAN_010,
				Message:     "Application not found",
				Description: fmt.Sprintf("No application found with ID: %s", id),
				HTTPStatus:  404,
			}
		}
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Check if application can be submitted
	if application.CurrentState != domain.StateInitiated {
		logger.Warn("Application cannot be submitted from current state",
			zap.String("current_state", string(application.CurrentState)))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_019,
			Message:     "Application cannot be submitted",
			Description: fmt.Sprintf("Application is in %s state and cannot be submitted", application.CurrentState),
			HTTPStatus:  400,
		}
	}

	// Update application state
	application.CurrentState = domain.StatePreQualified
	application.UpdatedAt = time.Now().UTC()

	// Save updated application
	if err := s.repo.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to update application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Failed to submit application",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Create state transition
	transition := &domain.StateTransition{
		ID:               uuid.New().String(),
		ApplicationID:    application.ID,
		FromState:        &application.CurrentState,
		ToState:          domain.StatePreQualified,
		TransitionReason: "Application submitted",
		Automated:        false,
		UserID:           &application.UserID,
		Metadata:         map[string]interface{}{"source": "api"},
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.repo.CreateStateTransition(ctx, transition); err != nil {
		logger.Warn("Failed to create state transition", zap.Error(err))
		// Don't fail the entire operation for this
	}

	// Start pre-qualification workflow when application is submitted
	if s.workflowOrchestrator != nil {
		logger.Info("Starting pre-qualification workflow for submitted application",
			zap.String("application_id", application.ID))

		// Create pre-qualification request
		preQualifyReq := &domain.PreQualifyRequest{
			LoanAmount:       application.LoanAmount,
			AnnualIncome:     application.AnnualIncome,
			MonthlyDebt:      application.MonthlyDebt,
			EmploymentStatus: application.EmploymentStatus,
		}

		workflowExecution, err := s.workflowOrchestrator.StartPreQualificationWorkflow(ctx, application.UserID, preQualifyReq)
		if err != nil {
			logger.Error("Failed to start pre-qualification workflow", zap.Error(err))
			// Don't fail the submission if workflow fails
		} else {
			// Update application with workflow ID
			application.WorkflowID = &workflowExecution.WorkflowID
			if err := s.repo.UpdateApplication(ctx, application); err != nil {
				logger.Error("Failed to update application with workflow ID", zap.Error(err))
			}

			logger.Info("Pre-qualification workflow started successfully",
				zap.String("application_id", application.ID),
				zap.String("workflow_id", workflowExecution.WorkflowID))
		}
	}

	logger.Info("Application submitted successfully",
		zap.String("application_id", id),
		zap.String("new_state", string(application.CurrentState)),
	)

	return application, nil
}

// GetApplicationStats retrieves application statistics
func (s *LoanService) GetApplicationStats(ctx context.Context) (map[string]interface{}, error) {
	logger := s.logger.With(
		zap.String("operation", "get_application_stats"),
	)

	// For now, return basic stats
	// In a real implementation, you would query the database for actual statistics
	stats := map[string]interface{}{
		"total_applications": 0,
		"pending_review":     0,
		"approved":           0,
		"denied":             0,
		"in_progress":        0,
	}

	logger.Info("Application stats retrieved")
	return stats, nil
}
