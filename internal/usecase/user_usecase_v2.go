package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"skeleton-go/internal/entity"
	"skeleton-go/internal/repository"
)

// UserUsecaseV2 wraps business rules for API v2 interactions.
type UserUsecaseV2 interface {
	ListUsers(ctx context.Context) ([]entity.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (entity.User, error)
	CreateUser(ctx context.Context, input CreateUserV2Input) (entity.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserV2Input) (entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// CreateUserV2Input accepts first/last name components.
type CreateUserV2Input struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// UpdateUserV2Input mirrors CreateUserV2Input for PUT requests.
type UpdateUserV2Input struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type userUsecaseV2 struct {
	repo repository.UserRepository
}

// NewUserUsecaseV2 instantiates the v2 implementation.
func NewUserUsecaseV2(repo repository.UserRepository) UserUsecaseV2 {
	return &userUsecaseV2{repo: repo}
}

func (u *userUsecaseV2) ListUsers(ctx context.Context) ([]entity.User, error) {
	return u.repo.List(ctx)
}

func (u *userUsecaseV2) GetUser(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *userUsecaseV2) CreateUser(ctx context.Context, input CreateUserV2Input) (entity.User, error) {
	if err := validateUserV2Payload(input.FirstName, input.LastName, input.Email); err != nil {
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
		Name:         composeFullName(input.FirstName, input.LastName),
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return u.repo.Create(ctx, user)
}

func (u *userUsecaseV2) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserV2Input) (entity.User, error) {
	if err := validateUserV2Payload(input.FirstName, input.LastName, input.Email); err != nil {
		return entity.User{}, err
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	existing.Name = composeFullName(input.FirstName, input.LastName)
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

func (u *userUsecaseV2) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func validateUserV2Payload(firstName, lastName, email string) error {
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return fmt.Errorf("first_name and last_name are required")
	}

	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}

	return nil
}

func composeFullName(firstName, lastName string) string {
	first := strings.TrimSpace(firstName)
	last := strings.TrimSpace(lastName)
	return strings.TrimSpace(fmt.Sprintf("%s %s", first, last))
}
