package playeruc

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID          uuid.UUID
	Aggregate   string // "player"
	AggregateID uuid.UUID
	Type        string // e.g. "player.status.changed"
	Key         string // Kafka key, usually AggregateID
	Payload     json.RawMessage
	CreatedAt   time.Time
}

func NewOutboxMessage(aggregate string, aggregateID uuid.UUID, typ string, key string, payload any, at time.Time) (OutboxMessage, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return OutboxMessage{}, err
	}
	return OutboxMessage{
		ID:          uuid.New(),
		Aggregate:   aggregate,
		AggregateID: aggregateID,
		Type:        typ,
		Key:         key,
		Payload:     raw,
		CreatedAt:   at,
	}, nil
}
