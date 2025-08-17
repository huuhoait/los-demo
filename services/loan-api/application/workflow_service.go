package application

import (
	"context"

	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/loan-api/infrastructure/workflow"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
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

// StartPreQualificationWorkflow starts a pre-qualification workflow
func (s *WorkflowService) StartPreQualificationWorkflow(
	ctx context.Context,
	userID string,
	request *domain.PreQualifyRequest,
) (*workflow.WorkflowExecution, error) {
	logger := s.logger.With(
		zap.String("user_id", userID),
		zap.String("operation", "start_prequalification_workflow"),
	)

	logger.Info("Starting pre-qualification workflow",
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

	return execution, nil
}

// StartLoanApplicationWorkflow starts a loan application workflow
func (s *WorkflowService) StartLoanApplicationWorkflow(
	ctx context.Context,
	application *domain.LoanApplication,
) (*workflow.WorkflowExecution, error) {
	logger := s.logger.With(
		zap.String("application_id", application.ID),
		zap.String("user_id", application.UserID),
		zap.String("operation", "start_loan_application_workflow"),
	)

	logger.Info("Starting loan application workflow",
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

// GetWorkflowStatus gets the status of a workflow
func (s *WorkflowService) GetWorkflowStatus(
	ctx context.Context,
	workflowID string,
) (*workflow.WorkflowStatus, error) {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_workflow_status"),
	)

	logger.Info("Getting workflow status")

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

	logger.Info("Workflow status retrieved",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return status, nil
}

// PauseWorkflow pauses a workflow
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

// TerminateWorkflow terminates a workflow
func (s *WorkflowService) TerminateWorkflow(
	ctx context.Context,
	workflowID string,
	reason string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "terminate_workflow"),
	)

	logger.Info("Terminating workflow", zap.String("reason", reason))

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
