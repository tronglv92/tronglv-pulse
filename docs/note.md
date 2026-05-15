
Pulse needs a domain-level Kafka producer in internal/kafka/ that implements the contract.KafkaPublisher interface. The helper package      │
│ helper/utils/queue/kafka/ already provides a fully-featured segmentio/kafka-go writer wrapper with OpenTelemetry tracing, snappy            │
│ compression, and batching. TASK-015 wraps this helper into a domain-specific producer with acks=all, idempotent batching, and JSON          │
│ serialization — ready for use by services, the outbox drainer (TASK-017), and the ingest handler (TASK-032).         


the worker binary (cmd/worker) needs to run Kafka consumers. The helper library (helper/utils/queue/kafka/consumer.go) already provides a
 full-featured consumer implementation: parallel goroutines, OpenTelemetry tracing, message parsing, commit management, and error handling.
 TASK-016 wraps this helper into domain-friendly constructors and adds a Supervisor with panic-recovery restart — the missing piece before
 domain consumers (enricher, rca, notify) can be wired.


 TASK-017 requires the outbox publisher system: a background goroutine in cmd/worker that polls the
 outbox_events table with FOR UPDATE SKIP LOCKED, publishes events to Kafka, and deduplicates via Redis.

 All production code is already implemented and wired. The three core files (publisher.go, repository.go,
 dedup.go) are complete, the worker binary creates and registers the publisher, and the registry exposes
 GetOutboxRepo() / GetOutboxAppender(). The remaining gap is test coverage — the publisher has zero tests,
  and the repository has only a constructor test.


  TASK-018: External API Clients + Test Fake                                                              │
│                                                                                                         │
│ Context                                                                                                 │
│                                                                                                         │
│ The Pulse platform needs concrete HTTP clients for its three external integrations: Anthropic (text     │
│ generation for RCA), OpenAI (embeddings for error similarity), and Slack (alert delivery). All contract │
│  interfaces (LLMClient, Notifier) and config structs (LLMConfig, SlackConfig) are already defined. The  │
│ external/ directory has empty stubs. A deterministic test fake is also needed so services consuming     │
│ contract.LLMClient can be tested without network calls.                                                 │
│                                                                                                         │
│ Design Decisions                                                                                        │
│                                                                                                         │
│ 1. External clients are thin HTTP wrappers — they expose methods matching the contract signatures but   │
│ do NOT individually implement the full contract.LLMClient interface (since Anthropic only does          │
│ Generate, OpenAI only does Embed). The Slack client directly implements contract.Notifier               │
│ (single-method interface).                                                                              │
│ 2. Budget guard lives in internal/llm/client.go (composite) — the external clients have no Redis        │
│ dependency. A composite llm.Client wires Anthropic + OpenAI together into contract.LLMClient and checks │
│  llm:disabled before each call.                                                                         │
│ 3. Manual JSON parsing for external APIs — httpc.ParseJsonBody returns a generic error for status >=    │
│ 400, which loses API-specific error messages. External clients will read the body and unmarshal JSON    │
│ manually.                                                                                               │
│ 4. Fake uses FNV-32a hashing — stdlib hash/fnv, zero deps, deterministic output for same input. 