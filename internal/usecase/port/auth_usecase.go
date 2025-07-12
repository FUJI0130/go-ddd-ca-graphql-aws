// internal/usecase/port/auth_usecase.go
package port

import (
	"context"
	"time"

	"github.com/FUJI0130/go-ddd-ca/internal/domain/entity"
)

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Token        string
	RefreshToken string
	User         *entity.User
	ExpiresAt    time.Time
}

type AuthUseCase interface {
	Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*entity.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}
