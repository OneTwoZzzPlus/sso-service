package auth

import (
	"context"
	"diaryhub/sso-service/internal/domain/models"
	"diaryhub/sso-service/internal/storage"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	tokenTTL     time.Duration
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	mailProvider MailProvider
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type MailProvider interface {
	Send(to string, access string) error
}

var (
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidAppId      = errors.New("invalid app id")
)

// New returns a new instance of the Auth service
func New(log *slog.Logger,
	tokenTTL time.Duration,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	mailProvider MailProvider,
) *Auth {
	return &Auth{
		log:          log,
		tokenTTL:     tokenTTL,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		mailProvider: mailProvider,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "services.auth.Login"
	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("Logining user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("User not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredential)
		}
		log.Error("Failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Info("Invalid password", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredential)
		}
		log.Error("Failed to match password", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("App not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		log.Error("Failed to found app", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User logged in succesfully")

	token, err := NewTokenJWT(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func NewTokenJWT(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "services.auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("Registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate password hash", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("User alredy exists", slog.String("error", err.Error()))
			return 0, fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}
		log.Error("Failed to save user", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("User registered")

	go func() {
		if err := a.mailProvider.Send(email, "empty"); err != nil {
			log.Error("Failed to send email", slog.String("error", err.Error()))
		} else {
			log.Info("Email sent successfully")
		}
	}()

	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "services.auth.IsAdmin"
	log := a.log.With(slog.String("op", op), slog.Int64("user_id", userID))
	log.Info("Checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
