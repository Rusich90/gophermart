package service

import (
	"context"
	"fmt"
	"time"

	domainuser "github.com/Rusich90/gophermart.git/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo domainuser.UserRepo
	secret   []byte
}

func NewAuthService(userRepo domainuser.UserRepo, secret []byte) *AuthService {
	return &AuthService{userRepo: userRepo, secret: secret}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (uuid.UUID, error) {
	exists, err := s.userRepo.LoginExists(ctx, login)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check if login exists: %w", err)
	}
	if exists {
		return uuid.Nil, domainuser.ErrLoginAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()

	user := domainuser.User{
		ID:        userID,
		Login:     login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return "", domainuser.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", domainuser.ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (s *AuthService) GenerateToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
