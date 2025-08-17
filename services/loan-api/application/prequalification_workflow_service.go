package application

import (
	"context"

	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/loan-api/infrastructure/workflow"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

// PreQualificationWorkflowService manages pre-qualification workflow operations
type PreQualificationWorkflowService struct {
	workflowOrchestrator *workflow.LoanWorkflowOrchestrator
	logger               *zap.Logger
	localizer            *i18n.Localizer
}

// NewPreQualificationWorkflowService creates a new pre-qualification workflow service
func NewPreQualificationWorkflowService(
	workflowOrchestrator *workflow.LoanWorkflowOrchestrator,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *PreQualificationWorkflowService {
	return &PreQualificationWorkflowService{
		workflowOrchestrator: workflowOrchestrator,
		logger:               logger,
		localizer:            localizer,
	}
}

// StartPreQualificationWorkflow starts a pre-qualification workflow
func (s *PreQualificationWorkflowService) StartPreQualificationWorkflow(
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

// GetPreQualificationStatus gets the status of a pre-qualification workflow
func (s *PreQualificationWorkflowService) GetPreQualificationStatus(
	ctx context.Context,
	workflowID string,
) (*workflow.WorkflowStatus, error) {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_prequalification_status"),
	)

	logger.Info("Getting pre-qualification workflow status")

	status, err := s.workflowOrchestrator.GetWorkflowStatus(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to get pre-qualification workflow status", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to get pre-qualification workflow status",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow status retrieved",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return status, nil
}
