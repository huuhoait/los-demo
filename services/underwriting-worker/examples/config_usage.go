package main

import (
	"fmt"
	"log"
	"os"

	"underwriting_worker/pkg/config"
)

func main() {
	// Example of loading configuration for different environments

	// Set environment (this would typically be set via environment variable)
	os.Setenv("ENVIRONMENT", "development")

	// Load configuration
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Print configuration summary
	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Service: %s v%s\n", cfg.Service.Name, cfg.Service.Version)
	fmt.Printf("Server Port: %d\n", cfg.Server.Port)
	fmt.Printf("Database: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("Conductor URL: %s\n", cfg.Conductor.BaseURL)
	fmt.Printf("Log Level: %s\n", cfg.Logging.Level)

	// Example of environment-specific behavior
	switch cfg.Environment {
	case "development":
		fmt.Printf("Development mode: Mock services enabled: %t\n", cfg.Application.EnableMockServices)
		fmt.Printf("Debug mode: %t\n", cfg.Application.DebugMode)

	case "uat":
		fmt.Printf("UAT mode: Using staging services\n")
		fmt.Printf("Performance monitoring: %t\n", cfg.Monitoring.MetricsEnabled)

	case "production":
		fmt.Printf("Production mode: High performance configuration\n")
		fmt.Printf("Worker pool size: %d\n", cfg.Conductor.WorkerPoolSize)
		fmt.Printf("Rate limiting: %t\n", cfg.Security.RateLimitingEnabled)
		if cfg.Backup.Enabled {
			fmt.Printf("Backup enabled: %s\n", cfg.Backup.Schedule)
		}
	}

	// Example of using service configuration
	fmt.Printf("\nExternal Services:\n")
	fmt.Printf("Credit Bureau: %s (%s)\n", cfg.Services.CreditBureau.Provider, cfg.Services.CreditBureau.BaseURL)
	fmt.Printf("Risk Scoring: %s (%s)\n", cfg.Services.RiskScoring.Provider, cfg.Services.RiskScoring.BaseURL)
	fmt.Printf("Income Verification: %s (%s)\n", cfg.Services.IncomeVerification.Provider, cfg.Services.IncomeVerification.BaseURL)
	fmt.Printf("Decision Engine: %s (%s)\n", cfg.Services.DecisionEngine.Provider, cfg.Services.DecisionEngine.BaseURL)

	// Example of application-specific settings
	fmt.Printf("\nApplication Settings:\n")
	fmt.Printf("Max Loan Amount: $%.2f\n", cfg.Application.MaxLoanAmount)
	fmt.Printf("Min Loan Amount: $%.2f\n", cfg.Application.MinLoanAmount)
	fmt.Printf("Max DTI Ratio: %.2f%%\n", cfg.Application.MaxDTIRatio*100)
	fmt.Printf("Default Interest Rate: %.2f%%\n", cfg.Application.DefaultInterestRate)
	fmt.Printf("Offer Expiration: %d hours\n", cfg.Application.OfferExpirationHours)
}

// Example of how to use configuration in different parts of the application
func exampleDatabaseConnection(cfg *config.Config) {
	fmt.Printf("Connecting to database: %s:%d/%s\n",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)

	// Use configuration values for database connection
	// db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
	//     cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
	//     cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode))
}

func exampleConductorClient(cfg *config.Config) {
	fmt.Printf("Initializing Conductor client: %s\n", cfg.Conductor.BaseURL)
	fmt.Printf("Worker pool size: %d\n", cfg.Conductor.WorkerPoolSize)
	fmt.Printf("Polling interval: %dms\n", cfg.Conductor.PollingIntervalMs)

	// Use configuration for Conductor client setup
	// client := conductor.NewClient(cfg.Conductor.BaseURL,
	//     conductor.WithTimeout(time.Duration(cfg.Conductor.Timeout)*time.Second),
	//     conductor.WithRetryAttempts(cfg.Conductor.RetryAttempts))
}

func exampleLoggingSetup(cfg *config.Config) {
	fmt.Printf("Setting up logging: level=%s, format=%s, output=%s\n",
		cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)

	if cfg.Logging.EnableFile {
		fmt.Printf("File logging enabled: %s\n", cfg.Logging.FilePath)
		fmt.Printf("Max file size: %dMB, Max age: %d days, Max backups: %d\n",
			cfg.Logging.MaxSize, cfg.Logging.MaxAge, cfg.Logging.MaxBackups)
	}

	// Use configuration for logging setup
	// logger := zap.NewProductionConfig()
	// logger.Level = zap.NewAtomicLevelAt(getZapLevel(cfg.Logging.Level))
	// logger.OutputPaths = []string{cfg.Logging.Output}
}

func exampleSecuritySetup(cfg *config.Config) {
	fmt.Printf("Security configuration:\n")
	fmt.Printf("JWT expiration: %d seconds\n", cfg.Security.JWTExpiration)
	fmt.Printf("BCrypt cost: %d\n", cfg.Security.BcryptCost)
	fmt.Printf("CORS enabled: %t\n", cfg.Server.CORS.Enabled)

	if cfg.Security.RateLimitingEnabled {
		fmt.Printf("Rate limiting: %d requests per minute\n", cfg.Security.MaxRequestsPerMinute)
	}

	// Use configuration for security setup
	// jwtMiddleware := middleware.JWT(cfg.Security.JWTSecret, cfg.Security.JWTExpiration)
	// corsMiddleware := middleware.CORS(cfg.Server.CORS.Origins, cfg.Server.CORS.Methods)
}

func exampleMonitoringSetup(cfg *config.Config) {
	fmt.Printf("Monitoring configuration:\n")
	fmt.Printf("Metrics enabled: %t\n", cfg.Monitoring.MetricsEnabled)
	fmt.Printf("Health check enabled: %t\n", cfg.Monitoring.HealthCheckEnabled)
	fmt.Printf("Tracing enabled: %t\n", cfg.Monitoring.TraceEnabled)

	if cfg.Monitoring.PrometheusEndpoint != "" {
		fmt.Printf("Prometheus endpoint: %s\n", cfg.Monitoring.PrometheusEndpoint)
	}

	if cfg.Monitoring.JaegerEndpoint != "" {
		fmt.Printf("Jaeger endpoint: %s\n", cfg.Monitoring.JaegerEndpoint)
	}

	// Use configuration for monitoring setup
	// if cfg.Monitoring.MetricsEnabled {
	//     prometheus.MustRegister(collectors.NewGoCollector())
	//     http.Handle(cfg.Monitoring.PrometheusEndpoint, promhttp.Handler())
	// }
}
