package playeruc

import (
	"context"
	"time"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type PlayerRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error)
	GetByEmail(ctx context.Context, email string) (*player.Player, error)
	Create(ctx context.Context, p *player.Player) error
	Update(ctx context.Context, p *player.Player) error

	// Admin-facing operations
	List(ctx context.Context, q ListPlayersQuery) ([]PlayerRow, int64, error)
	GetMany(ctx context.Context, ids []uuid.UUID) ([]PlayerRow, error)
	UpdateProfile(ctx context.Context, cmd UpdatePlayerProfileCmd) error
	UpdatePassword(ctx context.Context, cmd UpdatePlayerPasswordCmd) error
	UpdateLevel(ctx context.Context, cmd UpdatePlayerLevelCmd) error
	SetBan(ctx context.Context, id uuid.UUID, banned bool) error
	KickPlayers(ctx context.Context, cmd KickPlayersCmd) error
}

// ListPlayersQuery describes filters for admin list.
type ListPlayersQuery struct {
	Offset int
	Limit  int

	Search   string
	Country  string
	Currency string
	SortBy   string
	Order    string // asc|desc
}

// PlayerRow is a lightweight admin projection.
type PlayerRow struct {
	ID        uuid.UUID
	Login     string
	Email     string
	Phone     string
	Name      string
	Surname   string
	Nickname  string
	Currency  string
	Country   string
	IsBanned  bool
	Level     int
	CreatedAt time.Time
}

type UpdatePlayerProfileCmd struct {
	ID       uuid.UUID
	Login    *string
	Email    *string
	Phone    *string
	Name     *string
	Surname  *string
	Nickname *string
	Currency *string
	Country  *string
}

type UpdatePlayerPasswordCmd struct {
	ID          uuid.UUID
	NewPassword string
}

type UpdatePlayerLevelCmd struct {
	ID    uuid.UUID
	Level int
}

type KickPlayersCmd struct {
	PlayerIDs []uuid.UUID
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

type PlayerDocumentsRepository interface {
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) ([]*player.Document, error)
	GetByID(ctx context.Context, id uuid.UUID) (*player.Document, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status player.DocumentStatus, now time.Time) error
}

type UpdateDocumentStatusCmd struct {
	ID     uuid.UUID
	Status string
}

type PlayerRequisitesRepository interface {
	GetByPlayerID(ctx context.Context, playerID uuid.UUID) (*player.Requisites, error)
	Upsert(ctx context.Context, req *player.Requisites, now time.Time) error
}

type UpdatePlayerRequisitesCmd struct {
	PlayerID        uuid.UUID
	PaymentMethodID uuid.UUID
	FormData        map[string]any
}
