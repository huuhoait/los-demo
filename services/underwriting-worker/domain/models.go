package domain

import (
	"time"
)

// UnderwritingDecision represents the final underwriting decision
type UnderwritingDecision string

const (
	DecisionApproved     UnderwritingDecision = "approved"
	DecisionDenied       UnderwritingDecision = "denied"
	DecisionConditional  UnderwritingDecision = "conditional"
	DecisionManualReview UnderwritingDecision = "manual_review"
	DecisionCounterOffer UnderwritingDecision = "counter_offer"
)

// RiskLevel represents the risk assessment level
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// UnderwritingStatus represents the status of underwriting process
type UnderwritingStatus string

const (
	StatusPending    UnderwritingStatus = "pending"
	StatusInProgress UnderwritingStatus = "in_progress"
	StatusCompleted  UnderwritingStatus = "completed"
	StatusOnHold     UnderwritingStatus = "on_hold"
	StatusCancelled  UnderwritingStatus = "cancelled"
)

// CreditScoreRange represents credit score ranges
type CreditScoreRange string

const (
	CreditExcellent CreditScoreRange = "excellent" // 800-850
	CreditVeryGood  CreditScoreRange = "very_good" // 740-799
	CreditGood      CreditScoreRange = "good"      // 670-739
	CreditFair      CreditScoreRange = "fair"      // 580-669
	CreditPoor      CreditScoreRange = "poor"      // 300-579
)

// IncomeVerificationStatus represents income verification status
type IncomeVerificationStatus string

const (
	IncomeVerified   IncomeVerificationStatus = "verified"
	IncomeUnverified IncomeVerificationStatus = "unverified"
	IncomePartial    IncomeVerificationStatus = "partial"
	IncomeFailed     IncomeVerificationStatus = "failed"
)

// LoanApplication represents a loan application for underwriting
type LoanApplication struct {
	ID                       string                   `json:"id" db:"id"`
	UserID                   string                   `json:"user_id" db:"user_id"`
	ApplicationNumber        string                   `json:"application_number" db:"application_number"`
	LoanAmount               float64                  `json:"loan_amount" db:"loan_amount"`
	RequestedTerm            int                      `json:"requested_term_months" db:"requested_term_months"`
	LoanPurpose              string                   `json:"loan_purpose" db:"loan_purpose"`
	AnnualIncome             float64                  `json:"annual_income" db:"annual_income"`
	MonthlyIncome            float64                  `json:"monthly_income" db:"monthly_income"`
	MonthlyDebt              float64                  `json:"monthly_debt_payments" db:"monthly_debt_payments"`
	EmploymentStatus         string                   `json:"employment_status" db:"employment_status"`
	IncomeVerificationStatus IncomeVerificationStatus `json:"income_verification_status" db:"income_verification_status"`
	DTIRatio                 float64                  `json:"dti_ratio" db:"dti_ratio"`
	CurrentState             string                   `json:"current_state" db:"current_state"`
	Status                   string                   `json:"status" db:"status"`
	SubmittedAt              time.Time                `json:"submitted_at" db:"submitted_at"`
	CreatedAt                time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time                `json:"updated_at" db:"updated_at"`
}

// CreditReport represents a credit report from external credit bureau
type CreditReport struct {
	ID                  string                 `json:"id" db:"id"`
	ApplicationID       string                 `json:"application_id" db:"application_id"`
	UserID              string                 `json:"user_id" db:"user_id"`
	CreditScore         int                    `json:"credit_score" db:"credit_score"`
	CreditScoreRange    CreditScoreRange       `json:"credit_score_range" db:"credit_score_range"`
	ReportProvider      string                 `json:"report_provider" db:"report_provider"`
	ReportDate          time.Time              `json:"report_date" db:"report_date"`
	CreditAccounts      []CreditAccount        `json:"credit_accounts"`
	CreditInquiries     []CreditInquiry        `json:"credit_inquiries"`
	PublicRecords       []PublicRecord         `json:"public_records"`
	PaymentHistory      PaymentHistory         `json:"payment_history"`
	CreditUtilization   float64                `json:"credit_utilization" db:"credit_utilization"`
	TotalCreditLimit    float64                `json:"total_credit_limit" db:"total_credit_limit"`
	TotalCurrentBalance float64                `json:"total_current_balance" db:"total_current_balance"`
	DerogatoryCounts    DerogatoryCounts       `json:"derogatory_counts"`
	RiskFactors         []string               `json:"risk_factors"`
	CreditMix           []string               `json:"credit_mix"`
	ReportData          map[string]interface{} `json:"report_data" db:"report_data"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
}

// CreditAccount represents a credit account from credit report
type CreditAccount struct {
	AccountID        string    `json:"account_id"`
	AccountType      string    `json:"account_type"` // credit_card, mortgage, auto_loan, etc.
	Creditor         string    `json:"creditor"`
	OpenDate         time.Time `json:"open_date"`
	LastReportedDate time.Time `json:"last_reported_date"`
	CreditLimit      float64   `json:"credit_limit"`
	CurrentBalance   float64   `json:"current_balance"`
	PaymentStatus    string    `json:"payment_status"`
	PaymentHistory   string    `json:"payment_history"` // 24-month payment history
	MonthlyPayment   float64   `json:"monthly_payment"`
	AccountStatus    string    `json:"account_status"` // open, closed, charge_off, etc.
}

// CreditInquiry represents a credit inquiry
type CreditInquiry struct {
	InquiryID     string    `json:"inquiry_id"`
	InquiryDate   time.Time `json:"inquiry_date"`
	InquiryType   string    `json:"inquiry_type"` // hard, soft
	Creditor      string    `json:"creditor"`
	InquiryReason string    `json:"inquiry_reason"`
}

// PublicRecord represents public records (bankruptcies, liens, etc.)
type PublicRecord struct {
	RecordID     string    `json:"record_id"`
	RecordType   string    `json:"record_type"` // bankruptcy, lien, judgment
	FilingDate   time.Time `json:"filing_date"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	Court        string    `json:"court"`
	DocketNumber string    `json:"docket_number"`
}

// PaymentHistory represents payment history summary
type PaymentHistory struct {
	OnTimePayments  int     `json:"on_time_payments"`
	LatePayments30  int     `json:"late_payments_30"`
	LatePayments60  int     `json:"late_payments_60"`
	LatePayments90  int     `json:"late_payments_90"`
	LatePayments120 int     `json:"late_payments_120_plus"`
	ChargeOffs      int     `json:"charge_offs"`
	Collections     int     `json:"collections"`
	PaymentScore    float64 `json:"payment_score"` // 0-100
}

// DerogatoryCounts represents counts of derogatory items
type DerogatoryCounts struct {
	Bankruptcies int `json:"bankruptcies"`
	Liens        int `json:"liens"`
	Judgments    int `json:"judgments"`
	ChargeOffs   int `json:"charge_offs"`
	Collections  int `json:"collections"`
	LatePayments int `json:"late_payments"`
}

// RiskAssessment represents the risk assessment for an application
type RiskAssessment struct {
	ID                   string                 `json:"id" db:"id"`
	ApplicationID        string                 `json:"application_id" db:"application_id"`
	UserID               string                 `json:"user_id" db:"user_id"`
	OverallRiskLevel     RiskLevel              `json:"overall_risk_level" db:"overall_risk_level"`
	RiskScore            float64                `json:"risk_score" db:"risk_score"`
	CreditRiskScore      float64                `json:"credit_risk_score" db:"credit_risk_score"`
	IncomeRiskScore      float64                `json:"income_risk_score" db:"income_risk_score"`
	DebtRiskScore        float64                `json:"debt_risk_score" db:"debt_risk_score"`
	FraudRiskScore       float64                `json:"fraud_risk_score" db:"fraud_risk_score"`
	RiskFactors          []RiskFactor           `json:"risk_factors"`
	MitigatingFactors    []MitigatingFactor     `json:"mitigating_factors"`
	ProbabilityOfDefault float64                `json:"probability_of_default" db:"probability_of_default"`
	RecommendedAction    string                 `json:"recommended_action" db:"recommended_action"`
	ConfidenceLevel      float64                `json:"confidence_level" db:"confidence_level"`
	ModelVersion         string                 `json:"model_version" db:"model_version"`
	AssessmentData       map[string]interface{} `json:"assessment_data" db:"assessment_data"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
}

// RiskFactor represents a specific risk factor
type RiskFactor struct {
	FactorID    string  `json:"factor_id"`
	FactorType  string  `json:"factor_type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // high, medium, low
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
}

// MitigatingFactor represents factors that reduce risk
type MitigatingFactor struct {
	FactorID    string  `json:"factor_id"`
	FactorType  string  `json:"factor_type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
}

// IncomeVerification represents income verification details
type IncomeVerification struct {
	ID                    string                   `json:"id" db:"id"`
	ApplicationID         string                   `json:"application_id" db:"application_id"`
	UserID                string                   `json:"user_id" db:"user_id"`
	VerificationMethod    string                   `json:"verification_method" db:"verification_method"`
	VerificationStatus    IncomeVerificationStatus `json:"verification_status" db:"verification_status"`
	VerifiedAnnualIncome  float64                  `json:"verified_annual_income" db:"verified_annual_income"`
	VerifiedMonthlyIncome float64                  `json:"verified_monthly_income" db:"verified_monthly_income"`
	EmployerName          string                   `json:"employer_name" db:"employer_name"`
	JobTitle              string                   `json:"job_title" db:"job_title"`
	EmploymentStartDate   time.Time                `json:"employment_start_date" db:"employment_start_date"`
	EmploymentType        string                   `json:"employment_type" db:"employment_type"`
	PayFrequency          string                   `json:"pay_frequency" db:"pay_frequency"`
	LastPayStubDate       time.Time                `json:"last_pay_stub_date" db:"last_pay_stub_date"`
	TaxReturnYear         int                      `json:"tax_return_year" db:"tax_return_year"`
	W2Income              float64                  `json:"w2_income" db:"w2_income"`
	VerificationNotes     string                   `json:"verification_notes" db:"verification_notes"`
	DocumentsProvided     []string                 `json:"documents_provided"`
	VerificationData      map[string]interface{}   `json:"verification_data" db:"verification_data"`
	VerifiedAt            time.Time                `json:"verified_at" db:"verified_at"`
	CreatedAt             time.Time                `json:"created_at" db:"created_at"`
}

// UnderwritingResult represents the final underwriting result
type UnderwritingResult struct {
	ID                   string                  `json:"id" db:"id"`
	ApplicationID        string                  `json:"application_id" db:"application_id"`
	UserID               string                  `json:"user_id" db:"user_id"`
	Decision             UnderwritingDecision    `json:"decision" db:"decision"`
	Status               UnderwritingStatus      `json:"status" db:"status"`
	ApprovedAmount       float64                 `json:"approved_amount" db:"approved_amount"`
	ApprovedTerm         int                     `json:"approved_term_months" db:"approved_term_months"`
	InterestRate         float64                 `json:"interest_rate" db:"interest_rate"`
	APR                  float64                 `json:"apr" db:"apr"`
	MonthlyPayment       float64                 `json:"monthly_payment" db:"monthly_payment"`
	TotalInterest        float64                 `json:"total_interest" db:"total_interest"`
	TotalPayment         float64                 `json:"total_payment" db:"total_payment"`
	Conditions           []UnderwritingCondition `json:"conditions"`
	DecisionReasons      []DecisionReason        `json:"decision_reasons"`
	CounterOfferTerms    *CounterOfferTerms      `json:"counter_offer_terms,omitempty"`
	UnderwriterID        string                  `json:"underwriter_id" db:"underwriter_id"`
	UnderwriterName      string                  `json:"underwriter_name" db:"underwriter_name"`
	AutomatedDecision    bool                    `json:"automated_decision" db:"automated_decision"`
	ManualReviewRequired bool                    `json:"manual_review_required" db:"manual_review_required"`
	PolicyVersion        string                  `json:"policy_version" db:"policy_version"`
	ModelVersion         string                  `json:"model_version" db:"model_version"`
	OfferExpirationDate  time.Time               `json:"offer_expiration_date" db:"offer_expiration_date"`
	DecisionData         map[string]interface{}  `json:"decision_data" db:"decision_data"`
	ProcessingTime       time.Duration           `json:"processing_time"`
	CreatedAt            time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time               `json:"updated_at" db:"updated_at"`
}

// UnderwritingCondition represents conditions for loan approval
type UnderwritingCondition struct {
	ConditionID   string    `json:"condition_id"`
	ConditionType string    `json:"condition_type"` // prior_to_funding, prior_to_closing, ongoing
	Description   string    `json:"description"`
	Priority      string    `json:"priority"` // critical, high, medium, low
	Status        string    `json:"status"`   // pending, satisfied, waived
	DueDate       time.Time `json:"due_date"`
	SatisfiedDate time.Time `json:"satisfied_date,omitempty"`
	Notes         string    `json:"notes"`
}

// DecisionReason represents reasons for underwriting decision
type DecisionReason struct {
	ReasonCode  string  `json:"reason_code"`
	ReasonType  string  `json:"reason_type"` // approval, denial, condition
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // primary, secondary, informational
	Weight      float64 `json:"weight"`
}

// CounterOfferTerms represents alternative loan terms
type CounterOfferTerms struct {
	OfferedAmount   float64   `json:"offered_amount"`
	OfferedTerm     int       `json:"offered_term_months"`
	OfferedRate     float64   `json:"offered_rate"`
	OfferedAPR      float64   `json:"offered_apr"`
	MonthlyPayment  float64   `json:"monthly_payment"`
	TotalInterest   float64   `json:"total_interest"`
	OfferReason     string    `json:"offer_reason"`
	OfferConditions []string  `json:"offer_conditions"`
	ExpirationDate  time.Time `json:"expiration_date"`
}

// UnderwritingPolicy represents underwriting policies and rules
type UnderwritingPolicy struct {
	ID                     string                 `json:"id" db:"id"`
	PolicyName             string                 `json:"policy_name" db:"policy_name"`
	PolicyVersion          string                 `json:"policy_version" db:"policy_version"`
	EffectiveDate          time.Time              `json:"effective_date" db:"effective_date"`
	ExpirationDate         time.Time              `json:"expiration_date" db:"expiration_date"`
	MinCreditScore         int                    `json:"min_credit_score" db:"min_credit_score"`
	MaxDTIRatio            float64                `json:"max_dti_ratio" db:"max_dti_ratio"`
	MinAnnualIncome        float64                `json:"min_annual_income" db:"min_annual_income"`
	MaxLoanAmount          float64                `json:"max_loan_amount" db:"max_loan_amount"`
	MinLoanAmount          float64                `json:"min_loan_amount" db:"min_loan_amount"`
	AllowedLoanTerms       []int                  `json:"allowed_loan_terms"`
	AllowedLoanPurposes    []string               `json:"allowed_loan_purposes"`
	InterestRateMatrix     InterestRateMatrix     `json:"interest_rate_matrix"`
	AutoApprovalThresholds AutoApprovalThresholds `json:"auto_approval_thresholds"`
	ManualReviewTriggers   []string               `json:"manual_review_triggers"`
	PolicyRules            map[string]interface{} `json:"policy_rules" db:"policy_rules"`
	IsActive               bool                   `json:"is_active" db:"is_active"`
	CreatedBy              string                 `json:"created_by" db:"created_by"`
	CreatedAt              time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
}

// InterestRateMatrix represents interest rate based on risk factors
type InterestRateMatrix struct {
	BaseRate          float64                        `json:"base_rate"`
	RateRanges        map[CreditScoreRange]RateRange `json:"rate_ranges"`
	DTIAdjustments    map[string]float64             `json:"dti_adjustments"`
	IncomeAdjustments map[string]float64             `json:"income_adjustments"`
	RiskAdjustments   map[RiskLevel]float64          `json:"risk_adjustments"`
}

// RateRange represents interest rate range
type RateRange struct {
	MinRate float64 `json:"min_rate"`
	MaxRate float64 `json:"max_rate"`
}

// AutoApprovalThresholds represents thresholds for automatic approval
type AutoApprovalThresholds struct {
	MinCreditScore        int      `json:"min_credit_score"`
	MaxDTIRatio           float64  `json:"max_dti_ratio"`
	MinIncomeAmount       float64  `json:"min_income_amount"`
	MaxLoanAmount         float64  `json:"max_loan_amount"`
	MaxRiskScore          float64  `json:"max_risk_score"`
	RequiredVerifications []string `json:"required_verifications"`
}

// UnderwritingWorkflow represents the workflow state
type UnderwritingWorkflow struct {
	ID                  string                 `json:"id" db:"id"`
	ApplicationID       string                 `json:"application_id" db:"application_id"`
	WorkflowID          string                 `json:"workflow_id" db:"workflow_id"`
	CurrentStep         string                 `json:"current_step" db:"current_step"`
	Status              UnderwritingStatus     `json:"status" db:"status"`
	Steps               []WorkflowStep         `json:"steps"`
	StartedAt           time.Time              `json:"started_at" db:"started_at"`
	CompletedAt         *time.Time             `json:"completed_at" db:"completed_at"`
	EstimatedCompletion time.Time              `json:"estimated_completion" db:"estimated_completion"`
	WorkflowData        map[string]interface{} `json:"workflow_data" db:"workflow_data"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// WorkflowStep represents a step in the underwriting workflow
type WorkflowStep struct {
	StepID        string                 `json:"step_id"`
	StepName      string                 `json:"step_name"`
	StepType      string                 `json:"step_type"` // automated, manual, external
	Status        string                 `json:"status"`    // pending, in_progress, completed, failed, skipped
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	AssignedTo    string                 `json:"assigned_to,omitempty"`
	EstimatedTime time.Duration          `json:"estimated_time"`
	ActualTime    time.Duration          `json:"actual_time"`
	Dependencies  []string               `json:"dependencies"`
	Output        map[string]interface{} `json:"output"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	MaxRetries    int                    `json:"max_retries"`
}

// ValidationResult represents validation results
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   map[string]string `json:"errors,omitempty"`
	Warnings map[string]string `json:"warnings,omitempty"`
}

// UnderwritingError represents domain-specific errors
type UnderwritingError struct {
	Code        string
	Message     string
	Description string
	HTTPStatus  int
}

func (e *UnderwritingError) Error() string {
	return e.Message
}

// Error codes for underwriting service
const (
	ErrCodeInvalidApplication       = "UW_001"
	ErrCodeCreditCheckFailed        = "UW_002"
	ErrCodeIncomeVerificationFailed = "UW_003"
	ErrCodeRiskAssessmentFailed     = "UW_004"
	ErrCodePolicyViolation          = "UW_005"
	ErrCodeManualReviewRequired     = "UW_006"
	ErrCodeExternalServiceError     = "UW_007"
	ErrCodeInsufficientData         = "UW_008"
	ErrCodeDecisionEngineError      = "UW_009"
	ErrCodeWorkflowError            = "UW_010"
)

// NewUnderwritingError creates a new underwriting error
func NewUnderwritingError(code, message, description string, httpStatus int) *UnderwritingError {
	return &UnderwritingError{
		Code:        code,
		Message:     message,
		Description: description,
		HTTPStatus:  httpStatus,
	}
}

// Helper methods for validation and business logic

// CalculateDTI calculates debt-to-income ratio
func (app *LoanApplication) CalculateDTI() float64 {
	if app.MonthlyIncome <= 0 {
		return 0
	}
	return app.MonthlyDebt / app.MonthlyIncome
}

// GetCreditScoreRange returns the credit score range based on score
func GetCreditScoreRange(score int) CreditScoreRange {
	switch {
	case score >= 800:
		return CreditExcellent
	case score >= 740:
		return CreditVeryGood
	case score >= 670:
		return CreditGood
	case score >= 580:
		return CreditFair
	default:
		return CreditPoor
	}
}

// GetRiskLevel returns risk level based on risk score
func GetRiskLevel(score float64) RiskLevel {
	switch {
	case score >= 80:
		return RiskCritical
	case score >= 60:
		return RiskHigh
	case score >= 40:
		return RiskMedium
	default:
		return RiskLow
	}
}

// IsExpired checks if an underwriting result offer has expired
func (ur *UnderwritingResult) IsExpired() bool {
	return time.Now().After(ur.OfferExpirationDate)
}

// HasCriticalConditions checks if there are any critical conditions
func (ur *UnderwritingResult) HasCriticalConditions() bool {
	for _, condition := range ur.Conditions {
		if condition.Priority == "critical" && condition.Status == "pending" {
			return true
		}
	}
	return false
}

// GetDerogatoriesCount returns total count of derogatory items
func (cr *CreditReport) GetDerogatoriesCount() int {
	return cr.DerogatoryCounts.Bankruptcies +
		cr.DerogatoryCounts.Liens +
		cr.DerogatoryCounts.Judgments +
		cr.DerogatoryCounts.ChargeOffs +
		cr.DerogatoryCounts.Collections
}

// ValidateApplication validates loan application data
func (app *LoanApplication) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make(map[string]string),
		Warnings: make(map[string]string),
	}

	// Validate required fields
	if app.ID == "" {
		result.Valid = false
		result.Errors["id"] = "Application ID is required"
	}

	if app.UserID == "" {
		result.Valid = false
		result.Errors["user_id"] = "User ID is required"
	}

	if app.LoanAmount <= 0 {
		result.Valid = false
		result.Errors["loan_amount"] = "Loan amount must be greater than 0"
	}

	if app.AnnualIncome <= 0 {
		result.Valid = false
		result.Errors["annual_income"] = "Annual income must be greater than 0"
	}

	if app.RequestedTerm <= 0 {
		result.Valid = false
		result.Errors["requested_term"] = "Loan term must be greater than 0"
	}

	// Validate business rules
	dti := app.CalculateDTI()
	if dti > 0.5 {
		result.Warnings["dti_ratio"] = "DTI ratio is high (>50%)"
	}

	if app.LoanAmount > app.AnnualIncome {
		result.Warnings["loan_to_income"] = "Loan amount exceeds annual income"
	}

	return result
}
