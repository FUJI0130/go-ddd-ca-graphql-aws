// internal/usecase/interactor/auth_interactor.go
package interactor

import (
	"context"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
	"github.com/FUJI0130/go-ddd-ca/internal/domain/repository"
	"github.com/FUJI0130/go-ddd-ca/internal/infrastructure/auth"
	"github.com/FUJI0130/go-ddd-ca/internal/usecase/port"
	"github.com/FUJI0130/go-ddd-ca/support/customerrors"
)

type AuthInteractor struct {
	userRepo         repository.UserRepository
	jwtService       auth.JWTService
	passwordHash     auth.PasswordService
	refreshTokenRepo repository.RefreshTokenRepository
}

func NewAuthInteractor(
	userRepo repository.UserRepository,
	jwtService auth.JWTService,
	passwordHash auth.PasswordService,
	refreshTokenRepo repository.RefreshTokenRepository,
) port.AuthUseCase {
	return &AuthInteractor{
		userRepo:         userRepo,
		jwtService:       jwtService,
		passwordHash:     passwordHash,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (a *AuthInteractor) Login(ctx context.Context, request *port.LoginRequest) (*port.LoginResponse, error) {
	// 1. ユーザー名からユーザーを検索
	user, err := a.userRepo.FindByUsername(ctx, request.Username)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("ログインに失敗しました").WithContext(customerrors.Context{
			"username": request.Username,
			"error":    err.Error(),
		})
	}

	// 2. パスワードの検証
	err = a.passwordHash.VerifyPassword(request.Password, user.PasswordHash)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("パスワードが一致しません").WithContext(customerrors.Context{
			"username": request.Username,
		})
	}

	// 3. ユーザーの最終ログイン日時を更新
	err = a.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		// ログイン日時の更新失敗はログインプロセスを中断するほどの問題ではないため、
		// エラーをログに記録するだけでよい（実際の実装ではログ記録を追加）
	}

	// 4. JWTトークンの生成
	token, expiresAt, err := a.jwtService.GenerateToken(user)
	if err != nil {
		return nil, customerrors.NewInternalServerError("トークン生成に失敗しました").WithContext(customerrors.Context{
			"username": request.Username,
			"error":    err.Error(),
		})
	}

	// 5. リフレッシュトークンの生成
	refreshToken, _, err := a.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, customerrors.NewInternalServerError("リフレッシュトークン生成に失敗しました").WithContext(customerrors.Context{
			"username": request.Username,
			"error":    err.Error(),
		})
	}

	// 6. レスポンスの作成
	return &port.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    expiresAt,
	}, nil
}

func (a *AuthInteractor) ValidateToken(ctx context.Context, token string) (*entity.User, error) {
	// 1. トークンの検証
	userID, err := a.jwtService.ValidateToken(token)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("無効なトークンです").WithContext(customerrors.Context{
			"error": err.Error(),
		})
	}

	// 2. ユーザーの取得
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("ユーザーが見つかりません").WithContext(customerrors.Context{
			"userID": userID,
			"error":  err.Error(),
		})
	}

	return user, nil
}

func (a *AuthInteractor) RefreshToken(ctx context.Context, refreshToken string) (*port.LoginResponse, error) {
	// 1. リフレッシュトークンの検証
	userID, err := a.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("無効なリフレッシュトークンです").WithContext(customerrors.Context{
			"error": err.Error(),
		})
	}

	// 2. ユーザーの取得
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, customerrors.NewUnauthorizedError("ユーザーが見つかりません").WithContext(customerrors.Context{
			"userID": userID,
			"error":  err.Error(),
		})
	}

	// 3. 新しいJWTトークンの生成
	token, expiresAt, err := a.jwtService.GenerateToken(user)
	if err != nil {
		return nil, customerrors.NewInternalServerError("トークン生成に失敗しました").WithContext(customerrors.Context{
			"userID": userID,
			"error":  err.Error(),
		})
	}

	// 4. 新しいリフレッシュトークンを生成（オプション）
	// このステップはオプションで、古いリフレッシュトークンを継続して使用することも可能
	newRefreshToken, _, err := a.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, customerrors.NewInternalServerError("リフレッシュトークン生成に失敗しました").WithContext(customerrors.Context{
			"userID": userID,
			"error":  err.Error(),
		})
	}

	// 5. レスポンスの作成
	return &port.LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresAt:    expiresAt,
	}, nil
}

func (a *AuthInteractor) Logout(ctx context.Context, refreshToken string) error {
	// リフレッシュトークンリポジトリが設定されていない場合は、特に操作は不要
	if a.refreshTokenRepo == nil {
		return nil
	}

	// リフレッシュトークンを取得
	token, err := a.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		// トークンが見つからない場合でもエラーを返さない
		// ユーザーはすでにログアウトしていると見なす
		return nil
	}

	// トークンを無効化
	if token != nil {
		err = a.refreshTokenRepo.Revoke(ctx, token.ID)
		if err != nil {
			return customerrors.NewInternalServerError("ログアウト処理に失敗しました").WithContext(customerrors.Context{
				"tokenID": token.ID,
				"error":   err.Error(),
			})
		}
	}

	return nil
}
