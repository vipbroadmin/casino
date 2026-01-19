package player

import (
	"time"

	"github.com/google/uuid"
)

type Requisites struct {
	ID             uuid.UUID
	PlayerID       uuid.UUID
	PaymentMethodID uuid.UUID
	FormData       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewRequisites(playerID uuid.UUID, paymentMethodID uuid.UUID, formData map[string]any, now time.Time) *Requisites {
	return &Requisites{
		ID:              uuid.New(),
		PlayerID:        playerID,
		PaymentMethodID: paymentMethodID,
		FormData:        formData,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *Requisites) UpdateFormData(formData map[string]any, now time.Time) {
	r.FormData = formData
	r.UpdatedAt = now
}
