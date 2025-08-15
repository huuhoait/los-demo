package application

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"loan-service/domain"
	"loan-service/infrastructure/workflow"
	"loan-service/pkg/i18n"
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
	userRepo                 UserRepository
	repo                     LoanRepository
	workflowOrchestrator     *workflow.LoanWorkflowOrchestrator
	prequalificationWorkflow *PreQualificationWorkflowService
	logger                   *zap.Logger
	localizer                *i18n.Localizer
}

// NewLoanService creates a new loan service
func NewLoanService(userRepo UserRepository, repo LoanRepository, workflowOrchestrator *workflow.LoanWorkflowOrchestrator, logger *zap.Logger, localizer *i18n.Localizer) *LoanService {
	prequalificationWorkflow := NewPreQualificationWorkflowService(workflowOrchestrator, logger, localizer)

	return &LoanService{
		userRepo:                 userRepo,
		repo:                     repo,
		workflowOrchestrator:     workflowOrchestrator,
		prequalificationWorkflow: prequalificationWorkflow,
		logger:                   logger,
		localizer:                localizer,
	}
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
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

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
		logger.Info("Created new user", zap.String("user_id", userID))
	}

	// Check if user already has an active application
	existingApps, err := s.repo.GetApplicationsByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to check existing applications", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Check for active applications
	for _, app := range existingApps {
		if app.Status == domain.StatusSubmitted || app.Status == domain.StatusUnderReview {
			logger.Warn("User already has an active application", zap.String("existing_app_id", app.ID))
			return nil, &domain.LoanError{
				Code:        domain.LOAN_029,
				Message:     "Application already exists",
				Description: "You already have an active loan application",
				HTTPStatus:  409,
			}
		}
	}

	// Create application
	application := &domain.LoanApplication{
		ID:                uuid.New().String(),
		UserID:            userID,
		ApplicationNumber: s.generateApplicationNumber(),
		LoanAmount:        req.LoanAmount,
		LoanPurpose:       req.LoanPurpose,
		RequestedTerm:     req.RequestedTerm,
		AnnualIncome:      req.AnnualIncome,
		MonthlyIncome:     req.MonthlyIncome,
		EmploymentStatus:  req.EmploymentStatus,
		MonthlyDebt:       req.MonthlyDebt,
		CurrentState:      domain.StateInitiated,
		Status:            domain.StatusDraft,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save to database
	if err := s.repo.CreateApplication(ctx, application); err != nil {
		logger.Error("Failed to create application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Start initial workflow
	workflowExecution, err := s.workflowOrchestrator.StartLoanProcessingWorkflow(ctx, application)
	if err != nil {
		logger.Error("Failed to start workflow", zap.Error(err))
		// Don't fail the application creation if workflow fails
		// but log it for monitoring
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
			CreatedAt:     time.Now(),
		}
		if err := s.repo.SaveWorkflowExecution(ctx, workflowRecord); err != nil {
			logger.Error("Failed to save workflow execution", zap.Error(err))
		}
	}

	logger.Info("Application created successfully",
		zap.String("application_id", application.ID),
		zap.String("application_number", application.ApplicationNumber),
	)

	return application, nil
}

// GetApplication retrieves a loan application by ID
func (s *LoanService) GetApplication(ctx context.Context, applicationID string) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "get_application"),
	)

	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	return application, nil
}

// UpdateApplication updates a loan application
func (s *LoanService) UpdateApplication(ctx context.Context, applicationID string, req *domain.UpdateApplicationRequest) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "update_application"),
	)

	// Get existing application
	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	// Check if application can be updated
	if application.Status != domain.StatusDraft {
		logger.Warn("Attempted to update non-draft application")
		return nil, &domain.LoanError{
			Code:        domain.LOAN_019,
			Message:     "Invalid application status",
			Description: "Only draft applications can be updated",
			HTTPStatus:  400,
		}
	}

	// Apply updates
	if req.LoanAmount != nil {
		application.LoanAmount = *req.LoanAmount
	}
	if req.LoanPurpose != nil {
		application.LoanPurpose = *req.LoanPurpose
	}
	if req.RequestedTerm != nil {
		application.RequestedTerm = *req.RequestedTerm
	}
	if req.AnnualIncome != nil {
		application.AnnualIncome = *req.AnnualIncome
	}
	if req.MonthlyIncome != nil {
		application.MonthlyIncome = *req.MonthlyIncome
	}
	if req.EmploymentStatus != nil {
		application.EmploymentStatus = *req.EmploymentStatus
	}
	if req.MonthlyDebt != nil {
		application.MonthlyDebt = *req.MonthlyDebt
	}

	application.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to update application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Application updated successfully")
	return application, nil
}

// SubmitApplication submits a draft application for processing
func (s *LoanService) SubmitApplication(ctx context.Context, applicationID string) (*domain.LoanApplication, error) {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "submit_application"),
	)

	// Get application
	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	// Check if application can be submitted
	if application.Status != domain.StatusDraft {
		logger.Warn("Attempted to submit non-draft application")
		return nil, &domain.LoanError{
			Code:        domain.LOAN_019,
			Message:     "Invalid application status",
			Description: "Only draft applications can be submitted",
			HTTPStatus:  400,
		}
	}

	// Update status and state
	application.Status = domain.StatusSubmitted
	application.CurrentState = domain.StatePreQualified
	application.UpdatedAt = time.Now()

	// Save changes
	if err := s.repo.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to submit application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Record state transition
	fromState := domain.StateInitiated
	transition := &domain.StateTransition{
		ID:               uuid.New().String(),
		ApplicationID:    applicationID,
		FromState:        &fromState,
		ToState:          domain.StatePreQualified,
		TransitionReason: "Application submitted",
		Automated:        false,
		CreatedAt:        time.Now(),
	}

	if err := s.repo.CreateStateTransition(ctx, transition); err != nil {
		logger.Error("Failed to record state transition", zap.Error(err))
	}

	logger.Info("Application submitted successfully")
	return application, nil
}

// PreQualify performs pre-qualification check
func (s *LoanService) PreQualify(ctx context.Context, userID string, req *domain.PreQualifyRequest) (*domain.PreQualifyResult, error) {
	logger := s.logger.With(
		zap.String("user_id", userID),
		zap.String("operation", "pre_qualify"),
	)

	// Calculate DTI ratio
	monthlyDebtPayment := req.MonthlyDebt
	monthlyIncome := req.AnnualIncome / 12
	dtiRatio := monthlyDebtPayment / monthlyIncome

	logger.Info("Calculating pre-qualification",
		zap.Float64("dti_ratio", dtiRatio),
		zap.Float64("requested_amount", req.LoanAmount),
	)

	result := &domain.PreQualifyResult{
		DTIRatio: dtiRatio,
	}

	// Basic qualification rules
	if dtiRatio > 0.4 {
		result.Qualified = false
		result.Message = "Debt-to-income ratio too high"
		return result, nil
	}

	if req.AnnualIncome < 25000 {
		result.Qualified = false
		result.Message = "Minimum income requirement not met"
		return result, nil
	}

	// Calculate max loan amount (simplified calculation)
	maxMonthlyPayment := monthlyIncome * 0.25                        // 25% of income
	maxLoanAmount := s.calculateMaxLoanAmount(maxMonthlyPayment, 60) // 60 months term

	result.Qualified = true
	result.MaxLoanAmount = math.Min(maxLoanAmount, 50000) // Cap at $50k
	result.MinInterestRate = s.calculateInterestRate(req.AnnualIncome, dtiRatio, true)
	result.MaxInterestRate = s.calculateInterestRate(req.AnnualIncome, dtiRatio, false)
	result.RecommendedTerms = []int{36, 48, 60}
	result.Message = "You are pre-qualified for a loan"

	// Start pre-qualification workflow
	if _, err := s.prequalificationWorkflow.ProcessPreQualificationWorkflow(ctx, userID, req); err != nil {
		logger.Error("Failed to start pre-qualification workflow", zap.Error(err))
		// Don't fail the pre-qualification if workflow fails
	}

	logger.Info("Pre-qualification completed",
		zap.Bool("qualified", result.Qualified),
		zap.Float64("max_amount", result.MaxLoanAmount),
	)

	return result, nil
}

// GenerateOffer generates a loan offer for an application
func (s *LoanService) GenerateOffer(ctx context.Context, applicationID string) (*domain.LoanOffer, error) {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "generate_offer"),
	)

	// Get application
	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	// Check if application is in correct state for offer generation
	if application.CurrentState != domain.StateApproved {
		logger.Warn("Attempted to generate offer for non-approved application")
		return nil, &domain.LoanError{
			Code:        domain.LOAN_008,
			Message:     "Invalid state transition",
			Description: "Offers can only be generated for approved applications",
			HTTPStatus:  400,
		}
	}

	// Calculate offer terms
	dtiRatio := application.CalculateDTI()
	interestRate := s.calculateInterestRate(application.AnnualIncome, dtiRatio, true)
	monthlyPayment := s.calculateMonthlyPayment(application.LoanAmount, interestRate, application.RequestedTerm)
	totalInterest := (monthlyPayment * float64(application.RequestedTerm)) - application.LoanAmount
	apr := s.calculateAPR(interestRate, application.LoanAmount, application.RequestedTerm)

	// Create offer
	offer := &domain.LoanOffer{
		ID:             uuid.New().String(),
		ApplicationID:  applicationID,
		OfferAmount:    application.LoanAmount,
		InterestRate:   interestRate,
		TermMonths:     application.RequestedTerm,
		MonthlyPayment: monthlyPayment,
		TotalInterest:  totalInterest,
		APR:            apr,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	// Save offer
	if err := s.repo.CreateOffer(ctx, offer); err != nil {
		logger.Error("Failed to create offer", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Loan offer generated successfully",
		zap.String("offer_id", offer.ID),
		zap.Float64("offer_amount", offer.OfferAmount),
		zap.Float64("interest_rate", offer.InterestRate),
	)

	return offer, nil
}

// AcceptOffer accepts a loan offer
func (s *LoanService) AcceptOffer(ctx context.Context, applicationID string, req *domain.AcceptOfferRequest) error {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("offer_id", req.OfferID),
		zap.String("operation", "accept_offer"),
	)

	// Get offer
	offer, err := s.repo.GetOfferByApplicationID(ctx, applicationID)
	if err != nil || offer.ID != req.OfferID {
		logger.Error("Failed to get offer", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_009,
			Message:     "Offer not found or expired",
			Description: "The specified offer could not be found",
			HTTPStatus:  404,
		}
	}

	// Check if offer is expired
	if offer.IsExpired() {
		logger.Warn("Attempted to accept expired offer")
		return &domain.LoanError{
			Code:        domain.LOAN_009,
			Message:     "Offer has expired",
			Description: "This offer has expired and can no longer be accepted",
			HTTPStatus:  400,
		}
	}

	// Update offer status
	offer.Status = "accepted"
	if err := s.repo.UpdateOffer(ctx, offer); err != nil {
		logger.Error("Failed to update offer", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Update application state
	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	if err := s.TransitionState(ctx, applicationID, application.CurrentState, domain.StateDocumentsSigned); err != nil {
		logger.Error("Failed to transition application state", zap.Error(err))
		return err
	}

	logger.Info("Offer accepted successfully")
	return nil
}

// TransitionState transitions an application to a new state
func (s *LoanService) TransitionState(ctx context.Context, applicationID string, fromState, toState domain.ApplicationState) error {
	logger := s.logger.With(
		zap.String("application_id", applicationID),
		zap.String("from_state", string(fromState)),
		zap.String("to_state", string(toState)),
		zap.String("operation", "transition_state"),
	)

	// Get application
	application, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_010,
			Message:     "Application not found",
			Description: err.Error(),
			HTTPStatus:  404,
		}
	}

	// Check if transition is valid
	if !application.CanTransitionTo(toState) {
		logger.Warn("Invalid state transition attempted")
		return &domain.LoanError{
			Code:        domain.LOAN_008,
			Message:     "Invalid state transition",
			Description: fmt.Sprintf("Cannot transition from %s to %s", fromState, toState),
			HTTPStatus:  400,
		}
	}

	// Update application state
	application.CurrentState = toState
	application.UpdatedAt = time.Now()

	// Update status based on state
	switch toState {
	case domain.StateApproved:
		application.Status = domain.StatusApproved
	case domain.StateDenied:
		application.Status = domain.StatusDenied
	case domain.StateFunded:
		application.Status = domain.StatusFunded
	case domain.StateActive:
		application.Status = domain.StatusActive
	}

	// Save changes
	if err := s.repo.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to update application state", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_023,
			Message:     "Database error",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Record state transition
	transition := &domain.StateTransition{
		ID:               uuid.New().String(),
		ApplicationID:    applicationID,
		FromState:        &fromState,
		ToState:          toState,
		TransitionReason: "State transition",
		Automated:        true,
		CreatedAt:        time.Now(),
	}

	if err := s.repo.CreateStateTransition(ctx, transition); err != nil {
		logger.Error("Failed to record state transition", zap.Error(err))
	}

	// Handle workflow state transition
	if err := s.workflowOrchestrator.HandleStateTransition(ctx, applicationID, fromState, toState); err != nil {
		logger.Error("Failed to handle workflow state transition", zap.Error(err))
		// Don't fail the state transition if workflow fails
	}

	logger.Info("State transition completed successfully")
	return nil
}

// GetApplicationsByUserID retrieves all applications for a user
func (s *LoanService) GetApplicationsByUserID(ctx context.Context, userID string) ([]*domain.LoanApplication, error) {
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

// Helper methods

func (s *LoanService) generateApplicationNumber() string {
	return fmt.Sprintf("LOAN%d", time.Now().UnixNano()%1000000)
}

func (s *LoanService) calculateInterestRate(annualIncome, dtiRatio float64, best bool) float64 {
	baseRate := 8.0 // Base rate of 8%

	// Adjust based on income
	if annualIncome > 100000 {
		baseRate -= 1.0
	} else if annualIncome < 50000 {
		baseRate += 1.0
	}

	// Adjust based on DTI ratio
	if dtiRatio > 0.3 {
		baseRate += 1.5
	} else if dtiRatio < 0.2 {
		baseRate -= 0.5
	}

	if !best {
		baseRate += 2.0 // Higher rate for max calculation
	}

	// Cap rates
	if baseRate < 5.0 {
		baseRate = 5.0
	}
	if baseRate > 25.0 {
		baseRate = 25.0
	}

	return baseRate
}

func (s *LoanService) calculateMonthlyPayment(loanAmount, annualRate float64, termMonths int) float64 {
	monthlyRate := annualRate / 100 / 12
	if monthlyRate == 0 {
		return loanAmount / float64(termMonths)
	}

	payment := loanAmount * (monthlyRate * math.Pow(1+monthlyRate, float64(termMonths))) / (math.Pow(1+monthlyRate, float64(termMonths)) - 1)
	return math.Round(payment*100) / 100
}

func (s *LoanService) calculateMaxLoanAmount(maxMonthlyPayment float64, termMonths int) float64 {
	// Assuming average interest rate of 10%
	monthlyRate := 0.10 / 12
	maxAmount := maxMonthlyPayment * (math.Pow(1+monthlyRate, float64(termMonths)) - 1) / (monthlyRate * math.Pow(1+monthlyRate, float64(termMonths)))
	return math.Round(maxAmount*100) / 100
}

func (s *LoanService) calculateAPR(interestRate, loanAmount float64, termMonths int) float64 {
	// Simplified APR calculation (in reality, includes fees)
	// For now, just return the interest rate
	return interestRate
}
