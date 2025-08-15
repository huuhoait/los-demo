package domain

import (
	"time"
)

// DecisionRequest represents a loan decision request
type DecisionRequest struct {
	ApplicationID    string                 `json:"application_id" validate:"required"`
	UserID          string                 `json:"user_id" validate:"required"`
	LoanAmount      float64                `json:"loan_amount" validate:"required,min=1000,max=1000000"`
	AnnualIncome    float64                `json:"annual_income" validate:"required,min=0"`
	MonthlyIncome   float64                `json:"monthly_income" validate:"required,min=0"`
	MonthlyDebt     float64                `json:"monthly_debt" validate:"min=0"`
	CreditScore     int                    `json:"credit_score" validate:"required,min=300,max=850"`
	EmploymentType  EmploymentType         `json:"employment_type" validate:"required"`
	RequestedTerm   int                    `json:"requested_term" validate:"required,min=12,max=84"`
	LoanPurpose     LoanPurpose            `json:"loan_purpose" validate:"required"`
	AdditionalData  map[string]interface{} `json:"additional_data,omitempty"`
	RequestedAt     time.Time              `json:"requested_at"`
}

// DecisionResponse represents the decision engine response
type DecisionResponse struct {
	ApplicationID   string          `json:"application_id"`
	Decision        DecisionType    `json:"decision"`
	RiskScore       float64         `json:"risk_score"`
	RiskCategory    RiskCategory    `json:"risk_category"`
	InterestRate    float64         `json:"interest_rate"`
	ApprovedAmount  float64         `json:"approved_amount,omitempty"`
	DecisionReason  string          `json:"decision_reason"`
	RiskFactors     []RiskFactor    `json:"risk_factors"`
	Conditions      []string        `json:"conditions,omitempty"`
	RequiredDocs    []string        `json:"required_documents,omitempty"`
	DecisionDate    time.Time       `json:"decision_date"`
	ExpiresAt       *time.Time      `json:"expires_at,omitempty"`
	ReviewRequired  bool            `json:"review_required"`
	ReviewerNotes   string          `json:"reviewer_notes,omitempty"`
}

// RiskAssessment contains detailed risk analysis
type RiskAssessment struct {
	OverallScore     float64        `json:"overall_score"`
	CategoryScores   CategoryScores `json:"category_scores"`
	DTIRatio         float64        `json:"dti_ratio"`
	LTVRatio         float64        `json:"ltv_ratio,omitempty"`
	CreditUtilization float64       `json:"credit_utilization,omitempty"`
	PaymentHistory   PaymentHistory `json:"payment_history"`
	RiskFactors      []RiskFactor   `json:"risk_factors"`
	MitigatingFactors []string      `json:"mitigating_factors,omitempty"`
}

// CategoryScores represents risk scores by category
type CategoryScores struct {
	CreditRisk      float64 `json:"credit_risk"`
	IncomeRisk      float64 `json:"income_risk"`
	DebtRisk        float64 `json:"debt_risk"`
	EmploymentRisk  float64 `json:"employment_risk"`
	CollateralRisk  float64 `json:"collateral_risk,omitempty"`
}

// PaymentHistory represents credit payment history
type PaymentHistory struct {
	OnTimePayments   int     `json:"on_time_payments"`
	LatePayments     int     `json:"late_payments"`
	Defaults         int     `json:"defaults"`
	Bankruptcies     int     `json:"bankruptcies"`
	CreditAge        int     `json:"credit_age_months"`
	PaymentScore     float64 `json:"payment_score"`
}

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	Category    string  `json:"category"`
	Factor      string  `json:"factor"`
	Impact      string  `json:"impact"` // HIGH, MEDIUM, LOW
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// DecisionRule represents a business rule for decision making
type DecisionRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    RuleCategory           `json:"category"`
	Priority    int                    `json:"priority"`
	Conditions  []RuleCondition        `json:"conditions"`
	Action      RuleAction             `json:"action"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RuleCondition represents a condition in a decision rule
type RuleCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // gt, lt, eq, gte, lte, in, contains
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"` // number, string, boolean, array
}

// RuleAction represents the action to take when rule conditions are met
type RuleAction struct {
	Type        ActionType             `json:"type"`
	Decision    DecisionType           `json:"decision,omitempty"`
	Reason      string                 `json:"reason"`
	Adjustments map[string]interface{} `json:"adjustments,omitempty"`
	RequireReview bool                 `json:"require_review"`
}

// Enums and Constants
type DecisionType string

const (
	DecisionApprove       DecisionType = "APPROVE"
	DecisionDeny          DecisionType = "DENY"
	DecisionManualReview  DecisionType = "MANUAL_REVIEW"
	DecisionConditional   DecisionType = "CONDITIONAL"
	DecisionPending       DecisionType = "PENDING"
)

type RiskCategory string

const (
	RiskLow       RiskCategory = "LOW"
	RiskMedium    RiskCategory = "MEDIUM"
	RiskHigh      RiskCategory = "HIGH"
	RiskCritical  RiskCategory = "CRITICAL"
)

type EmploymentType string

const (
	EmploymentFullTime   EmploymentType = "full_time"
	EmploymentPartTime   EmploymentType = "part_time"
	EmploymentContract   EmploymentType = "contract"
	EmploymentSelfEmployed EmploymentType = "self_employed"
	EmploymentUnemployed EmploymentType = "unemployed"
	EmploymentRetired    EmploymentType = "retired"
)

type LoanPurpose string

const (
	PurposePersonal      LoanPurpose = "personal"
	PurposeDebtConsolidation LoanPurpose = "debt_consolidation"
	PurposeHomeImprovement LoanPurpose = "home_improvement"
	PurposeBusiness      LoanPurpose = "business"
	PurposeEducation     LoanPurpose = "education"
	PurposeMedical       LoanPurpose = "medical"
	PurposeVacation      LoanPurpose = "vacation"
	PurposeOther         LoanPurpose = "other"
)

type RuleCategory string

const (
	RuleCategoryCredit     RuleCategory = "CREDIT"
	RuleCategoryIncome     RuleCategory = "INCOME"
	RuleCategoryDebt       RuleCategory = "DEBT"
	RuleCategoryEmployment RuleCategory = "EMPLOYMENT"
	RuleCategoryGeneral    RuleCategory = "GENERAL"
)

type ActionType string

const (
	ActionDecision       ActionType = "DECISION"
	ActionAdjustment     ActionType = "ADJUSTMENT"
	ActionRequirement    ActionType = "REQUIREMENT"
	ActionFlag           ActionType = "FLAG"
)

// Domain Services Interfaces
type RiskAssessmentService interface {
	AssessRisk(request *DecisionRequest) (*RiskAssessment, error)
	CalculateRiskScore(assessment *RiskAssessment) float64
	CategorizeRisk(score float64) RiskCategory
}

type DecisionEngineService interface {
	MakeDecision(request *DecisionRequest) (*DecisionResponse, error)
	ApplyBusinessRules(request *DecisionRequest, assessment *RiskAssessment) (*DecisionResponse, error)
	ValidateRequest(request *DecisionRequest) error
}

type RulesEngineService interface {
	EvaluateRules(request *DecisionRequest, assessment *RiskAssessment) (*DecisionResponse, error)
	GetActiveRules() ([]DecisionRule, error)
	AddRule(rule *DecisionRule) error
	UpdateRule(rule *DecisionRule) error
	DeleteRule(ruleID string) error
}

// Repository Interfaces
type DecisionRepository interface {
	SaveDecision(response *DecisionResponse) error
	GetDecision(applicationID string) (*DecisionResponse, error)
	GetDecisionHistory(userID string) ([]DecisionResponse, error)
	UpdateDecision(response *DecisionResponse) error
}

type RulesRepository interface {
	GetRules() ([]DecisionRule, error)
	GetRule(ruleID string) (*DecisionRule, error)
	SaveRule(rule *DecisionRule) error
	UpdateRule(rule *DecisionRule) error
	DeleteRule(ruleID string) error
	GetRulesByCategory(category RuleCategory) ([]DecisionRule, error)
}

// Value Objects
type CreditScoreRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type IncomeRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type DTIThresholds struct {
	Excellent float64 `json:"excellent"` // < 0.2
	Good      float64 `json:"good"`      // 0.2 - 0.35
	Fair      float64 `json:"fair"`      // 0.35 - 0.45
	Poor      float64 `json:"poor"`      // > 0.45
}

// Business Rules Constants
var (
	MinCreditScore    = 600
	MaxDTIRatio      = 0.45
	MinAnnualIncome  = 25000.0
	
	CreditScoreRanges = map[RiskCategory]CreditScoreRange{
		RiskLow:      {740, 850},
		RiskMedium:   {670, 739},
		RiskHigh:     {600, 669},
		RiskCritical: {300, 599},
	}
	
	DefaultDTIThresholds = DTIThresholds{
		Excellent: 0.20,
		Good:      0.35,
		Fair:      0.45,
		Poor:      1.00,
	}
)

// Validation Methods
func (dr *DecisionRequest) CalculateDTI() float64 {
	if dr.MonthlyIncome <= 0 {
		return 0
	}
	return dr.MonthlyDebt / dr.MonthlyIncome
}

func (dr *DecisionRequest) IsValidCreditScore() bool {
	return dr.CreditScore >= 300 && dr.CreditScore <= 850
}

func (dr *DecisionRequest) GetLoanToIncomeRatio() float64 {
	if dr.AnnualIncome <= 0 {
		return 0
	}
	return dr.LoanAmount / dr.AnnualIncome
}

// Domain Errors
type DecisionError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	HTTPStatus  int    `json:"http_status"`
}

func (e *DecisionError) Error() string {
	return e.Message
}

const (
	ERROR_INVALID_REQUEST      = "DECISION_001"
	ERROR_INSUFFICIENT_DATA    = "DECISION_002"
	ERROR_RULE_EVALUATION      = "DECISION_003"
	ERROR_RISK_ASSESSMENT      = "DECISION_004"
	ERROR_DATABASE_ERROR       = "DECISION_005"
	ERROR_EXTERNAL_SERVICE     = "DECISION_006"
	ERROR_BUSINESS_RULE        = "DECISION_007"
)
