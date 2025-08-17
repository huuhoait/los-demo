package tasks

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"underwriting-worker/application/usecases"
	"underwriting-worker/domain"
)

// UnderwritingDecisionTaskHandler handles final underwriting decision tasks
type UnderwritingDecisionTaskHandler struct {
	logger                 *zap.Logger
	underwritingUseCase    *usecases.UnderwritingUseCase
	loanApplicationRepo    domain.LoanApplicationRepository
	creditReportRepo       domain.CreditReportRepository
	riskAssessmentRepo     domain.RiskAssessmentRepository
	incomeVerificationRepo domain.IncomeVerificationRepository
	underwritingResultRepo domain.UnderwritingResultRepository
	underwritingPolicyRepo domain.UnderwritingPolicyRepository
	decisionEngineService  domain.DecisionEngineService
}

// NewUnderwritingDecisionTaskHandler creates a new underwriting decision task handler
func NewUnderwritingDecisionTaskHandler(
	logger *zap.Logger,
	underwritingUseCase *usecases.UnderwritingUseCase,
	loanApplicationRepo domain.LoanApplicationRepository,
	creditReportRepo domain.CreditReportRepository,
	riskAssessmentRepo domain.RiskAssessmentRepository,
	incomeVerificationRepo domain.IncomeVerificationRepository,
	underwritingResultRepo domain.UnderwritingResultRepository,
	underwritingPolicyRepo domain.UnderwritingPolicyRepository,
	decisionEngineService domain.DecisionEngineService,
) *UnderwritingDecisionTaskHandler {
	return &UnderwritingDecisionTaskHandler{
		logger:                 logger,
		underwritingUseCase:    underwritingUseCase,
		loanApplicationRepo:    loanApplicationRepo,
		creditReportRepo:       creditReportRepo,
		riskAssessmentRepo:     riskAssessmentRepo,
		incomeVerificationRepo: incomeVerificationRepo,
		underwritingResultRepo: underwritingResultRepo,
		underwritingPolicyRepo: underwritingPolicyRepo,
		decisionEngineService:  decisionEngineService,
	}
}

// Execute makes the final underwriting decision
func (h *UnderwritingDecisionTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	logger := h.logger.With(zap.String("operation", "underwriting_decision"))

	logger.Info("Starting underwriting decision task")

	// Extract input parameters
	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	userID, ok := input["userId"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get all required data
	application, creditReport, riskAssessment, incomeVerification, policy, err := h.gatherUnderwritingData(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to gather underwriting data", zap.Error(err))
		return h.createFailureResponse(applicationID, err), nil
	}

	logger.Info("Gathered underwriting data",
		zap.String("application_id", applicationID),
		zap.Int("credit_score", creditReport.CreditScore),
		zap.String("risk_level", string(riskAssessment.OverallRiskLevel)),
		zap.String("income_status", string(incomeVerification.VerificationStatus)))

	// Make underwriting decision
	decision, err := h.makeUnderwritingDecision(ctx, application, creditReport, riskAssessment, incomeVerification, policy)
	if err != nil {
		logger.Error("Failed to make underwriting decision", zap.Error(err))
		return h.createFailureResponse(applicationID, err), nil
	}

	// Save underwriting result
	if err := h.underwritingResultRepo.Create(ctx, decision); err != nil {
		logger.Error("Failed to save underwriting result", zap.Error(err))
		// Don't fail the process, just log the error
	}

	processingTime := time.Since(startTime)
	decision.ProcessingTime = processingTime

	logger.Info("Underwriting decision completed",
		zap.String("application_id", applicationID),
		zap.String("decision", string(decision.Decision)),
		zap.Float64("approved_amount", decision.ApprovedAmount),
		zap.Float64("interest_rate", decision.InterestRate),
		zap.Bool("manual_review", decision.ManualReviewRequired),
		zap.Duration("processing_time", processingTime))

	// Create comprehensive response
	return h.createSuccessResponse(decision, application, creditReport, riskAssessment, incomeVerification), nil
}

// gatherUnderwritingData gathers all required data for decision making
func (h *UnderwritingDecisionTaskHandler) gatherUnderwritingData(ctx context.Context, applicationID string) (
	*domain.LoanApplication,
	*domain.CreditReport,
	*domain.RiskAssessment,
	*domain.IncomeVerification,
	*domain.UnderwritingPolicy,
	error,
) {
	// Get loan application
	application, err := h.loanApplicationRepo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get loan application: %w", err)
	}

	// Get credit report
	creditReport, err := h.creditReportRepo.GetByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get credit report: %w", err)
	}

	// Get risk assessment
	riskAssessment, err := h.riskAssessmentRepo.GetByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get risk assessment: %w", err)
	}

	// Get income verification
	incomeVerification, err := h.incomeVerificationRepo.GetByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get income verification: %w", err)
	}

	// Get active underwriting policy
	policy, err := h.underwritingPolicyRepo.GetActive(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get underwriting policy: %w", err)
	}

	return application, creditReport, riskAssessment, incomeVerification, policy, nil
}

// makeUnderwritingDecision makes the final underwriting decision
func (h *UnderwritingDecisionTaskHandler) makeUnderwritingDecision(
	ctx context.Context,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	incomeVerification *domain.IncomeVerification,
	policy *domain.UnderwritingPolicy,
) (*domain.UnderwritingResult, error) {
	// Create decision request
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

	// Use decision engine service if available, otherwise use built-in logic
	var decisionResponse *domain.DecisionResponse
	var err error

	if h.decisionEngineService != nil && h.decisionEngineService.IsAvailable(ctx) {
		decisionResponse, err = h.decisionEngineService.MakeDecision(ctx, request)
		if err != nil {
			h.logger.Warn("Decision engine service failed, using built-in logic", zap.Error(err))
			decisionResponse = h.makeBuiltInDecision(application, creditReport, riskAssessment, incomeVerification, policy)
		}
	} else {
		decisionResponse = h.makeBuiltInDecision(application, creditReport, riskAssessment, incomeVerification, policy)
	}

	// Create underwriting result
	result := &domain.UnderwritingResult{
		ID:                   application.ID + "_underwriting_result",
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
		ModelVersion:         riskAssessment.ModelVersion,
		OfferExpirationDate:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		DecisionData:         decisionResponse.DecisionData,
		ProcessingTime:       decisionResponse.ProcessingTime,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Calculate financial details
	h.calculateFinancialDetails(result)

	return result, nil
}

// makeBuiltInDecision makes a decision using built-in logic when external service is unavailable
func (h *UnderwritingDecisionTaskHandler) makeBuiltInDecision(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	incomeVerification *domain.IncomeVerification,
	policy *domain.UnderwritingPolicy,
) *domain.DecisionResponse {
	response := &domain.DecisionResponse{
		Decision:             domain.DecisionDenied,
		ApprovedAmount:       0,
		ApprovedTerm:         0,
		InterestRate:         0,
		APR:                  0,
		MonthlyPayment:       0,
		Conditions:           []domain.UnderwritingCondition{},
		Reasons:              []domain.DecisionReason{},
		CounterOffer:         nil,
		ManualReviewRequired: false,
		PolicyVersion:        policy.PolicyVersion,
		DecisionData:         make(map[string]interface{}),
		ProcessingTime:       0,
	}

	// Apply policy rules
	policyCheck := h.checkPolicyCompliance(application, creditReport, riskAssessment, incomeVerification, policy)

	if !policyCheck.Compliant {
		response.Decision = domain.DecisionDenied
		response.Reasons = h.convertPolicyViolationsToReasons(policyCheck.Violations)
		return response
	}

	// Risk-based decision logic
	switch riskAssessment.OverallRiskLevel {
	case domain.RiskLow:
		response = h.makeApprovalDecision(application, creditReport, riskAssessment, policy)
	case domain.RiskMedium:
		response = h.makeConditionalDecision(application, creditReport, riskAssessment, policy)
	case domain.RiskHigh:
		response = h.makeManualReviewDecision(application, creditReport, riskAssessment, policy)
	case domain.RiskCritical:
		response = h.makeDenialDecision(application, creditReport, riskAssessment, policy)
	}

	// Apply income verification requirements
	if incomeVerification.VerificationStatus != domain.IncomeVerified {
		response.ManualReviewRequired = true
		response.Conditions = append(response.Conditions, domain.UnderwritingCondition{
			ConditionID:   "income_verification_required",
			ConditionType: "prior_to_funding",
			Description:   "Income verification must be completed",
			Priority:      "critical",
			Status:        "pending",
			DueDate:       time.Now().Add(7 * 24 * time.Hour),
		})
	}

	return response
}

// Policy compliance check
func (h *UnderwritingDecisionTaskHandler) checkPolicyCompliance(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	incomeVerification *domain.IncomeVerification,
	policy *domain.UnderwritingPolicy,
) PolicyComplianceResult {
	result := PolicyComplianceResult{
		Compliant:  true,
		Violations: []PolicyViolation{},
	}

	// Check minimum credit score
	if creditReport.CreditScore < policy.MinCreditScore {
		result.Compliant = false
		result.Violations = append(result.Violations, PolicyViolation{
			RuleID:      "min_credit_score",
			Description: fmt.Sprintf("Credit score %d below minimum %d", creditReport.CreditScore, policy.MinCreditScore),
			Severity:    "critical",
		})
	}

	// Check maximum DTI ratio
	dti := application.CalculateDTI()
	if dti > policy.MaxDTIRatio {
		result.Compliant = false
		result.Violations = append(result.Violations, PolicyViolation{
			RuleID:      "max_dti_ratio",
			Description: fmt.Sprintf("DTI ratio %.1f%% exceeds maximum %.1f%%", dti*100, policy.MaxDTIRatio*100),
			Severity:    "critical",
		})
	}

	// Check minimum annual income
	if application.AnnualIncome < policy.MinAnnualIncome {
		result.Compliant = false
		result.Violations = append(result.Violations, PolicyViolation{
			RuleID:      "min_annual_income",
			Description: fmt.Sprintf("Annual income $%.0f below minimum $%.0f", application.AnnualIncome, policy.MinAnnualIncome),
			Severity:    "critical",
		})
	}

	// Check loan amount limits
	if application.LoanAmount > policy.MaxLoanAmount {
		result.Compliant = false
		result.Violations = append(result.Violations, PolicyViolation{
			RuleID:      "max_loan_amount",
			Description: fmt.Sprintf("Loan amount $%.0f exceeds maximum $%.0f", application.LoanAmount, policy.MaxLoanAmount),
			Severity:    "critical",
		})
	}

	if application.LoanAmount < policy.MinLoanAmount {
		result.Compliant = false
		result.Violations = append(result.Violations, PolicyViolation{
			RuleID:      "min_loan_amount",
			Description: fmt.Sprintf("Loan amount $%.0f below minimum $%.0f", application.LoanAmount, policy.MinLoanAmount),
			Severity:    "critical",
		})
	}

	return result
}

// Decision type methods
func (h *UnderwritingDecisionTaskHandler) makeApprovalDecision(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) *domain.DecisionResponse {
	interestRate := h.calculateInterestRate(creditReport, riskAssessment, policy)

	return &domain.DecisionResponse{
		Decision:       domain.DecisionApproved,
		ApprovedAmount: application.LoanAmount,
		ApprovedTerm:   application.RequestedTerm,
		InterestRate:   interestRate,
		APR:            interestRate + 0.5, // Simplified APR calculation
		Conditions:     []domain.UnderwritingCondition{},
		Reasons: []domain.DecisionReason{
			{
				ReasonCode:  "low_risk_approval",
				ReasonType:  "approval",
				Description: "Application meets all requirements for standard approval",
				Impact:      "primary",
				Weight:      1.0,
			},
		},
		CounterOffer:         nil,
		ManualReviewRequired: false,
		PolicyVersion:        policy.PolicyVersion,
		DecisionData: map[string]interface{}{
			"risk_level":    string(riskAssessment.OverallRiskLevel),
			"credit_score":  creditReport.CreditScore,
			"decision_type": "automated_approval",
		},
	}
}

func (h *UnderwritingDecisionTaskHandler) makeConditionalDecision(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) *domain.DecisionResponse {
	interestRate := h.calculateInterestRate(creditReport, riskAssessment, policy)
	conditions := h.generateConditions(riskAssessment)

	return &domain.DecisionResponse{
		Decision:       domain.DecisionConditional,
		ApprovedAmount: application.LoanAmount,
		ApprovedTerm:   application.RequestedTerm,
		InterestRate:   interestRate,
		APR:            interestRate + 0.5,
		Conditions:     conditions,
		Reasons: []domain.DecisionReason{
			{
				ReasonCode:  "conditional_approval",
				ReasonType:  "approval",
				Description: "Application approved subject to conditions",
				Impact:      "primary",
				Weight:      1.0,
			},
		},
		CounterOffer:         nil,
		ManualReviewRequired: false,
		PolicyVersion:        policy.PolicyVersion,
		DecisionData: map[string]interface{}{
			"risk_level":       string(riskAssessment.OverallRiskLevel),
			"credit_score":     creditReport.CreditScore,
			"decision_type":    "conditional_approval",
			"conditions_count": len(conditions),
		},
	}
}

func (h *UnderwritingDecisionTaskHandler) makeManualReviewDecision(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) *domain.DecisionResponse {
	return &domain.DecisionResponse{
		Decision:       domain.DecisionManualReview,
		ApprovedAmount: 0,
		ApprovedTerm:   0,
		InterestRate:   0,
		APR:            0,
		Conditions:     []domain.UnderwritingCondition{},
		Reasons: []domain.DecisionReason{
			{
				ReasonCode:  "manual_review_required",
				ReasonType:  "condition",
				Description: "Application requires manual underwriter review due to elevated risk",
				Impact:      "primary",
				Weight:      1.0,
			},
		},
		CounterOffer:         nil,
		ManualReviewRequired: true,
		PolicyVersion:        policy.PolicyVersion,
		DecisionData: map[string]interface{}{
			"risk_level":    string(riskAssessment.OverallRiskLevel),
			"credit_score":  creditReport.CreditScore,
			"decision_type": "manual_review",
		},
	}
}

func (h *UnderwritingDecisionTaskHandler) makeDenialDecision(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) *domain.DecisionResponse {
	return &domain.DecisionResponse{
		Decision:       domain.DecisionDenied,
		ApprovedAmount: 0,
		ApprovedTerm:   0,
		InterestRate:   0,
		APR:            0,
		Conditions:     []domain.UnderwritingCondition{},
		Reasons: []domain.DecisionReason{
			{
				ReasonCode:  "high_risk_denial",
				ReasonType:  "denial",
				Description: "Application denied due to high risk profile",
				Impact:      "primary",
				Weight:      1.0,
			},
		},
		CounterOffer:         h.generateCounterOffer(application, creditReport, riskAssessment, policy),
		ManualReviewRequired: false,
		PolicyVersion:        policy.PolicyVersion,
		DecisionData: map[string]interface{}{
			"risk_level":    string(riskAssessment.OverallRiskLevel),
			"credit_score":  creditReport.CreditScore,
			"decision_type": "automated_denial",
		},
	}
}

// Helper calculation methods
func (h *UnderwritingDecisionTaskHandler) calculateInterestRate(
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) float64 {
	baseRate := policy.InterestRateMatrix.BaseRate

	// Credit score adjustment
	scoreRange := domain.GetCreditScoreRange(creditReport.CreditScore)
	if rateRange, exists := policy.InterestRateMatrix.RateRanges[scoreRange]; exists {
		baseRate = rateRange.MinRate
	}

	// Risk level adjustment
	if riskAdjustment, exists := policy.InterestRateMatrix.RiskAdjustments[riskAssessment.OverallRiskLevel]; exists {
		baseRate += riskAdjustment
	}

	// Ensure rate is within reasonable bounds
	if baseRate < 5.0 {
		baseRate = 5.0
	}
	if baseRate > 25.0 {
		baseRate = 25.0
	}

	return math.Round(baseRate*100) / 100
}

func (h *UnderwritingDecisionTaskHandler) generateConditions(riskAssessment *domain.RiskAssessment) []domain.UnderwritingCondition {
	conditions := []domain.UnderwritingCondition{}

	for _, riskFactor := range riskAssessment.RiskFactors {
		if riskFactor.Impact == "high" {
			conditions = append(conditions, domain.UnderwritingCondition{
				ConditionID:   riskFactor.FactorID + "_condition",
				ConditionType: "prior_to_funding",
				Description:   "Address " + riskFactor.Description,
				Priority:      "high",
				Status:        "pending",
				DueDate:       time.Now().Add(14 * 24 * time.Hour),
			})
		}
	}

	return conditions
}

func (h *UnderwritingDecisionTaskHandler) generateCounterOffer(
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	policy *domain.UnderwritingPolicy,
) *domain.CounterOfferTerms {
	// Generate a counter offer with reduced amount or different terms
	reducedAmount := application.LoanAmount * 0.7 // 70% of requested amount

	if reducedAmount < policy.MinLoanAmount {
		return nil // Can't make a counter offer below minimum
	}

	interestRate := h.calculateInterestRate(creditReport, riskAssessment, policy) + 2.0 // Higher rate

	return &domain.CounterOfferTerms{
		OfferedAmount:   reducedAmount,
		OfferedTerm:     application.RequestedTerm,
		OfferedRate:     interestRate,
		OfferedAPR:      interestRate + 0.5,
		OfferReason:     "Reduced amount to mitigate risk",
		OfferConditions: []string{"Additional income verification required"},
		ExpirationDate:  time.Now().Add(7 * 24 * time.Hour),
	}
}

func (h *UnderwritingDecisionTaskHandler) calculateFinancialDetails(result *domain.UnderwritingResult) {
	if result.ApprovedAmount > 0 && result.ApprovedTerm > 0 && result.InterestRate > 0 {
		// Calculate monthly payment
		principal := result.ApprovedAmount
		monthlyRate := result.InterestRate / 12 / 100
		termMonths := float64(result.ApprovedTerm)

		if monthlyRate > 0 {
			result.MonthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, termMonths)) /
				(math.Pow(1+monthlyRate, termMonths) - 1)
		} else {
			result.MonthlyPayment = principal / termMonths
		}

		// Calculate total payment and interest
		result.TotalPayment = result.MonthlyPayment * termMonths
		result.TotalInterest = result.TotalPayment - result.ApprovedAmount

		// Round to 2 decimal places
		result.MonthlyPayment = math.Round(result.MonthlyPayment*100) / 100
		result.TotalPayment = math.Round(result.TotalPayment*100) / 100
		result.TotalInterest = math.Round(result.TotalInterest*100) / 100
	}
}

// Response creation methods
func (h *UnderwritingDecisionTaskHandler) createSuccessResponse(
	result *domain.UnderwritingResult,
	application *domain.LoanApplication,
	creditReport *domain.CreditReport,
	riskAssessment *domain.RiskAssessment,
	incomeVerification *domain.IncomeVerification,
) map[string]interface{} {
	return map[string]interface{}{
		"success":       true,
		"applicationId": application.ID,
		"userId":        application.UserID,
		"underwritingResult": map[string]interface{}{
			"resultId":             result.ID,
			"decision":             string(result.Decision),
			"status":               string(result.Status),
			"approvedAmount":       result.ApprovedAmount,
			"approvedTerm":         result.ApprovedTerm,
			"interestRate":         result.InterestRate,
			"apr":                  result.APR,
			"monthlyPayment":       result.MonthlyPayment,
			"totalInterest":        result.TotalInterest,
			"totalPayment":         result.TotalPayment,
			"automatedDecision":    result.AutomatedDecision,
			"manualReviewRequired": result.ManualReviewRequired,
			"offerExpirationDate":  result.OfferExpirationDate.Format(time.RFC3339),
			"policyVersion":        result.PolicyVersion,
			"modelVersion":         result.ModelVersion,
		},
		"conditions":      h.formatConditions(result.Conditions),
		"decisionReasons": h.formatDecisionReasons(result.DecisionReasons),
		"counterOffer":    h.formatCounterOffer(result.CounterOfferTerms),
		"inputSummary": map[string]interface{}{
			"requestedAmount": application.LoanAmount,
			"requestedTerm":   application.RequestedTerm,
			"creditScore":     creditReport.CreditScore,
			"riskLevel":       string(riskAssessment.OverallRiskLevel),
			"incomeVerified":  incomeVerification.VerificationStatus == domain.IncomeVerified,
			"dtiRatio":        application.CalculateDTI(),
		},
		"processingTime": result.ProcessingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}
}

func (h *UnderwritingDecisionTaskHandler) createFailureResponse(applicationID string, err error) map[string]interface{} {
	return map[string]interface{}{
		"success":       false,
		"applicationId": applicationID,
		"error":         err.Error(),
		"underwritingResult": map[string]interface{}{
			"decision":             string(domain.DecisionManualReview),
			"manualReviewRequired": true,
		},
		"completedAt": time.Now().UTC().Format(time.RFC3339),
	}
}

// Format methods
func (h *UnderwritingDecisionTaskHandler) formatConditions(conditions []domain.UnderwritingCondition) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(conditions))
	for i, condition := range conditions {
		formatted[i] = map[string]interface{}{
			"conditionId":   condition.ConditionID,
			"conditionType": condition.ConditionType,
			"description":   condition.Description,
			"priority":      condition.Priority,
			"status":        condition.Status,
			"dueDate":       condition.DueDate.Format(time.RFC3339),
		}
	}
	return formatted
}

func (h *UnderwritingDecisionTaskHandler) formatDecisionReasons(reasons []domain.DecisionReason) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(reasons))
	for i, reason := range reasons {
		formatted[i] = map[string]interface{}{
			"reasonCode":  reason.ReasonCode,
			"reasonType":  reason.ReasonType,
			"description": reason.Description,
			"impact":      reason.Impact,
			"weight":      reason.Weight,
		}
	}
	return formatted
}

func (h *UnderwritingDecisionTaskHandler) formatCounterOffer(counterOffer *domain.CounterOfferTerms) map[string]interface{} {
	if counterOffer == nil {
		return nil
	}

	return map[string]interface{}{
		"offeredAmount":   counterOffer.OfferedAmount,
		"offeredTerm":     counterOffer.OfferedTerm,
		"offeredRate":     counterOffer.OfferedRate,
		"offeredAPR":      counterOffer.OfferedAPR,
		"monthlyPayment":  counterOffer.MonthlyPayment,
		"totalInterest":   counterOffer.TotalInterest,
		"offerReason":     counterOffer.OfferReason,
		"offerConditions": counterOffer.OfferConditions,
		"expirationDate":  counterOffer.ExpirationDate.Format(time.RFC3339),
	}
}

// Helper methods
func (h *UnderwritingDecisionTaskHandler) convertPolicyViolationsToReasons(violations []PolicyViolation) []domain.DecisionReason {
	reasons := make([]domain.DecisionReason, len(violations))
	for i, violation := range violations {
		reasons[i] = domain.DecisionReason{
			ReasonCode:  violation.RuleID,
			ReasonType:  "denial",
			Description: violation.Description,
			Impact:      "primary",
			Weight:      1.0,
		}
	}
	return reasons
}

// Supporting types
type PolicyComplianceResult struct {
	Compliant  bool              `json:"compliant"`
	Violations []PolicyViolation `json:"violations"`
}

type PolicyViolation struct {
	RuleID      string `json:"rule_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}
