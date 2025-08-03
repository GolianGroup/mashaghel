package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	viper.SetConfigFile(path)

	// Set default values
	viper.SetDefault("environment", "development")

	viper.AutomaticEnv()
	// viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config values
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// func SetViperEnvMappings() {
// 	// Set environment variable mappings
// 	viper.SetEnvPrefix("APP") // Optional: adds APP_ prefix to all env variables

// 	// Map environment variables to config fields
// 	viper.BindEnv("server.port", "APP_SERVER_PORT")
// 	viper.BindEnv("server.host", "APP_SERVER_HOST")
// 	viper.BindEnv("server.mode", "APP_SERVER_MODE")

// 	viper.BindEnv("db.host", "APP_DB_HOST")
// 	viper.BindEnv("db.port", "APP_DB_PORT")
// 	viper.BindEnv("db.user", "APP_DB_USER")
// 	viper.BindEnv("db.password", "APP_DB_PASSWORD")
// 	viper.BindEnv("db.dbname", "APP_DB_NAME")
// 	viper.BindEnv("db.sslmode", "APP_DB_SSLMODE")

// 	viper.BindEnv("redis.host", "APP_REDIS_HOST")
// 	viper.BindEnv("redis.port", "APP_REDIS_PORT")
// 	viper.BindEnv("redis.password", "APP_REDIS_PASSWORD")
// 	viper.BindEnv("redis.db", "APP_REDIS_DB")

// 	viper.BindEnv("jwt.secret", "APP_JWT_SECRET")
// 	viper.BindEnv("jwt.expire_hour", "APP_JWT_EXPIRE_HOUR")

// 	viper.BindEnv("log_level", "APP_LOG_LEVEL")
// }

// GetRedisAddr returns redis connection address
func GetRedisAddr(cfg *RedisConfig) string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

// NewProductionEncoderConfig returns an opinionated EncoderConfig for
// production environments.
//
// for more information about fields check the documentation
func NewLoggerEncoderConfig(cfg *LoggerEncoderConfig) zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       cfg.LevelKey, // The logging level (e.g. "info", "error").
		NameKey:        cfg.NameKey,
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     cfg.MessageKey, // The message passed to the log statement.
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
func GetArangoStrings(cfg *ArangoConfig) ([]string, error) {
	connections := strings.Split(cfg.ConnStrs, ",")

	allowedProtocols := []string{"tcp", "http", "https", "ssl", "unix", "http+tcp", "http+srv", "http+ssl", "http+unix"}

	for _, conn := range connections {
		parts := strings.SplitN(conn, "://", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid connection string: %s", conn)
		}
		if !slices.Contains(allowedProtocols, parts[0]) {
			return nil, fmt.Errorf("invalid protocol: %s in connection string: %s", parts[0], conn)
		}
	}

	return connections, nil
}

func GetNatsURL(cfg *NatsConfig) string {
	return fmt.Sprintf("nats://%s:%d/",
		cfg.Host,
		cfg.ClientPort,
	)
}

func SystemServerAddr(cfg *ServerConfig) string {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return addr
}
