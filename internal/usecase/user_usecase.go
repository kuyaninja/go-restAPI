package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"skeleton-go/internal/entity"
	"skeleton-go/internal/repository"
)

// UserUsecase wraps business rules around user interactions.
type UserUsecase interface {
	ListUsers(ctx context.Context) ([]entity.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (entity.User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (entity.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// CreateUserInput collects data required to create a new user.
type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserInput collects the fields that can be updated.
type UpdateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userUsecase struct {
	repo repository.UserRepository
}

// NewUserUsecase instantiates the default implementation.
func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{repo: repo}
}

func (u *userUsecase) ListUsers(ctx context.Context) ([]entity.User, error) {
	return u.repo.List(ctx)
}

func (u *userUsecase) GetUser(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *userUsecase) CreateUser(ctx context.Context, input CreateUserInput) (entity.User, error) {
	if err := validateUserPayload(input.Name, input.Email); err != nil {
		return entity.User{}, err
	}

	if err := validatePassword(input.Password); err != nil {
		return entity.User{}, err
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		return entity.User{}, err
	}

	user := entity.User{
		ID:           uuid.New(),
		Name:         strings.TrimSpace(input.Name),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return u.repo.Create(ctx, user)
}

func (u *userUsecase) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (entity.User, error) {
	if err := validateUserPayload(input.Name, input.Email); err != nil {
		return entity.User{}, err
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	existing.Name = strings.TrimSpace(input.Name)
	existing.Email = strings.TrimSpace(input.Email)
	existing.UpdatedAt = time.Now().UTC()
	if strings.TrimSpace(input.Password) != "" {
		if err := validatePassword(input.Password); err != nil {
			return entity.User{}, err
		}
		hash, err := hashPassword(input.Password)
		if err != nil {
			return entity.User{}, err
		}
		existing.PasswordHash = hash
	}

	return u.repo.Update(ctx, existing)
}

func (u *userUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func validateUserPayload(name, email string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		return errors.New("name and email are required")
	}
	return nil
}

func validatePassword(password string) error {
	if len(strings.TrimSpace(password)) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
