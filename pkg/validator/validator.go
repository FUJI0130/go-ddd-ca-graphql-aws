package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator はアプリケーション全体で使用するバリデーターを提供します
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator は新しいCustomValidatorインスタンスを作成します
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// struct タグをJSONタグから取得するように設定
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// カスタムバリデーションの登録
	registerCustomValidations(v)

	return &CustomValidator{
		validator: v,
	}
}

// Validate は構造体のバリデーションを実行します
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// バリデーションエラーを適切な形式に変換
		return cv.translateError(err)
	}
	return nil
}

// translateError はバリデーションエラーを日本語のメッセージに変換します
func (cv *CustomValidator) translateError(err error) error {
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) == 0 {
		return err
	}

	errorMessages := make([]string, 0, len(validationErrors))
	for _, e := range validationErrors {
		message := cv.translateSingleError(e)
		errorMessages = append(errorMessages, message)
	}

	return fmt.Errorf("バリデーションエラー: %s", strings.Join(errorMessages, "; "))
}

// translateSingleError は単一のバリデーションエラーを日本語メッセージに変換します
func (cv *CustomValidator) translateSingleError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%sは必須です", field)
	case "min":
		return fmt.Sprintf("%sは%s文字以上である必要があります", field, e.Param())
	case "max":
		return fmt.Sprintf("%sは%s文字以下である必要があります", field, e.Param())
	case "gtfield":
		return fmt.Sprintf("%sは%sよりも後の日付である必要があります", field, e.Param())
	default:
		return fmt.Sprintf("%sは%sルールを満たしていません", field, e.Tag())
	}
}

// registerCustomValidations はカスタムバリデーションルールを登録します
func registerCustomValidations(v *validator.Validate) {
	// 今後、カスタムバリデーションを追加する場合はここに実装
	// 例: TestSuite特有のバリデーションルールなど
}
