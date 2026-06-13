package services

import (
	"errors"
	"os"
	"time"

	"backend/internal/application/responses"
	"backend/internal/infrastructure/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthServices struct {
	UserRepository *repositories.UserRepository
}

func NewAuthServices(
	userRepository *repositories.UserRepository,
) *AuthServices {
	return &AuthServices{
		UserRepository: userRepository,
	}
}

const defaultJWTSecret = "resume-optimizer-dev-secret-key"

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return defaultJWTSecret
}

func (s *AuthServices) Login(
	email, password string,
) (responses.LoginResponse, error) {

	user, err := s.UserRepository.GetByEmail(email)
	if err != nil {
		return responses.LoginResponse{}, errors.New("credenciais inválidas")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return responses.LoginResponse{}, errors.New("credenciais inválidas")
	}

	secret := jwtSecret()

	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"userID": user.ID.String(),
		"sub":    user.ID.String(),
		"exp":    expiresAt.Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return responses.LoginResponse{}, errors.New("erro ao gerar token")
	}

	return responses.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthServices) ValidateToken(
	tokenString string,
) (jwt.MapClaims, error) {

	secret := jwtSecret()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New("token inválido ou expirado")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
