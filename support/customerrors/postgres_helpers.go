package customerrors

import (
	"database/sql"

	"github.com/lib/pq"
)

// PostgreSQL error codes
const (
	PgUniqueViolationCode     = "23505" // 一意性制約違反
	PgForeignKeyViolationCode = "23503" // 外部キー制約違反
	PgCheckViolationCode      = "23514" // チェック制約違反
	PgNotNullViolationCode    = "23502" // NOT NULL制約違反
)

/*
テストコード更新ガイド：

1. エラー型のインポートを変更:
   - pkg/errors → support/customerrors

2. エラータイプの検証を更新:
   - errors.IsNotFoundError(err) → customerrors.IsNotFoundError(err)

3. エラーメッセージ検証の調整:
   - エラーメッセージの形式が変わっている可能性があるため、
     テストの期待値を新しいフォーマットに合わせる

4. コンテキスト情報の検証を追加:
   - 新しいエラーはコンテキスト情報を持つため、
     適切な場合はコンテキスト情報も検証する
*/

// IsPgUniqueViolation はPostgreSQLの一意性制約違反エラーかチェック
func IsPgUniqueViolation(err error) (*pq.Error, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PgUniqueViolationCode {
		return pqErr, true
	}
	return nil, false
}

// IsPgForeignKeyViolation はPostgreSQLの外部キー制約違反エラーかチェック
func IsPgForeignKeyViolation(err error) (*pq.Error, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PgForeignKeyViolationCode {
		return pqErr, true
	}
	return nil, false
}

// IsPgCheckViolation はPostgreSQLのチェック制約違反エラーかチェック
func IsPgCheckViolation(err error) (*pq.Error, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PgCheckViolationCode {
		return pqErr, true
	}
	return nil, false
}

// IsPgNotNullViolation はPostgreSQLのNOT NULL制約違反エラーかチェック
func IsPgNotNullViolation(err error) (*pq.Error, bool) {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == PgNotNullViolationCode {
		return pqErr, true
	}
	return nil, false
}

// DatabaseError はデータベース操作エラーのヘルパー関数
func DatabaseError(operation, entity string, err error) *InternalServerError {
	message := "データベース操作中にエラーが発生しました"
	internal := NewInternalServerError(message)

	// コンテキスト情報の追加
	ctx := Context{
		"operation": operation,
		"entity":    entity,
		"error":     err.Error(),
	}

	return internal.WithContext(ctx)
}

// EntityNotFoundError はエンティティ未検出エラーのヘルパー関数
func EntityNotFoundError(entityType, id string) *NotFoundError {
	message := ""
	if id != "" {
		message = "ID " + id + " の" + entityType + "が見つかりません"
	} else {
		message = entityType + "が見つかりません"
	}

	return NewNotFoundError(message).WithContext(Context{
		"entity_type": entityType,
		"id":          id,
	})
}

// EntityConflictError はエンティティ競合エラーのヘルパー関数
func EntityConflictError(entityType, id, reason string) *ConflictError {
	message := entityType + " (ID: " + id + ") " + reason

	return NewConflictError(message).WithContext(Context{
		"entity_type": entityType,
		"id":          id,
		"reason":      reason,
	})
}

// ConvertDBError はデータベースエラーを適切なドメインエラーに変換します
func ConvertDBError(err error, operation string, entity string, id string) error {
	// 一意性制約違反
	if pqErr, ok := IsPgUniqueViolation(err); ok {
		return EntityConflictError(
			entity,
			id,
			"は既に存在しています",
		).WithContext(Context{
			"constraint": pqErr.Constraint,
			"operation":  operation,
		})
	}

	// 外部キー制約違反
	if pqErr, ok := IsPgForeignKeyViolation(err); ok {
		// 操作に応じてエラー型を分ける
		if operation == "delete" {
			// 削除操作の場合はConflictError
			return EntityConflictError(
				entity,
				id,
				"は関連するデータが存在するため削除できません",
			).WithContext(Context{
				"constraint": pqErr.Constraint,
				"operation":  operation,
			})
		} else {
			// その他の操作（作成、更新など）の場合はValidationError
			return NewValidationError(
				"関連するリソースが存在しません",
				map[string]string{
					"id":         id,
					"constraint": pqErr.Constraint,
				},
			).WithContext(Context{
				"operation": operation,
				"entity":    entity,
			})
		}
	}

	// sql.ErrNoRows (SQL行なし)
	if err == sql.ErrNoRows {
		return EntityNotFoundError(entity, id)
	}

	// その他のデータベースエラー
	return DatabaseError(operation, entity, err).WithContext(Context{
		"id": id,
	})
}

// NotFound は EntityNotFoundError のショートハンド
func NotFound(entityType, id string) *NotFoundError {
	return EntityNotFoundError(entityType, id)
}

// Conflict は EntityConflictError のショートハンド
func Conflict(entityType, id, reason string) *ConflictError {
	return EntityConflictError(entityType, id, reason)
}

// DBError は DatabaseError のショートハンド
func DBError(operation, entity string, err error) *InternalServerError {
	return DatabaseError(operation, entity, err)
}

// Validation は NewValidationError のショートハンド
func Validation(message string, details map[string]string) *ValidationError {
	return NewValidationError(message, details)
}
