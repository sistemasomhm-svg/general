package models

import "time"

// Entry representa la estructura que viaja entre cliente y servidor
type Entry struct {
	ID            string    `json:"id"`
	EncryptedData string    `json:"encrypted_data"`
	Version       int64     `json:"version"`
	IsDeleted     bool      `json:"is_deleted"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

type SyncRequest struct {
	LastSyncVersion int64   `json:"last_sync_version"`
	Changes         []Entry `json:"changes"`
}

type SyncResponse struct {
	NewVersion int64   `json:"new_version"`
	Changes    []Entry `json:"changes"`
}
