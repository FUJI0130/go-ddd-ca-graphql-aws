package entity_test

import (
    "testing"
    "time"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
    "github.com/FUJI0130/go-ddd-ca/internal/domain/valueobject"
)

func TestNewTestSuite(t *testing.T) {
    tests := []struct {
        name                string
        id                  string
        suiteName          string
        description        string
        estimatedStartDate time.Time
        estimatedEndDate   time.Time
        requireEffortComment bool
        wantErr           bool
    }{
        {
            name:                "正常な入力で作成成功",
            id:                  "TS001-202412",
            suiteName:          "テストスイート1",
            description:        "説明文",
            estimatedStartDate: time.Now(),
            estimatedEndDate:   time.Now().Add(24 * time.Hour),
            requireEffortComment: true,
            wantErr:           false,
        },
        {
            name:                "IDが空の場合はエラー",
            id:                  "",
            suiteName:          "テストスイート1",
            description:        "説明文",
            estimatedStartDate: time.Now(),
            estimatedEndDate:   time.Now().Add(24 * time.Hour),
            requireEffortComment: true,
            wantErr:           true,
        },
        {
            name:                "終了予定日が開始予定日より前の場合はエラー",
            id:                  "TS001-202412",
            suiteName:          "テストスイート1",
            description:        "説明文",
            estimatedStartDate: time.Now().Add(24 * time.Hour),
            estimatedEndDate:   time.Now(),
            requireEffortComment: true,
            wantErr:           true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := entity.NewTestSuite(
                tt.id,
                tt.suiteName,
                tt.description,
                tt.estimatedStartDate,
                tt.estimatedEndDate,
                tt.requireEffortComment,
            )

            if (err != nil) != tt.wantErr {
                t.Errorf("NewTestSuite() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr {
                if got == nil {
                    t.Error("NewTestSuite() returned nil but wanted a TestSuite")
                    return
                }
                
                // 各フィールドの検証
                if got.ID != tt.id {
                    t.Errorf("NewTestSuite().ID = %v, want %v", got.ID, tt.id)
                }
                if got.Name != tt.suiteName {
                    t.Errorf("NewTestSuite().Name = %v, want %v", got.Name, tt.suiteName)
                }
                if got.Status != valueobject.SuiteStatusPreparation {
                    t.Errorf("NewTestSuite().Status = %v, want %v", got.Status, valueobject.SuiteStatusPreparation)
                }
            }
        })
    }
}

func TestTestSuite_UpdateStatus(t *testing.T) {
    // 初期データの作成
    now := time.Now()
    suite, err := entity.NewTestSuite(
        "TS001-202412",
        "テストスイート1",
        "説明文",
        now,
        now.Add(24 * time.Hour),
        true,
    )
    if err != nil {
        t.Fatalf("Failed to create test suite: %v", err)
    }

    tests := []struct {
        name      string
        newStatus valueobject.SuiteStatus
        wantErr   bool
    }{
        {
            name:      "有効なステータスに更新",
            newStatus: valueobject.SuiteStatusInProgress,
            wantErr:   false,
        },
        {
            name:      "無効なステータスで更新失敗",
            newStatus: "不正な値",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := suite.UpdateStatus(tt.newStatus)
            if (err != nil) != tt.wantErr {
                t.Errorf("TestSuite.UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && suite.Status != tt.newStatus {
                t.Errorf("TestSuite.Status = %v, want %v", suite.Status, tt.newStatus)
            }
        })
    }
}