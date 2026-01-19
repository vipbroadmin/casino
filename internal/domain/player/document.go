package player

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
)

func (s DocumentStatus) String() string {
	return string(s)
}

func ParseDocumentStatus(v string) (DocumentStatus, error) {
	switch v {
	case "pending":
		return DocumentStatusPending, nil
	case "approved":
		return DocumentStatusApproved, nil
	case "rejected":
		return DocumentStatusRejected, nil
	default:
		return DocumentStatusPending, fmt.Errorf("%w: invalid document status: %s", ErrValidation, v)
	}
}

type Document struct {
	ID        uuid.UUID
	PlayerID  uuid.UUID
	Type      string
	Status    DocumentStatus
	FileURL   string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDocument(playerID uuid.UUID, docType string, fileURL string, metadata map[string]any, now time.Time) *Document {
	return &Document{
		ID:        uuid.New(),
		PlayerID:  playerID,
		Type:      docType,
		Status:    DocumentStatusPending,
		FileURL:   fileURL,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (d *Document) UpdateStatus(status DocumentStatus, now time.Time) error {
	if status == "" {
		return fmt.Errorf("%w: status cannot be empty", ErrValidation)
	}
	d.Status = status
	d.UpdatedAt = now
	return nil
}
