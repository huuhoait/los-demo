package workflow

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

// ConductorClient interface for Netflix Conductor workflow engine
type ConductorClient interface {
	StartWorkflow(ctx context.Context, workflowName string, version int, input map[string]interface{}) (*WorkflowExecution, error)
	GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
	TerminateWorkflow(ctx context.Context, workflowID string, reason string) error
	PauseWorkflow(ctx context.Context, workflowID string, reason string) error
	ResumeWorkflow(ctx context.Context, workflowID string) error
	UpdateTask(ctx context.Context, taskID string, workflowInstanceId string, referenceTaskName string, status string, output map[string]interface{}) error
	GetBaseURL() string
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	WorkflowID    string                 `json:"workflowId"`
	Status        string                 `json:"status"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	CorrelationID string                 `json:"correlationId"`
	StartTime     time.Time              `json:"startTime"`
	EndTime       *time.Time             `json:"endTime,omitempty"`
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus struct {
	WorkflowID string                 `json:"workflowId"`
	Status     string                 `json:"status"`
	Tasks      []TaskStatus           `json:"tasks"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
}

// TaskStatus represents the status of a workflow task
type TaskStatus struct {
	TaskID            string                 `json:"taskId"`
	TaskType          string                 `json:"taskType"`
	Status            string                 `json:"status"`
	ReferenceTaskName string                 `json:"referenceTaskName"`
	Input             map[string]interface{} `json:"inputData"`
	Output            map[string]interface{} `json:"outputData"`
	StartTime         time.Time              `json:"startTime"`
	EndTime           *time.Time             `json:"endTime,omitempty"`
}

// LoanWorkflowOrchestrator manages loan processing workflows using Netflix Conductor
type LoanWorkflowOrchestrator struct {
	conductorClient ConductorClient
	logger          *zap.Logger
	localizer       *i18n.Localizer
}

// NewLoanWorkflowOrchestrator creates a new workflow orchestrator
func NewLoanWorkflowOrchestrator(conductorClient ConductorClient, logger *zap.Logger, localizer *i18n.Localizer) *LoanWorkflowOrchestrator {
	return &LoanWorkflowOrchestrator{
		conductorClient: conductorClient,
		logger:          logger,
		localizer:       localizer,
	}
}

// StartLoanProcessingWorkflow starts the main loan processing workflow
func (o *LoanWorkflowOrchestrator) StartLoanProcessingWorkflow(ctx context.Context, application *domain.LoanApplication) (*WorkflowExecution, error) {
	logger := o.logger.With(
		zap.String("application_id", application.ID),
		zap.String("user_id", application.UserID),
		zap.String("operation", "start_loan_workflow"),
	)

	workflowInput := map[string]interface{}{
		"applicationId": application.ID,
		"userId":        application.UserID,
		"loanAmount":    application.LoanAmount,
		"loanPurpose":   application.LoanPurpose,
		"annualIncome":  application.AnnualIncome,
		"monthlyIncome": application.MonthlyIncome,
		"monthlyDebt":   application.MonthlyDebt,
		"requestedTerm": application.RequestedTerm,
		"currentState":  application.CurrentState,
		"startTime":     time.Now().UTC(),
	}

	logger.Info("Starting loan processing workflow",
		zap.Float64("loan_amount", application.LoanAmount),
		zap.String("loan_purpose", string(application.LoanPurpose)),
	)

	execution, err := o.conductorClient.StartWorkflow(ctx, "loan_processing_workflow", 1, workflowInput)
	if err != nil {
		logger.Error("Failed to start loan processing workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Loan processing workflow started successfully",
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("correlation_id", execution.CorrelationID),
	)

	return execution, nil
}

// StartPreQualificationWorkflow starts a pre-qualification workflow
func (o *LoanWorkflowOrchestrator) StartPreQualificationWorkflow(ctx context.Context, userID string, request *domain.PreQualifyRequest) (*WorkflowExecution, error) {
	logger := o.logger.With(
		zap.String("user_id", userID),
		zap.String("operation", "start_prequalification_workflow"),
	)

	workflowInput := map[string]interface{}{
		"userId":           userID,
		"loanAmount":       request.LoanAmount,
		"annualIncome":     request.AnnualIncome,
		"monthlyDebt":      request.MonthlyDebt,
		"employmentStatus": request.EmploymentStatus,
		"startTime":        time.Now().UTC(),
	}

	logger.Info("Starting pre-qualification workflow",
		zap.Float64("loan_amount", request.LoanAmount),
		zap.Float64("annual_income", request.AnnualIncome),
	)

	execution, err := o.conductorClient.StartWorkflow(ctx, "prequalification_workflow", 1, workflowInput)
	if err != nil {
		logger.Error("Failed to start pre-qualification workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start pre-qualification workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Pre-qualification workflow started successfully",
		zap.String("workflow_id", execution.WorkflowID),
	)

	return execution, nil
}

// StartUnderwritingWorkflow starts the underwriting workflow
func (o *LoanWorkflowOrchestrator) StartUnderwritingWorkflow(ctx context.Context, application *domain.LoanApplication) (*WorkflowExecution, error) {
	logger := o.logger.With(
		zap.String("application_id", application.ID),
		zap.String("operation", "start_underwriting_workflow"),
	)

	workflowInput := map[string]interface{}{
		"applicationId": application.ID,
		"userId":        application.UserID,
		"loanAmount":    application.LoanAmount,
		"annualIncome":  application.AnnualIncome,
		"monthlyIncome": application.MonthlyIncome,
		"monthlyDebt":   application.MonthlyDebt,
		"dtiRatio":      application.CalculateDTI(),
		"riskScore":     application.RiskScore,
		"startTime":     time.Now().UTC(),
	}

	logger.Info("Starting underwriting workflow")

	execution, err := o.conductorClient.StartWorkflow(ctx, "underwriting_workflow", 1, workflowInput)
	if err != nil {
		logger.Error("Failed to start underwriting workflow", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_011,
			Message:     "Failed to start underwriting workflow",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Info("Underwriting workflow started successfully",
		zap.String("workflow_id", execution.WorkflowID),
	)

	return execution, nil
}

// HandleStateTransition handles state transitions triggered by workflow events
func (o *LoanWorkflowOrchestrator) HandleStateTransition(ctx context.Context, applicationID string, fromState, toState domain.ApplicationState) error {
	logger := o.logger.With(
		zap.String("application_id", applicationID),
		zap.String("from_state", string(fromState)),
		zap.String("to_state", string(toState)),
		zap.String("operation", "handle_state_transition"),
	)

	// Define workflow actions based on state transitions
	workflowActions := map[string]map[string]string{
		string(domain.StateInitiated): {
			string(domain.StatePreQualified): "prequalification_task",
		},
		string(domain.StatePreQualified): {
			string(domain.StateDocumentsSubmitted): "document_collection_task",
		},
		string(domain.StateDocumentsSubmitted): {
			string(domain.StateIdentityVerified): "identity_verification_task",
		},
		string(domain.StateIdentityVerified): {
			string(domain.StateUnderwriting): "automated_underwriting_task",
		},
		string(domain.StateUnderwriting): {
			string(domain.StateApproved):     "approval_task",
			string(domain.StateDenied):       "denial_task",
			string(domain.StateManualReview): "manual_review_task",
		},
		string(domain.StateManualReview): {
			string(domain.StateApproved): "manual_approval_task",
			string(domain.StateDenied):   "manual_denial_task",
		},
		string(domain.StateApproved): {
			string(domain.StateDocumentsSigned): "document_signing_task",
		},
		string(domain.StateDocumentsSigned): {
			string(domain.StateFunded): "funding_task",
		},
		string(domain.StateFunded): {
			string(domain.StateActive): "loan_activation_task",
		},
	}

	// Get the task to trigger for this transition
	fromStateActions, exists := workflowActions[string(fromState)]
	if !exists {
		logger.Warn("No workflow actions defined for source state")
		return nil
	}

	taskName, exists := fromStateActions[string(toState)]
	if !exists {
		logger.Warn("No workflow action defined for state transition")
		return nil
	}

	logger.Info("Triggering workflow task for state transition",
		zap.String("task_name", taskName),
	)

	// Here you would typically trigger a specific workflow task or signal
	// For now, we'll log the action that should be taken
	logger.Info("Workflow action would be triggered",
		zap.String("action", taskName),
	)

	return nil
}

// GetWorkflowStatus gets the current status of a workflow
func (o *LoanWorkflowOrchestrator) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error) {
	logger := o.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_workflow_status"),
	)

	status, err := o.conductorClient.GetWorkflowStatus(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to get workflow status", zap.Error(err))
		return nil, &domain.LoanError{
			Code:        domain.LOAN_012,
			Message:     "Failed to get workflow status",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Debug("Retrieved workflow status",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return status, nil
}

// TerminateWorkflow terminates a running workflow
func (o *LoanWorkflowOrchestrator) TerminateWorkflow(ctx context.Context, workflowID string, reason string) error {
	logger := o.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "terminate_workflow"),
	)

	err := o.conductorClient.TerminateWorkflow(ctx, workflowID, reason)
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
func (o *LoanWorkflowOrchestrator) PauseWorkflow(ctx context.Context, workflowID string) error {
	logger := o.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "pause_workflow"),
	)

	err := o.conductorClient.PauseWorkflow(ctx, workflowID, "Paused by user request")
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
func (o *LoanWorkflowOrchestrator) ResumeWorkflow(ctx context.Context, workflowID string) error {
	logger := o.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "resume_workflow"),
	)

	err := o.conductorClient.ResumeWorkflow(ctx, workflowID)
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
