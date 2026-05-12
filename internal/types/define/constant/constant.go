package constant

import "time"

// Kafka topic names
const (
	TopicRawLogs    = "pulse.logs.raw"
	TopicAnomalies  = "pulse.anomalies"
	TopicRCARequest = "pulse.rca.request"
	TopicRCAResult  = "pulse.rca.result"
	TopicOutbox     = "pulse.outbox"
)

// Redis key patterns — use fmt.Sprintf to interpolate values
const (
	RedisKeyRCA         = "rca:%s"   // rca:{fingerprint}
	RedisKeyLLMDisabled = "llm:disabled"
	RedisKeyRateLimit   = "rl:%d:%s" // rl:{tenantID}:{window}
)

// Redis TTLs
const (
	RCAL1TTL     = 30 * time.Minute
	RateLimitTTL = time.Minute
)

// LLM budget (default daily cap in USD; overridden by LLM_DAILY_BUDGET_USD env var)
const DefaultLLMDailyBudgetUSD = 5.00
