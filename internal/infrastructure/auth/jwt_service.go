package auth

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

// JWTクレームの構造体
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Role   string   `json:"role"`
	Perms  []string `json:"perms,omitempty"`
}

// JWTService はJWTトークンの生成と検証を行うインターフェース
type JWTService interface {
	// GenerateToken はユーザー情報からJWTトークンを生成する
	GenerateToken(user *entity.User) (string, time.Time, error)

	// ValidateToken はトークンを検証し、ユーザーIDを返す
	ValidateToken(token string) (string, error)

	// GenerateRefreshToken はリフレッシュトークンを生成する
	GenerateRefreshToken(userID string) (string, time.Time, error)

	// ValidateRefreshToken はリフレッシュトークンを検証する
	ValidateRefreshToken(token string) (string, error)
}

// JWTServiceImpl はJWTサービスの実装
type JWTServiceImpl struct {
	secretKey              []byte
	accessTokenDuration    time.Duration
	refreshTokenDuration   time.Duration
	refreshTokenRepository repository.RefreshTokenRepository
}

// NewJWTService は新しいJWTServiceを作成する
func NewJWTService(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) JWTService {
	return &JWTServiceImpl{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func NewJWTServiceWithRepo(
	secretKey string,
	accessTokenDuration,
	refreshTokenDuration time.Duration,
	refreshTokenRepo repository.RefreshTokenRepository,
) JWTService {
	return &JWTServiceImpl{
		secretKey:              []byte(secretKey),
		accessTokenDuration:    accessTokenDuration,
		refreshTokenDuration:   refreshTokenDuration,
		refreshTokenRepository: refreshTokenRepo,
	}
}

// GenerateToken はユーザー情報からJWTトークンを生成する
func (s *JWTServiceImpl) GenerateToken(user *entity.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, customerrors.NewValidationError("user cannot be nil", nil)
	}

	expirationTime := time.Now().Add(s.accessTokenDuration)

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
		UserID: user.ID,
		Role:   user.Role.String(),
		// Permsフィールドは将来の拡張のために準備
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, customerrors.WrapInternalServerError(err, "failed to sign token")
	}

	return tokenString, expirationTime, nil
}

// ValidateToken はトークンを検証し、ユーザーIDを返す
func (s *JWTServiceImpl) ValidateToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", customerrors.NewValidationError("token cannot be empty", nil)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// アルゴリズムの検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, customerrors.NewUnauthorizedErrorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		// エラーメッセージに "expired" が含まれている場合は期限切れエラーとして処理
		if err == jwt.ErrTokenExpired || strings.Contains(err.Error(), "expired") {
			return "", customerrors.NewUnauthorizedError("token expired")
		}
		return "", customerrors.WrapUnauthorizedError(err, "failed to parse token")
	}

	if !token.Valid {
		return "", customerrors.NewUnauthorizedError("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return "", customerrors.NewUnauthorizedError("invalid token claims")
	}
	if claims.UserID == "" {
		return "", customerrors.NewUnauthorizedError("invalid token: missing user ID")
	}

	return claims.UserID, nil
}

// GenerateRefreshToken はリフレッシュトークンを生成する
func (s *JWTServiceImpl) GenerateRefreshToken(userID string) (string, time.Time, error) {
	if userID == "" {
		return "", time.Time{}, customerrors.NewValidationError("user ID cannot be empty", nil)
	}

	expirationTime := time.Now().Add(s.refreshTokenDuration)

	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, customerrors.WrapInternalServerError(err, "failed to sign token")
	}

	// リポジトリが設定されている場合、トークンを保存
	if s.refreshTokenRepository != nil {
		refreshToken := &entity.RefreshToken{
			ID:         uuid.New().String(),
			Token:      tokenString,
			UserID:     userID,
			IssuedAt:   time.Now(),
			ExpiresAt:  expirationTime,
			LastUsedAt: time.Now(),
			IsRevoked:  false,
		}

		if err := s.refreshTokenRepository.Store(context.Background(), refreshToken); err != nil {
			return "", time.Time{}, customerrors.WrapInternalServerError(err, "failed to store refresh token")
		}
	}

	return tokenString, expirationTime, nil
}

// ValidateRefreshToken はリフレッシュトークンを検証する
func (s *JWTServiceImpl) ValidateRefreshToken(tokenString string) (string, error) {

	// リポジトリが設定されている場合、トークンの存在と有効性を確認
	if s.refreshTokenRepository != nil {
		token, err := s.refreshTokenRepository.GetByToken(context.Background(), tokenString)
		if err != nil {
			return "", customerrors.WrapInternalServerError(err, "failed to retrieve refresh token")
		}

		if token == nil {
			return "", customerrors.NewNotFoundError("refresh token not found")
		}

		if !token.IsValid() {
			return "", customerrors.NewUnauthorizedError("refresh token is invalid or expired")
		}

		// 最終使用日時を更新（第2フェーズ機能）
		// この機能は最初のフェーズでは省略し、エラーハンドリングを簡略化

		return token.UserID, nil
	}

	// リフレッシュトークンの検証はアクセストークンと同じロジック
	return s.ValidateToken(tokenString)
}
