package playeruc

import (
	"context"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type PlayerRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error)
	GetByEmail(ctx context.Context, email string) (*player.Player, error)
	Create(ctx context.Context, p *player.Player) error
	Update(ctx context.Context, p *player.Player) error
}

type PlayerStatusEventRepository interface {
	Append(ctx context.Context, ev player.PlayerStatusEvent) error
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, msg OutboxMessage) error
}

// UnitOfWork defines transaction boundary (TBD): one usecase == one transaction.
type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
