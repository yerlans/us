package services

import (
	"auth/internal/domain/models"
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type UserStorage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	GetUser(ctx context.Context, email string) (models.User, error)
}

type Auth struct {
	log     *slog.Logger
	storage UserStorage
}

func New(log *slog.Logger,
	storage UserStorage,
) *Auth {
	return &Auth{
		log:     log,
		storage: storage,
	}
}

func (u *Auth) SaveUser(ctx context.Context, email string, pass string) (uid int64, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		u.log.Error("failed to hash password", slog.String("err", err.Error()))
		return 0, err
	}

	// Save the user
	userID, err := u.storage.SaveUser(ctx, email, hashedPassword)
	if err != nil {
		u.log.Error("failed to save user", slog.String("err", err.Error()))
		return 0, err
	}

	return userID, nil
}
func (u *Auth) GetUser(ctx context.Context, email string) (models.User, error) {
	user, err := u.storage.GetUser(ctx, email)
	if err != nil {
		u.log.Error("failed to get user", slog.String("err", err.Error()))
		return models.User{}, err
	}

	return user, nil
}

type Claims struct {
	Email  string `json:"email"`
	UserID int64  `json:"userId"`
	jwt.StandardClaims
}

var jwtSecret = []byte("wzqGcrxSmzR1P9iKuHI1RvO9X_Zl4Vtz1JT-LAjHCaY=")

func (u *Auth) GenerateJWT(user models.User) (string, error) {
	// Create the JWT claims, which includes the user information and expiry time
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email:  user.Email,
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		u.log.Error("failed to generate JWT", slog.String("err", err.Error()))
		return "", err
	}

	return tokenString, nil
}
func (u *Auth) ValidateJWT(tokenString string) (models.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		u.log.Error("failed to parse JWT", slog.String("err", err.Error()))
		return models.User{}, err
	}
	if !token.Valid {
		u.log.Error("invalid JWT token")
		return models.User{}, errors.New("invalid token")
	}

	// Get the user from the claims
	user, err := u.GetUser(context.Background(), claims.Email)
	if err != nil {
		u.log.Error("failed to get user from claims", slog.String("err", err.Error()))
		return models.User{}, err
	}

	return user, nil
}
