package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8082"

	fmt.Println("🧪 Testing Real Underwriting Workflow with Conductor...")

	// Step 1: Register task definitions
	fmt.Println("\n1. Registering underwriting task definitions...")

	taskDefs := []map[string]interface{}{
		{
			"name":                   "credit_check",
			"description":            "Performs credit check and analysis",
			"timeoutSeconds":         300,
			"responseTimeoutSeconds": 280,
			"retryCount":             3,
			"inputKeys":              []string{"applicationId", "userId"},
			"outputKeys":             []string{"creditScore", "creditDecision", "riskAnalysis"},
		},
		{
			"name":                   "income_verification",
			"description":            "Verifies applicant income and employment",
			"timeoutSeconds":         300,
			"responseTimeoutSeconds": 280,
			"retryCount":             3,
			"inputKeys":              []string{"applicationId", "userId"},
			"outputKeys":             []string{"incomeVerification", "incomeAnalysis"},
		},
		{
			"name":                   "risk_assessment",
			"description":            "Performs comprehensive risk assessment",
			"timeoutSeconds":         180,
			"responseTimeoutSeconds": 160,
			"retryCount":             3,
			"inputKeys":              []string{"applicationId", "userId"},
			"outputKeys":             []string{"riskAssessment", "riskLevel", "riskScore"},
		},
		{
			"name":                   "underwriting_decision",
			"description":            "Makes final underwriting decision",
			"timeoutSeconds":         120,
			"responseTimeoutSeconds": 100,
			"retryCount":             2,
			"inputKeys":              []string{"applicationId", "userId"},
			"outputKeys":             []string{"decision", "approvedAmount", "interestRate", "conditions"},
		},
		{
			"name":                   "update_application_state",
			"description":            "Updates loan application state",
			"timeoutSeconds":         60,
			"responseTimeoutSeconds": 50,
			"retryCount":             3,
			"inputKeys":              []string{"applicationId", "newState"},
			"outputKeys":             []string{"success", "stateTransition"},
		},
	}

	for _, taskDef := range taskDefs {
		jsonData, _ := json.Marshal([]interface{}{taskDef})
		resp, err := http.Post(baseURL+"/api/metadata/taskdefs", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("❌ Failed to register task %s: %v", taskDef["name"], err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 || resp.StatusCode == 409 {
			fmt.Printf("✅ Task %s: %d\n", taskDef["name"], resp.StatusCode)
		} else {
			fmt.Printf("❌ Task %s failed: %d - %s\n", taskDef["name"], resp.StatusCode, string(body))
		}
	}

	// Step 2: Register workflow definition
	fmt.Println("\n2. Registering underwriting workflow...")

	workflowDef := map[string]interface{}{
		"name":        "underwriting_workflow",
		"description": "Complete loan underwriting workflow",
		"version":     1,
		"tasks": []map[string]interface{}{
			{
				"name":              "credit_check",
				"taskReferenceName": "credit_check_task",
				"type":              "SIMPLE",
				"inputParameters": map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				"name":              "income_verification",
				"taskReferenceName": "income_verification_task",
				"type":              "SIMPLE",
				"inputParameters": map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				"name":              "risk_assessment",
				"taskReferenceName": "risk_assessment_task",
				"type":              "SIMPLE",
				"inputParameters": map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				"name":              "underwriting_decision",
				"taskReferenceName": "underwriting_decision_task",
				"type":              "SIMPLE",
				"inputParameters": map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				"name":              "update_application_state",
				"taskReferenceName": "update_state_task",
				"type":              "SIMPLE",
				"inputParameters": map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"newState":      "underwriting_completed",
				},
			},
		},
		"inputParameters": []string{"applicationId", "userId"},
		"outputParameters": map[string]interface{}{
			"decision":       "${underwriting_decision_task.output.decision}",
			"approvedAmount": "${underwriting_decision_task.output.approvedAmount}",
			"interestRate":   "${underwriting_decision_task.output.interestRate}",
		},
		"schemaVersion": 2,
	}

	jsonData, _ := json.Marshal(workflowDef)
	resp, err := http.Post(baseURL+"/api/metadata/workflow", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("❌ Failed to register workflow: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Workflow registration: %d - %s\n", resp.StatusCode, string(body))

	// Step 3: Start workflow
	fmt.Println("\n3. Starting underwriting workflow...")

	workflowInput := map[string]interface{}{
		"applicationId": "APP-TEST-001",
		"userId":        "USER-TEST-001",
	}

	jsonData, _ = json.Marshal(workflowInput)
	resp, err = http.Post(baseURL+"/api/workflow/underwriting_workflow", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("❌ Failed to start workflow: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("❌ Workflow start failed: %d - %s", resp.StatusCode, string(body))
	}

	workflowID := string(bytes.Trim(body, "\""))
	fmt.Printf("✅ Workflow started: %s\n", workflowID)

	// Step 4: Monitor workflow execution
	fmt.Println("\n4. Monitoring workflow execution...")

	for i := 0; i < 30; i++ { // Monitor for 30 seconds
		time.Sleep(1 * time.Second)

		// Get workflow status
		resp, err := http.Get(fmt.Sprintf("%s/api/workflow/%s", baseURL, workflowID))
		if err != nil {
			fmt.Printf("❌ Failed to get workflow status: %v\n", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("❌ Get workflow status failed: %d\n", resp.StatusCode)
			continue
		}

		var workflow map[string]interface{}
		if err := json.Unmarshal(body, &workflow); err != nil {
			fmt.Printf("❌ Failed to parse workflow: %v\n", err)
			continue
		}

		status := workflow["status"].(string)
		fmt.Printf("📊 Workflow status [%ds]: %s\n", i+1, status)

		// Print task statuses
		if tasks, ok := workflow["tasks"].([]interface{}); ok {
			for _, taskInterface := range tasks {
				if task, ok := taskInterface.(map[string]interface{}); ok {
					taskName := task["taskType"].(string)
					taskStatus := task["status"].(string)
					fmt.Printf("   📝 %s: %s\n", taskName, taskStatus)
				}
			}
		}

		if status == "COMPLETED" {
			fmt.Println("🎉 Workflow completed successfully!")

			// Print final output
			if output, ok := workflow["output"].(map[string]interface{}); ok {
				fmt.Println("\n📄 Final Output:")
				for key, value := range output {
					fmt.Printf("   %s: %v\n", key, value)
				}
			}
			break
		} else if status == "FAILED" || status == "TERMINATED" {
			fmt.Printf("❌ Workflow failed with status: %s\n", status)

			// Print failure reason
			if reason, ok := workflow["reasonForIncompletion"].(string); ok && reason != "" {
				fmt.Printf("   Reason: %s\n", reason)
			}
			break
		}

		fmt.Println() // Empty line for readability
	}

	fmt.Println("\n🏁 Workflow monitoring completed!")
}
