package application

import (
	"context"
	"time"

	"go.uber.org/zap"

	"loan-api/domain"
	"loan-api/infrastructure/workflow"
	"loan-api/pkg/i18n"
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

// ProcessPreQualificationWorkflow processes a pre-qualification request through the workflow engine
func (s *PreQualificationWorkflowService) ProcessPreQualificationWorkflow(
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
		zap.String("employment_status", string(request.EmploymentStatus)),
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

	// Wait for workflow completion or timeout
	result, err := s.waitForWorkflowCompletion(ctx, execution.WorkflowID)
	if err != nil {
		logger.Error("Workflow execution failed", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Pre-qualification workflow execution failed",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow completed successfully",
		zap.String("workflow_id", execution.WorkflowID),
		zap.Bool("qualified", result.Qualified),
	)

	return result, nil
}

// waitForWorkflowCompletion waits for the workflow to complete and returns the result
func (s *PreQualificationWorkflowService) waitForWorkflowCompletion(
	ctx context.Context,
	workflowID string,
) (*domain.PreQualifyResult, error) {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "wait_for_workflow_completion"),
	)

	timeout := 5 * time.Minute
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-time.After(timeout):
			logger.Warn("Workflow execution timed out")
			return nil, &domain.LoanError{
				Code:        domain.LOAN_012,
				Message:     "Pre-qualification workflow timed out",
				Description: "Workflow execution exceeded the maximum allowed time",
				HTTPStatus:  408,
			}

		case <-ticker.C:
			// Check workflow status
			status, err := s.workflowOrchestrator.GetWorkflowStatus(ctx, workflowID)
			if err != nil {
				logger.Error("Failed to get workflow status", zap.Error(err))
				continue
			}

			logger.Debug("Workflow status check",
				zap.String("status", status.Status),
				zap.Int("task_count", len(status.Tasks)),
			)

			// Check if workflow is completed
			if status.Status == "COMPLETED" {
				return s.extractWorkflowResult(status)
			}

			// Check if workflow failed
			if status.Status == "FAILED" {
				logger.Error("Workflow execution failed")
				return nil, &domain.LoanError{
					Code:        domain.LOAN_012,
					Message:     "Pre-qualification workflow failed",
					Description: "Workflow execution encountered an error",
					HTTPStatus:  500,
				}
			}

			// Check if we've been waiting too long
			if time.Since(startTime) > timeout {
				logger.Warn("Workflow execution exceeded timeout")
				return nil, &domain.LoanError{
					Code:        domain.LOAN_012,
					Message:     "Pre-qualification workflow timed out",
					Description: "Workflow execution exceeded the maximum allowed time",
					HTTPStatus:  408,
				}
			}
		}
	}
}

// extractWorkflowResult extracts the pre-qualification result from workflow output
func (s *PreQualificationWorkflowService) extractWorkflowResult(
	status *workflow.WorkflowStatus,
) (*domain.PreQualifyResult, error) {
	logger := s.logger.With(
		zap.String("workflow_id", status.WorkflowID),
		zap.String("operation", "extract_workflow_result"),
	)

	// Look for the finalize_prequalification task output
	var finalizeTaskOutput map[string]interface{}
	for _, task := range status.Tasks {
		if task.ReferenceTaskName == "finalize_prequalification" && task.Status == "COMPLETED" {
			finalizeTaskOutput = task.Output
			break
		}
	}

	if finalizeTaskOutput == nil {
		logger.Error("Finalize prequalification task output not found")
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Workflow result extraction failed",
			Description: "Could not find finalize prequalification task output",
			HTTPStatus:  500,
		}
	}

	// Extract result fields
	result := &domain.PreQualifyResult{}

	// Extract qualified status
	if qualified, ok := finalizeTaskOutput["qualified"].(bool); ok {
		result.Qualified = qualified
	}

	// Extract max loan amount
	if maxAmount, ok := finalizeTaskOutput["maxLoanAmount"].(float64); ok {
		result.MaxLoanAmount = maxAmount
	}

	// Extract interest rate range
	if minRate, ok := finalizeTaskOutput["minInterestRate"].(float64); ok {
		result.MinInterestRate = minRate
	}
	if maxRate, ok := finalizeTaskOutput["maxInterestRate"].(float64); ok {
		result.MaxInterestRate = maxRate
	}

	// Extract recommended terms
	if terms, ok := finalizeTaskOutput["recommendedTerms"].([]interface{}); ok {
		result.RecommendedTerms = make([]int, len(terms))
		for i, term := range terms {
			if termInt, ok := term.(int); ok {
				result.RecommendedTerms[i] = termInt
			}
		}
	}

	// Extract DTI ratio
	if dtiRatio, ok := finalizeTaskOutput["dtiRatio"].(float64); ok {
		result.DTIRatio = dtiRatio
	}

	// Extract message
	if message, ok := finalizeTaskOutput["message"].(string); ok {
		result.Message = message
	}

	logger.Info("Workflow result extracted successfully",
		zap.Bool("qualified", result.Qualified),
		zap.Float64("max_loan_amount", result.MaxLoanAmount),
	)

	return result, nil
}

// GetWorkflowStatus retrieves the status of a pre-qualification workflow
func (s *PreQualificationWorkflowService) GetWorkflowStatus(
	ctx context.Context,
	workflowID string,
) (*workflow.WorkflowStatus, error) {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_prequalification_workflow_status"),
	)

	logger.Debug("Retrieving pre-qualification workflow status")

	status, err := s.workflowOrchestrator.GetWorkflowStatus(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to get workflow status", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to get pre-qualification workflow status",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Debug("Pre-qualification workflow status retrieved",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return status, nil
}

// TerminateWorkflow terminates a running pre-qualification workflow
func (s *PreQualificationWorkflowService) TerminateWorkflow(
	ctx context.Context,
	workflowID string,
	reason string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "terminate_prequalification_workflow"),
	)

	logger.Info("Terminating pre-qualification workflow")

	err := s.workflowOrchestrator.TerminateWorkflow(ctx, workflowID, reason)
	if err != nil {
		logger.Error("Failed to terminate workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to terminate pre-qualification workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow terminated successfully")
	return nil
}

// PauseWorkflow pauses a running pre-qualification workflow
func (s *PreQualificationWorkflowService) PauseWorkflow(
	ctx context.Context,
	workflowID string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "pause_prequalification_workflow"),
	)

	logger.Info("Pausing pre-qualification workflow")

	err := s.workflowOrchestrator.PauseWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to pause workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to pause pre-qualification workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow paused successfully")
	return nil
}

// ResumeWorkflow resumes a paused pre-qualification workflow
func (s *PreQualificationWorkflowService) ResumeWorkflow(
	ctx context.Context,
	workflowID string,
) error {
	logger := s.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "resume_prequalification_workflow"),
	)

	logger.Info("Resuming pre-qualification workflow")

	err := s.workflowOrchestrator.ResumeWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to resume workflow", zap.Error(err))
		return &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to resume pre-qualification workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow resumed successfully")
	return nil
}
