package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting_worker/pkg/config"
)

// UnderwritingTaskWorker handles all underwriting-related workflow tasks
type UnderwritingTaskWorker struct {
	logger                        *zap.Logger
	config                        *config.Config
	conductorClient               *HTTPConductorClient
	mockConductorClient           *MockConductorClient
	useMockConductor              bool
	creditCheckHandler            *CreditCheckTaskHandler
	incomeVerificationHandler     *IncomeVerificationTaskHandler
	riskAssessmentHandler         *RiskAssessmentTaskHandler
	underwritingDecisionHandler   *UnderwritingDecisionTaskHandler
	updateApplicationStateHandler *UpdateApplicationStateTaskHandler
}

// NewUnderwritingTaskWorker creates a new underwriting task worker
func NewUnderwritingTaskWorker(logger *zap.Logger, cfg *config.Config) *UnderwritingTaskWorker {
	// Try to initialize real HTTP Conductor client first
	httpConductorClient, err := NewHTTPConductorClient(logger, cfg)
	var mockConductorClient *MockConductorClient
	useMockConductor := false

	if err != nil {
		logger.Warn("Failed to initialize HTTP Conductor client, falling back to mock", zap.Error(err))
		// Fallback to mock conductor
		mockConductorClient = NewMockConductorClient(logger, cfg.Conductor.WorkerPoolSize)
		useMockConductor = true
	}

	worker := &UnderwritingTaskWorker{
		logger:              logger,
		config:              cfg,
		conductorClient:     httpConductorClient,
		mockConductorClient: mockConductorClient,
		useMockConductor:    useMockConductor,
	}

	// Initialize task handlers
	worker.initializeTaskHandlers()

	return worker
}

// initializeTaskHandlers initializes all task handlers
func (w *UnderwritingTaskWorker) initializeTaskHandlers() {
	// Note: In a real implementation, these handlers would be initialized with actual repositories and services
	// For this example, we'll create mock implementations

	w.logger.Info("Initializing underwriting task handlers")

	// Initialize handlers with mock dependencies
	// In a real implementation, these would be properly injected
	w.creditCheckHandler = NewCreditCheckTaskHandler(
		w.logger.With(zap.String("handler", "credit_check")),
		nil, // creditService - would be injected
		nil, // underwritingUseCase - would be injected
		nil, // loanApplicationRepo - would be injected
		nil, // creditReportRepo - would be injected
	)

	w.incomeVerificationHandler = NewIncomeVerificationTaskHandler(
		w.logger.With(zap.String("handler", "income_verification")),
		nil, // underwritingUseCase - would be injected
		nil, // loanApplicationRepo - would be injected
		nil, // incomeVerificationRepo - would be injected
		nil, // incomeVerificationService - would be injected
	)

	w.riskAssessmentHandler = NewRiskAssessmentTaskHandler(
		w.logger.With(zap.String("handler", "risk_assessment")),
		nil, // underwritingUseCase - would be injected
		nil, // loanApplicationRepo - would be injected
		nil, // creditReportRepo - would be injected
		nil, // riskAssessmentRepo - would be injected
		nil, // riskScoringService - would be injected
	)

	w.underwritingDecisionHandler = NewUnderwritingDecisionTaskHandler(
		w.logger.With(zap.String("handler", "underwriting_decision")),
		nil, // underwritingUseCase - would be injected
		nil, // loanApplicationRepo - would be injected
		nil, // creditReportRepo - would be injected
		nil, // riskAssessmentRepo - would be injected
		nil, // incomeVerificationRepo - would be injected
		nil, // underwritingResultRepo - would be injected
		nil, // underwritingPolicyRepo - would be injected
		nil, // decisionEngineService - would be injected
	)

	w.updateApplicationStateHandler = NewUpdateApplicationStateTaskHandler(
		w.logger.With(zap.String("handler", "update_application_state")),
		nil, // loanApplicationRepo - would be injected
	)

	w.logger.Info("All underwriting task handlers initialized successfully")
}

// Start starts the task worker
func (w *UnderwritingTaskWorker) Start(ctx context.Context) error {
	clientType := "real Conductor"
	if w.useMockConductor {
		clientType = "mock Conductor"
	}

	w.logger.Info("Starting underwriting task worker",
		zap.String("conductor_url", w.config.Conductor.ServerURL),
		zap.String("client_type", clientType),
		zap.Int("worker_pool_size", w.config.Conductor.WorkerPoolSize),
		zap.Int("polling_interval_ms", w.config.Conductor.PollingInterval))

	// Register underwriting workflow tasks
	w.registerUnderwritingTasks()

	// Register workflow and task definitions with real Conductor
	if !w.useMockConductor {
		if err := w.registerWorkflowDefinitions(); err != nil {
			w.logger.Error("Failed to register workflow definitions", zap.Error(err))
			// Don't fail startup, just log the error
		}
	}

	// Start the appropriate conductor client
	if w.useMockConductor {
		return w.mockConductorClient.StartPolling()
	}
	return w.conductorClient.StartPolling()
}

// Stop stops the task worker
func (w *UnderwritingTaskWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping underwriting task worker")

	if w.useMockConductor {
		w.mockConductorClient.StopPolling()
	} else {
		w.conductorClient.StopPolling()
	}

	return nil
}

// registerUnderwritingTasks registers all underwriting-related tasks
func (w *UnderwritingTaskWorker) registerUnderwritingTasks() {
	w.logger.Info("Registering underwriting workflow tasks")

	// Register credit check task
	w.registerWorker("credit_check", w.wrapTaskHandler("credit_check", w.creditCheckHandler.Execute))
	w.logger.Info("Registered task: credit_check")

	// Register income verification task
	w.registerWorker("income_verification", w.wrapTaskHandler("income_verification", w.incomeVerificationHandler.Execute))
	w.logger.Info("Registered task: income_verification")

	// Register risk assessment task
	w.registerWorker("risk_assessment", w.wrapTaskHandler("risk_assessment", w.riskAssessmentHandler.Execute))
	w.logger.Info("Registered task: risk_assessment")

	// Register underwriting decision task
	w.registerWorker("underwriting_decision", w.wrapTaskHandler("underwriting_decision", w.underwritingDecisionHandler.Execute))
	w.logger.Info("Registered task: underwriting_decision")

	// Register application state update task
	w.registerWorker("update_application_state", w.wrapTaskHandler("update_application_state", w.updateApplicationStateHandler.Execute))
	w.logger.Info("Registered task: update_application_state")

	// Register additional underwriting tasks
	w.registerAdditionalUnderwritingTasks()

	w.logger.Info("All underwriting tasks registered successfully")
}

// registerAdditionalUnderwritingTasks registers additional underwriting tasks
func (w *UnderwritingTaskWorker) registerAdditionalUnderwritingTasks() {
	// Register policy compliance check task
	w.registerWorker("policy_compliance_check", w.wrapTaskHandler("policy_compliance_check", w.handlePolicyComplianceCheck))
	w.logger.Info("Registered task: policy_compliance_check")

	// Register fraud detection task
	w.registerWorker("fraud_detection", w.wrapTaskHandler("fraud_detection", w.handleFraudDetection))
	w.logger.Info("Registered task: fraud_detection")

	// Register interest rate calculation task
	w.registerWorker("calculate_interest_rate", w.wrapTaskHandler("calculate_interest_rate", w.handleInterestRateCalculation))
	w.logger.Info("Registered task: calculate_interest_rate")

	// Register final approval task
	w.registerWorker("final_approval", w.wrapTaskHandler("final_approval", w.handleFinalApproval))
	w.logger.Info("Registered task: final_approval")

	// Register denial processing task
	w.registerWorker("process_denial", w.wrapTaskHandler("process_denial", w.handleDenialProcessing))
	w.logger.Info("Registered task: process_denial")

	// Register manual review assignment task
	w.registerWorker("assign_manual_review", w.wrapTaskHandler("assign_manual_review", w.handleManualReviewAssignment))
	w.logger.Info("Registered task: assign_manual_review")

	// Register conditional approval task
	w.registerWorker("process_conditional_approval", w.wrapTaskHandler("process_conditional_approval", w.handleConditionalApproval))
	w.logger.Info("Registered task: process_conditional_approval")

	// Register counter offer task
	w.registerWorker("generate_counter_offer", w.wrapTaskHandler("generate_counter_offer", w.handleCounterOfferGeneration))
	w.logger.Info("Registered task: generate_counter_offer")
}

// wrapTaskHandler wraps a task handler with common logging and error handling
func (w *UnderwritingTaskWorker) wrapTaskHandler(taskName string, handler func(context.Context, map[string]interface{}) (map[string]interface{}, error)) TaskHandler {
	return func(task *MockTask) (*MockTaskResult, error) {
		startTime := time.Now()
		logger := w.logger.With(
			zap.String("task_name", taskName),
			zap.String("task_id", task.TaskID),
			zap.String("workflow_instance_id", task.WorkflowInstanceID),
		)

		logger.Info("Executing underwriting task",
			zap.Any("input_data", task.InputData))

		// Execute the task handler
		ctx := context.Background()
		outputData, err := handler(ctx, task.InputData)

		processingTime := time.Since(startTime)

		if err != nil {
			logger.Error("Task execution failed",
				zap.Error(err),
				zap.Duration("processing_time", processingTime))

			return &MockTaskResult{
				TaskID:                task.TaskID,
				Status:                "FAILED",
				ReasonForIncompletion: err.Error(),
				OutputData: map[string]interface{}{
					"error":           err.Error(),
					"processing_time": processingTime.String(),
					"failed_at":       time.Now().UTC().Format(time.RFC3339),
				},
				WorkerID:      fmt.Sprintf("underwriting-worker-%d", time.Now().Unix()),
				CompletedTime: time.Now(),
			}, nil
		}

		logger.Info("Task execution completed successfully",
			zap.Duration("processing_time", processingTime),
			zap.Any("output_data", outputData))

		return &MockTaskResult{
			TaskID:        task.TaskID,
			Status:        "COMPLETED",
			OutputData:    outputData,
			WorkerID:      fmt.Sprintf("underwriting-worker-%d", time.Now().Unix()),
			CompletedTime: time.Now(),
		}, nil
	}
}

// registerWorker registers a worker with the appropriate client
func (w *UnderwritingTaskWorker) registerWorker(taskType string, handler TaskHandler) {
	if w.useMockConductor {
		w.mockConductorClient.RegisterWorker(taskType, handler)
	} else {
		w.conductorClient.RegisterWorker(taskType, handler)
	}
}

// registerWorkflowDefinitions registers workflow and task definitions with Conductor
func (w *UnderwritingTaskWorker) registerWorkflowDefinitions() error {
	if w.useMockConductor {
		return nil // No need to register definitions with mock
	}

	w.logger.Info("Registering workflow and task definitions with Conductor")

	// Register task definitions
	taskDefs := w.conductorClient.CreateTaskDefinitions()
	successfulRegistrations := 0
	totalTasks := len(taskDefs)

	for _, taskDef := range taskDefs {
		// Convert to HTTP client format
		httpTaskDef := &TaskDefinition{
			Name:                   taskDef.Name,
			Description:            taskDef.Description,
			TimeoutSeconds:         taskDef.TimeoutSeconds,
			ResponseTimeoutSeconds: taskDef.ResponseTimeoutSeconds,
			RetryCount:             taskDef.RetryCount,
			InputKeys:              taskDef.InputKeys,
			OutputKeys:             taskDef.OutputKeys,
		}

		if err := w.conductorClient.RegisterTaskDefinition(httpTaskDef); err != nil {
			w.logger.Error("Failed to register task definition",
				zap.String("task_name", httpTaskDef.Name),
				zap.Error(err))
			// Continue with other tasks but track failures
		} else {
			w.logger.Info("Registered task definition", zap.String("task_name", httpTaskDef.Name))
			successfulRegistrations++
		}
	}

	w.logger.Info("Task definition registration summary",
		zap.Int("successful", successfulRegistrations),
		zap.Int("total", totalTasks),
		zap.Int("failed", totalTasks-successfulRegistrations))

	// Ensure at least the core tasks are registered
	if successfulRegistrations < 3 {
		w.logger.Warn("Very few task definitions registered successfully, this may cause issues")
	}

	// Register workflow definition
	workflowDef := w.conductorClient.CreateUnderwritingWorkflowDefinition()
	if err := w.conductorClient.RegisterWorkflowDefinition(workflowDef); err != nil {
		w.logger.Error("Failed to register workflow definition", zap.Error(err))
		return err
	}

	w.logger.Info("Successfully registered workflow definition",
		zap.String("workflow_name", workflowDef.Name))

	// Add a small delay to ensure definitions are propagated in Conductor
	w.logger.Info("Waiting for task definitions to propagate in Conductor...")
	time.Sleep(2 * time.Second)

	return nil
}

// StartWorkflow starts an underwriting workflow (for testing/manual execution)
func (w *UnderwritingTaskWorker) StartWorkflow(applicationID, userID string) (string, error) {
	if w.useMockConductor {
		// For mock conductor, simulate workflow execution
		w.logger.Info("Simulating workflow start with mock conductor",
			zap.String("application_id", applicationID),
			zap.String("user_id", userID))

		// Submit individual tasks to mock conductor
		tasks := []string{"credit_check", "income_verification", "risk_assessment", "underwriting_decision"}
		for _, taskType := range tasks {
			taskID := w.mockConductorClient.SubmitTask(taskType, map[string]interface{}{
				"applicationId": applicationID,
				"userId":        userID,
			})
			w.logger.Info("Submitted task to mock conductor",
				zap.String("task_id", taskID),
				zap.String("task_type", taskType))
		}

		return fmt.Sprintf("mock-workflow-%s", applicationID), nil
	}

	// Start real workflow
	input := map[string]interface{}{
		"applicationId": applicationID,
		"userId":        userID,
	}

	workflowId, err := w.conductorClient.StartWorkflow("underwriting_workflow", input)
	if err != nil {
		return "", fmt.Errorf("failed to start underwriting workflow: %w", err)
	}

	w.logger.Info("Started underwriting workflow",
		zap.String("workflow_id", workflowId),
		zap.String("application_id", applicationID),
		zap.String("user_id", userID))

	return workflowId, nil
}

// GetWorkflowStatus gets the status of a workflow
func (w *UnderwritingTaskWorker) GetWorkflowStatus(workflowID string) (map[string]interface{}, error) {
	if w.useMockConductor {
		// For mock conductor, return simulated status
		return map[string]interface{}{
			"workflowId": workflowID,
			"status":     "COMPLETED",
			"output": map[string]interface{}{
				"decision":       "APPROVED",
				"approvedAmount": 25000,
				"interestRate":   8.5,
			},
		}, nil
	}

	// Get real workflow status
	workflow, err := w.conductorClient.GetWorkflowStatus(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	return map[string]interface{}{
		"workflowId": workflow["workflowId"],
		"status":     workflow["status"],
		"startTime":  workflow["startTime"],
		"endTime":    workflow["endTime"],
		"output":     workflow["output"],
	}, nil
}

// Additional task handlers

// handlePolicyComplianceCheck handles policy compliance checking
func (w *UnderwritingTaskWorker) handlePolicyComplianceCheck(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "policy_compliance_check"))
	logger.Info("Performing policy compliance check")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	// Mock policy compliance check
	compliant := true
	violations := []string{}

	// Simulate policy checks
	if creditScore, ok := input["creditScore"].(float64); ok && creditScore < 580 {
		compliant = false
		violations = append(violations, "Credit score below minimum threshold")
	}

	if dtiRatio, ok := input["dtiRatio"].(float64); ok && dtiRatio > 0.43 {
		compliant = false
		violations = append(violations, "DTI ratio exceeds maximum allowed")
	}

	logger.Info("Policy compliance check completed",
		zap.String("application_id", applicationID),
		zap.Bool("compliant", compliant),
		zap.Strings("violations", violations))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"compliant":     compliant,
		"violations":    violations,
		"completedAt":   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleFraudDetection handles fraud detection analysis
func (w *UnderwritingTaskWorker) handleFraudDetection(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "fraud_detection"))
	logger.Info("Performing fraud detection analysis")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	// Mock fraud detection
	fraudRiskScore := 15.0 // Low fraud risk
	fraudIndicators := []string{}

	// Simulate fraud indicators
	if loanAmount, ok := input["loanAmount"].(float64); ok && loanAmount > 100000 {
		fraudRiskScore += 10
		fraudIndicators = append(fraudIndicators, "High loan amount requested")
	}

	fraudRiskLevel := "low"
	if fraudRiskScore > 50 {
		fraudRiskLevel = "high"
	} else if fraudRiskScore > 25 {
		fraudRiskLevel = "medium"
	}

	logger.Info("Fraud detection completed",
		zap.String("application_id", applicationID),
		zap.Float64("fraud_risk_score", fraudRiskScore),
		zap.String("fraud_risk_level", fraudRiskLevel))

	return map[string]interface{}{
		"success":         true,
		"applicationId":   applicationID,
		"fraudRiskScore":  fraudRiskScore,
		"fraudRiskLevel":  fraudRiskLevel,
		"fraudIndicators": fraudIndicators,
		"completedAt":     time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleInterestRateCalculation handles interest rate calculation
func (w *UnderwritingTaskWorker) handleInterestRateCalculation(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "calculate_interest_rate"))
	logger.Info("Calculating interest rate")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	// Get input parameters
	creditScore, _ := input["creditScore"].(float64)
	riskLevel, _ := input["riskLevel"].(string)

	// Calculate interest rate based on credit score and risk
	baseRate := 8.0

	switch {
	case creditScore >= 800:
		baseRate = 5.5
	case creditScore >= 740:
		baseRate = 6.5
	case creditScore >= 670:
		baseRate = 8.0
	case creditScore >= 620:
		baseRate = 12.0
	default:
		baseRate = 18.0
	}

	// Risk adjustment
	switch riskLevel {
	case "low":
		baseRate -= 0.5
	case "high":
		baseRate += 2.0
	case "critical":
		baseRate += 5.0
	}

	apr := baseRate + 0.5 // Add margin for APR

	logger.Info("Interest rate calculated",
		zap.String("application_id", applicationID),
		zap.Float64("credit_score", creditScore),
		zap.String("risk_level", riskLevel),
		zap.Float64("interest_rate", baseRate),
		zap.Float64("apr", apr))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"interestRate":  baseRate,
		"apr":           apr,
		"rateFactors": []map[string]interface{}{
			{
				"factor":      "credit_score",
				"impact":      creditScore,
				"description": fmt.Sprintf("Credit score: %.0f", creditScore),
			},
			{
				"factor":      "risk_level",
				"impact":      riskLevel,
				"description": fmt.Sprintf("Risk level: %s", riskLevel),
			},
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleFinalApproval handles final loan approval processing
func (w *UnderwritingTaskWorker) handleFinalApproval(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "final_approval"))
	logger.Info("Processing final approval")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	// Extract approval details
	approvedAmount, _ := input["approvedAmount"].(float64)
	interestRate, _ := input["interestRate"].(float64)
	term, _ := input["approvedTerm"].(float64)

	// Generate loan terms
	offerExpirationDate := time.Now().Add(7 * 24 * time.Hour)
	loanNumber := fmt.Sprintf("UW-%s-%d", applicationID[:8], time.Now().Unix())

	logger.Info("Final approval processed",
		zap.String("application_id", applicationID),
		zap.String("loan_number", loanNumber),
		zap.Float64("approved_amount", approvedAmount),
		zap.Float64("interest_rate", interestRate))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"loanNumber":    loanNumber,
		"approvalDetails": map[string]interface{}{
			"approvedAmount":      approvedAmount,
			"interestRate":        interestRate,
			"term":                term,
			"offerExpirationDate": offerExpirationDate.Format(time.RFC3339),
		},
		"nextSteps": []string{
			"Send approval notification to customer",
			"Generate loan documents",
			"Schedule loan closing",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleDenialProcessing handles loan denial processing
func (w *UnderwritingTaskWorker) handleDenialProcessing(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "process_denial"))
	logger.Info("Processing loan denial")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	denialReasons, _ := input["denialReasons"].([]interface{})

	logger.Info("Denial processed",
		zap.String("application_id", applicationID),
		zap.Any("denial_reasons", denialReasons))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"denialReasons": denialReasons,
		"nextSteps": []string{
			"Send denial notification to customer",
			"Provide adverse action notice",
			"Offer credit counseling resources",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleManualReviewAssignment handles manual review assignment
func (w *UnderwritingTaskWorker) handleManualReviewAssignment(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "assign_manual_review"))
	logger.Info("Assigning manual review")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	// Assign to next available underwriter
	assignedUnderwriter := "underwriter_1" // Mock assignment
	reviewPriority := "normal"

	if riskLevel, ok := input["riskLevel"].(string); ok && riskLevel == "high" {
		reviewPriority = "high"
	}

	dueDate := time.Now().Add(24 * time.Hour) // 1 day for review

	logger.Info("Manual review assigned",
		zap.String("application_id", applicationID),
		zap.String("assigned_underwriter", assignedUnderwriter),
		zap.String("priority", reviewPriority))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"assignedTo":    assignedUnderwriter,
		"priority":      reviewPriority,
		"dueDate":       dueDate.Format(time.RFC3339),
		"reviewInstructions": []string{
			"Review credit history in detail",
			"Verify income documentation",
			"Assess risk factors",
			"Make final underwriting decision",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleConditionalApproval handles conditional approval processing
func (w *UnderwritingTaskWorker) handleConditionalApproval(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "process_conditional_approval"))
	logger.Info("Processing conditional approval")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	conditions, _ := input["conditions"].([]interface{})
	approvedAmount, _ := input["approvedAmount"].(float64)

	logger.Info("Conditional approval processed",
		zap.String("application_id", applicationID),
		zap.Float64("approved_amount", approvedAmount),
		zap.Any("conditions", conditions))

	return map[string]interface{}{
		"success":        true,
		"applicationId":  applicationID,
		"approvedAmount": approvedAmount,
		"conditions":     conditions,
		"nextSteps": []string{
			"Send conditional approval to customer",
			"Track condition fulfillment",
			"Schedule final approval upon completion",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// handleCounterOfferGeneration handles counter offer generation
func (w *UnderwritingTaskWorker) handleCounterOfferGeneration(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	logger := w.logger.With(zap.String("operation", "generate_counter_offer"))
	logger.Info("Generating counter offer")

	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	requestedAmount, _ := input["requestedAmount"].(float64)

	// Generate reduced counter offer
	counterOfferAmount := requestedAmount * 0.75 // 75% of requested
	higherRate := 12.5                           // Higher interest rate

	expirationDate := time.Now().Add(7 * 24 * time.Hour)

	logger.Info("Counter offer generated",
		zap.String("application_id", applicationID),
		zap.Float64("requested_amount", requestedAmount),
		zap.Float64("counter_offer_amount", counterOfferAmount),
		zap.Float64("offered_rate", higherRate))

	return map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"counterOffer": map[string]interface{}{
			"offeredAmount":  counterOfferAmount,
			"offeredRate":    higherRate,
			"offerReason":    "Reduced amount to mitigate risk profile",
			"expirationDate": expirationDate.Format(time.RFC3339),
		},
		"nextSteps": []string{
			"Send counter offer to customer",
			"Await customer response",
			"Process acceptance or counter-negotiation",
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}
