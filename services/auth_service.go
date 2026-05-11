package services

import (
	"context"
	"fmt"
	"vault-backend/models"
	"vault-backend/repositories"
	"vault-backend/security"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(ctx context.Context, email, authHash string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error en login: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("credenciales inválidas")
	}

	// Verificar AuthHash contra el StoredHash en DB
	match, err := security.VerifyPassword(authHash, user.AuthHash)
	if err != nil || !match {
		return "", fmt.Errorf("credenciales inválidas")
	}

	// Generar JWT
	token, err := security.GenerateJWT(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, user *models.User) error {
	// Lógica adicional de validación si es necesario
	return s.userRepo.Create(ctx, user)
}
