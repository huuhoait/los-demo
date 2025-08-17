package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// BaseConfig contains common configuration fields for all services
type BaseConfig struct {
	Environment string         `yaml:"environment" json:"environment"`
	Service     ServiceConfig  `yaml:"service" json:"service"`
	Server      ServerConfig   `yaml:"server" json:"server"`
	Database    DatabaseConfig `yaml:"database" json:"database"`
	Redis       RedisConfig    `yaml:"redis" json:"redis"`
	Logging     LoggingConfig  `yaml:"logging" json:"logging"`
	I18n        I18nConfig     `yaml:"i18n" json:"i18n"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	Port    string `yaml:"port" json:"port"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string        `yaml:"port" json:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	CORS         CORSConfig    `yaml:"cors" json:"cors"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Origins     []string `yaml:"origins" json:"origins"`
	Methods     []string `yaml:"methods" json:"methods"`
	Headers     []string `yaml:"headers" json:"headers"`
	Credentials bool     `yaml:"credentials" json:"credentials"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL             string        `yaml:"url" json:"url"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	AutoMigrate     bool          `yaml:"auto_migrate" json:"auto_migrate"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
	PoolSize int    `yaml:"pool_size" json:"pool_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}

// I18nConfig holds internationalization configuration
type I18nConfig struct {
	DefaultLanguage string   `yaml:"default_language" json:"default_language"`
	SupportedLangs  []string `yaml:"supported_languages" json:"supported_languages"`
	FallbackLang    string   `yaml:"fallback_language" json:"fallback_language"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(filename string, config interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv(config *BaseConfig) {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.Service.Name = serviceName
	}

	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.Service.Version = serviceVersion
	}

	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}

	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}

	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		config.Redis.Port = redisPort
	}

	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}

	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if db, err := strconv.Atoi(redisDB); err == nil {
			config.Redis.DB = db
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		config.Server.CORS.Origins = strings.Split(corsOrigins, ",")
	}
}

// GetString gets a string environment variable with a default value
func GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt gets an integer environment variable with a default value
func GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetBool gets a boolean environment variable with a default value
func GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDuration gets a duration environment variable with a default value
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetStringSlice gets a comma-separated string environment variable as a slice
func GetStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// SetDefaults sets default values for common configuration fields
func SetDefaults(config *BaseConfig) {
	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.Service.Name == "" {
		config.Service.Name = "unknown-service"
	}

	if config.Service.Version == "" {
		config.Service.Version = "1.0.0"
	}

	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}

	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}

	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}

	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}

	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}

	if config.Redis.Port == "" {
		config.Redis.Port = "6379"
	}

	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	if config.I18n.DefaultLanguage == "" {
		config.I18n.DefaultLanguage = "en"
	}

	if len(config.I18n.SupportedLangs) == 0 {
		config.I18n.SupportedLangs = []string{"en", "vi"}
	}

	if config.I18n.FallbackLang == "" {
		config.I18n.FallbackLang = "en"
	}

	if len(config.Server.CORS.Origins) == 0 {
		config.Server.CORS.Origins = []string{"*"}
	}

	if len(config.Server.CORS.Methods) == 0 {
		config.Server.CORS.Methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	if len(config.Server.CORS.Headers) == 0 {
		config.Server.CORS.Headers = []string{"*"}
	}
}
