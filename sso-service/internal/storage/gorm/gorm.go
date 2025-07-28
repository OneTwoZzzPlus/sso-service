package orm

import (
	"context"
	"diaryhub/sso-service/internal/domain/models"
	"diaryhub/sso-service/internal/storage"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	ID       int64  `gorm:"primaryKey"`
	Email    string `gorm:"type:text;not null;unique"`
	PassHash []byte `gorm:"type:bytea;not null"`
	IsAdmin  bool   `gorm:"not null;default:false"`
}

type App struct {
	gorm.Model
	ID     int    `gorm:"primaryKey"`
	Name   string `gorm:"type:text;not null;unique"`
	Secret string `gorm:"type:text;not null;unique"`
}

func New(connStr string) (*Storage, error) {
	const op = "storage.postgresql.New"

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&App{})

	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	fmt.Print("")
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.gorm.SaveUser"

	user := User{Email: email, PassHash: passHash}

	// Check if email already exists
	err := s.db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.gorm.User"

	var user User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.User{ID: user.ID, Email: user.Email, PassHash: user.PassHash}, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.gorm.IsAdmin"

	var user User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return user.IsAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.gorm.App"

	var app App
	if err := s.db.WithContext(ctx).Where("id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.App{ID: app.ID, Name: app.Name, Secret: app.Secret}, nil
}
