//go:build test_conductor
// +build test_conductor

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

	fmt.Println("🔍 Testing Conductor API connectivity...")

	// Test 1: Health check
	fmt.Println("\n1. Testing health endpoint...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✅ Health check: %d\n", resp.StatusCode)

	// Test 2: Get metadata
	fmt.Println("\n2. Testing metadata endpoint...")
	resp, err = http.Get(baseURL + "/api/metadata/taskdefs")
	if err != nil {
		log.Printf("❌ Metadata check failed: %v", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Metadata endpoint: %d - %d task definitions found\n", resp.StatusCode, countTaskDefs(body))

	// Test 3: Register a simple task definition
	fmt.Println("\n3. Testing task definition registration...")
	taskDef := map[string]interface{}{
		"name":           "test_task",
		"description":    "Test task for debugging",
		"timeoutSeconds": 60,
		"retryCount":     2,
	}

	taskDefArray := []map[string]interface{}{taskDef}
	jsonData, _ := json.Marshal(taskDefArray)

	resp, err = http.Post(baseURL+"/api/metadata/taskdefs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Task definition registration failed: %v", err)
		return
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Task definition registration: %d - %s\n", resp.StatusCode, string(body))

	// Test 4: Register a simple workflow
	fmt.Println("\n4. Testing workflow definition registration...")
	workflowDef := map[string]interface{}{
		"name":        "test_workflow",
		"description": "Test workflow for debugging",
		"version":     1,
		"tasks": []map[string]interface{}{
			{
				"name":              "test_task",
				"taskReferenceName": "test_task_ref",
				"type":              "SIMPLE",
			},
		},
		"schemaVersion": 2,
	}

	jsonData, _ = json.Marshal(workflowDef)
	resp, err = http.Post(baseURL+"/api/metadata/workflow", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Workflow definition registration failed: %v", err)
		return
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Workflow definition registration: %d - %s\n", resp.StatusCode, string(body))

	// Test 5: Poll for tasks
	fmt.Println("\n5. Testing task polling...")
	resp, err = http.Get(baseURL + "/api/tasks/poll/test_task?workerid=test-worker")
	if err != nil {
		log.Printf("❌ Task polling failed: %v", err)
		return
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Task polling: %d - Response length: %d bytes\n", resp.StatusCode, len(body))

	if len(body) > 0 {
		fmt.Printf("   Response: %s\n", string(body))
	}

	// Test 6: Start a workflow
	fmt.Println("\n6. Testing workflow execution...")
	workflowInput := map[string]interface{}{
		"test": "data",
	}
	jsonData, _ = json.Marshal(workflowInput)

	resp, err = http.Post(baseURL+"/api/workflow/test_workflow", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Workflow start failed: %v", err)
		return
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Workflow start: %d - Workflow ID: %s\n", resp.StatusCode, string(body))

	// Test 7: Poll again for any created tasks
	fmt.Println("\n7. Polling for tasks after workflow start...")
	time.Sleep(1 * time.Second)
	resp, err = http.Get(baseURL + "/api/tasks/poll/test_task?workerid=test-worker")
	if err != nil {
		log.Printf("❌ Task polling after workflow failed: %v", err)
		return
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("✅ Task polling after workflow: %d - Response length: %d bytes\n", resp.StatusCode, len(body))

	if len(body) > 0 {
		fmt.Printf("   Task found: %s\n", string(body))

		// Parse the task and complete it
		var task map[string]interface{}
		if err := json.Unmarshal(body, &task); err == nil {
			if taskID, ok := task["taskId"].(string); ok && taskID != "" {
				fmt.Printf("\n8. Completing task %s...\n", taskID)

				taskResult := map[string]interface{}{
					"taskId":   taskID,
					"status":   "COMPLETED",
					"workerId": "test-worker",
					"outputData": map[string]interface{}{
						"result": "Test task completed successfully",
					},
				}

				jsonData, _ = json.Marshal(taskResult)
				resp, err = http.Post(baseURL+"/api/tasks", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("❌ Task completion failed: %v", err)
					return
				}
				body, _ = io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("✅ Task completion: %d - %s\n", resp.StatusCode, string(body))
			}
		}
	}

	fmt.Println("\n🎉 Conductor API test completed!")
}

func countTaskDefs(body []byte) int {
	var taskDefs []interface{}
	if err := json.Unmarshal(body, &taskDefs); err != nil {
		return 0
	}
	return len(taskDefs)
}
