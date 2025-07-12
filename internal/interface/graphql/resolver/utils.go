package resolver

import (
	"github.com/FUJI0130/go-ddd-ca/internal/interface/graphql/model"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/dto"
)

//-----------------------------------------------------------------------------
// Status Mapping Functions
//-----------------------------------------------------------------------------

// mapStatusToEnum はドメインのステータス文字列をGraphQLのenum型に変換します
func mapStatusToEnum(status string) model.SuiteStatus {
	switch status {
	case "準備中":
		return model.SuiteStatusPreparation
	case "実行中":
		return model.SuiteStatusInProgress
	case "完了":
		return model.SuiteStatusCompleted
	case "中断":
		return model.SuiteStatusSuspended
	default:
		return model.SuiteStatusPreparation // デフォルト値
	}
}

// mapEnumToStatus はGraphQLのenum型をドメインのステータス文字列に変換します
func mapEnumToStatus(status model.SuiteStatus) string {
	switch status {
	case model.SuiteStatusPreparation:
		return "準備中"
	case model.SuiteStatusInProgress:
		return "実行中"
	case model.SuiteStatusCompleted:
		return "完了"
	case model.SuiteStatusSuspended:
		return "中断"
	default:
		return "準備中" // デフォルト値
	}
}

// mapGroupStatusToEnum はグループステータス文字列をGraphQLのenum型に変換します
// 現在はテストスイートと同じステータス体系を使用しています
func mapGroupStatusToEnum(status string) model.SuiteStatus {
	return mapStatusToEnum(status)
}

// mapTestStatusToEnum はテストステータス文字列をGraphQLのenum型に変換します
func mapTestStatusToEnum(status string) model.TestStatus {
	switch status {
	case "作成":
		return model.TestStatusCreated
	case "テスト":
		return model.TestStatusTesting
	case "修正":
		return model.TestStatusFixing
	case "レビュー待ち":
		return model.TestStatusReviewWaiting
	case "レビュー中":
		return model.TestStatusReviewing
	case "完了":
		return model.TestStatusCompleted
	case "再テスト":
		return model.TestStatusRetesting
	default:
		return model.TestStatusCreated // デフォルト値
	}
}

// mapPriorityToEnum は優先度文字列をGraphQLのenum型に変換します
func mapPriorityToEnum(priority string) model.Priority {
	switch priority {
	case "Critical":
		return model.PriorityCritical
	case "High":
		return model.PriorityHigh
	case "Medium":
		return model.PriorityMedium
	case "Low":
		return model.PriorityLow
	default:
		return model.PriorityMedium // デフォルト値
	}
}

// mapCaseStatusToEnum はケースステータス文字列をGraphQLのenum型に変換します
// 現在はテストステータスと同じ変換ロジックを使用しています
func mapCaseStatusToEnum(status string) model.TestStatus {
	return mapTestStatusToEnum(status)
}

//-----------------------------------------------------------------------------
// Cursor Helper Functions
//-----------------------------------------------------------------------------

// getStartCursor はエッジのリストから最初のカーソルを取得します
// エッジが空の場合はnilを返します
func getStartCursor(edges interface{}) *string {
	switch e := edges.(type) {
	case []*model.TestSuiteEdge:
		if len(e) == 0 {
			return nil
		}
		return &e[0].Cursor
	// 必要に応じて他のエッジタイプを追加
	default:
		return nil
	}
}

// getEndCursor はエッジのリストから最後のカーソルを取得します
// エッジが空の場合はnilを返します
func getEndCursor(edges interface{}) *string {
	switch e := edges.(type) {
	case []*model.TestSuiteEdge:
		if len(e) == 0 {
			return nil
		}
		return &e[len(e)-1].Cursor
	// 必要に応じて他のエッジタイプを追加
	default:
		return nil
	}
}

//-----------------------------------------------------------------------------
// Model Conversion Functions
//-----------------------------------------------------------------------------

// TestSuiteDTOToModel はテストスイートのDTOをGraphQLモデルに変換します
func TestSuiteDTOToModel(dto *dto.TestSuiteResponseDTO) *model.TestSuite {
	if dto == nil {
		return nil
	}

	return &model.TestSuite{
		ID:                   dto.ID,
		Name:                 dto.Name,
		Description:          dto.Description,
		Status:               mapStatusToEnum(dto.Status),
		EstimatedStartDate:   dto.EstimatedStartDate,
		EstimatedEndDate:     dto.EstimatedEndDate,
		RequireEffortComment: dto.RequireEffortComment,
		Progress:             dto.Progress,
		CreatedAt:            dto.CreatedAt,
		UpdatedAt:            dto.UpdatedAt,
	}
}

// TestGroupDTOToModel はテストグループのDTOをGraphQLモデルに変換します
func TestGroupDTOToModel(dto *dto.TestGroupResponseDTO) *model.TestGroup {
	if dto == nil {
		return nil
	}

	return &model.TestGroup{
		ID:           dto.ID,
		Name:         dto.Name,
		Description:  dto.Description,
		DisplayOrder: dto.DisplayOrder,
		SuiteID:      dto.SuiteID,
		Status:       mapStatusToEnum(dto.Status),
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}
}

// TestCaseDTOToModel はテストケースのDTOをGraphQLモデルに変換します
func TestCaseDTOToModel(dto *dto.TestCaseResponseDTO) *model.TestCase {
	if dto == nil {
		return nil
	}

	var plannedEffort, actualEffort *float64
	var delayDays *int

	// オプショナルフィールドの処理
	if dto.PlannedEffort > 0 {
		pe := dto.PlannedEffort
		plannedEffort = &pe
	}

	if dto.ActualEffort > 0 {
		ae := dto.ActualEffort
		actualEffort = &ae
	}

	if dto.DelayDays > 0 {
		dd := dto.DelayDays
		delayDays = &dd
	}

	return &model.TestCase{
		ID:            dto.ID,
		Title:         dto.Title,
		Description:   dto.Description,
		Status:        mapCaseStatusToEnum(dto.Status),
		Priority:      mapPriorityToEnum(dto.Priority),
		PlannedEffort: plannedEffort,
		ActualEffort:  actualEffort,
		IsDelayed:     dto.IsDelayed,
		DelayDays:     delayDays,
		GroupID:       dto.GroupID,
		CreatedAt:     dto.CreatedAt,
		UpdatedAt:     dto.UpdatedAt,
	}
}
