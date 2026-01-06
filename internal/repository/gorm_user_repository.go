package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"skeleton-go/internal/entity"
	"skeleton-go/pkg/logger"
)

const (
	userListLogKey       = "user:list"
	userGetLogKey        = "user:get"
	userCreateLogKey     = "user:create"
	userUpdateLogKey     = "user:update"
	userGetByEmailLogKey = "user:get_by_email"
	userDeleteLogKey     = "user:delete"
)

// GormUserRepository persists users using GORM.
type GormUserRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewGormUserRepository constructs a UserRepository backed by GORM.
func NewGormUserRepository(db *gorm.DB, log *logger.Logger) UserRepository {
	return &GormUserRepository{db: db, logger: log}
}

func (r *GormUserRepository) List(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		r.logDBError(ctx, userListLogKey, err, nil)
		return nil, err
	}

	r.logDB(ctx, userListLogKey, logger.Fields{"count": len(users)})
	return users, nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logDBError(ctx, userGetLogKey, ErrUserNotFound, logger.Fields{"user_id": id.String()})
			return entity.User{}, ErrUserNotFound
		}
		r.logDBError(ctx, userGetLogKey, err, logger.Fields{"user_id": id.String()})
		return entity.User{}, err
	}

	r.logDB(ctx, userGetLogKey, logger.Fields{"user_id": id.String()})
	return user, nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logDBError(ctx, userGetByEmailLogKey, ErrUserNotFound, logger.Fields{"email": email})
			return entity.User{}, ErrUserNotFound
		}
		r.logDBError(ctx, userGetByEmailLogKey, err, logger.Fields{"email": email})
		return entity.User{}, err
	}

	r.logDB(ctx, userGetByEmailLogKey, logger.Fields{"email": email})
	return user, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		r.logDBError(ctx, userCreateLogKey, err, logger.Fields{"user_id": user.ID.String()})
		return entity.User{}, err
	}

	r.logDB(ctx, userCreateLogKey, logger.Fields{"user_id": user.ID.String()})
	return user, nil
}

func (r *GormUserRepository) Update(ctx context.Context, user entity.User) (entity.User, error) {
	result := r.db.WithContext(ctx).Save(&user)
	if err := result.Error; err != nil {
		r.logDBError(ctx, userUpdateLogKey, err, logger.Fields{"user_id": user.ID.String()})
		return entity.User{}, err
	}

	if result.RowsAffected == 0 {
		r.logDBError(ctx, userUpdateLogKey, ErrUserNotFound, logger.Fields{"user_id": user.ID.String()})
		return entity.User{}, ErrUserNotFound
	}

	r.logDB(ctx, userUpdateLogKey, logger.Fields{"user_id": user.ID.String()})
	return user, nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id)
	if err := result.Error; err != nil {
		r.logDBError(ctx, userDeleteLogKey, err, logger.Fields{"user_id": id.String()})
		return err
	}

	if result.RowsAffected == 0 {
		r.logDBError(ctx, userDeleteLogKey, ErrUserNotFound, logger.Fields{"user_id": id.String()})
		return ErrUserNotFound
	}

	r.logDB(ctx, userDeleteLogKey, logger.Fields{"user_id": id.String()})
	return nil
}

func (r *GormUserRepository) logDB(ctx context.Context, message string, fields logger.Fields) {
	if r.logger != nil {
		r.logger.DB(ctx, message, fields)
	}
}

func (r *GormUserRepository) logDBError(ctx context.Context, message string, err error, fields logger.Fields) {
	if r.logger != nil {
		r.logger.Error(ctx, message, err, fields)
	}
}
