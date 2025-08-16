package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lendingplatform/los/services/loan-api/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting loan API service",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	// Setup HTTP server
	router := setupRouter(logger)

	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("addr", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.GracefulShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// initLogger initializes the zap logger
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: cfg.IsDevelopment(),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: cfg.Logging.Format,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service": cfg.Application.Name,
			"version": cfg.Application.Version,
		},
	}

	return zapConfig.Build()
}

// setupRouter sets up the Gin router with middleware and routes
func setupRouter(logger *zap.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggerMiddleware(logger))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "loan-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	v1 := router.Group("/v1")
	{
		// Loan applications
		loans := v1.Group("/loans")
		{
			loans.POST("/applications", handleCreateApplication)
			loans.GET("/applications", handleGetApplications)
			loans.GET("/applications/:id", handleGetApplication)
			loans.PUT("/applications/:id", handleUpdateApplication)
			loans.POST("/applications/:id/submit", handleSubmitApplication)

			// Pre-qualification
			loans.POST("/prequalify", handlePreQualify)

			// Offers
			loans.POST("/applications/:id/offer", handleGenerateOffer)
			loans.POST("/applications/:id/accept-offer", handleAcceptOffer)

			// Admin endpoints
			loans.POST("/applications/:id/transition", handleTransitionState)
			loans.GET("/stats", handleGetApplicationStats)

			// Document management
			loans.POST("/documents/upload", handleUploadDocument)
			loans.GET("/applications/:id/documents/status", handleGetDocumentStatus)
			loans.POST("/applications/:id/documents/complete", handleCompleteDocumentCollection)
		}

		// Workflow management
		workflows := v1.Group("/workflows")
		{
			workflows.GET("/:id/status", handleGetWorkflowStatus)
			workflows.POST("/:id/pause", handlePauseWorkflow)
			workflows.POST("/:id/resume", handleResumeWorkflow)
			workflows.POST("/:id/terminate", handleTerminateWorkflow)
		}
	}

	return router
}

// Handler functions for loan API endpoints
func handleCreateApplication(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application created successfully (mock)",
		"data": gin.H{
			"application_id": "mock-app-123",
			"status":         "draft",
		},
	})
}

func handleGetApplications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []gin.H{},
		"message": "No applications found",
	})
}

func handleGetApplication(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":     id,
			"status": "draft",
		},
	})
}

func handleUpdateApplication(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application updated successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

func handleSubmitApplication(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application submitted successfully",
		"data": gin.H{
			"id":     id,
			"status": "submitted",
		},
	})
}

func handlePreQualify(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pre-qualification completed",
		"data": gin.H{
			"qualified":  true,
			"max_amount": 50000,
		},
	})
}

func handleGenerateOffer(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Offer generated successfully",
		"data": gin.H{
			"application_id": id,
			"offer_id":       "mock-offer-123",
			"amount":         25000,
			"interest_rate":  8.5,
		},
	})
}

func handleAcceptOffer(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Offer accepted successfully",
		"data": gin.H{
			"application_id": id,
			"status":         "approved",
		},
	})
}

func handleTransitionState(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "State transition completed",
		"data": gin.H{
			"application_id": id,
			"new_state":      "underwriting",
		},
	})
}

func handleGetApplicationStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_applications": 0,
			"pending_review":     0,
			"approved":           0,
			"denied":             0,
		},
	})
}

func handleUploadDocument(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document uploaded successfully",
		"data": gin.H{
			"document_id": "mock-doc-123",
		},
	})
}

func handleGetDocumentStatus(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"application_id":      id,
			"documents_required":  3,
			"documents_submitted": 0,
			"status":              "pending",
		},
	})
}

func handleCompleteDocumentCollection(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Document collection completed",
		"data": gin.H{
			"application_id": id,
			"status":         "documents_submitted",
		},
	})
}

func handleGetWorkflowStatus(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"workflow_id": id,
			"status":      "RUNNING",
		},
	})
}

func handlePauseWorkflow(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Workflow paused successfully",
		"data": gin.H{
			"workflow_id": id,
			"status":      "PAUSED",
		},
	})
}

func handleResumeWorkflow(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Workflow resumed successfully",
		"data": gin.H{
			"workflow_id": id,
			"status":      "RUNNING",
		},
	})
}

func handleTerminateWorkflow(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Workflow terminated successfully",
		"data": gin.H{
			"workflow_id": id,
			"status":      "TERMINATED",
		},
	})
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Language, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// loggerMiddleware logs HTTP requests
func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithWriter(gin.DefaultWriter)
}
