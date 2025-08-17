//go:build !conductor && !test_conductor && !test_workflow
// +build !conductor,!test_conductor,!test_workflow

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("🚀 Starting Underwriting Worker Service")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Environment: development")
	fmt.Println()

	// Initialize mock task worker
	worker := NewMockUnderwritingWorker()

	// Start the worker
	ctx := context.Background()
	if err := worker.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	fmt.Println("✅ Underwriting worker started successfully!")
	fmt.Println("📋 Registered Tasks:")
	fmt.Println("   - credit_check")
	fmt.Println("   - income_verification")
	fmt.Println("   - risk_assessment")
	fmt.Println("   - underwriting_decision")
	fmt.Println("   - update_application_state")
	fmt.Println("   - policy_compliance_check")
	fmt.Println("   - fraud_detection")
	fmt.Println("   - calculate_interest_rate")
	fmt.Println("   - final_approval")
	fmt.Println("   - process_denial")
	fmt.Println("   - assign_manual_review")
	fmt.Println("   - process_conditional_approval")
	fmt.Println("   - generate_counter_offer")
	fmt.Println()

	// Simulate task processing
	go worker.SimulateTaskProcessing()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down underwriting worker...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := worker.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("✅ Underwriting worker exited cleanly")
}

// MockUnderwritingWorker represents a mock implementation
type MockUnderwritingWorker struct {
	running bool
}

// NewMockUnderwritingWorker creates a new mock worker
func NewMockUnderwritingWorker() *MockUnderwritingWorker {
	return &MockUnderwritingWorker{}
}

// Start starts the mock worker
func (w *MockUnderwritingWorker) Start(ctx context.Context) error {
	w.running = true
	return nil
}

// Stop stops the mock worker
func (w *MockUnderwritingWorker) Stop(ctx context.Context) error {
	w.running = false
	return nil
}

// SimulateTaskProcessing simulates processing underwriting tasks
func (w *MockUnderwritingWorker) SimulateTaskProcessing() {
	tasks := []string{
		"credit_check",
		"income_verification",
		"risk_assessment",
		"underwriting_decision",
	}

	applications := []string{"APP-001", "APP-002", "APP-003"}

	for w.running {
		time.Sleep(10 * time.Second)

		if !w.running {
			break
		}

		// Simulate processing a random application
		appID := applications[time.Now().Second()%len(applications)]
		task := tasks[time.Now().Second()%len(tasks)]

		fmt.Printf("🔄 Processing %s for %s...\n", task, appID)

		// Simulate processing time
		time.Sleep(2 * time.Second)

		result := "✅ COMPLETED"
		if time.Now().Second()%7 == 0 {
			result = "⚠️  MANUAL_REVIEW_REQUIRED"
		}

		fmt.Printf("   %s %s completed: %s\n", result, task, appID)

		// Simulate different outcomes
		switch task {
		case "credit_check":
			fmt.Printf("   Credit Score: %d, Risk Level: %s\n",
				600+time.Now().Second()%200, getRiskLevel(time.Now().Second()))
		case "income_verification":
			fmt.Printf("   Income Status: %s, Verified Amount: $%d\n",
				getIncomeStatus(time.Now().Second()), 50000+time.Now().Second()*1000)
		case "risk_assessment":
			fmt.Printf("   Risk Score: %.1f, PD: %.2f%%\n",
				float64(time.Now().Second()%100)/2, float64(time.Now().Second()%20)/100)
		case "underwriting_decision":
			decision := getDecision(time.Now().Second())
			fmt.Printf("   Decision: %s\n", decision)
			if decision == "APPROVED" {
				fmt.Printf("   Amount: $%d, Rate: %.2f%%\n",
					25000+time.Now().Second()*500, 5.5+float64(time.Now().Second()%10)/2)
			}
		}
		fmt.Println()
	}
}

func getRiskLevel(seed int) string {
	levels := []string{"LOW", "MEDIUM", "HIGH"}
	return levels[seed%len(levels)]
}

func getIncomeStatus(seed int) string {
	statuses := []string{"VERIFIED", "PARTIAL", "FAILED"}
	return statuses[seed%len(statuses)]
}

func getDecision(seed int) string {
	decisions := []string{"APPROVED", "DENIED", "CONDITIONAL", "MANUAL_REVIEW"}
	return decisions[seed%len(decisions)]
}
