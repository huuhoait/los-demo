package tasks

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MockTask represents a task in the mock conductor
type MockTask struct {
	TaskID             string                 `json:"taskId"`
	TaskType           string                 `json:"taskType"`
	WorkflowInstanceID string                 `json:"workflowInstanceId"`
	InputData          map[string]interface{} `json:"inputData"`
	Status             string                 `json:"status"`
	CreatedTime        time.Time              `json:"createdTime"`
	UpdatedTime        time.Time              `json:"updatedTime"`
}

// MockTaskResult represents the result of task execution
type MockTaskResult struct {
	TaskID                string                 `json:"taskId"`
	Status                string                 `json:"status"`
	OutputData            map[string]interface{} `json:"outputData"`
	ReasonForIncompletion string                 `json:"reasonForIncompletion,omitempty"`
	WorkerID              string                 `json:"workerId"`
	CompletedTime         time.Time              `json:"completedTime"`
}

// TaskHandler represents a function that can handle a task
type TaskHandler func(task *MockTask) (*MockTaskResult, error)

// MockConductorClient is a mock implementation of Conductor client
type MockConductorClient struct {
	logger     *zap.Logger
	workers    map[string]TaskHandler
	tasks      map[string]*MockTask
	polling    bool
	pollMutex  sync.RWMutex
	taskMutex  sync.RWMutex
	workerPool int
}

// NewMockConductorClient creates a new mock conductor client
func NewMockConductorClient(logger *zap.Logger, workerPool int) *MockConductorClient {
	return &MockConductorClient{
		logger:     logger,
		workers:    make(map[string]TaskHandler),
		tasks:      make(map[string]*MockTask),
		polling:    false,
		workerPool: workerPool,
	}
}

// RegisterWorker registers a worker for a specific task type
func (c *MockConductorClient) RegisterWorker(taskType string, handler TaskHandler) {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	c.workers[taskType] = handler
	c.logger.Info("Registered worker for task type", zap.String("task_type", taskType))
}

// StartPolling starts polling for tasks
func (c *MockConductorClient) StartPolling() error {
	c.pollMutex.Lock()
	defer c.pollMutex.Unlock()

	if c.polling {
		return fmt.Errorf("polling already started")
	}

	c.polling = true
	c.logger.Info("Started polling for tasks", zap.Int("worker_pool", c.workerPool))

	// Start worker goroutines
	for i := 0; i < c.workerPool; i++ {
		go c.workerLoop(fmt.Sprintf("worker-%d", i))
	}

	return nil
}

// StopPolling stops polling for tasks
func (c *MockConductorClient) StopPolling() {
	c.pollMutex.Lock()
	defer c.pollMutex.Unlock()

	c.polling = false
	c.logger.Info("Stopped polling for tasks")
}

// SubmitTask submits a new task for processing (for testing)
func (c *MockConductorClient) SubmitTask(taskType string, inputData map[string]interface{}) string {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())

	task := &MockTask{
		TaskID:             taskID,
		TaskType:           taskType,
		WorkflowInstanceID: fmt.Sprintf("workflow-%d", time.Now().Unix()),
		InputData:          inputData,
		Status:             "IN_PROGRESS",
		CreatedTime:        time.Now(),
		UpdatedTime:        time.Now(),
	}

	c.tasks[taskID] = task

	c.logger.Info("Submitted task",
		zap.String("task_id", taskID),
		zap.String("task_type", taskType))

	return taskID
}

// workerLoop simulates a worker polling for tasks
func (c *MockConductorClient) workerLoop(workerID string) {
	logger := c.logger.With(zap.String("worker_id", workerID))

	for {
		c.pollMutex.RLock()
		if !c.polling {
			c.pollMutex.RUnlock()
			break
		}
		c.pollMutex.RUnlock()

		// Simulate polling interval
		time.Sleep(1 * time.Second)

		// Check for tasks to process
		task := c.getNextTask()
		if task != nil {
			c.processTask(task, workerID, logger)
		}
	}

	logger.Info("Worker loop ended")
}

// getNextTask gets the next available task
func (c *MockConductorClient) getNextTask() *MockTask {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	for _, task := range c.tasks {
		if task.Status == "IN_PROGRESS" {
			return task
		}
	}

	return nil
}

// processTask processes a task with the registered handler
func (c *MockConductorClient) processTask(task *MockTask, workerID string, logger *zap.Logger) {
	startTime := time.Now()

	logger.Info("Processing task",
		zap.String("task_id", task.TaskID),
		zap.String("task_type", task.TaskType))

	c.taskMutex.RLock()
	handler, exists := c.workers[task.TaskType]
	c.taskMutex.RUnlock()

	if !exists {
		logger.Warn("No handler registered for task type", zap.String("task_type", task.TaskType))
		c.updateTaskStatus(task.TaskID, "FAILED", "No handler registered")
		return
	}

	// Execute the task handler
	result, err := handler(task)

	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Task execution failed",
			zap.String("task_id", task.TaskID),
			zap.Error(err),
			zap.Duration("processing_time", processingTime))

		c.updateTaskStatus(task.TaskID, "FAILED", err.Error())
		return
	}

	logger.Info("Task execution completed",
		zap.String("task_id", task.TaskID),
		zap.String("status", result.Status),
		zap.Duration("processing_time", processingTime))

	c.updateTaskResult(result)
}

// updateTaskStatus updates the status of a task
func (c *MockConductorClient) updateTaskStatus(taskID, status, reason string) {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if task, exists := c.tasks[taskID]; exists {
		task.Status = status
		task.UpdatedTime = time.Now()

		if status == "FAILED" && reason != "" {
			// Store failure reason in task data
			if task.InputData == nil {
				task.InputData = make(map[string]interface{})
			}
			task.InputData["failure_reason"] = reason
		}
	}
}

// updateTaskResult updates task with result data
func (c *MockConductorClient) updateTaskResult(result *MockTaskResult) {
	c.taskMutex.Lock()
	defer c.taskMutex.Unlock()

	if task, exists := c.tasks[result.TaskID]; exists {
		task.Status = result.Status
		task.UpdatedTime = time.Now()

		// Store output data
		if task.InputData == nil {
			task.InputData = make(map[string]interface{})
		}
		task.InputData["output_data"] = result.OutputData
		task.InputData["worker_id"] = result.WorkerID
		task.InputData["completed_time"] = result.CompletedTime
	}
}
