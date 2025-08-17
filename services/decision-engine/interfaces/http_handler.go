package interfaces

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huuhoait/los-demo/services/decision-engine/application"
	"github.com/huuhoait/los-demo/services/decision-engine/domain"
	"go.uber.org/zap"
)

// DecisionHandler handles HTTP requests for decision engine
type DecisionHandler struct {
	decisionService *application.DecisionEngineService
	logger          *zap.Logger
}

// NewDecisionHandler creates a new decision handler
func NewDecisionHandler(decisionService *application.DecisionEngineService, logger *zap.Logger) *DecisionHandler {
	return &DecisionHandler{
		decisionService: decisionService,
		logger:          logger,
	}
}

// MakeDecision handles POST /api/v1/decisions
func (h *DecisionHandler) MakeDecision(c *gin.Context) {
	logger := h.logger.With(
		zap.String("endpoint", "make_decision"),
		zap.String("method", "POST"),
	)

	var request domain.DecisionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		logger.Error("Request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Request validation failed",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Processing decision request",
		zap.String("application_id", request.ApplicationID),
		zap.String("customer_id", request.CustomerID),
		zap.Float64("loan_amount", request.LoanAmount),
	)

	// Process decision
	response, err := h.decisionService.MakeDecision(c.Request.Context(), &request)
	if err != nil {
		logger.Error("Failed to process decision", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process decision",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Decision processed successfully",
		zap.String("application_id", request.ApplicationID),
		zap.String("decision", string(response.Decision)),
		zap.Float64("confidence_score", response.ConfidenceScore),
	)

	c.JSON(http.StatusOK, response)
}

// GetDecision handles GET /api/v1/decisions/:applicationId
func (h *DecisionHandler) GetDecision(c *gin.Context) {
	applicationID := c.Param("applicationId")

	logger := h.logger.With(
		zap.String("endpoint", "get_decision"),
		zap.String("method", "GET"),
		zap.String("application_id", applicationID),
	)

	if applicationID == "" {
		logger.Error("Missing application ID")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	logger.Info("Retrieving decision")

	response, err := h.decisionService.GetDecision(c.Request.Context(), applicationID)
	if err != nil {
		logger.Error("Failed to retrieve decision", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Decision not found",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Decision retrieved successfully")
	c.JSON(http.StatusOK, response)
}

// GetDecisionHistory handles GET /api/v1/customers/:customerId/decisions
func (h *DecisionHandler) GetDecisionHistory(c *gin.Context) {
	customerID := c.Param("customerId")
	limitStr := c.DefaultQuery("limit", "10")

	logger := h.logger.With(
		zap.String("endpoint", "get_decision_history"),
		zap.String("method", "GET"),
		zap.String("customer_id", customerID),
	)

	if customerID == "" {
		logger.Error("Missing customer ID")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Customer ID is required",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10 // Default limit
	}

	logger.Info("Retrieving decision history", zap.Int("limit", limit))

	decisions, err := h.decisionService.GetDecisionHistory(c.Request.Context(), customerID)
	if err != nil {
		logger.Error("Failed to retrieve decision history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve decision history",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Decision history retrieved successfully", zap.Int("count", len(decisions)))

	c.JSON(http.StatusOK, gin.H{
		"customer_id": customerID,
		"decisions":   decisions,
		"count":       len(decisions),
	})
}

// GetStatistics handles GET /api/v1/decisions/statistics
func (h *DecisionHandler) GetStatistics(c *gin.Context) {
	dateFromStr := c.DefaultQuery("date_from", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	dateToStr := c.DefaultQuery("date_to", time.Now().Format("2006-01-02"))

	logger := h.logger.With(
		zap.String("endpoint", "get_statistics"),
		zap.String("method", "GET"),
		zap.String("date_from", dateFromStr),
		zap.String("date_to", dateToStr),
	)

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		logger.Error("Invalid date_from format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date_from format. Use YYYY-MM-DD",
		})
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		logger.Error("Invalid date_to format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date_to format. Use YYYY-MM-DD",
		})
		return
	}

	// Add 24 hours to dateTo to include the entire day
	dateTo = dateTo.Add(24 * time.Hour)

	logger.Info("Retrieving decision statistics")

	stats, err := h.decisionService.GetStatistics(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		logger.Error("Failed to retrieve statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve statistics",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Statistics retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"period":     gin.H{"from": dateFromStr, "to": dateToStr},
		"statistics": stats,
	})
}

// HealthCheck handles GET /health
func (h *DecisionHandler) HealthCheck(c *gin.Context) {
	logger := h.logger.With(
		zap.String("endpoint", "health_check"),
		zap.String("method", "GET"),
	)

	logger.Info("Health check requested")

	status := gin.H{
		"status":    "healthy",
		"service":   "decision-engine",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	// In production, you might want to check database connectivity, etc.
	logger.Info("Health check completed")
	c.JSON(http.StatusOK, status)
}

// ValidateDecisionRequest handles POST /api/v1/decisions/validate
func (h *DecisionHandler) ValidateDecisionRequest(c *gin.Context) {
	logger := h.logger.With(
		zap.String("endpoint", "validate_decision_request"),
		zap.String("method", "POST"),
	)

	var request domain.DecisionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Validating decision request",
		zap.String("application_id", request.ApplicationID),
	)

	// Validate request
	if err := request.Validate(); err != nil {
		logger.Error("Request validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   "Request validation failed",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Request validation successful")

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Request is valid",
		"request": request,
	})
}

// GetDecisionRules handles GET /api/v1/decisions/rules
func (h *DecisionHandler) GetDecisionRules(c *gin.Context) {
	logger := h.logger.With(
		zap.String("endpoint", "get_decision_rules"),
		zap.String("method", "GET"),
	)

	logger.Info("Retrieving decision rules")

	rules, err := h.decisionService.GetDecisionRules(c.Request.Context())
	if err != nil {
		logger.Error("Failed to retrieve decision rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve decision rules",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Decision rules retrieved successfully", zap.Int("count", len(rules)))

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// ErrorHandler is a middleware for handling panics and errors
func (h *DecisionHandler) ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger := h.logger.With(
			zap.String("endpoint", c.FullPath()),
			zap.String("method", c.Request.Method),
			zap.Any("panic", recovered),
		)

		logger.Error("Panic occurred during request processing")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// RequestLogger is a middleware for request logging
func (h *DecisionHandler) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithWriter(gin.DefaultWriter, "/health")
}

// CORSMiddleware handles CORS headers
func (h *DecisionHandler) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SetupRoutes configures all routes for the decision engine API
func (h *DecisionHandler) SetupRoutes(router *gin.Engine) {
	// Middleware
	router.Use(h.RequestLogger())
	router.Use(h.ErrorHandler())
	router.Use(h.CORSMiddleware())

	// Health check
	router.GET("/health", h.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		decisions := v1.Group("/decisions")
		{
			decisions.POST("", h.MakeDecision)
			decisions.POST("/validate", h.ValidateDecisionRequest)
			decisions.GET("/rules", h.GetDecisionRules)
			decisions.GET("/statistics", h.GetStatistics)
			decisions.GET("/:applicationId", h.GetDecision)
		}

		customers := v1.Group("/customers")
		{
			customers.GET("/:customerId/decisions", h.GetDecisionHistory)
		}
	}
}

// RegisterRoutes registers all routes with the router
func (h *DecisionHandler) RegisterRoutes(router *gin.Engine) {
	h.SetupRoutes(router)
}
