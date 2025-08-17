package domain

import (
	"context"
	"time"
)

// LoanApplicationRepository defines the interface for loan application data access
type LoanApplicationRepository interface {
	GetByID(ctx context.Context, id string) (*LoanApplication, error)
	GetByApplicationNumber(ctx context.Context, applicationNumber string) (*LoanApplication, error)
	Update(ctx context.Context, app *LoanApplication) error
	UpdateStatus(ctx context.Context, id string, status string) error
	List(ctx context.Context, filter ApplicationFilter) ([]*LoanApplication, error)
	Count(ctx context.Context, filter ApplicationFilter) (int, error)
}

// CreditReportRepository defines the interface for credit report data access
type CreditReportRepository interface {
	Create(ctx context.Context, report *CreditReport) error
	GetByApplicationID(ctx context.Context, applicationID string) (*CreditReport, error)
	GetByUserID(ctx context.Context, userID string) ([]*CreditReport, error)
	Update(ctx context.Context, report *CreditReport) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter CreditReportFilter) ([]*CreditReport, error)
}

// RiskAssessmentRepository defines the interface for risk assessment data access
type RiskAssessmentRepository interface {
	Create(ctx context.Context, assessment *RiskAssessment) error
	GetByApplicationID(ctx context.Context, applicationID string) (*RiskAssessment, error)
	GetByUserID(ctx context.Context, userID string) ([]*RiskAssessment, error)
	Update(ctx context.Context, assessment *RiskAssessment) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter RiskAssessmentFilter) ([]*RiskAssessment, error)
}

// IncomeVerificationRepository defines the interface for income verification data access
type IncomeVerificationRepository interface {
	Create(ctx context.Context, verification *IncomeVerification) error
	GetByApplicationID(ctx context.Context, applicationID string) (*IncomeVerification, error)
	GetByUserID(ctx context.Context, userID string) ([]*IncomeVerification, error)
	Update(ctx context.Context, verification *IncomeVerification) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter IncomeVerificationFilter) ([]*IncomeVerification, error)
}

// UnderwritingResultRepository defines the interface for underwriting results data access
type UnderwritingResultRepository interface {
	Create(ctx context.Context, result *UnderwritingResult) error
	GetByApplicationID(ctx context.Context, applicationID string) (*UnderwritingResult, error)
	GetByUserID(ctx context.Context, userID string) ([]*UnderwritingResult, error)
	GetByID(ctx context.Context, id string) (*UnderwritingResult, error)
	Update(ctx context.Context, result *UnderwritingResult) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter UnderwritingResultFilter) ([]*UnderwritingResult, error)
	GetPendingReviews(ctx context.Context) ([]*UnderwritingResult, error)
	GetApprovedOffers(ctx context.Context, userID string) ([]*UnderwritingResult, error)
}

// UnderwritingPolicyRepository defines the interface for underwriting policies data access
type UnderwritingPolicyRepository interface {
	Create(ctx context.Context, policy *UnderwritingPolicy) error
	GetByID(ctx context.Context, id string) (*UnderwritingPolicy, error)
	GetActive(ctx context.Context) (*UnderwritingPolicy, error)
	GetByVersion(ctx context.Context, version string) (*UnderwritingPolicy, error)
	Update(ctx context.Context, policy *UnderwritingPolicy) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter PolicyFilter) ([]*UnderwritingPolicy, error)
	SetActive(ctx context.Context, id string) error
}

// UnderwritingWorkflowRepository defines the interface for workflow data access
type UnderwritingWorkflowRepository interface {
	Create(ctx context.Context, workflow *UnderwritingWorkflow) error
	GetByApplicationID(ctx context.Context, applicationID string) (*UnderwritingWorkflow, error)
	GetByWorkflowID(ctx context.Context, workflowID string) (*UnderwritingWorkflow, error)
	Update(ctx context.Context, workflow *UnderwritingWorkflow) error
	UpdateStep(ctx context.Context, workflowID string, stepID string, step WorkflowStep) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter WorkflowFilter) ([]*UnderwritingWorkflow, error)
	GetActiveWorkflows(ctx context.Context) ([]*UnderwritingWorkflow, error)
}

// CreditBureauService defines the interface for credit bureau integration
type CreditBureauService interface {
	GetCreditReport(ctx context.Context, request *CreditReportRequest) (*CreditReport, error)
	GetCreditScore(ctx context.Context, userID string, ssn string) (*CreditScore, error)
	RefreshCreditReport(ctx context.Context, applicationID string) (*CreditReport, error)
	GetServiceName() string
	IsAvailable(ctx context.Context) bool
	GetRateLimits() RateLimits
}

// RiskScoringService defines the interface for risk scoring
type RiskScoringService interface {
	CalculateRiskScore(ctx context.Context, application *LoanApplication, creditReport *CreditReport) (*RiskAssessment, error)
	CalculateFraudScore(ctx context.Context, application *LoanApplication) (*FraudScore, error)
	GetRiskFactors(ctx context.Context, application *LoanApplication) ([]RiskFactor, error)
	GetServiceName() string
	GetModelVersion() string
	IsAvailable(ctx context.Context) bool
}

// IncomeVerificationService defines the interface for income verification
type IncomeVerificationService interface {
	VerifyIncome(ctx context.Context, request *IncomeVerificationRequest) (*IncomeVerification, error)
	VerifyEmployment(ctx context.Context, request *EmploymentVerificationRequest) (*EmploymentVerification, error)
	GetSupportedVerificationMethods() []string
	GetServiceName() string
	IsAvailable(ctx context.Context) bool
}

// DecisionEngineService defines the interface for decision engine
type DecisionEngineService interface {
	MakeDecision(ctx context.Context, request *DecisionRequest) (*DecisionResponse, error)
	CalculateInterestRate(ctx context.Context, request *InterestRateRequest) (*InterestRateResponse, error)
	ApplyPolicy(ctx context.Context, application *LoanApplication, policy *UnderwritingPolicy) (*PolicyResult, error)
	GetServiceName() string
	GetPolicyVersion() string
	IsAvailable(ctx context.Context) bool
}

// WorkflowOrchestrator defines the interface for workflow orchestration
type WorkflowOrchestrator interface {
	StartUnderwritingWorkflow(ctx context.Context, applicationID string) (*WorkflowExecution, error)
	GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
	UpdateWorkflowData(ctx context.Context, workflowID string, data map[string]interface{}) error
	CancelWorkflow(ctx context.Context, workflowID string, reason string) error
	RetryWorkflow(ctx context.Context, workflowID string) error
	GetServiceName() string
	IsAvailable(ctx context.Context) bool
}

// NotificationService defines the interface for notifications
type NotificationService interface {
	SendUnderwritingUpdate(ctx context.Context, notification *UnderwritingNotification) error
	SendDecisionNotification(ctx context.Context, notification *DecisionNotification) error
	SendManualReviewAlert(ctx context.Context, alert *ManualReviewAlert) error
	GetServiceName() string
	IsAvailable(ctx context.Context) bool
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	LogUnderwritingEvent(ctx context.Context, event *UnderwritingEvent) error
	LogDecisionEvent(ctx context.Context, event *DecisionEvent) error
	LogAccessEvent(ctx context.Context, event *AccessEvent) error
	GetEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error)
}

// Filter types for repositories
type ApplicationFilter struct {
	UserID        string
	Status        string
	State         string
	DateFrom      *time.Time
	DateTo        *time.Time
	LoanAmountMin *float64
	LoanAmountMax *float64
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type CreditReportFilter struct {
	UserID        string
	ApplicationID string
	Provider      string
	DateFrom      *time.Time
	DateTo        *time.Time
	MinScore      *int
	MaxScore      *int
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type RiskAssessmentFilter struct {
	UserID        string
	ApplicationID string
	RiskLevel     string
	DateFrom      *time.Time
	DateTo        *time.Time
	MinScore      *float64
	MaxScore      *float64
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type IncomeVerificationFilter struct {
	UserID        string
	ApplicationID string
	Status        string
	Method        string
	DateFrom      *time.Time
	DateTo        *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type UnderwritingResultFilter struct {
	UserID        string
	ApplicationID string
	Decision      string
	Status        string
	UnderwriterID string
	DateFrom      *time.Time
	DateTo        *time.Time
	Automated     *bool
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type PolicyFilter struct {
	PolicyName string
	Version    string
	IsActive   *bool
	DateFrom   *time.Time
	DateTo     *time.Time
	CreatedBy  string
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

type WorkflowFilter struct {
	ApplicationID string
	Status        string
	CurrentStep   string
	DateFrom      *time.Time
	DateTo        *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

type AuditFilter struct {
	UserID        string
	ApplicationID string
	EventType     string
	DateFrom      *time.Time
	DateTo        *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}

// Service request/response types
type CreditReportRequest struct {
	UserID        string
	ApplicationID string
	SSN           string
	FirstName     string
	LastName      string
	DateOfBirth   time.Time
	Address       Address
	ReportType    string // full, summary, monitoring
	Permissible   string // loan_application, account_review, etc.
}

type CreditScore struct {
	Score        int
	ScoreRange   CreditScoreRange
	Provider     string
	ScoreDate    time.Time
	ScoreFactors []string
}

type RateLimits struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstLimit        int
}

type FraudScore struct {
	Score        float64
	RiskLevel    RiskLevel
	Indicators   []FraudIndicator
	Confidence   float64
	ModelVersion string
}

type FraudIndicator struct {
	Type        string
	Description string
	Severity    string
	Score       float64
}

type IncomeVerificationRequest struct {
	UserID             string
	ApplicationID      string
	EmployerName       string
	EmployerPhone      string
	EmployerAddress    Address
	JobTitle           string
	HireDate           time.Time
	AnnualSalary       float64
	PayFrequency       string
	VerificationMethod string // voe, paystub, tax_return, bank_statement
	Documents          []Document
}

type EmploymentVerificationRequest struct {
	UserID          string
	ApplicationID   string
	EmployerName    string
	EmployerContact string
	JobTitle        string
	EmploymentType  string
	StartDate       time.Time
	EndDate         *time.Time
	Salary          float64
}

type EmploymentVerification struct {
	Verified           bool
	EmployerName       string
	JobTitle           string
	EmploymentType     string
	StartDate          time.Time
	EndDate            *time.Time
	Salary             float64
	Status             string
	VerifiedAt         time.Time
	VerificationMethod string
	Notes              string
}

type DecisionRequest struct {
	ApplicationID      string
	LoanApplication    *LoanApplication
	CreditReport       *CreditReport
	RiskAssessment     *RiskAssessment
	IncomeVerification *IncomeVerification
	Policy             *UnderwritingPolicy
	RequestedAmount    float64
	RequestedTerm      int
	Purpose            string
}

type DecisionResponse struct {
	Decision             UnderwritingDecision
	ApprovedAmount       float64
	ApprovedTerm         int
	InterestRate         float64
	APR                  float64
	MonthlyPayment       float64
	Conditions           []UnderwritingCondition
	Reasons              []DecisionReason
	CounterOffer         *CounterOfferTerms
	ManualReviewRequired bool
	PolicyVersion        string
	DecisionData         map[string]interface{}
	ProcessingTime       time.Duration
}

type InterestRateRequest struct {
	CreditScore    int
	DTIRatio       float64
	LoanAmount     float64
	LoanTerm       int
	RiskScore      float64
	LoanPurpose    string
	IncomeVerified bool
}

type InterestRateResponse struct {
	BaseRate       float64
	AdjustedRate   float64
	APR            float64
	RateFactors    []RateFactor
	MonthlyPayment float64
	TotalInterest  float64
	TotalPayment   float64
}

type RateFactor struct {
	Factor      string
	Adjustment  float64
	Description string
}

type PolicyResult struct {
	Compliant       bool
	Violations      []PolicyViolation
	Recommendations []PolicyRecommendation
	AutoApproval    bool
	ManualReview    bool
}

type PolicyViolation struct {
	RuleID      string
	RuleName    string
	Description string
	Severity    string
	Value       interface{}
	Threshold   interface{}
}

type PolicyRecommendation struct {
	Type        string
	Description string
	Priority    string
	Action      string
}

type WorkflowExecution struct {
	WorkflowID     string
	ExecutionID    string
	Status         string
	StartTime      time.Time
	Input          map[string]interface{}
	Output         map[string]interface{}
	CurrentTask    string
	CompletedTasks []string
	FailedTasks    []string
}

type WorkflowStatus struct {
	WorkflowID          string
	ExecutionID         string
	Status              string
	CurrentStep         string
	Progress            float64
	StartTime           time.Time
	EndTime             *time.Time
	EstimatedCompletion *time.Time
	ExecutionData       map[string]interface{}
	ErrorMessage        string
}

// Notification types
type UnderwritingNotification struct {
	ApplicationID    string
	UserID           string
	NotificationType string
	Title            string
	Message          string
	Data             map[string]interface{}
	Channels         []string // email, sms, push, webhook
	Priority         string
	ScheduledFor     *time.Time
}

type DecisionNotification struct {
	ApplicationID   string
	UserID          string
	Decision        UnderwritingDecision
	DecisionReasons []DecisionReason
	ApprovedAmount  float64
	InterestRate    float64
	MonthlyPayment  float64
	Conditions      []UnderwritingCondition
	CounterOffer    *CounterOfferTerms
	ExpirationDate  time.Time
	NextSteps       []string
}

type ManualReviewAlert struct {
	ApplicationID   string
	UserID          string
	ReviewType      string
	Priority        string
	Reason          string
	AssignedTo      string
	DueDate         time.Time
	ReviewData      map[string]interface{}
	EscalationRules []EscalationRule
}

type EscalationRule struct {
	Condition  string
	Action     string
	DelayHours int
	AssignTo   string
}

// Audit event types
type UnderwritingEvent struct {
	EventID       string
	ApplicationID string
	UserID        string
	EventType     string
	EventData     map[string]interface{}
	Timestamp     time.Time
	UnderwriterID string
	IPAddress     string
	UserAgent     string
}

type DecisionEvent struct {
	EventID       string
	ApplicationID string
	UserID        string
	Decision      UnderwritingDecision
	DecisionData  map[string]interface{}
	PolicyVersion string
	ModelVersion  string
	Timestamp     time.Time
	UnderwriterID string
	Automated     bool
}

type AccessEvent struct {
	EventID       string
	UserID        string
	ApplicationID string
	Action        string
	Resource      string
	IPAddress     string
	UserAgent     string
	Timestamp     time.Time
	Success       bool
	ErrorMessage  string
}

type AuditEvent struct {
	ID            string
	EventType     string
	ApplicationID string
	UserID        string
	EventData     map[string]interface{}
	Timestamp     time.Time
	Source        string
	IPAddress     string
	UserAgent     string
}

// Address represents an address
type Address struct {
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
	Country       string `json:"country"`
}

// Document represents a document
type Document struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	URL      string                 `json:"url"`
	Size     int64                  `json:"size"`
	MimeType string                 `json:"mime_type"`
	Metadata map[string]interface{} `json:"metadata"`
}
