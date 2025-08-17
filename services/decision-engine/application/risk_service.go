package application

import (
	"context"
	"fmt"
	"math"

	"decision-engine/domain"
	"go.uber.org/zap"
)

// RiskAssessmentService implements risk assessment logic
type RiskAssessmentService struct {
	logger *zap.Logger
}

// NewRiskAssessmentService creates a new risk assessment service
func NewRiskAssessmentService(logger *zap.Logger) *RiskAssessmentService {
	return &RiskAssessmentService{
		logger: logger,
	}
}

// AssessRisk performs comprehensive risk assessment
func (s *RiskAssessmentService) AssessRisk(request *domain.DecisionRequest) (*domain.RiskAssessment, error) {
	logger := s.logger.With(
		zap.String("application_id", request.ApplicationID),
		zap.String("operation", "assess_risk"),
	)

	logger.Info("Starting risk assessment")

	assessment := &domain.RiskAssessment{
		DTIRatio: request.CalculateDTI(),
	}

	// Calculate category scores
	assessment.CategoryScores = s.calculateCategoryScores(request)

	// Assess payment history (simulated - in real implementation, this would come from credit bureau)
	assessment.PaymentHistory = s.assessPaymentHistory(request.CreditScore)

	// Identify risk factors
	assessment.RiskFactors = s.identifyRiskFactors(request, assessment)

	// Calculate overall risk score
	assessment.OverallScore = s.CalculateRiskScore(assessment)

	// Identify mitigating factors
	assessment.MitigatingFactors = s.identifyMitigatingFactors(request, assessment)

	logger.Info("Risk assessment completed",
		zap.Float64("overall_score", assessment.OverallScore),
		zap.Float64("dti_ratio", assessment.DTIRatio),
		zap.Int("risk_factors", len(assessment.RiskFactors)),
	)

	return assessment, nil
}

// calculateCategoryScores calculates risk scores for each category
func (s *RiskAssessmentService) calculateCategoryScores(request *domain.DecisionRequest) domain.CategoryScores {
	return domain.CategoryScores{
		CreditRisk:     s.calculateCreditRisk(request),
		IncomeRisk:     s.calculateIncomeRisk(request),
		DebtRisk:       s.calculateDebtRisk(request),
		EmploymentRisk: s.calculateEmploymentRisk(request),
		CollateralRisk: 0.0, // Not applicable for unsecured loans
	}
}

// calculateCreditRisk calculates credit-based risk score (0-1, higher is riskier)
func (s *RiskAssessmentService) calculateCreditRisk(request *domain.DecisionRequest) float64 {
	creditScore := float64(request.CreditScore)

	// Normalize credit score to risk score (invert so higher credit = lower risk)
	normalizedScore := (850 - creditScore) / (850 - 300)

	// Apply additional penalties for very low scores
	if creditScore < 600 {
		normalizedScore += 0.2
	} else if creditScore < 650 {
		normalizedScore += 0.1
	}

	return math.Min(normalizedScore, 1.0)
}

// calculateIncomeRisk calculates income-based risk score
func (s *RiskAssessmentService) calculateIncomeRisk(request *domain.DecisionRequest) float64 {
	// Loan-to-income ratio risk
	ltiRatio := request.GetLoanToIncomeRatio()

	var riskScore float64

	switch {
	case ltiRatio <= 0.5:
		riskScore = 0.1 // Very low risk
	case ltiRatio <= 1.0:
		riskScore = 0.3 // Low risk
	case ltiRatio <= 2.0:
		riskScore = 0.6 // Medium risk
	case ltiRatio <= 3.0:
		riskScore = 0.8 // High risk
	default:
		riskScore = 1.0 // Very high risk
	}

	// Adjust for absolute income level
	if request.AnnualIncome < 30000 {
		riskScore += 0.2
	} else if request.AnnualIncome < 50000 {
		riskScore += 0.1
	}

	// Verify monthly vs annual income consistency
	expectedMonthly := request.AnnualIncome / 12
	actualMonthly := request.MonthlyIncome

	if math.Abs(expectedMonthly-actualMonthly)/expectedMonthly > 0.2 {
		riskScore += 0.15 // Inconsistent income data
	}

	return math.Min(riskScore, 1.0)
}

// calculateDebtRisk calculates debt-based risk score
func (s *RiskAssessmentService) calculateDebtRisk(request *domain.DecisionRequest) float64 {
	dtiRatio := request.CalculateDTI()

	switch {
	case dtiRatio <= 0.20:
		return 0.1 // Excellent
	case dtiRatio <= 0.35:
		return 0.3 // Good
	case dtiRatio <= 0.40:
		return 0.6 // Fair
	case dtiRatio <= 0.45:
		return 0.8 // Poor
	default:
		return 1.0 // Very poor
	}
}

// calculateEmploymentRisk calculates employment-based risk score
func (s *RiskAssessmentService) calculateEmploymentRisk(request *domain.DecisionRequest) float64 {
	riskScores := map[domain.EmploymentType]float64{
		domain.EmploymentFullTime:     0.1,
		domain.EmploymentPartTime:     0.4,
		domain.EmploymentContract:     0.5,
		domain.EmploymentSelfEmployed: 0.7,
		domain.EmploymentRetired:      0.3,
		domain.EmploymentUnemployed:   1.0,
	}

	if score, exists := riskScores[request.EmploymentType]; exists {
		return score
	}
	return 0.8 // Default for unknown employment types
}

// assessPaymentHistory simulates payment history assessment
func (s *RiskAssessmentService) assessPaymentHistory(creditScore int) domain.PaymentHistory {
	// In real implementation, this would come from credit bureau data
	// This is a simulation based on credit score

	var history domain.PaymentHistory

	switch {
	case creditScore >= 750:
		history = domain.PaymentHistory{
			OnTimePayments: 95,
			LatePayments:   2,
			Defaults:       0,
			Bankruptcies:   0,
			CreditAge:      120, // 10 years
			PaymentScore:   0.95,
		}
	case creditScore >= 700:
		history = domain.PaymentHistory{
			OnTimePayments: 85,
			LatePayments:   8,
			Defaults:       0,
			Bankruptcies:   0,
			CreditAge:      84, // 7 years
			PaymentScore:   0.85,
		}
	case creditScore >= 650:
		history = domain.PaymentHistory{
			OnTimePayments: 75,
			LatePayments:   15,
			Defaults:       1,
			Bankruptcies:   0,
			CreditAge:      60, // 5 years
			PaymentScore:   0.70,
		}
	case creditScore >= 600:
		history = domain.PaymentHistory{
			OnTimePayments: 65,
			LatePayments:   20,
			Defaults:       2,
			Bankruptcies:   0,
			CreditAge:      48, // 4 years
			PaymentScore:   0.60,
		}
	default:
		history = domain.PaymentHistory{
			OnTimePayments: 50,
			LatePayments:   30,
			Defaults:       5,
			Bankruptcies:   1,
			CreditAge:      36, // 3 years
			PaymentScore:   0.40,
		}
	}

	return history
}

// identifyRiskFactors identifies specific risk factors
func (s *RiskAssessmentService) identifyRiskFactors(request *domain.DecisionRequest, assessment *domain.RiskAssessment) []domain.RiskFactor {
	var factors []domain.RiskFactor

	// Credit score factors
	if request.CreditScore < 650 {
		impact := "HIGH"
		if request.CreditScore >= 600 {
			impact = "MEDIUM"
		}
		factors = append(factors, domain.RiskFactor{
			Category:    "CREDIT",
			Factor:      "Low Credit Score",
			Impact:      impact,
			Score:       assessment.CategoryScores.CreditRisk,
			Description: fmt.Sprintf("Credit score of %d is below optimal range", request.CreditScore),
		})
	}

	// DTI factors
	if assessment.DTIRatio > 0.35 {
		impact := "MEDIUM"
		if assessment.DTIRatio > 0.40 {
			impact = "HIGH"
		}
		factors = append(factors, domain.RiskFactor{
			Category:    "DEBT",
			Factor:      "High Debt-to-Income Ratio",
			Impact:      impact,
			Score:       assessment.CategoryScores.DebtRisk,
			Description: fmt.Sprintf("DTI ratio of %.2f exceeds recommended threshold", assessment.DTIRatio),
		})
	}

	// Income factors
	if request.AnnualIncome < 40000 {
		factors = append(factors, domain.RiskFactor{
			Category:    "INCOME",
			Factor:      "Low Income",
			Impact:      "MEDIUM",
			Score:       assessment.CategoryScores.IncomeRisk,
			Description: fmt.Sprintf("Annual income of $%.0f may limit repayment capacity", request.AnnualIncome),
		})
	}

	// Loan amount factors
	ltiRatio := request.GetLoanToIncomeRatio()
	if ltiRatio > 2.0 {
		factors = append(factors, domain.RiskFactor{
			Category:    "INCOME",
			Factor:      "High Loan-to-Income Ratio",
			Impact:      "HIGH",
			Score:       ltiRatio,
			Description: fmt.Sprintf("Loan amount is %.1fx annual income", ltiRatio),
		})
	}

	// Employment factors
	if request.EmploymentType == domain.EmploymentSelfEmployed {
		factors = append(factors, domain.RiskFactor{
			Category:    "EMPLOYMENT",
			Factor:      "Self-Employment",
			Impact:      "MEDIUM",
			Score:       assessment.CategoryScores.EmploymentRisk,
			Description: "Self-employed income may be less stable",
		})
	}

	if request.EmploymentType == domain.EmploymentContract {
		factors = append(factors, domain.RiskFactor{
			Category:    "EMPLOYMENT",
			Factor:      "Contract Employment",
			Impact:      "MEDIUM",
			Score:       assessment.CategoryScores.EmploymentRisk,
			Description: "Contract employment may have limited duration",
		})
	}

	// Payment history factors
	if assessment.PaymentHistory.PaymentScore < 0.7 {
		factors = append(factors, domain.RiskFactor{
			Category:    "CREDIT",
			Factor:      "Poor Payment History",
			Impact:      "HIGH",
			Score:       1.0 - assessment.PaymentHistory.PaymentScore,
			Description: fmt.Sprintf("Payment score of %.2f indicates past payment issues", assessment.PaymentHistory.PaymentScore),
		})
	}

	return factors
}

// identifyMitigatingFactors identifies factors that reduce risk
func (s *RiskAssessmentService) identifyMitigatingFactors(request *domain.DecisionRequest, assessment *domain.RiskAssessment) []string {
	var factors []string

	// Excellent credit score
	if request.CreditScore >= 750 {
		factors = append(factors, "Excellent credit score demonstrates strong credit management")
	}

	// Low DTI
	if assessment.DTIRatio <= 0.20 {
		factors = append(factors, "Low debt-to-income ratio indicates strong debt management")
	}

	// High income
	if request.AnnualIncome >= 100000 {
		factors = append(factors, "High income provides strong repayment capacity")
	}

	// Stable employment
	if request.EmploymentType == domain.EmploymentFullTime {
		factors = append(factors, "Full-time employment provides income stability")
	}

	// Conservative loan amount
	if request.GetLoanToIncomeRatio() <= 0.5 {
		factors = append(factors, "Conservative loan amount relative to income")
	}

	// Excellent payment history
	if assessment.PaymentHistory.PaymentScore >= 0.9 {
		factors = append(factors, "Excellent payment history demonstrates reliability")
	}

	// Debt consolidation purpose (generally lower risk)
	if request.LoanPurpose == domain.PurposeDebtConsolidation {
		factors = append(factors, "Debt consolidation may improve overall financial position")
	}

	return factors
}

// CalculateRiskScore calculates the overall risk score
func (s *RiskAssessmentService) CalculateRiskScore(assessment *domain.RiskAssessment) float64 {
	// Weighted average of category scores
	weights := map[string]float64{
		"credit":     0.35, // Credit risk is most important
		"debt":       0.25, // Debt management is crucial
		"income":     0.20, // Income capacity matters
		"employment": 0.15, // Employment stability
		"collateral": 0.05, // Minimal for unsecured loans
	}

	score := assessment.CategoryScores.CreditRisk*weights["credit"] +
		assessment.CategoryScores.DebtRisk*weights["debt"] +
		assessment.CategoryScores.IncomeRisk*weights["income"] +
		assessment.CategoryScores.EmploymentRisk*weights["employment"] +
		assessment.CategoryScores.CollateralRisk*weights["collateral"]

	// Adjust for payment history
	paymentAdjustment := (1.0 - assessment.PaymentHistory.PaymentScore) * 0.1
	score += paymentAdjustment

	// Apply mitigating factors discount
	mitigatingDiscount := float64(len(assessment.MitigatingFactors)) * 0.02
	score -= mitigatingDiscount

	// Ensure score is between 0 and 1
	return math.Max(0.0, math.Min(1.0, score))
}

// CategorizeRisk categorizes risk level based on score
func (s *RiskAssessmentService) CategorizeRisk(score float64) domain.RiskCategory {
	switch {
	case score <= 0.3:
		return domain.RiskLow
	case score <= 0.6:
		return domain.RiskMedium
	case score <= 0.8:
		return domain.RiskHigh
	default:
		return domain.RiskCritical
	}
}
