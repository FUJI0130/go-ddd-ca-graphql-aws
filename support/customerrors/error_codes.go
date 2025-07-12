package customerrors

// エラーコード定数
const (
	// HTTPステータスコードに対応
	StatusCodeBadRequest          = 400
	StatusCodeUnauthorized        = 401
	StatusCodeForbidden           = 403
	StatusCodeNotFound            = 404
	StatusCodeConflict            = 409
	StatusCodeUnprocessableEntity = 422
	StatusCodeInternalServerError = 500
)
