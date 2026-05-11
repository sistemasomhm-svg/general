package repositories

import (
	"context"
	"fmt"
	"vault-backend/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryRepository struct {
	pool *pgxpool.Pool
}

func NewEntryRepository(pool *pgxpool.Pool) *EntryRepository {
	return &EntryRepository{pool: pool}
}

// GetDeltas obtiene las entradas que han cambiado desde la última sincronización del cliente
func (r *EntryRepository) GetDeltas(ctx context.Context, userID string, lastVersion int64) ([]models.VaultEntry, error) {
	query := `SELECT id, user_id, encrypted_data, version, is_deleted, updated_at 
			  FROM vault_entries 
			  WHERE user_id = $1 AND version > $2 
			  ORDER BY version ASC`

	rows, err := r.pool.Query(ctx, query, userID, lastVersion)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo deltas: %w", err)
	}
	defer rows.Close()

	var entries []models.VaultEntry
	for rows.Next() {
		var e models.VaultEntry
		err := rows.Scan(&e.ID, &e.UserID, &e.EncryptedData, &e.Version, &e.IsDeleted, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("error escaneando delta: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, nil
}

// UpsertFromClient maneja la lógica de "Last Write Wins" basada en versiones
func (r *EntryRepository) UpsertFromClient(ctx context.Context, entry *models.VaultEntry) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var existingVersion int64
		queryCheck := `SELECT version FROM vault_entries WHERE id = $1`
		err := tx.QueryRow(ctx, queryCheck, entry.ID).Scan(&existingVersion)

		if err != nil {
			if err == pgx.ErrNoRows {
				// Es nueva, la insertamos
				queryInsert := `INSERT INTO vault_entries (id, user_id, encrypted_data, version, is_deleted) 
								VALUES ($1, $2, $3, $4, $5)`
				_, err = tx.Exec(ctx, queryInsert, entry.ID, entry.UserID, entry.EncryptedData, entry.Version, entry.IsDeleted)
				if err != nil {
					return fmt.Errorf("error insertando entrada: %w", err)
				}
				return nil
			}
			return fmt.Errorf("error verificando existencia de entrada: %w", err)
		}

		// Si ya existe, solo actualizamos si la versión enviada es mayor
		if entry.Version > existingVersion {
			queryUpdate := `UPDATE vault_entries 
							SET encrypted_data = $1, version = $2, is_deleted = $3, updated_at = NOW() 
							WHERE id = $4`
			_, err = tx.Exec(ctx, queryUpdate, entry.EncryptedData, entry.Version, entry.IsDeleted, entry.ID)
			if err != nil {
				return fmt.Errorf("error actualizando entrada: %w", err)
			}
		}

		return nil
	})
}

func (r *EntryRepository) GetMaxVersion(ctx context.Context, userID string) (int64, error) {
	var maxVersion int64
	query := `SELECT COALESCE(MAX(version), 0) FROM vault_entries WHERE user_id = $1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo max version: %w", err)
	}
	return maxVersion, nil
}
