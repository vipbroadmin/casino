package playeruc

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type Service struct {
	uow     UnitOfWork
	players PlayerRepository
	events  PlayerStatusEventRepository
	outbox  OutboxRepository // optional, can be nil
	clock   ClockReal
}

type ClockReal interface {
	Now() time.Time
}

func New(uow UnitOfWork, players PlayerRepository, events PlayerStatusEventRepository, outbox OutboxRepository, clock ClockReal) *Service {
	return &Service{
		uow:     uow,
		players: players,
		events:  events,
		outbox:  outbox,
		clock:   clock,
	}
}

type CreatePlayerCmd struct {
	Email          string
	Phone          string
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Gender         string
	CountryCode    string
	Locale         string
	TimeZone       string
	RegistrationIP string // text from HTTP
	Metadata       map[string]any
	RegisteredAt   time.Time
}

func (s *Service) CreatePlayer(ctx context.Context, cmd CreatePlayerCmd) (*player.Player, error) {
	now := s.clock.Now()

	addr, err := player.NewAddress(strings.ToUpper(cmd.CountryCode), cmd.Locale, cmd.TimeZone)
	if err != nil {
		return nil, err
	}

	g, err := player.ParseGender(strings.ToLower(strings.TrimSpace(cmd.Gender)))
	if err != nil {
		return nil, err
	}

	var ip net.IP
	if strings.TrimSpace(cmd.RegistrationIP) != "" {
		ip = net.ParseIP(strings.TrimSpace(cmd.RegistrationIP))
		if ip == nil {
			return nil, player.ErrValidation
		}
	}

	p, err := player.NewPlayer(player.CreateParams{
		Email:          cmd.Email,
		Phone:          cmd.Phone,
		FirstName:      cmd.FirstName,
		LastName:       cmd.LastName,
		BirthDate:      cmd.BirthDate,
		Gender:         g,
		Address:        addr,
		RegistrationIP: ip,
		Metadata:       cmd.Metadata,
		RegisteredAt:   cmd.RegisteredAt,
	}, now)
	if err != nil {
		return nil, err
	}

	err = s.uow.WithinTx(ctx, func(ctx context.Context) error {
		ex, err := s.players.GetByEmail(ctx, p.Email)
		if err != nil && !errors.Is(err, player.ErrNotFound) {
			return err
		}
		if ex != nil {
			return player.ErrConflict
		}
		return s.players.Create(ctx, p)
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

type ChangeStatusCmd struct {
	PlayerID uuid.UUID
	ToStatus string
	Reason   string
	Actor    player.ActorType
}

func (s *Service) ChangeStatus(ctx context.Context, cmd ChangeStatusCmd) (*player.Player, player.PlayerStatusEvent, error) {
	now := s.clock.Now()

	to, err := player.ParseStatus(strings.ToLower(strings.TrimSpace(cmd.ToStatus)))
	if err != nil {
		return nil, player.PlayerStatusEvent{}, err
	}

	var updated *player.Player
	var ev player.PlayerStatusEvent

	err = s.uow.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.players.GetByID(ctx, cmd.PlayerID)
		if err != nil {
			return err
		}

		event, err := p.ChangeStatus(to, cmd.Reason, cmd.Actor, now)
		if err != nil {
			return err
		}

		if err := s.players.Update(ctx, p); err != nil {
			return err
		}
		if err := s.events.Append(ctx, event); err != nil {
			return err
		}

		// Outbox pattern (optional) — enqueue message in the same tx.
		if s.outbox != nil {
			msg, err := NewOutboxMessage(
				"player",
				p.ID,
				"player.status.changed",
				p.ID.String(),
				map[string]any{
					"id":          event.ID.String(),
					"player_id":   event.PlayerID.String(),
					"from_status": event.From.String(),
					"to_status":   event.To.String(),
					"reason":      event.Reason,
					"actor_type":  event.ActorType.String(),
					"created_at":  event.CreatedAt.Format(time.RFC3339Nano),
				},
				now,
			)
			if err != nil {
				return err
			}
			if err := s.outbox.Enqueue(ctx, msg); err != nil {
				return err
			}
		}

		updated = p
		ev = event
		return nil
	})
	if err != nil {
		return nil, player.PlayerStatusEvent{}, err
	}

	return updated, ev, nil
}

func (s *Service) GetPlayer(ctx context.Context, id uuid.UUID) (*player.Player, error) {
	return s.players.GetByID(ctx, id)
}
