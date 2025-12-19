package player

import (
	"time"

	"github.com/google/uuid"
)

type PlayerStatusEvent struct {
	ID        uuid.UUID
	PlayerID  uuid.UUID
	From      Status
	To        Status
	Reason    string
	ActorType ActorType
	CreatedAt time.Time
}

func NewPlayerStatusEvent(playerID uuid.UUID, from, to Status, reason string, actor ActorType, at time.Time) PlayerStatusEvent {
	return PlayerStatusEvent{
		ID:        uuid.New(),
		PlayerID:  playerID,
		From:      from,
		To:        to,
		Reason:    reason,
		ActorType: actor,
		CreatedAt: at,
	}
}
