package postgres

// PostgreSQL error codes
const (
	PgUniqueViolationCode     = "23505" // 一意性制約違反
	PgForeignKeyViolationCode = "23503" // 外部キー制約違反
	PgCheckViolationCode      = "23514" // チェック制約違反
	PgNotNullViolationCode    = "23502" // NOT NULL制約違反
)
