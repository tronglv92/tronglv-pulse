package entity

// All returns all GORM entities for AutoMigrate registration.
// Schema-managed tables (all except users) are authoritative via golang-migrate SQL files;
// AutoMigrate is additive-only and safe to run on startup.
func All() []any {
	return []any{
		&Tenant{},
		&APIKey{},
		&User{},
		&LogEntry{},
		&AnomalyEvent{},
		&Incident{},
		&ErrorEmbedding{},
		&LLMCall{},
		&RCACache{},
		&AuditLog{},
		&OutboxEvent{},
		&SagaInstance{},
	}
}
