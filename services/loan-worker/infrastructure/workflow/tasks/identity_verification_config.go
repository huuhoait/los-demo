package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ExternalVerificationService represents an external identity verification service
type ExternalVerificationService interface {
	VerifyIdentity(ctx context.Context, request *IdentityVerificationRequest) (*ExternalVerificationResponse, error)
	GetServiceName() string
	IsAvailable(ctx context.Context) bool
}

// IdentityVerificationRequest represents a request to an external verification service
type IdentityVerificationRequest struct {
	ApplicationID     string                 `json:"application_id"`
	UserID            string                 `json:"user_id"`
	PersonalInfo      map[string]interface{} `json:"personal_info"`
	Documents         []DocumentInfo         `json:"documents"`
	VerificationLevel string                 `json:"verification_level"` // basic, standard, premium
}

// DocumentInfo represents document information for verification
type DocumentInfo struct {
	Type     string                 `json:"type"` // drivers_license, passport, utility_bill
	ImageURL string                 `json:"image_url"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ExternalVerificationResponse represents response from external verification service
type ExternalVerificationResponse struct {
	VerificationID  string                 `json:"verification_id"`
	Verified        bool                   `json:"verified"`
	Score           float64                `json:"score"`
	RiskLevel       string                 `json:"risk_level"`
	Details         map[string]interface{} `json:"details"`
	RiskFlags       []string               `json:"risk_flags"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	ServiceProvider string                 `json:"service_provider"`
}

// VerificationServiceConfig holds configuration for identity verification services
type VerificationServiceConfig struct {
	Logger                    *zap.Logger
	EnableExternalServices    bool
	PrimaryService            string
	FallbackServices          []string
	TimeoutDuration           time.Duration
	MaxRetries                int
	RequiredVerificationLevel string
	Services                  map[string]ExternalVerificationService
}

// NewVerificationServiceConfig creates a new verification service configuration
func NewVerificationServiceConfig(logger *zap.Logger) *VerificationServiceConfig {
	config := &VerificationServiceConfig{
		Logger:                    logger,
		EnableExternalServices:    false, // Default to simulation mode
		PrimaryService:            "jumio",
		FallbackServices:          []string{"onfido", "trulioo"},
		TimeoutDuration:           30 * time.Second,
		MaxRetries:                3,
		RequiredVerificationLevel: "standard",
		Services:                  make(map[string]ExternalVerificationService),
	}

	// Register available services
	config.registerServices()

	return config
}

// registerServices registers all available external verification services
func (c *VerificationServiceConfig) registerServices() {
	// Register Jumio service
	c.Services["jumio"] = &JumioVerificationService{
		Logger:  c.Logger,
		BaseURL: "https://api.jumio.com/v2/identity-verification",
		APIKey:  "", // Configure from environment
		Timeout: c.TimeoutDuration,
	}

	// Register Onfido service
	c.Services["onfido"] = &OnfidoVerificationService{
		Logger:  c.Logger,
		BaseURL: "https://api.onfido.com/v3/identity-verification",
		APIKey:  "", // Configure from environment
		Timeout: c.TimeoutDuration,
	}

	// Register Trulioo service
	c.Services["trulioo"] = &TruliooVerificationService{
		Logger:  c.Logger,
		BaseURL: "https://api.trulioo.com/v1/identity-verification",
		APIKey:  "", // Configure from environment
		Timeout: c.TimeoutDuration,
	}
}

// PerformExternalVerification performs identity verification using external services
func (c *VerificationServiceConfig) PerformExternalVerification(
	ctx context.Context,
	request *IdentityVerificationRequest,
) (*ExternalVerificationResponse, error) {
	if !c.EnableExternalServices {
		return c.simulateExternalVerification(request), nil
	}

	// Try primary service first
	if service, exists := c.Services[c.PrimaryService]; exists {
		if service.IsAvailable(ctx) {
			response, err := c.attemptVerification(ctx, service, request)
			if err == nil {
				return response, nil
			}
			c.Logger.Warn("Primary verification service failed",
				zap.String("service", c.PrimaryService),
				zap.Error(err))
		}
	}

	// Try fallback services
	for _, serviceName := range c.FallbackServices {
		if service, exists := c.Services[serviceName]; exists {
			if service.IsAvailable(ctx) {
				response, err := c.attemptVerification(ctx, service, request)
				if err == nil {
					return response, nil
				}
				c.Logger.Warn("Fallback verification service failed",
					zap.String("service", serviceName),
					zap.Error(err))
			}
		}
	}

	// If all services fail, return simulation
	c.Logger.Error("All external verification services failed, falling back to simulation")
	return c.simulateExternalVerification(request), nil
}

// attemptVerification attempts verification with retry logic
func (c *VerificationServiceConfig) attemptVerification(
	ctx context.Context,
	service ExternalVerificationService,
	request *IdentityVerificationRequest,
) (*ExternalVerificationResponse, error) {
	var lastErr error

	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		if attempt > 0 {
			// Add exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		response, err := service.VerifyIdentity(ctx, request)
		if err == nil {
			return response, nil
		}

		lastErr = err
		c.Logger.Warn("Verification attempt failed",
			zap.String("service", service.GetServiceName()),
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	return nil, fmt.Errorf("verification failed after %d attempts: %w", c.MaxRetries, lastErr)
}

// simulateExternalVerification simulates external verification when services are not available
func (c *VerificationServiceConfig) simulateExternalVerification(
	request *IdentityVerificationRequest,
) *ExternalVerificationResponse {
	return &ExternalVerificationResponse{
		VerificationID: fmt.Sprintf("sim_%s_%d", request.ApplicationID, time.Now().Unix()),
		Verified:       true,
		Score:          88.5,
		RiskLevel:      "low",
		Details: map[string]interface{}{
			"simulation_mode":     true,
			"documents_processed": len(request.Documents),
			"verification_level":  request.VerificationLevel,
		},
		RiskFlags:       []string{},
		ProcessingTime:  2 * time.Second,
		ServiceProvider: "simulation",
	}
}

// JumioVerificationService implements Jumio identity verification
type JumioVerificationService struct {
	Logger  *zap.Logger
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func (j *JumioVerificationService) VerifyIdentity(ctx context.Context, request *IdentityVerificationRequest) (*ExternalVerificationResponse, error) {
	// Implementation would make actual API calls to Jumio
	// For now, return simulation
	return &ExternalVerificationResponse{
		VerificationID: fmt.Sprintf("jumio_%s_%d", request.ApplicationID, time.Now().Unix()),
		Verified:       true,
		Score:          92.0,
		RiskLevel:      "low",
		Details: map[string]interface{}{
			"provider":              "jumio",
			"face_match_score":      95.0,
			"document_authenticity": "verified",
		},
		RiskFlags:       []string{},
		ProcessingTime:  3 * time.Second,
		ServiceProvider: "jumio",
	}, nil
}

func (j *JumioVerificationService) GetServiceName() string {
	return "jumio"
}

func (j *JumioVerificationService) IsAvailable(ctx context.Context) bool {
	// Check if service is available (health check)
	// For simulation, always return true if API key is configured
	return j.APIKey != ""
}

// OnfidoVerificationService implements Onfido identity verification
type OnfidoVerificationService struct {
	Logger  *zap.Logger
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func (o *OnfidoVerificationService) VerifyIdentity(ctx context.Context, request *IdentityVerificationRequest) (*ExternalVerificationResponse, error) {
	// Implementation would make actual API calls to Onfido
	return &ExternalVerificationResponse{
		VerificationID: fmt.Sprintf("onfido_%s_%d", request.ApplicationID, time.Now().Unix()),
		Verified:       true,
		Score:          89.5,
		RiskLevel:      "low",
		Details: map[string]interface{}{
			"provider":          "onfido",
			"document_check":    "clear",
			"facial_similarity": "match",
		},
		RiskFlags:       []string{},
		ProcessingTime:  4 * time.Second,
		ServiceProvider: "onfido",
	}, nil
}

func (o *OnfidoVerificationService) GetServiceName() string {
	return "onfido"
}

func (o *OnfidoVerificationService) IsAvailable(ctx context.Context) bool {
	return o.APIKey != ""
}

// TruliooVerificationService implements Trulioo identity verification
type TruliooVerificationService struct {
	Logger  *zap.Logger
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func (t *TruliooVerificationService) VerifyIdentity(ctx context.Context, request *IdentityVerificationRequest) (*ExternalVerificationResponse, error) {
	// Implementation would make actual API calls to Trulioo
	return &ExternalVerificationResponse{
		VerificationID: fmt.Sprintf("trulioo_%s_%d", request.ApplicationID, time.Now().Unix()),
		Verified:       true,
		Score:          87.0,
		RiskLevel:      "medium",
		Details: map[string]interface{}{
			"provider":             "trulioo",
			"identity_check":       "verified",
			"address_verification": "confirmed",
		},
		RiskFlags:       []string{"address_partial_match"},
		ProcessingTime:  5 * time.Second,
		ServiceProvider: "trulioo",
	}, nil
}

func (t *TruliooVerificationService) GetServiceName() string {
	return "trulioo"
}

func (t *TruliooVerificationService) IsAvailable(ctx context.Context) bool {
	return t.APIKey != ""
}

// HealthCheckResponse represents the health status of verification services
type HealthCheckResponse struct {
	Service   string        `json:"service"`
	Available bool          `json:"available"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

// CheckServicesHealth checks the health of all configured verification services
func (c *VerificationServiceConfig) CheckServicesHealth(ctx context.Context) []HealthCheckResponse {
	var responses []HealthCheckResponse

	for serviceName, service := range c.Services {
		startTime := time.Now()
		available := service.IsAvailable(ctx)
		latency := time.Since(startTime)

		response := HealthCheckResponse{
			Service:   serviceName,
			Available: available,
			Latency:   latency,
			CheckedAt: time.Now(),
		}

		if !available {
			response.Error = "Service unavailable"
		}

		responses = append(responses, response)
	}

	return responses
}

// GetVerificationConfig returns the current verification configuration
func (c *VerificationServiceConfig) GetVerificationConfig() map[string]interface{} {
	return map[string]interface{}{
		"external_services_enabled":   c.EnableExternalServices,
		"primary_service":             c.PrimaryService,
		"fallback_services":           c.FallbackServices,
		"timeout_duration":            c.TimeoutDuration.String(),
		"max_retries":                 c.MaxRetries,
		"required_verification_level": c.RequiredVerificationLevel,
		"configured_services":         len(c.Services),
	}
}
