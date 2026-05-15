package config

import "github.com/zeromicro/go-zero/core/conf"

// Load reads a YAML config file with ${ENV_VAR} substitution into any config type.
func Load[T any](file *string) T {
	var c T
	conf.MustLoad(*file, &c, conf.UseEnv())
	return c
}

// ── Shared sub-configs ────────────────────────────────────────────────────────

// DBConfig holds the Postgres/TimescaleDB DSN.
// YAML key: DB.DataSource (go-pg DSN format).
type DBConfig struct {
	DataSource string `json:"DataSource"`
}

// CacheConfig holds Redis connection settings.
type CacheConfig struct {
	Host string `json:"Host"`
	Pass string `json:"Pass,optional"`
}

// GrpcConfig holds the gRPC server listen address.
type GrpcConfig struct {
	ListenOn string `json:"ListenOn"`
}

// KafkaConfig holds broker addresses and per-consumer group names.
type KafkaConfig struct {
	Brokers []string    `json:"Brokers"`
	Groups  KafkaGroups `json:"Groups,optional"`
}

// KafkaGroups names each consumer group used by the worker binary.
type KafkaGroups struct {
	Enricher string `json:"Enricher,default=pulse-enricher"`
	Rca      string `json:"Rca,default=pulse-rca"`
	Notify   string `json:"Notify,default=pulse-notify"`
}

// AuthConfig holds secrets for JWT and HMAC ingest authentication.
type AuthConfig struct {
	JwtSecret    string `json:"JwtSecret,optional"`
	HmacSecret   string `json:"HmacSecret,optional"`
	ApiKeyHeader string `json:"ApiKeyHeader,default=X-API-Key"`
}

// LLMConfig holds Anthropic/OpenAI credentials and the daily spend cap.
type LLMConfig struct {
	AnthropicApiKey string  `json:"AnthropicApiKey,optional"`
	AnthropicModel  string  `json:"AnthropicModel,default=claude-sonnet-4-6"`
	OpenAiApiKey    string  `json:"OpenAiApiKey,optional"`
	DailyBudgetUsd  float64 `json:"DailyBudgetUsd,default=5.0"`
}

// SlackConfig holds the incoming webhook URL for alert delivery.
type SlackConfig struct {
	WebhookUrl string `json:"WebhookUrl,optional"`
}

// TelemetryConfig holds the OpenTelemetry collector endpoint.
type TelemetryConfig struct {
	Endpoint string `json:"Endpoint,optional"`
}

// ── Per-binary configs ────────────────────────────────────────────────────────

// APIConfig is loaded by cmd/api from etc/api.yaml.
// Serves REST (8000) + gRPC (8001). Requires DB, Cache, Kafka, Auth, LLM.
type APIConfig struct {
	Name      string          `json:"Name"`
	Host      string          `json:"Host,default=0.0.0.0"`
	Port      int             `json:"Port"`
	Grpc      GrpcConfig      `json:"Grpc"`
	DB        DBConfig        `json:"DB"`
	Cache     CacheConfig     `json:"Cache"`
	Kafka     KafkaConfig     `json:"Kafka"`
	Auth      AuthConfig      `json:"Auth"`
	LLM       LLMConfig       `json:"LLM"`
	Telemetry TelemetryConfig `json:"Telemetry,optional"`
}

// IngestConfig is loaded by cmd/ingest from etc/ingest.yaml.
// Serves HTTP-only (8002). Only needs Kafka and HMAC auth — no DB direct writes.
type IngestConfig struct {
	Name  string      `json:"Name"`
	Host  string      `json:"Host,default=0.0.0.0"`
	Port  int         `json:"Port"`
	Cache CacheConfig `json:"Cache"`
	Kafka KafkaConfig `json:"Kafka"`
	Auth  AuthConfig  `json:"Auth"`
}

// WorkerConfig is loaded by cmd/worker from etc/worker.yaml.
// Runs Kafka consumers (enricher, rca, notify) + outbox drainer. Needs full LLM stack.
type WorkerConfig struct {
	Name  string      `json:"Name"`
	DB    DBConfig    `json:"DB"`
	Cache CacheConfig `json:"Cache"`
	Kafka KafkaConfig `json:"Kafka"`
	LLM   LLMConfig   `json:"LLM"`
}

// CronConfig is loaded by cmd/cron from etc/cron.yaml.
// Runs scheduled jobs: anomaly scan, outbox cleanup, retention, LLM budget reset.
type CronConfig struct {
	Name  string      `json:"Name"`
	DB    DBConfig    `json:"DB"`
	Cache CacheConfig `json:"Cache"`
	Kafka KafkaConfig `json:"Kafka"`
	LLM   LLMConfig   `json:"LLM"`
	Slack SlackConfig `json:"Slack,optional"`
}
