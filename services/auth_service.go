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
		return "", fmt.Errorf("usuario no encontrado")
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

func (s *AuthService) GetSalt(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("usuario no existe")
	}
	return user.ClientSalt, nil
}

func (s *AuthService) Register(ctx context.Context, email, authHash, clientSalt string) error {
	// Verificar si ya existe
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return fmt.Errorf("el usuario ya está registrado")
	}

	// Hashear el authHash para guardarlo de forma segura (doble hashing)
	hashedAuth, err := security.HashPassword(authHash)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:      email,
		AuthHash:   hashedAuth,
		ClientSalt: clientSalt,
	}

	return s.userRepo.Create(ctx, user)
}
