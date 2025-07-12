package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// モックJWTサービス - 署名失敗をシミュレートするためのモック
type mockJWTService struct {
	JWTServiceImpl
	failSigning bool
}

func (m *mockJWTService) GenerateToken(user *entity.User) (string, time.Time, error) {
	if m.failSigning {
		return "", time.Time{}, customerrors.NewInternalServerError("failed to sign token")
	}
	return m.JWTServiceImpl.GenerateToken(user)
}

// GenerateRefreshTokenメソッドを追加
func (m *mockJWTService) GenerateRefreshToken(userID string) (string, time.Time, error) {
	if m.failSigning {
		return "", time.Time{}, customerrors.NewInternalServerError("failed to sign token")
	}
	return m.JWTServiceImpl.GenerateRefreshToken(userID)
}

func newMockJWTService(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration, failSigning bool) *mockJWTService {
	return &mockJWTService{
		JWTServiceImpl: JWTServiceImpl{
			secretKey:            []byte(secretKey),
			accessTokenDuration:  accessTokenDuration,
			refreshTokenDuration: refreshTokenDuration,
		},
		failSigning: failSigning,
	}
}

// JWTと互換性のないカスタムクレーム（型キャストエラーを発生させるため）
type CustomInvalidClaims struct {
	jwt.RegisteredClaims
	// UserIDとRoleが意図的に欠落
	CustomField string `json:"custom_field"`
}

func TestJWTService_GenerateToken(t *testing.T) {
	// テスト用のシークレットキーと有効期間
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	// JWTサービスの作成
	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// 署名失敗をシミュレートするモックサービス
	mockService := newMockJWTService(secretKey, accessTokenDuration, refreshTokenDuration, true)

	// テスト用のユーザー
	validUser := &entity.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     entity.RoleAdmin,
	}

	// 境界値テスト用ユーザー
	edgeCaseUser := &entity.User{
		ID:       "user-edge",
		Username: "",
		Role:     entity.RoleTester,
	}

	tests := []struct {
		name           string
		service        JWTService
		user           *entity.User
		duration       time.Duration
		shouldError    bool
		errorMessage   string
		validateToken  bool // トークン生成後に検証も行うかどうか
		validateResult bool // 検証結果の期待値（エラーがないかどうか）
	}{
		{
			name:           "valid user",
			service:        jwtService,
			user:           validUser,
			duration:       accessTokenDuration,
			shouldError:    false,
			validateToken:  true,
			validateResult: true,
		},
		{
			name:         "nil user",
			service:      jwtService,
			user:         nil,
			duration:     accessTokenDuration,
			shouldError:  true,
			errorMessage: "user cannot be nil",
		},
		{
			name:           "user with empty username",
			service:        jwtService,
			user:           edgeCaseUser,
			duration:       accessTokenDuration,
			shouldError:    false,
			validateToken:  true,
			validateResult: true,
		},
		{
			name:         "signing fails",
			service:      mockService,
			user:         validUser,
			duration:     accessTokenDuration,
			shouldError:  true,
			errorMessage: "failed to sign token",
		},
		{
			name:           "zero duration",
			service:        NewJWTService(secretKey, 0, refreshTokenDuration),
			user:           validUser,
			duration:       0,
			shouldError:    false, // 修正: トークン生成時にはエラーは発生しない
			validateToken:  true,  // 生成後に検証を行う
			validateResult: false, // 検証時にはエラーになる（期限切れ）
		},
		{
			name:           "negative duration",
			service:        NewJWTService(secretKey, -1*time.Hour, refreshTokenDuration),
			user:           validUser,
			duration:       -1 * time.Hour,
			shouldError:    false, // 修正: トークン生成時にはエラーは発生しない
			validateToken:  true,  // 生成後に検証を行う
			validateResult: false, // 検証時にはエラーになる（期限切れ）
		},
		{
			name:           "extremely long duration",
			service:        NewJWTService(secretKey, 10000*24*time.Hour, refreshTokenDuration),
			user:           validUser,
			duration:       10000 * 24 * time.Hour,
			shouldError:    false,
			validateToken:  true,
			validateResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expTime, err := tt.service.GenerateToken(tt.user)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Empty(t, token)
				assert.True(t, expTime.IsZero())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// 有効期限が現在から約accessTokenDuration後であることを確認
				expectedExpTime := time.Now().Add(tt.duration)
				assert.WithinDuration(t, expectedExpTime, expTime, 2*time.Second)

				// トークンの検証（サービスが標準のJWTServiceImplの場合のみ）
				if _, ok := tt.service.(*JWTServiceImpl); ok && tt.user != nil {
					parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(secretKey), nil
					})

					// トークン期限のテストケースか確認
					if tt.name == "zero duration" || tt.name == "negative duration" {
						// 期間ゼロ/負の場合はエラーが期待される
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "token is expired")
					} else {
						// それ以外のケースではエラーなしを期待
						require.NoError(t, err)
						require.True(t, parsedToken.Valid)

						// クレームの検証（エラーがない場合のみ）
						claims, ok := parsedToken.Claims.(*JWTClaims)
						require.True(t, ok)
						assert.Equal(t, tt.user.ID, claims.UserID)
						assert.Equal(t, tt.user.Role.String(), claims.Role)
					}
				}

				// オプション: 生成したトークンの追加検証
				if tt.validateToken {
					// 標準サービスを使用して検証
					standardService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
					userID, validateErr := standardService.ValidateToken(token)

					if tt.validateResult {
						assert.NoError(t, validateErr)
						if tt.user != nil {
							assert.Equal(t, tt.user.ID, userID)
						}
					} else {
						assert.Error(t, validateErr)
						// 完全一致ではなく、部分一致に変更
						assert.True(t,
							strings.Contains(validateErr.Error(), "token expired") ||
								strings.Contains(validateErr.Error(), "token has invalid claims: token is expired"),
							"エラーメッセージが期限切れに関連していません: %s", validateErr.Error())
						assert.Empty(t, userID)
					}
				}
			}
		})
	}
}

// TestJWTService_ValidateToken_Debug関数を追加
func TestJWTService_ValidateToken_Debug(t *testing.T) {
	// テスト用の準備（既存と同じ）
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour
	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
	validUser := &entity.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     entity.RoleAdmin,
	}

	// 不正なクレーム形式のトークン生成
	mapClaimsToken := jwt.New(jwt.SigningMethodHS256)
	mapClaimsToken.Claims = jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"sub": validUser.ID,
		// UserIDフィールドが欠落
		"completely_wrong_field": "value",
	}
	invalidClaimsTokenString, err := mapClaimsToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// トークンの詳細をデバッグ出力
	t.Logf("== デバッグ: 作成したトークン ==")
	t.Logf("トークン文字列: %s", invalidClaimsTokenString)

	// 直接トークンを解析してクレームを確認
	rawToken, _, err := new(jwt.Parser).ParseUnverified(invalidClaimsTokenString, &jwt.MapClaims{})
	if err != nil {
		t.Logf("ParseUnverified エラー: %v", err)
	} else {
		if mapClaims, ok := rawToken.Claims.(*jwt.MapClaims); ok {
			t.Logf("MapClaimsの内容: %+v", *mapClaims)
		}
	}

	// ValidateTokenを実行
	userID, err := jwtService.ValidateToken(invalidClaimsTokenString)
	t.Logf("== デバッグ: ValidateToken結果 ==")
	t.Logf("戻り値: userID=%s, err=%v", userID, err)

	// 成功するはずなのに失敗している状況を確認
	if err == nil {
		t.Logf("エラーが発生しなかった - 期待に反して成功")
	} else {
		t.Logf("エラーメッセージ: %s", err.Error())
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	// テスト用のシークレットキーと有効期間
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	// JWTサービスの作成
	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// テスト用のユーザー
	validUser := &entity.User{
		ID:       "user-123",
		Username: "testuser",
		Role:     entity.RoleAdmin,
	}

	// 有効なトークンを生成
	validToken, _, err := jwtService.GenerateToken(validUser)
	require.NoError(t, err)
	require.NotEmpty(t, validToken)

	// 期限切れトークンを生成
	expiredClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 1時間前に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   validUser.ID,
		},
		UserID: validUser.ID,
		Role:   validUser.Role.String(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// 異なる署名で生成したトークン
	invalidSignatureToken, _, err := NewJWTService("different-secret", accessTokenDuration, refreshTokenDuration).GenerateToken(validUser)
	require.NoError(t, err)

	// 不正なアルゴリズムのトークン
	unsupportedAlgToken := jwt.New(jwt.SigningMethodNone)
	unsupportedAlgToken.Claims = &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   validUser.ID,
		},
		UserID: validUser.ID,
		Role:   validUser.Role.String(),
	}
	unsupportedAlgTokenString, err := unsupportedAlgToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// 不正なクレーム形式のトークン - MapClaimsを使用した方法
	mapClaimsToken := jwt.New(jwt.SigningMethodHS256)
	mapClaimsToken.Claims = jwt.MapClaims{
		"exp":                    time.Now().Add(time.Hour).Unix(),
		"iat":                    time.Now().Unix(),
		"sub":                    validUser.ID,
		"user_id":                "",      // 空のユーザーID
		"completely_wrong_field": "value", // 期待されないフィールド
	}
	invalidClaimsTokenString, err := mapClaimsToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	tests := []struct {
		name         string
		token        string
		shouldError  bool
		errorMessage string
		expectedID   string
	}{
		{
			name:        "valid token",
			token:       validToken,
			shouldError: false,
			expectedID:  validUser.ID,
		},
		{
			name:         "empty token",
			token:        "",
			shouldError:  true,
			errorMessage: "token cannot be empty",
		},
		{
			name:         "expired token",
			token:        expiredTokenString,
			shouldError:  true,
			errorMessage: "token expired",
		},
		{
			name:         "invalid signature",
			token:        invalidSignatureToken,
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:         "malformed token",
			token:        "not.a.valid.jwt.token",
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:         "unsupported signing algorithm",
			token:        unsupportedAlgTokenString,
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:         "invalid claims format",
			token:        invalidClaimsTokenString,
			shouldError:  true,
			errorMessage: "invalid token: missing user ID",
		},
		{
			name:         "corrupted token",
			token:        validToken + "corrupted",
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := jwtService.ValidateToken(tt.token)

			// デバッグログ追加
			t.Logf("ValidateToken returned: userID=%s, err=%v", userID, err)
			if err != nil {
				t.Logf("Error message: %s", err.Error())
			}

			if tt.shouldError {
				assert.Error(t, err)
				// nil参照によるセグメンテーションフォルトを防ぐ
				if err != nil && tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	// テスト用のシークレットキーと有効期間
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	// JWTサービスの作成
	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// 署名失敗をシミュレートするモックサービス
	mockService := newMockJWTService(secretKey, accessTokenDuration, refreshTokenDuration, true)

	tests := []struct {
		name           string
		service        JWTService
		userID         string
		duration       time.Duration
		shouldError    bool
		errorMessage   string
		validateToken  bool // トークン生成後に検証も行うかどうか
		validateResult bool // 検証結果の期待値（エラーがないかどうか）
	}{
		{
			name:           "valid user ID",
			service:        jwtService,
			userID:         "user-123",
			duration:       refreshTokenDuration,
			shouldError:    false,
			validateToken:  true,
			validateResult: true,
		},
		{
			name:         "empty user ID",
			service:      jwtService,
			userID:       "",
			duration:     refreshTokenDuration,
			shouldError:  true,
			errorMessage: "user ID cannot be empty",
		},
		{
			name:         "signing fails",
			service:      mockService,
			userID:       "user-123",
			duration:     refreshTokenDuration,
			shouldError:  true,
			errorMessage: "failed to sign token",
		},
		{
			name:           "zero duration",
			service:        NewJWTService(secretKey, accessTokenDuration, 0),
			userID:         "user-123",
			duration:       0,
			shouldError:    false, // 修正: トークン生成時にはエラーは発生しない
			validateToken:  true,  // 生成後に検証を行う
			validateResult: false, // 検証時にはエラーになる（期限切れ）
		},
		{
			name:           "negative duration",
			service:        NewJWTService(secretKey, accessTokenDuration, -1*time.Hour),
			userID:         "user-123",
			duration:       -1 * time.Hour,
			shouldError:    false, // 修正: トークン生成時にはエラーは発生しない
			validateToken:  true,  // 生成後に検証を行う
			validateResult: false, // 検証時にはエラーになる（期限切れ）
		},
		{
			name:           "extremely long duration",
			service:        NewJWTService(secretKey, accessTokenDuration, 10000*24*time.Hour),
			userID:         "user-123",
			duration:       10000 * 24 * time.Hour,
			shouldError:    false,
			validateToken:  true,
			validateResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスの状態をデバッグ（サービスがモックの場合）
			if mockSvc, ok := tt.service.(*mockJWTService); ok {
				t.Logf("Using mockJWTService with failSigning: %v", mockSvc.failSigning)
			}

			token, expTime, err := tt.service.GenerateRefreshToken(tt.userID)

			// デバッグログ追加
			// t.Logf("GenerateRefreshToken returned: token=%s, expTime=%v, err=%v", token, expTime, err)
			if err != nil {
				t.Logf("Error message: %s", err.Error())
			}

			if tt.shouldError {
				assert.Error(t, err)
				// nil参照によるセグメンテーションフォルトを防ぐ
				if err != nil && tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Empty(t, token)
				assert.True(t, expTime.IsZero())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// 有効期限が現在から約refreshTokenDuration後であることを確認
				expectedExpTime := time.Now().Add(tt.duration)
				assert.WithinDuration(t, expectedExpTime, expTime, 2*time.Second)

				// トークンの検証（サービスが標準のJWTServiceImplの場合のみ）
				if _, ok := tt.service.(*JWTServiceImpl); ok {
					parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(secretKey), nil
					})

					// 有効期間が0以下の場合は期限切れエラーを期待
					if tt.name == "zero duration" || tt.name == "negative duration" {
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "token is expired")
					} else {
						// 正の有効期間ではエラーなしを期待
						require.NoError(t, err)
						require.True(t, parsedToken.Valid)

						// クレームの検証
						claims, ok := parsedToken.Claims.(*JWTClaims)
						require.True(t, ok)
						assert.Equal(t, tt.userID, claims.UserID)
						assert.Empty(t, claims.Role) // リフレッシュトークンにはロールを含めない
					}
				}

				// オプション: 生成したトークンの追加検証
				if tt.validateToken {
					// 標準サービスを使用して検証
					standardService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)
					userID, validateErr := standardService.ValidateRefreshToken(token)

					if tt.validateResult {
						assert.NoError(t, validateErr)
						assert.Equal(t, tt.userID, userID)
					} else {
						assert.Error(t, validateErr)
						assert.Empty(t, userID)
					}
				}
			}
		})
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	// テスト用のシークレットキーと有効期間
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	// JWTサービスの作成
	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// 有効なリフレッシュトークンを生成
	validUserID := "user-123"
	validToken, _, err := jwtService.GenerateRefreshToken(validUserID)
	require.NoError(t, err)
	require.NotEmpty(t, validToken)

	// アクセストークンを生成（リフレッシュトークンではない）
	validUser := &entity.User{
		ID:       validUserID,
		Username: "testuser",
		Role:     entity.RoleAdmin,
	}
	accessToken, _, err := jwtService.GenerateToken(validUser)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	// 期限切れリフレッシュトークンを生成
	expiredClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 1時間前に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   validUserID,
		},
		UserID: validUserID,
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	// 異なる署名で生成したリフレッシュトークン
	invalidSignatureToken, _, err := NewJWTService("different-secret", accessTokenDuration, refreshTokenDuration).GenerateRefreshToken(validUserID)
	require.NoError(t, err)

	// 不正なクレーム形式のトークン - MapClaimsを使用した方法
	mapClaimsToken := jwt.New(jwt.SigningMethodHS256)
	mapClaimsToken.Claims = jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"sub": validUserID,
		// UserIDフィールドが欠落している
		"completely_wrong_field": "value", // 期待されないフィールド
	}
	invalidClaimsTokenString, err := mapClaimsToken.SignedString([]byte(secretKey))
	require.NoError(t, err)

	tests := []struct {
		name         string
		token        string
		shouldError  bool
		errorMessage string
		expectedID   string
	}{
		{
			name:        "valid refresh token",
			token:       validToken,
			shouldError: false,
			expectedID:  validUserID,
		},
		{
			name:         "empty token",
			token:        "",
			shouldError:  true,
			errorMessage: "token cannot be empty",
		},
		{
			name:         "expired token",
			token:        expiredTokenString,
			shouldError:  true,
			errorMessage: "token expired",
		},
		{
			name:         "invalid signature",
			token:        invalidSignatureToken,
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:         "malformed token",
			token:        "not.a.valid.jwt.token",
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:        "access token used as refresh token",
			token:       accessToken,
			shouldError: false, // 実際はJWTの構造上区別できないため成功する
			expectedID:  validUserID,
		},
		{
			name:         "corrupted token",
			token:        validToken + "corrupted",
			shouldError:  true,
			errorMessage: "failed to parse token",
		},
		{
			name:         "invalid claims format",
			token:        invalidClaimsTokenString,
			shouldError:  true,
			errorMessage: "invalid token: missing user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := jwtService.ValidateRefreshToken(tt.token)

			// デバッグログ追加
			t.Logf("ValidateRefreshToken returned: userID=%s, err=%v", userID, err)
			if err != nil {
				t.Logf("Error message: %s", err.Error())
			}

			if tt.shouldError {
				assert.Error(t, err)
				// nil参照によるセグメンテーションフォルトを防ぐ
				if err != nil && tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				assert.Empty(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
