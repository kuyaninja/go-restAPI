package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"skeleton-go/internal/entity"
)

// Common repository errors.
var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository describes the persistence behavior for users.
type UserRepository interface {
	List(ctx context.Context) ([]entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	Create(ctx context.Context, user entity.User) (entity.User, error)
	Update(ctx context.Context, user entity.User) (entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
