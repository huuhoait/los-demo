package application

import (
	"context"
	"fmt"
	"math"
	"time"

	"decision-engine/domain"
	"go.uber.org/zap"
)

// DecisionEngineService implements the core decision making logic
type DecisionEngineService struct {
	riskService    domain.RiskAssessmentService
	rulesService   domain.RulesEngineService
	decisionRepo   domain.DecisionRepository
	logger         *zap.Logger
}

// NewDecisionEngineService creates a new decision engine service
func NewDecisionEngineService(
	riskService domain.RiskAssessmentService,
	rulesService domain.RulesEngineService,
	decisionRepo domain.DecisionRepository,
	logger *zap.Logger,
) *DecisionEngineService {
	return &DecisionEngineService{
		riskService:  riskService,
		rulesService: rulesService,
		decisionRepo: decisionRepo,
		logger:       logger,
	}
}

// MakeDecision processes a loan decision request
func (s *DecisionEngineService) MakeDecision(ctx context.Context, request *domain.DecisionRequest) (*domain.DecisionResponse, error) {
	logger := s.logger.With(
		zap.String("application_id", request.ApplicationID),
		zap.String("user_id", request.UserID),
		zap.Float64("loan_amount", request.LoanAmount),
	)

	logger.Info("Processing decision request")

	// Validate request
	if err := s.ValidateRequest(request); err != nil {
		logger.Error("Request validation failed", zap.Error(err))
		return nil, err
	}

	// Perform risk assessment
	riskAssessment, err := s.riskService.AssessRisk(request)
	if err != nil {
		logger.Error("Risk assessment failed", zap.Error(err))
		return nil, &domain.DecisionError{
			Code:        domain.ERROR_RISK_ASSESSMENT,
			Message:     "Risk assessment failed",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Apply business rules
	decision, err := s.rulesService.EvaluateRules(request, riskAssessment)
	if err != nil {
		logger.Error("Rules evaluation failed", zap.Error(err))
		return nil, &domain.DecisionError{
			Code:        domain.ERROR_RULE_EVALUATION,
			Message:     "Rules evaluation failed",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	// Enhance decision with additional logic
	s.enhanceDecision(decision, request, riskAssessment)

	// Save decision
	if err := s.decisionRepo.SaveDecision(decision); err != nil {
		logger.Error("Failed to save decision", zap.Error(err))
		// Don't fail the request if saving fails
	}

	logger.Info("Decision completed",
		zap.String("decision", string(decision.Decision)),
		zap.Float64("risk_score", decision.RiskScore),
		zap.String("risk_category", string(decision.RiskCategory)),
	)

	return decision, nil
}

// ValidateRequest validates the decision request
func (s *DecisionEngineService) ValidateRequest(request *domain.DecisionRequest) error {
	if request.ApplicationID == "" {
		return &domain.DecisionError{
			Code:        domain.ERROR_INVALID_REQUEST,
			Message:     "Application ID is required",
			HTTPStatus:  400,
		}
	}

	if request.UserID == "" {
		return &domain.DecisionError{
			Code:        domain.ERROR_INVALID_REQUEST,
			Message:     "User ID is required",
			HTTPStatus:  400,
		}
	}

	if request.LoanAmount <= 0 || request.LoanAmount > 1000000 {
		return &domain.DecisionError{
			Code:        domain.ERROR_INVALID_REQUEST,
			Message:     "Invalid loan amount",
			HTTPStatus:  400,
		}
	}

	if request.AnnualIncome <= 0 {
		return &domain.DecisionError{
			Code:        domain.ERROR_INVALID_REQUEST,
			Message:     "Annual income must be positive",
			HTTPStatus:  400,
		}
	}

	if !request.IsValidCreditScore() {
		return &domain.DecisionError{
			Code:        domain.ERROR_INVALID_REQUEST,
			Message:     "Invalid credit score",
			HTTPStatus:  400,
		}
	}

	return nil
}

// enhanceDecision adds additional business logic to the decision
func (s *DecisionEngineService) enhanceDecision(
	decision *domain.DecisionResponse,
	request *domain.DecisionRequest,
	assessment *domain.RiskAssessment,
) {
	// Set expiration date for approvals
	if decision.Decision == domain.DecisionApprove {
		expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
		decision.ExpiresAt = &expiresAt
	}

	// Add conditions based on risk factors
	s.addConditions(decision, assessment)

	// Set required documents
	s.setRequiredDocuments(decision, request, assessment)

	// Adjust interest rate based on additional factors
	s.adjustInterestRate(decision, request, assessment)
}

// addConditions adds loan conditions based on risk assessment
func (s *DecisionEngineService) addConditions(decision *domain.DecisionResponse, assessment *domain.RiskAssessment) {
	conditions := []string{}

	// DTI ratio conditions
	if assessment.DTIRatio > 0.35 && assessment.DTIRatio <= 0.40 {
		conditions = append(conditions, "Debt-to-income ratio monitoring required")
	}

	// Credit score conditions
	if assessment.CategoryScores.CreditRisk > 0.6 {
		conditions = append(conditions, "Credit improvement plan required")
	}

	// Employment conditions
	if assessment.CategoryScores.EmploymentRisk > 0.5 {
		conditions = append(conditions, "Employment verification within 30 days")
	}

	decision.Conditions = conditions
}

// setRequiredDocuments sets required documents based on risk factors
func (s *DecisionEngineService) setRequiredDocuments(
	decision *domain.DecisionResponse,
	request *domain.DecisionRequest,
	assessment *domain.RiskAssessment,
) {
	docs := []string{}

	// Always require basic documents
	docs = append(docs, "Government ID", "Proof of income")

	// Additional documents based on employment type
	switch request.EmploymentType {
	case domain.EmploymentSelfEmployed:
		docs = append(docs, "Tax returns (2 years)", "Business license", "Bank statements (6 months)")
	case domain.EmploymentContract:
		docs = append(docs, "Contract agreement", "Previous year tax return")
	}

	// Additional documents based on risk
	if assessment.CategoryScores.IncomeRisk > 0.5 {
		docs = append(docs, "Additional income verification", "Employment letter")
	}

	if assessment.DTIRatio > 0.35 {
		docs = append(docs, "Debt statements", "Monthly budget plan")
	}

	decision.RequiredDocs = docs
}

// adjustInterestRate adjusts interest rate based on comprehensive risk analysis
func (s *DecisionEngineService) adjustInterestRate(
	decision *domain.DecisionResponse,
	request *domain.DecisionRequest,
	assessment *domain.RiskAssessment,
) {
	baseRate := s.getBaseInterestRate(request.LoanPurpose)
	
	// Risk adjustment
	riskAdjustment := assessment.OverallScore * 0.05 // Up to 5% adjustment

	// Credit score adjustment
	creditAdjustment := s.getCreditScoreAdjustment(request.CreditScore)

	// DTI adjustment
	dtiAdjustment := s.getDTIAdjustment(assessment.DTIRatio)

	// Employment type adjustment
	employmentAdjustment := s.getEmploymentAdjustment(request.EmploymentType)

	// Calculate final rate
	finalRate := baseRate + riskAdjustment + creditAdjustment + dtiAdjustment + employmentAdjustment

	// Apply floor and ceiling
	finalRate = math.Max(finalRate, 5.0)  // Minimum 5%
	finalRate = math.Min(finalRate, 25.0) // Maximum 25%

	decision.InterestRate = math.Round(finalRate*100) / 100 // Round to 2 decimal places
}

// getBaseInterestRate returns base interest rate by loan purpose
func (s *DecisionEngineService) getBaseInterestRate(purpose domain.LoanPurpose) float64 {
	rates := map[domain.LoanPurpose]float64{
		domain.PurposePersonal:           8.5,
		domain.PurposeDebtConsolidation: 7.5,
		domain.PurposeHomeImprovement:   7.0,
		domain.PurposeBusiness:          9.0,
		domain.PurposeEducation:         6.5,
		domain.PurposeMedical:           8.0,
		domain.PurposeVacation:          10.0,
		domain.PurposeOther:             9.5,
	}

	if rate, exists := rates[purpose]; exists {
		return rate
	}
	return 9.0 // Default rate
}

// getCreditScoreAdjustment returns interest rate adjustment based on credit score
func (s *DecisionEngineService) getCreditScoreAdjustment(creditScore int) float64 {
	switch {
	case creditScore >= 750:
		return -1.5 // Excellent credit discount
	case creditScore >= 700:
		return -0.5 // Good credit discount
	case creditScore >= 650:
		return 0.0  // No adjustment
	case creditScore >= 600:
		return 1.0  // Fair credit premium
	default:
		return 2.5  // Poor credit premium
	}
}

// getDTIAdjustment returns interest rate adjustment based on DTI ratio
func (s *DecisionEngineService) getDTIAdjustment(dtiRatio float64) float64 {
	switch {
	case dtiRatio <= 0.20:
		return -0.5 // Low DTI discount
	case dtiRatio <= 0.35:
		return 0.0  // No adjustment
	case dtiRatio <= 0.40:
		return 0.5  // Moderate DTI premium
	default:
		return 1.5  // High DTI premium
	}
}

// getEmploymentAdjustment returns interest rate adjustment based on employment type
func (s *DecisionEngineService) getEmploymentAdjustment(employmentType domain.EmploymentType) float64 {
	adjustments := map[domain.EmploymentType]float64{
		domain.EmploymentFullTime:    0.0,  // No adjustment
		domain.EmploymentPartTime:    0.5,  // Small premium
		domain.EmploymentContract:    1.0,  // Moderate premium
		domain.EmploymentSelfEmployed: 1.5, // Higher premium
		domain.EmploymentRetired:     0.25, // Small premium
		domain.EmploymentUnemployed:  5.0,  // High premium (should rarely be approved)
	}

	if adjustment, exists := adjustments[employmentType]; exists {
		return adjustment
	}
	return 1.0 // Default premium
}

// GetDecision retrieves a saved decision
func (s *DecisionEngineService) GetDecision(ctx context.Context, applicationID string) (*domain.DecisionResponse, error) {
	logger := s.logger.With(zap.String("application_id", applicationID))

	decision, err := s.decisionRepo.GetDecision(applicationID)
	if err != nil {
		logger.Error("Failed to retrieve decision", zap.Error(err))
		return nil, &domain.DecisionError{
			Code:        domain.ERROR_DATABASE_ERROR,
			Message:     "Failed to retrieve decision",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Debug("Decision retrieved successfully")
	return decision, nil
}

// GetDecisionHistory retrieves decision history for a user
func (s *DecisionEngineService) GetDecisionHistory(ctx context.Context, userID string) ([]domain.DecisionResponse, error) {
	logger := s.logger.With(zap.String("user_id", userID))

	decisions, err := s.decisionRepo.GetDecisionHistory(userID)
	if err != nil {
		logger.Error("Failed to retrieve decision history", zap.Error(err))
		return nil, &domain.DecisionError{
			Code:        domain.ERROR_DATABASE_ERROR,
			Message:     "Failed to retrieve decision history",
			Description: err.Error(),
			HTTPStatus:  500,
		}
	}

	logger.Debug("Decision history retrieved", zap.Int("count", len(decisions)))
	return decisions, nil
}
