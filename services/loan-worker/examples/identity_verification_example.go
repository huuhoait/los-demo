package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/loan-worker/infrastructure/workflow/tasks"
)

// IdentityVerificationWorkflowExample demonstrates the complete identity verification workflow
type IdentityVerificationWorkflowExample struct {
	logger             *zap.Logger
	identityHandler    *tasks.IdentityVerificationTaskHandler
	updateStateHandler *tasks.UpdateApplicationStateTaskHandler
	verificationConfig *tasks.VerificationServiceConfig
}

// NewIdentityVerificationWorkflowExample creates a new workflow example
func NewIdentityVerificationWorkflowExample(logger *zap.Logger, loanRepository tasks.LoanRepository) *IdentityVerificationWorkflowExample {
	return &IdentityVerificationWorkflowExample{
		logger:             logger,
		identityHandler:    tasks.NewIdentityVerificationTaskHandler(logger),
		updateStateHandler: tasks.NewUpdateApplicationStateTaskHandlerWithRepository(logger, loanRepository),
		verificationConfig: tasks.NewVerificationServiceConfig(logger),
	}
}

// ExecuteCompleteWorkflow demonstrates the complete identity verification workflow
func (e *IdentityVerificationWorkflowExample) ExecuteCompleteWorkflow(ctx context.Context) error {
	e.logger.Info("Starting complete identity verification workflow example")

	// Step 1: Create sample application data
	applicationData := e.createSampleApplicationData()

	// Step 2: Execute identity verification task
	identityResult, err := e.executeIdentityVerification(ctx, applicationData)
	if err != nil {
		return fmt.Errorf("identity verification failed: %w", err)
	}

	// Step 3: Update application state based on verification result
	stateUpdateResult, err := e.executeStateUpdate(ctx, applicationData, identityResult)
	if err != nil {
		return fmt.Errorf("state update failed: %w", err)
	}

	// Step 4: Log complete workflow results
	e.logWorkflowResults(applicationData, identityResult, stateUpdateResult)

	e.logger.Info("Complete identity verification workflow completed successfully")
	return nil
}

// createSampleApplicationData creates sample application data for testing
func (e *IdentityVerificationWorkflowExample) createSampleApplicationData() map[string]interface{} {
	return map[string]interface{}{
		"applicationId": "12345678-1234-1234-1234-123456789012",
		"userId":        "user_98765432-1234-1234-1234-123456789012",
		"personalInfo": map[string]interface{}{
			"first_name":    "John",
			"last_name":     "Doe",
			"date_of_birth": "1990-01-15",
			"ssn":           "123456789",
			"email":         "john.doe@example.com",
			"phone":         "+1234567890",
			"address": map[string]interface{}{
				"street_address": "123 Main Street",
				"city":           "San Francisco",
				"state":          "CA",
				"zip_code":       "94105",
				"country":        "USA",
			},
		},
		"documents": []interface{}{
			map[string]interface{}{
				"type":      "drivers_license",
				"image_url": "https://example.com/dl_front.jpg",
				"metadata":  map[string]interface{}{"side": "front"},
			},
			map[string]interface{}{
				"type":      "drivers_license",
				"image_url": "https://example.com/dl_back.jpg",
				"metadata":  map[string]interface{}{"side": "back"},
			},
			map[string]interface{}{
				"type":      "utility_bill",
				"image_url": "https://example.com/utility_bill.jpg",
				"metadata":  map[string]interface{}{"date": "2024-01-15"},
			},
		},
	}
}

// executeIdentityVerification executes the identity verification task
func (e *IdentityVerificationWorkflowExample) executeIdentityVerification(
	ctx context.Context,
	applicationData map[string]interface{},
) (map[string]interface{}, error) {
	e.logger.Info("Executing identity verification task")

	// Prepare input for identity verification
	input := map[string]interface{}{
		"applicationId": applicationData["applicationId"],
		"userId":        applicationData["userId"],
		"personalInfo":  applicationData["personalInfo"],
		"documents":     applicationData["documents"],
	}

	// Execute identity verification
	result, err := e.identityHandler.Execute(ctx, input)
	if err != nil {
		e.logger.Error("Identity verification failed", zap.Error(err))
		return nil, err
	}

	e.logger.Info("Identity verification completed",
		zap.Bool("verified", result["verified"].(bool)),
		zap.Float64("score", result["verificationScore"].(float64)))

	return result, nil
}

// executeStateUpdate updates the application state based on verification results
func (e *IdentityVerificationWorkflowExample) executeStateUpdate(
	ctx context.Context,
	applicationData map[string]interface{},
	identityResult map[string]interface{},
) (map[string]interface{}, error) {
	e.logger.Info("Executing state update task")

	verified, _ := identityResult["verified"].(bool)
	var toState, reason string

	if verified {
		toState = "identity_verified"
		reason = "Identity verification completed successfully"
	} else {
		toState = "manual_review"
		reason = "Identity verification requires manual review"
	}

	// Prepare input for state update
	input := map[string]interface{}{
		"applicationId": applicationData["applicationId"],
		"fromState":     "documents_submitted",
		"toState":       toState,
		"reason":        reason,
		"automated":     true,
		"metadata": map[string]interface{}{
			"verification_score":   identityResult["verificationScore"],
			"verification_methods": identityResult["verificationMethods"],
			"risk_flags":           identityResult["riskFlags"],
		},
	}

	// Execute state update
	result, err := e.updateStateHandler.Execute(ctx, input)
	if err != nil {
		e.logger.Error("State update failed", zap.Error(err))
		return nil, err
	}

	e.logger.Info("State update completed",
		zap.String("new_state", result["newState"].(string)),
		zap.Bool("success", result["success"].(bool)))

	return result, nil
}

// logWorkflowResults logs the complete workflow results
func (e *IdentityVerificationWorkflowExample) logWorkflowResults(
	applicationData map[string]interface{},
	identityResult map[string]interface{},
	stateResult map[string]interface{},
) {
	e.logger.Info("=== IDENTITY VERIFICATION WORKFLOW RESULTS ===")

	// Application Info
	e.logger.Info("Application Details",
		zap.String("application_id", applicationData["applicationId"].(string)),
		zap.String("user_id", applicationData["userId"].(string)))

	// Verification Results
	verified, _ := identityResult["verified"].(bool)
	score, _ := identityResult["verificationScore"].(float64)
	riskFlags, _ := identityResult["riskFlags"].([]string)

	e.logger.Info("Identity Verification Results",
		zap.Bool("verified", verified),
		zap.Float64("verification_score", score),
		zap.Strings("risk_flags", riskFlags))

	// Verification Details
	if verificationDetails, ok := identityResult["verificationDetails"].(map[string]interface{}); ok {
		if methodResults, ok := verificationDetails["method_results"].(map[string]interface{}); ok {
			for method, result := range methodResults {
				if methodResult, ok := result.(map[string]interface{}); ok {
					methodScore, _ := methodResult["score"].(float64)
					methodStatus, _ := methodResult["status"].(string)
					e.logger.Info("Verification Method Result",
						zap.String("method", method),
						zap.Float64("score", methodScore),
						zap.String("status", methodStatus))
				}
			}
		}
	}

	// State Update Results
	newState, _ := stateResult["newState"].(string)
	previousState, _ := stateResult["previousState"].(string)
	success, _ := stateResult["success"].(bool)

	e.logger.Info("State Update Results",
		zap.Bool("update_success", success),
		zap.String("previous_state", previousState),
		zap.String("new_state", newState))

	e.logger.Info("=== WORKFLOW COMPLETED ===")
}

// DemonstrateTaskIntegration shows how tasks integrate with external services
func (e *IdentityVerificationWorkflowExample) DemonstrateTaskIntegration(ctx context.Context) error {
	e.logger.Info("Demonstrating task integration with external services")

	// 1. Check service health
	healthStatus := e.verificationConfig.CheckServicesHealth(ctx)
	e.logger.Info("Verification services health check completed",
		zap.Int("services_checked", len(healthStatus)))

	for _, status := range healthStatus {
		e.logger.Info("Service health status",
			zap.String("service", status.Service),
			zap.Bool("available", status.Available),
			zap.Duration("latency", status.Latency))
	}

	// 2. Show configuration
	config := e.verificationConfig.GetVerificationConfig()
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	e.logger.Info("Current verification configuration",
		zap.String("config", string(configJSON)))

	// 3. Demonstrate different verification levels
	return e.demonstrateVerificationLevels(ctx)
}

// demonstrateVerificationLevels shows different verification scenarios
func (e *IdentityVerificationWorkflowExample) demonstrateVerificationLevels(ctx context.Context) error {
	scenarios := []struct {
		name         string
		documents    []interface{}
		personalInfo map[string]interface{}
		expectedTier string
	}{
		{
			name: "Premium Verification - Complete Documents",
			documents: []interface{}{
				map[string]interface{}{"type": "drivers_license", "image_url": "front.jpg"},
				map[string]interface{}{"type": "drivers_license", "image_url": "back.jpg"},
				map[string]interface{}{"type": "passport", "image_url": "passport.jpg"},
				map[string]interface{}{"type": "utility_bill", "image_url": "bill.jpg"},
			},
			personalInfo: map[string]interface{}{
				"ssn": "123456789",
				"address": map[string]interface{}{
					"street_address": "123 Main St",
					"city":           "San Francisco",
					"state":          "CA",
					"zip_code":       "94105",
				},
				"biometric_data": map[string]interface{}{"face_scan": "available"},
			},
			expectedTier: "premium",
		},
		{
			name: "Standard Verification - Basic Documents",
			documents: []interface{}{
				map[string]interface{}{"type": "drivers_license", "image_url": "front.jpg"},
				map[string]interface{}{"type": "utility_bill", "image_url": "bill.jpg"},
			},
			personalInfo: map[string]interface{}{
				"ssn": "123456789",
				"address": map[string]interface{}{
					"street_address": "123 Main St",
					"city":           "San Francisco",
					"state":          "CA",
					"zip_code":       "94105",
				},
			},
			expectedTier: "standard",
		},
		{
			name: "Basic Verification - Minimal Documents",
			documents: []interface{}{
				map[string]interface{}{"type": "drivers_license", "image_url": "front.jpg"},
			},
			personalInfo: map[string]interface{}{
				"ssn": "123456789",
				"address": map[string]interface{}{
					"street_address": "123 Main St",
					"city":           "San Francisco",
				},
			},
			expectedTier: "basic",
		},
	}

	for _, scenario := range scenarios {
		e.logger.Info("Testing verification scenario", zap.String("scenario", scenario.name))

		input := map[string]interface{}{
			"applicationId": fmt.Sprintf("test_%d", time.Now().Unix()),
			"userId":        fmt.Sprintf("user_%d", time.Now().Unix()),
			"personalInfo":  scenario.personalInfo,
			"documents":     scenario.documents,
		}

		result, err := e.identityHandler.Execute(ctx, input)
		if err != nil {
			e.logger.Error("Verification scenario failed",
				zap.String("scenario", scenario.name),
				zap.Error(err))
			continue
		}

		verified, _ := result["verified"].(bool)
		score, _ := result["verificationScore"].(float64)

		if verificationDetails, ok := result["verificationDetails"].(map[string]interface{}); ok {
			tier, _ := verificationDetails["verification_tier"].(string)
			confidence, _ := verificationDetails["confidence_level"].(string)

			e.logger.Info("Verification scenario results",
				zap.String("scenario", scenario.name),
				zap.Bool("verified", verified),
				zap.Float64("score", score),
				zap.String("tier", tier),
				zap.String("confidence", confidence),
				zap.String("expected_tier", scenario.expectedTier))
		}
	}

	return nil
}

// GetTaskHandlers returns the task handlers for external use
func (e *IdentityVerificationWorkflowExample) GetTaskHandlers() (
	*tasks.IdentityVerificationTaskHandler,
	*tasks.UpdateApplicationStateTaskHandler,
) {
	return e.identityHandler, e.updateStateHandler
}

// GetVerificationConfig returns the verification configuration
func (e *IdentityVerificationWorkflowExample) GetVerificationConfig() *tasks.VerificationServiceConfig {
	return e.verificationConfig
}
