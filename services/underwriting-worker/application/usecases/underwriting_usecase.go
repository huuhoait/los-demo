package usecases

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting-worker/domain"
)

// UnderwritingUseCase implements the main underwriting business logic
type UnderwritingUseCase struct {
	logger                    *zap.Logger
	loanApplicationRepo       domain.LoanApplicationRepository
	creditReportRepo          domain.CreditReportRepository
	riskAssessmentRepo        domain.RiskAssessmentRepository
	incomeVerificationRepo    domain.IncomeVerificationRepository
	underwritingResultRepo    domain.UnderwritingResultRepository
	underwritingPolicyRepo    domain.UnderwritingPolicyRepository
	underwritingWorkflowRepo  domain.UnderwritingWorkflowRepository
	creditBureauService       domain.CreditBureauService
	riskScoringService        domain.RiskScoringService
	incomeVerificationService domain.IncomeVerificationService
	decisionEngineService     domain.DecisionEngineService
	workflowOrchestrator      domain.WorkflowOrchestrator
	notificationService       domain.NotificationService
	auditLogger               domain.AuditLogger
}

// NewUnderwritingUseCase creates a new underwriting use case
func NewUnderwritingUseCase(
	logger *zap.Logger,
	loanApplicationRepo domain.LoanApplicationRepository,
	creditReportRepo domain.CreditReportRepository,
	riskAssessmentRepo domain.RiskAssessmentRepository,
	incomeVerificationRepo domain.IncomeVerificationRepository,
	underwritingResultRepo domain.UnderwritingResultRepository,
	underwritingPolicyRepo domain.UnderwritingPolicyRepository,
	underwritingWorkflowRepo domain.UnderwritingWorkflowRepository,
	creditBureauService domain.CreditBureauService,
	riskScoringService domain.RiskScoringService,
	incomeVerificationService domain.IncomeVerificationService,
	decisionEngineService domain.DecisionEngineService,
	workflowOrchestrator domain.WorkflowOrchestrator,
	notificationService domain.NotificationService,
	auditLogger domain.AuditLogger,
) *UnderwritingUseCase {
	return &UnderwritingUseCase{
		logger:                    logger,
		loanApplicationRepo:       loanApplicationRepo,
		creditReportRepo:          creditReportRepo,
		riskAssessmentRepo:        riskAssessmentRepo,
		incomeVerificationRepo:    incomeVerificationRepo,
		underwritingResultRepo:    underwritingResultRepo,
		underwritingPolicyRepo:    underwritingPolicyRepo,
		underwritingWorkflowRepo:  underwritingWorkflowRepo,
		creditBureauService:       creditBureauService,
		riskScoringService:        riskScoringService,
		incomeVerificationService: incomeVerificationService,
		decisionEngineService:     decisionEngineService,
		workflowOrchestrator:      workflowOrchestrator,
		notificationService:       notificationService,
		auditLogger:               auditLogger,
	}
}

// ProcessUnderwritingRequest processes a complete underwriting request
func (uc *UnderwritingUseCase) ProcessUnderwritingRequest(ctx context.Context, applicationID string) (*domain.UnderwritingResult, error) {
	startTime := time.Now()
	logger := uc.logger.With(
		zap.String("application_id", applicationID),
		zap.String("operation", "process_underwriting_request"),
	)

	logger.Info("Starting underwriting process")

	// 1. Get loan application
	application, err := uc.loanApplicationRepo.GetByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get loan application", zap.Error(err))
		return nil, fmt.Errorf("failed to get loan application: %w", err)
	}

	// 2. Validate application
	validation := application.Validate()
	if !validation.Valid {
		logger.Error("Application validation failed", zap.Any("errors", validation.Errors))
		return nil, domain.NewUnderwritingError(
			domain.ErrCodeInvalidApplication,
			"Application validation failed",
			fmt.Sprintf("Validation errors: %v", validation.Errors),
			400,
		)
	}

	// 3. Get or create workflow
	workflow, err := uc.getOrCreateWorkflow(ctx, application)
	if err != nil {
		logger.Error("Failed to get or create workflow", zap.Error(err))
		return nil, fmt.Errorf("failed to get or create workflow: %w", err)
	}

	// 4. Get active underwriting policy
	policy, err := uc.underwritingPolicyRepo.GetActive(ctx)
	if err != nil {
		logger.Error("Failed to get active underwriting policy", zap.Error(err))
		return nil, fmt.Errorf("failed to get active underwriting policy: %w", err)
	}

	// 5. Perform underwriting steps
	result, err := uc.performUnderwritingSteps(ctx, application, policy, workflow)
	if err != nil {
		logger.Error("Underwriting process failed", zap.Error(err))

		// Update workflow with failure
		uc.updateWorkflowStatus(ctx, workflow.ID, domain.StatusCancelled, err.Error())

		return nil, fmt.Errorf("underwriting process failed: %w", err)
	}

	// 6. Save underwriting result
	result.ProcessingTime = time.Since(startTime)
	if err := uc.underwritingResultRepo.Create(ctx, result); err != nil {
		logger.Error("Failed to save underwriting result", zap.Error(err))
		return nil, fmt.Errorf("failed to save underwriting result: %w", err)
	}

	// 7. Update workflow completion
	uc.updateWorkflowStatus(ctx, workflow.ID, domain.StatusCompleted, "")

	// 8. Send notifications
	uc.sendDecisionNotification(ctx, application, result)

	// 9. Log audit event
	uc.logUnderwritingEvent(ctx, application, result)

	logger.Info("Underwriting process completed",
		zap.String("decision", string(result.Decision)),
		zap.Float64("approved_amount", result.ApprovedAmount),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// performUnderwritingSteps performs the main underwriting steps
func (uc *UnderwritingUseCase) performUnderwritingSteps(
	ctx context.Context,
	application *domain.LoanApplication,
	policy *domain.UnderwritingPolicy,
	workflow *domain.UnderwritingWorkflow,
) (*domain.UnderwritingResult, error) {
	logger := uc.logger.With(zap.String("application_id", application.ID))

	// Step 1: Credit Check
	logger.Info("Performing credit check")
	creditReport, err := uc.performCreditCheck(ctx, application)
	if err != nil {
		return nil, fmt.Errorf("credit check failed: %w", err)
	}

	// Step 2: Income Verification
	logger.Info("Performing income verification")
	incomeVerification, err := uc.performIncomeVerification(ctx, application)
	if err != nil {
		return nil, fmt.Errorf("income verification failed: %w", err)
	}

	// Step 3: Risk Assessment
	logger.Info("Performing risk assessment")
	riskAssessment, err := uc.performRiskAssessment(ctx, application, creditReport)
	if err != nil {
		return nil, fmt.Errorf("risk assessment failed: %w", err)
	}

	// Step 4: Policy Compliance Check
	logger.Info("Checking policy compliance")
	policyResult, err := uc.checkPolicyCompliance(ctx, application, policy, creditReport, riskAssessment)
	if err != nil {
		return nil, fmt.Errorf("policy compliance check failed: %w", err)
	}

	// Step 5: Make Underwriting Decision
	logger.Info("Making underwriting decision")
	decision, err := uc.makeUnderwritingDecision(ctx, application, creditReport, riskAssessment, incomeVerification, policy, policyResult)
	if err != nil {
		return nil, fmt.Errorf("underwriting decision failed: %w", err)
	}

	return decision, nil
}

// performCreditCheck performs credit check and returns credit report
func (uc *UnderwritingUseCase) performCreditCheck(ctx context.Context, application *domain.LoanApplication) (*domain.CreditReport, error) {
	// Check if we already have a recent credit report
	existingReport, err := uc.creditReportRepo.GetByApplicationID(ctx, application.ID)
	if err == nil && time.Since(existingReport.ReportDate) < 30*24*time.Hour {
		uc.logger.Info("Using existing credit report",
			zap.String("report_id", existingReport.ID),
			zap.Time("report_date", existingReport.ReportDate))
		return existingReport, nil
	}

	// Request new credit report
	request := &domain.CreditReportRequest{
		UserID:        application.UserID,
		ApplicationID: application.ID,
		SSN:           "", // Would be retrieved from user profile
		ReportType:    "full",
		Permissible:   "loan_application",
	}

	creditReport, err := uc.creditBureauService.GetCreditReport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit report: %w", err)
	}

	// Save credit report
	if err := uc.creditReportRepo.Create(ctx, creditReport); err != nil {
		uc.logger.Error("Failed to save credit report", zap.Error(err))
		// Don't fail the process, just log the error
	}

	return creditReport, nil
}

// performIncomeVerification performs income verification
func (uc *UnderwritingUseCase) performIncomeVerification(ctx context.Context, application *domain.LoanApplication) (*domain.IncomeVerification, error) {
	// Check if we already have income verification
	existingVerification, err := uc.incomeVerificationRepo.GetByApplicationID(ctx, application.ID)
	if err == nil && existingVerification.VerificationStatus == domain.IncomeVerified {
		uc.logger.Info("Using existing income verification",
			zap.String("verification_id", existingVerification.ID))
		return existingVerification, nil
	}

	// Request income verification
	request := &domain.IncomeVerificationRequest{
		UserID:             application.UserID,
		ApplicationID:      application.ID,
		AnnualSalary:       application.AnnualIncome,
		VerificationMethod: "automated", // Could be enhanced to support multiple methods
	}

	verification, err := uc.incomeVerificationService.VerifyIncome(ctx, request)
	if err != nil {
		// If automated verification fails, mark for manual review
		verification = &domain.IncomeVerification{
			ID:                   application.ID + "_income",
			ApplicationID:        application.ID,
			UserID:               application.UserID,
			VerificationMethod:   "manual_review_required",
			VerificationStatus:   domain.IncomeUnverified,
			VerifiedAnnualIncome: 0,
			VerificationNotes:    fmt.Sprintf("Automated verification failed: %v", err),
			CreatedAt:            time.Now(),
		}
	}

	// Save income verification
	if err := uc.incomeVerificationRepo.Create(ctx, verification); err != nil {
		uc.logger.Error("Failed to save income verification", zap.Error(err))
	}

	return verification, nil
}

// performRiskAssessment performs risk assessment
func (uc *UnderwritingUseCase) performRiskAssessment(ctx context.Context, application *domain.LoanApplication, creditReport *domain.CreditReport) (*domain.RiskAssessment, error) {
	// Check if we already have a recent risk assessment
	existingAssessment, err := uc.riskAssessmentRepo.GetByApplicationID(ctx, application.ID)
	if err == nil && time.Since(existingAssessment.CreatedAt) < 24*time.Hour {
		uc.logger.Info("Using existing risk assessment",
			zap.String("assessment_id", existingAssessment.ID))
		return existingAssessment, nil
	}

	// Calculate risk assessment
	riskAssessment, err := uc.riskScoringService.CalculateRiskScore(ctx, application, creditReport)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate risk score: %w", err)
	}

	// Save risk assessment
	if err := uc.riskAssessmentRepo.Create(ctx, riskAssessment); err != nil {
		uc.logger.Error("Failed to save risk assessment", zap.Error(err))
	}

	return riskAssessment, nil
}

// checkPolicyCompliance checks compliance with underwriting policy
func (uc *UnderwritingUseCase) checkPolicyCompliance(
	ctx context.Context,
	application *domain.LoanApplication,
	policy *domain.UnderwritingPolicy,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
) (*domain.PolicyResult, error) {
	return uc.decisionEngineService.ApplyPolicy(ctx, application, policy)
}

// makeUnderwritingDecision makes the final underwriting decision
func (uc *UnderwritingUseCase) makeUnderwritingDecision(
	ctx context.Context,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	incomeVerification *domain.IncomeVerification,
	policy *domain.UnderwritingPolicy,
	policyResult *domain.PolicyResult,
) (*domain.UnderwritingResult, error) {
	// Prepare decision request
	request := &domain.DecisionRequest{
		ApplicationID:      application.ID,
		LoanApplication:    application,
		CreditReport:       creditReport,
		RiskAssessment:     riskAssessment,
		IncomeVerification: incomeVerification,
		Policy:             policy,
		RequestedAmount:    application.LoanAmount,
		RequestedTerm:      application.RequestedTerm,
		Purpose:            application.LoanPurpose,
	}

	// Get decision from decision engine
	decisionResponse, err := uc.decisionEngineService.MakeDecision(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("decision engine failed: %w", err)
	}

	// Create underwriting result
	result := &domain.UnderwritingResult{
		ID:                   application.ID + "_result",
		ApplicationID:        application.ID,
		UserID:               application.UserID,
		Decision:             decisionResponse.Decision,
		Status:               domain.StatusCompleted,
		ApprovedAmount:       decisionResponse.ApprovedAmount,
		ApprovedTerm:         decisionResponse.ApprovedTerm,
		InterestRate:         decisionResponse.InterestRate,
		APR:                  decisionResponse.APR,
		MonthlyPayment:       decisionResponse.MonthlyPayment,
		Conditions:           decisionResponse.Conditions,
		DecisionReasons:      decisionResponse.Reasons,
		CounterOfferTerms:    decisionResponse.CounterOffer,
		AutomatedDecision:    !decisionResponse.ManualReviewRequired,
		ManualReviewRequired: decisionResponse.ManualReviewRequired,
		PolicyVersion:        decisionResponse.PolicyVersion,
		ModelVersion:         uc.riskScoringService.GetModelVersion(),
		OfferExpirationDate:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		DecisionData:         decisionResponse.DecisionData,
		ProcessingTime:       decisionResponse.ProcessingTime,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Calculate additional financial details
	uc.calculateFinancialDetails(result)

	return result, nil
}

// calculateFinancialDetails calculates additional financial details
func (uc *UnderwritingUseCase) calculateFinancialDetails(result *domain.UnderwritingResult) {
	if result.ApprovedAmount > 0 && result.ApprovedTerm > 0 && result.InterestRate > 0 {
		// Calculate monthly payment if not already set
		if result.MonthlyPayment == 0 {
			monthlyRate := result.InterestRate / 12 / 100
			termMonths := float64(result.ApprovedTerm)
			principal := result.ApprovedAmount

			if monthlyRate > 0 {
				result.MonthlyPayment = principal * (monthlyRate * func() float64 {
					return 1 / (1 - func() float64 {
						base := 1 + monthlyRate
						power := -termMonths
						result := 1.0
						for i := 0; i < int(-power); i++ {
							result /= base
						}
						return result
					}())
				}())
			} else {
				result.MonthlyPayment = principal / termMonths
			}
		}

		// Calculate total interest and payment
		result.TotalPayment = result.MonthlyPayment * float64(result.ApprovedTerm)
		result.TotalInterest = result.TotalPayment - result.ApprovedAmount
	}
}

// getOrCreateWorkflow gets existing workflow or creates a new one
func (uc *UnderwritingUseCase) getOrCreateWorkflow(ctx context.Context, application *domain.LoanApplication) (*domain.UnderwritingWorkflow, error) {
	// Try to get existing workflow
	workflow, err := uc.underwritingWorkflowRepo.GetByApplicationID(ctx, application.ID)
	if err == nil {
		return workflow, nil
	}

	// Create new workflow
	workflow = &domain.UnderwritingWorkflow{
		ID:                  application.ID + "_workflow",
		ApplicationID:       application.ID,
		WorkflowID:          application.ID + "_conductor_workflow",
		CurrentStep:         "credit_check",
		Status:              domain.StatusInProgress,
		StartedAt:           time.Now(),
		EstimatedCompletion: time.Now().Add(2 * time.Hour),
		WorkflowData:        make(map[string]interface{}),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := uc.underwritingWorkflowRepo.Create(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// updateWorkflowStatus updates workflow status
func (uc *UnderwritingUseCase) updateWorkflowStatus(ctx context.Context, workflowID string, status domain.UnderwritingStatus, errorMessage string) {
	workflow, err := uc.underwritingWorkflowRepo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		uc.logger.Error("Failed to get workflow for status update", zap.Error(err))
		return
	}

	workflow.Status = status
	workflow.UpdatedAt = time.Now()

	if status == domain.StatusCompleted {
		now := time.Now()
		workflow.CompletedAt = &now
	}

	if errorMessage != "" {
		if workflow.WorkflowData == nil {
			workflow.WorkflowData = make(map[string]interface{})
		}
		workflow.WorkflowData["error_message"] = errorMessage
	}

	if err := uc.underwritingWorkflowRepo.Update(ctx, workflow); err != nil {
		uc.logger.Error("Failed to update workflow status", zap.Error(err))
	}
}

// sendDecisionNotification sends decision notification to user
func (uc *UnderwritingUseCase) sendDecisionNotification(ctx context.Context, application *domain.LoanApplication, result *domain.UnderwritingResult) {
	notification := &domain.DecisionNotification{
		ApplicationID:   application.ID,
		UserID:          application.UserID,
		Decision:        result.Decision,
		DecisionReasons: result.DecisionReasons,
		ApprovedAmount:  result.ApprovedAmount,
		InterestRate:    result.InterestRate,
		MonthlyPayment:  result.MonthlyPayment,
		Conditions:      result.Conditions,
		CounterOffer:    result.CounterOfferTerms,
		ExpirationDate:  result.OfferExpirationDate,
	}

	if err := uc.notificationService.SendDecisionNotification(ctx, notification); err != nil {
		uc.logger.Error("Failed to send decision notification", zap.Error(err))
	}
}

// logUnderwritingEvent logs underwriting event for audit
func (uc *UnderwritingUseCase) logUnderwritingEvent(ctx context.Context, application *domain.LoanApplication, result *domain.UnderwritingResult) {
	event := &domain.UnderwritingEvent{
		EventID:       result.ID + "_event",
		ApplicationID: application.ID,
		UserID:        application.UserID,
		EventType:     "underwriting_completed",
		EventData: map[string]interface{}{
			"decision":        result.Decision,
			"approved_amount": result.ApprovedAmount,
			"interest_rate":   result.InterestRate,
			"automated":       result.AutomatedDecision,
			"processing_time": result.ProcessingTime.String(),
		},
		Timestamp:     time.Now(),
		UnderwriterID: result.UnderwriterID,
	}

	if err := uc.auditLogger.LogUnderwritingEvent(ctx, event); err != nil {
		uc.logger.Error("Failed to log underwriting event", zap.Error(err))
	}
}

// GetUnderwritingStatus gets the current status of underwriting for an application
func (uc *UnderwritingUseCase) GetUnderwritingStatus(ctx context.Context, applicationID string) (*domain.UnderwritingWorkflow, error) {
	return uc.underwritingWorkflowRepo.GetByApplicationID(ctx, applicationID)
}

// GetUnderwritingResult gets the underwriting result for an application
func (uc *UnderwritingUseCase) GetUnderwritingResult(ctx context.Context, applicationID string) (*domain.UnderwritingResult, error) {
	return uc.underwritingResultRepo.GetByApplicationID(ctx, applicationID)
}

// ReprocessUnderwriting reprocesses underwriting for an application
func (uc *UnderwritingUseCase) ReprocessUnderwriting(ctx context.Context, applicationID string, reason string) (*domain.UnderwritingResult, error) {
	uc.logger.Info("Reprocessing underwriting",
		zap.String("application_id", applicationID),
		zap.String("reason", reason))

	// Mark existing result as superseded if it exists
	if existingResult, err := uc.underwritingResultRepo.GetByApplicationID(ctx, applicationID); err == nil {
		existingResult.Status = domain.StatusCancelled
		existingResult.UpdatedAt = time.Now()
		uc.underwritingResultRepo.Update(ctx, existingResult)
	}

	// Process new underwriting request
	return uc.ProcessUnderwritingRequest(ctx, applicationID)
}
