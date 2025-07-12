package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
	"github.com/FUJI0130/go-ddd-ca/pkg/errors"
	"github.com/FUJI0130/go-ddd-ca/pkg/validator"
	"github.com/gorilla/mux"
)

// TestSuiteHandler はテストスイート関連のHTTPハンドラーを提供します
type TestSuiteHandler struct {
	useCase   TestSuiteUseCase
	validator *validator.CustomValidator
}

// TestSuiteUseCase はテストスイートのユースケースインターフェースを定義します
type TestSuiteUseCase interface {
	CreateTestSuite(dto *dto.TestSuiteCreateDTO) (*dto.TestSuiteResponseDTO, error)
	GetTestSuite(id string) (*dto.TestSuiteResponseDTO, error)
	ListTestSuites(params *dto.TestSuiteQueryParamDTO) (*dto.TestSuiteListResponseDTO, error)
	UpdateTestSuite(id string, updateDTO *dto.TestSuiteUpdateDTO) (*dto.TestSuiteResponseDTO, error)
	UpdateTestSuiteStatus(id string, statusDTO *dto.TestSuiteStatusUpdateDTO) (*dto.TestSuiteResponseDTO, error) // 追加
}

// NewTestSuiteHandler は新しいTestSuiteHandlerインスタンスを作成します
func NewTestSuiteHandler(useCase TestSuiteUseCase) *TestSuiteHandler {
	return &TestSuiteHandler{
		useCase:   useCase,
		validator: validator.NewCustomValidator(),
	}
}

// Create はテストスイートを作成するハンドラーです
func (h *TestSuiteHandler) Create(w http.ResponseWriter, r *http.Request) {
	// リクエストボディのデコード
	var createDTO dto.TestSuiteCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&createDTO); err != nil {
		apiErr := errors.NewBadRequestError("不正なリクエスト形式です")
		apiErr.WriteToResponse(w)
		return
	}

	// バリデーション
	if err := h.validator.Validate(createDTO); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// ユースケースの実行
	response, err := h.useCase.CreateTestSuite(&createDTO)
	if err != nil {
		// 拡張された変換関数を使用
		apiErr := errors.ConvertToAPIError(err)
		apiErr.WriteToResponse(w)
		return
	}

	// 成功レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTestSuite は指定されたIDのテストスイートを取得するハンドラーです
func (h *TestSuiteHandler) GetTestSuite(w http.ResponseWriter, r *http.Request) {
	// リクエストからIDを取得
	id := mux.Vars(r)["id"]

	// ユースケースの実行
	testSuite, err := h.useCase.GetTestSuite(id)
	if err != nil {
		// 拡張された変換関数を使用
		apiErr := errors.ConvertToAPIError(err, id)
		apiErr.WriteToResponse(w)
		return
	}

	// 成功レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(testSuite)
}

// テストスイート固有のエラー定義
var (
	ErrDuplicateTestSuite = errors.NewConflictError("duplicate test suite")
)

// List はテストスイート一覧を取得するハンドラーです
func (h *TestSuiteHandler) List(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータの解析
	params := &dto.TestSuiteQueryParamDTO{
		Status:   stringPtr(r.URL.Query().Get("status")),
		Page:     intPtr(parseIntParam(r.URL.Query().Get("page"), 1)),
		PageSize: intPtr(parseIntParam(r.URL.Query().Get("pageSize"), 10)),
	}

	// バリデーション
	if err := h.validator.Validate(params); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// ユースケースの実行
	response, err := h.useCase.ListTestSuites(params)
	if err != nil {
		// 拡張された変換関数を使用
		apiErr := errors.ConvertToAPIError(err)
		apiErr.WriteToResponse(w)
		return
	}

	// 成功レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	return value
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	return &i
}

// Update はテストスイートを更新するハンドラーです
func (h *TestSuiteHandler) Update(w http.ResponseWriter, r *http.Request) {
	// IDの取得
	id := mux.Vars(r)["id"]

	// リクエストボディのデコード
	var updateDTO dto.TestSuiteUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		apiErr := errors.NewBadRequestError("不正なリクエスト形式です")
		apiErr.WriteToResponse(w)
		return
	}

	// バリデーション
	if err := h.validator.Validate(updateDTO); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// カスタムバリデーション
	if err := updateDTO.Validate(); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// ユースケースの実行
	response, err := h.useCase.UpdateTestSuite(id, &updateDTO)
	if err != nil {
		// 拡張された変換関数を使用し、リソースIDも渡す
		apiErr := errors.ConvertToAPIError(err, id)
		apiErr.WriteToResponse(w)
		return
	}

	// 成功レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateStatus はテストスイートのステータスを更新するハンドラーです
func (h *TestSuiteHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// IDの取得
	id := mux.Vars(r)["id"]

	// リクエストボディのデコード
	var statusDTO dto.TestSuiteStatusUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&statusDTO); err != nil {
		apiErr := errors.NewBadRequestError("不正なリクエスト形式です")
		apiErr.WriteToResponse(w)
		return
	}

	// バリデーション
	if err := h.validator.Validate(statusDTO); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// カスタムバリデーション
	if err := statusDTO.Validate(); err != nil {
		apiErr := errors.NewValidationError(err.Error(), nil)
		apiErr.WriteToResponse(w)
		return
	}

	// ユースケースの実行
	response, err := h.useCase.UpdateTestSuiteStatus(id, &statusDTO)
	if err != nil {
		// 拡張された変換関数を使用
		apiErr := errors.ConvertToAPIError(err, id)
		apiErr.WriteToResponse(w)
		return
	}

	// 成功レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
