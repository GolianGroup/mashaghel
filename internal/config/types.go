package config

import "time"

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig     `mapstructure:"server" validate:"required"`
	Redis       RedisConfig      `mapstructure:"redis" validate:"required"`
	JWT         JWTConfig        `mapstructure:"jwt" validate:"required"`
	Logger      LoggerConfig     `mapstructure:"logger" validate:"required"`
	ArangoDB    ArangoConfig     `mapstructure:"arango" validate:"required"`
	Tracer      TracerConfig     `mapstructure:"tracer" validate:"required"`
	LogLevel    string           `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
	GRPC        GRPCConfig       `mapstructure:"grpc" validate:"required"`
	Environment string           `mapstructure:"environment" validate:"required,oneof=development production testing"`
	ScyllaDB    ScyllaDBConfig   `mapstructure:"scylladb" validate:"required"`
	Nats        NatsConfig       `mapstructure:"nats" validate:"required"`
	WorkerPool  WorkerPoolConfig `mapstructure:"worker_pool" validate:"required"`
}

// ServerConfig holds all server related configuration
type ServerConfig struct {
	Port         string `mapstructure:"port" validate:"required,number"`
	Host         string `mapstructure:"host" validate:"required,hostname|ip"`
	Mode         string `mapstructure:"mode" validate:"required,oneof=development production testing"`
	ReadTimeout  int    `mapstructure:"read_timeout" validate:"required,min=1"`
	WriteTimeout int    `mapstructure:"write_timeout" validate:"required,min=1"`
}

// RedisConfig holds all redis related configuration
type RedisConfig struct {
	Host         string `mapstructure:"host" validate:"required,hostname|ip"`
	Port         int    `mapstructure:"port" validate:"required,number"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	MaxRetries   int    `mapstructure:"max_retries" validate:"required,min=1"`
	PoolSize     int    `mapstructure:"pool_size" validate:"required,min=1"`
	MinIdleConns int    `mapstructure:"min_idle_conns" validate:"required,min=1"`
	DialTimeout  int    `mapstructure:"dial_time_out" validate:"required,min=1"`
	ReadTimeout  int    `mapstructure:"read_time_out" validate:"required,min=1"`
	WriteTimeout int    `mapstructure:"write_time_out" validate:"required,min=1"`
	IdleTimeout  int    `mapstructure:"idle_time_out" validate:"required,min=1"`
}

// JWTConfig holds all JWT related configuration
type JWTConfig struct {
	Secret           string `mapstructure:"secret" validate:"required,min=32"`
	ExpireHour       int    `mapstructure:"expire_hour" validate:"required,min=1"`
	RefreshExpireDay int    `mapstructure:"refresh_expire_day" validate:"required,min=1"`
}

type LoggerConfig struct {
	Level         string              `mapstructure:"level" validate:"required,oneof=debug info warn error panic"`
	EncoderConfig LoggerEncoderConfig `mapstructure:"encoder_config"`
	Rotation      RotationConfig      `mapstructure:"rotation_config"`
	Fluentbit     FluentbitConfig     `mapstructure:"fluentbit_config" validate:"required"`
}

type LoggerEncoderConfig struct {
	MessageKey string `mapstructure:"message_key" validate:"required"`
	LevelKey   string `mapstructure:"level_key" validate:"required"`
	NameKey    string `mapstructure:"name_key" validate:"required"`
}

type RotationConfig struct {
	Filename   string `mapstruct:"filename" validate:"required"`
	MaxSize    int    `mapstruct:"mazsize"` // megabytes
	MaxBackups int    `mapstruct:"max_backups"`
	MaxAge     int    `mapstruct:"max_ages"` // days
}
type FluentbitConfig struct {
	Host string `mapstructure:"host" validate:"required"`
	Port int    `mapstructure:"port" validate:"required"`
	Tag  string `mapstructure:"tag" validate:"required"`
}
type ArangoConfig struct {
	ConnStrs           string `mapstructure:"conn_strs" validate:"required"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify" validate:"required"`
	DBName             string `mapstructure:"db_name" validate:"required"`
	User               string `mapstructure:"user" validate:"required"`
	Pass               string `mapstructure:"password" validate:"required"`
}

// Signoz Otel tracer configuration
type TracerConfig struct {
	ServiceName  string `mapstructure:"service_name" validate:"required"`
	CollectorUrl string `mapstructure:"collector_url" validate:"required"`
	Insecure     string `mapstructure:"insecure" validate:"required"`
}
type GRPCConfig struct {
	Host string `mapstructure:"grpc_host" validate:"required,hostname|ip"` // gRPC server host
	Port string `mapstructure:"grpc_port" validate:"required,number"`      // gRPC server port
}

// ScyllaDBConfig holds the configuration for ScyllaDB
type ScyllaDBConfig struct {
	Hosts             []string `mapstructure:"hosts" validate:"required"`
	Keyspace          string   `mapstructure:"keyspace" validate:"required"`
	Username          string   `mapstructure:"username" validate:"required"`
	Password          string   `mapstructure:"password" validate:"required"`
	ReplicationClass  string   `mapstructure:"replication_class" validate:"oneof=SimpleStrategy NetworkTopologyStrategy"`
	ReplicationFactor int      `mapstructure:"replication_factor" validate:"required,min=1"`
}

type WorkerPoolConfig struct {
	WorkerPoolSize int         `mapstructure:"worker_pool_size" validate:"required,min=1"`
	TasksConfig    TasksConfig `mapstructure:"tasks_config" validate:"required"`
}

type TasksConfig struct {
	WatchCooldownDuration int `mapstructure:"watch_cooldown_duration" validate:"required,min=10"`
	WatchAgeLimit         int `mapstructure:"watch_age_limit" validate:"required"`
}

type NatsConfig struct {
	ClientPort   int          `mapstructure:"client_port" validate:"required,min=1"`
	ServerPort   int          `mapstructure:"server_port" validate:"required,min=1"`
	Username     string       `mapstructure:"username" validate:"required"`
	Password     string       `mapstructure:"password" validate:"required"`
	Host         string       `mapstructure:"host" validate:"required"`
	StreamConfig StreamConfig `mapstructure:"stream_config" validate:"required"`
	Queue        string       `mapstructure:"queue" validate:"required"`
}

type StreamConfig struct {
	NoAck        bool          `mapstructure:"no_ack"`
	Name         string        `mapstructure:"name" validate:"required"`
	Retention    string        `mapstructure:"retention" validate:"required,oneof=limits interest workqueue"`
	Discard      string        `mapstructure:"discard" validate:"required,oneof=old new"`
	Storage      string        `mapstructure:"storage" validate:"required,oneof=file memory"`
	MaxConsumers int           `mapstructure:"max_consumers" validate:"required,min=1"`
	MaxAge       time.Duration `mapstructure:"max_age" validate:"required,min=1"`
}
