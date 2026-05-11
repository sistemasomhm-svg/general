package services

import (
	"context"
	"fmt"
	"time"
	"vault-backend/models"
	"vault-backend/repositories"

	"github.com/google/uuid"
)

type SyncService struct {
	entryRepo *repositories.EntryRepository
}

func NewSyncService(entryRepo *repositories.EntryRepository) *SyncService {
	return &SyncService{entryRepo: entryRepo}
}

func (s *SyncService) ProcessSync(ctx context.Context, userID string, req models.SyncRequest) (*models.SyncResponse, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// 1. Persistir cambios que vienen del cliente (Push)
	for _, entry := range req.Changes {
		eID, err := uuid.Parse(entry.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid entry id %s: %w", entry.ID, err)
		}

		vaultEntry := models.VaultEntry{
			ID:            eID,
			UserID:        uID,
			EncryptedData: entry.EncryptedData,
			Version:       entry.Version,
			IsDeleted:     entry.IsDeleted,
		}

		if err := s.entryRepo.UpsertFromClient(ctx, &vaultEntry); err != nil {
			return nil, fmt.Errorf("error persistiendo cambio %s: %w", entry.ID, err)
		}
	}

	// 2. Obtener cambios del servidor para el cliente (Pull)
	serverChanges, err := s.entryRepo.GetDeltas(ctx, userID, req.LastSyncVersion)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo deltas: %w", err)
	}

	// 3. Obtener la nueva versión máxima para el usuario
	maxVersion, err := s.entryRepo.GetMaxVersion(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo versión máxima: %w", err)
	}

	// Mapear VaultEntry de vuelta a Entry para la respuesta
	var responseChanges []models.Entry
	for _, se := range serverChanges {
		responseChanges = append(responseChanges, models.Entry{
			ID:            se.ID.String(),
			EncryptedData: se.EncryptedData,
			Version:       se.Version,
			IsDeleted:     se.IsDeleted,
			UpdatedAt:     se.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &models.SyncResponse{
		NewVersion: maxVersion,
		Changes:    responseChanges,
	}, nil
}
