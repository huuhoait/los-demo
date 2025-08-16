package application

import (
	"context"
	"time"

	"go.uber.org/zap"

	"loan-api/domain"
	"loan-api/infrastructure/workflow"
	"loan-api/pkg/i18n"
)

// WorkflowService manages workflow operations for loan applications
type WorkflowService struct {
	workflowOrchestrator *workflow.LoanWorkflowOrchestrator
	logger               *zap.Logger
	localizer            *i18n.Localizer
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(
	workflowOrchestrator *workflow.LoanWorkflowOrchestrator,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *WorkflowService {
	return &WorkflowService{
		workflowOrchestrator: workflowOrchestrator,
		logger:               logger,
		localizer:            localizer,
	}
}

// ProcessPreQualificationWorkflow processes a pre-qualification request through workflow
func (s *WorkflowService) ProcessPreQualificationWorkflow(
	ctx context.Context,
	userID string,
	request *domain.PreQualifyRequest,
) (*domain.PreQualifyResult, error) {
	logger := s.logger.With(
		zap.String("user_id", userID),
		zap.String("operation", "process_prequalification_workflow"),
	)

	logger.Info("Starting pre-qualification workflow processing",
		zap.Float64("loan_amount", request.LoanAmount),
		zap.Float64("annual_income", request.AnnualIncome),
	)

	// Start the pre-qualification workflow
	execution, err := s.workflowOrchestrator.StartPreQualificationWorkflow(ctx, userID, request)
	if err != nil {
		logger.Error("Failed to start pre-qualification workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start pre-qualification workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow started",
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("correlation_id", execution.CorrelationID),
	)

	// For now, we'll simulate the workflow result
	// In a real implementation, you would poll the workflow status
	// and wait for completion or handle async callbacks

	// Simulate workflow processing time
	time.Sleep(2 * time.Second)

	// Generate pre-qualification result based on business rules
	result := s.generatePreQualificationResult(request)

	logger.Info("Pre-qualification workflow completed",
		zap.String("workflow_id", execution.WorkflowID),
		zap.Bool("qualified", result.Qualified),
	)

	return result, nil
}

// ProcessLoanApplicationWorkflow processes a loan application through workflow
func (s *WorkflowService) ProcessLoanApplicationWorkflow(
	ctx context.Context,
	application *domain.LoanApplication,
) (*workflow.WorkflowExecution, error) {
	logger := s.logger.With(
		zap.String("application_id", application.ID),
		zap.String("user_id", application.UserID),
		zap.String("operation", "process_loan_application_workflow"),
	)

	logger.Info("Starting loan application workflow processing",
		zap.Float64("loan_amount", application.LoanAmount),
		zap.String("loan_purpose", string(application.LoanPurpose)),
	)

	// Start the loan processing workflow
	execution, err := s.workflowOrchestrator.StartLoanProcessingWorkflow(ctx, application)
	if err != nil {
		logger.Error("Failed to start loan application workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start loan application workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Loan application workflow started",
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("correlation_id", execution.CorrelationID),
	)

	return execution, nil
}

// ProcessUnderwritingWorkflow processes underwriting through workflow
func (s *WorkflowService) ProcessUnderwritingWorkflow(
	ctx context.Context,
	application *domain.LoanApplication,
) (*workflow.WorkflowExecution, error) {
	logger := s.logger.With(
		zap.String("application_id", application.ID),
		zap.String("user_id", application.UserID),
		zap.String("operation", "process_underwriting_workflow"),
	)

	logger.Info("Starting underwriting workflow processing",
		zap.Float64("loan_amount", application.LoanAmount),
		zap.String("current_state", string(application.CurrentState)),
	)

	// Start the underwriting workflow
	execution, err := s.workflowOrchestrator.StartUnderwritingWorkflow(ctx, application)
	if err != nil {
		logger.Error("Failed to start underwriting workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start underwriting workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Underwriting workflow started",
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("correlation_id", execution.CorrelationID),
	)

	return execution, nil
}

// GetWorkflowStatus retrieves the status of a workflow
func (s *WorkflowService) GetWorkflowStatus(
	ctx context.Context,
	workflowID string,
) (*workflow.WorkflowStatus, error) {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_workflow_status"),
	)

	logger.Debug("Retrieving workflow status")

	status, err := s.workflowOrchestrator.GetWorkflowStatus(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to get workflow status", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to get workflow status",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Debug("Workflow status retrieved",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return status, nil
}

// TerminateWorkflow terminates a running workflow
func (s *WorkflowService) TerminateWorkflow(
	ctx context.Context,
	workflowID string,
	reason string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "terminate_workflow"),
	)

	logger.Info("Terminating workflow")

	err := s.workflowOrchestrator.TerminateWorkflow(ctx, workflowID, reason)
	if err != nil {
		logger.Error("Failed to terminate workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to terminate workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Workflow terminated successfully")
	return nil
}

// PauseWorkflow pauses a running workflow
func (s *WorkflowService) PauseWorkflow(
	ctx context.Context,
	workflowID string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "pause_workflow"),
	)

	logger.Info("Pausing workflow")

	err := s.workflowOrchestrator.PauseWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to pause workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to pause workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Workflow paused successfully")
	return nil
}

// ResumeWorkflow resumes a paused workflow
func (s *WorkflowService) ResumeWorkflow(
	ctx context.Context,
	workflowID string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "resume_workflow"),
	)

	logger.Info("Resuming workflow")

	err := s.workflowOrchestrator.ResumeWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to resume workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to resume workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Workflow resumed successfully")
	return nil
}

// generatePreQualificationResult generates a pre-qualification result based on business rules
func (s *WorkflowService) generatePreQualificationResult(request *domain.PreQualifyRequest) *domain.PreQualifyResult {
	// Calculate DTI ratio
	dtiRatio := request.MonthlyDebt / (request.AnnualIncome / 12)

	// Basic qualification logic
	qualified := dtiRatio <= 0.43 && request.AnnualIncome >= 30000

	var maxLoanAmount float64
	var minInterestRate, maxInterestRate float64
	var recommendedTerms []int
	var message string

	if qualified {
		// Calculate max loan amount based on income and DTI
		maxLoanAmount = (request.AnnualIncome * 0.43) / 12 * 60 // 5-year term max
		if maxLoanAmount > 50000 {
			maxLoanAmount = 50000
		}

		// Interest rates based on DTI and income
		if dtiRatio <= 0.30 {
			minInterestRate = 5.0
			maxInterestRate = 8.5
		} else if dtiRatio <= 0.40 {
			minInterestRate = 6.5
			maxInterestRate = 10.5
		} else {
			minInterestRate = 8.5
			maxInterestRate = 12.0
		}

		// Recommended terms
		if request.AnnualIncome >= 50000 {
			recommendedTerms = []int{36, 48, 60, 72}
		} else {
			recommendedTerms = []int{36, 48, 60}
		}

		message = "You are pre-qualified for a loan"
	} else {
		maxLoanAmount = 0
		minInterestRate = 0
		maxInterestRate = 0
		recommendedTerms = []int{}

		if dtiRatio > 0.43 {
			message = "Your debt-to-income ratio is too high"
		} else {
			message = "Your annual income is below the minimum requirement"
		}
	}

	return &domain.PreQualifyResult{
		Qualified:        qualified,
		MaxLoanAmount:    maxLoanAmount,
		MinInterestRate:  minInterestRate,
		MaxInterestRate:  maxInterestRate,
		RecommendedTerms: recommendedTerms,
		DTIRatio:         dtiRatio,
		Message:          message,
	}
}

// HandleWorkflowEvent handles workflow events and updates application state
func (s *WorkflowService) HandleWorkflowEvent(
	ctx context.Context,
	eventType string,
	workflowID string,
	applicationID string,
	eventData map[string]interface{},
) error {
	logger := s.logger.With(
		zap.String("event_type", eventType),
		zap.String("workflow_id", workflowID),
		zap.String("application_id", applicationID),
		zap.String("operation", "handle_workflow_event"),
	)

	logger.Info("Handling workflow event",
		zap.String("event_type", eventType),
	)

	switch eventType {
	case "workflow_completed":
		return s.handleWorkflowCompleted(ctx, workflowID, applicationID, eventData)
	case "workflow_failed":
		return s.handleWorkflowFailed(ctx, workflowID, applicationID, eventData)
	case "task_completed":
		return s.handleTaskCompleted(ctx, workflowID, applicationID, eventData)
	case "state_transition":
		return s.handleStateTransition(ctx, workflowID, applicationID, eventData)
	default:
		logger.Warn("Unknown workflow event type", zap.String("event_type", eventType))
		return nil
	}
}

// handleWorkflowCompleted handles workflow completion events
func (s *WorkflowService) handleWorkflowCompleted(
	ctx context.Context,
	workflowID string,
	applicationID string,
	eventData map[string]interface{},
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("application_id", applicationID),
		zap.String("operation", "handle_workflow_completed"),
	)

	logger.Info("Workflow completed",
		zap.String("workflow_id", workflowID),
	)

	// Here you would typically:
	// 1. Update application status in database
	// 2. Send notifications
	// 3. Trigger next steps
	// 4. Update audit logs

	return nil
}

// handleWorkflowFailed handles workflow failure events
func (s *WorkflowService) handleWorkflowFailed(
	ctx context.Context,
	workflowID string,
	applicationID string,
	eventData map[string]interface{},
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("application_id", applicationID),
		zap.String("operation", "handle_workflow_failed"),
	)

	logger.Error("Workflow failed",
		zap.String("workflow_id", workflowID),
		zap.Any("event_data", eventData),
	)

	// Here you would typically:
	// 1. Update application status to failed
	// 2. Send failure notifications
	// 3. Log error details
	// 4. Trigger manual intervention if needed

	return nil
}

// handleTaskCompleted handles task completion events
func (s *WorkflowService) handleTaskCompleted(
	ctx context.Context,
	workflowID string,
	applicationID string,
	eventData map[string]interface{},
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("application_id", applicationID),
		zap.String("operation", "handle_task_completed"),
	)

	taskName, _ := eventData["task_name"].(string)
	logger.Info("Task completed",
		zap.String("workflow_id", workflowID),
		zap.String("task_name", taskName),
	)

	// Here you would typically:
	// 1. Update application state based on task completion
	// 2. Trigger next workflow steps
	// 3. Update progress tracking
	// 4. Send progress notifications

	return nil
}

// handleStateTransition handles state transition events
func (s *WorkflowService) handleStateTransition(
	ctx context.Context,
	workflowID string,
	applicationID string,
	eventData map[string]interface{},
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("application_id", applicationID),
		zap.String("operation", "handle_state_transition"),
	)

	fromState, _ := eventData["from_state"].(string)
	toState, _ := eventData["to_state"].(string)

	logger.Info("State transition",
		zap.String("workflow_id", workflowID),
		zap.String("from_state", fromState),
		zap.String("to_state", toState),
	)

	// Here you would typically:
	// 1. Update application state in database
	// 2. Log state transition
	// 3. Send state change notifications
	// 4. Trigger state-specific actions

	return nil
}
