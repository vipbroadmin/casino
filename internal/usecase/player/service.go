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
	uow        UnitOfWork
	players    PlayerRepository
	events     PlayerStatusEventRepository
	outbox     OutboxRepository // optional, can be nil
	clock      ClockReal
	documents  PlayerDocumentsRepository
	requisites PlayerRequisitesRepository
}

type ClockReal interface {
	Now() time.Time
}

func New(uow UnitOfWork, players PlayerRepository, events PlayerStatusEventRepository, outbox OutboxRepository, clock ClockReal, documents PlayerDocumentsRepository, requisites PlayerRequisitesRepository) *Service {
	return &Service{
		uow:        uow,
		players:    players,
		events:     events,
		outbox:     outbox,
		clock:      clock,
		documents:  documents,
		requisites: requisites,
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

// --- admin operations ---

func (s *Service) ListPlayers(ctx context.Context, q ListPlayersQuery) ([]PlayerRow, int64, error) {
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 20
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return s.players.List(ctx, q)
}

func (s *Service) GetPlayersInfo(ctx context.Context, ids []uuid.UUID) ([]PlayerRow, error) {
	if len(ids) == 0 {
		return []PlayerRow{}, nil
	}
	return s.players.GetMany(ctx, ids)
}

func (s *Service) UpdatePlayerProfile(ctx context.Context, cmd UpdatePlayerProfileCmd) error {
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		// ensure player exists
		if _, err := s.players.GetByID(ctx, cmd.ID); err != nil {
			return err
		}
		return s.players.UpdateProfile(ctx, cmd)
	})
}

func (s *Service) UpdatePlayerPassword(ctx context.Context, cmd UpdatePlayerPasswordCmd) error {
	// hashing delegated to repository or helper; keep usecase simple for now
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		// ensure player exists
		if _, err := s.players.GetByID(ctx, cmd.ID); err != nil {
			return err
		}
		return s.players.UpdatePassword(ctx, cmd)
	})
}

func (s *Service) BanPlayer(ctx context.Context, id uuid.UUID) error {
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.players.GetByID(ctx, id); err != nil {
			return err
		}
		return s.players.SetBan(ctx, id, true)
	})
}

func (s *Service) UnbanPlayer(ctx context.Context, id uuid.UUID) error {
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.players.GetByID(ctx, id); err != nil {
			return err
		}
		return s.players.SetBan(ctx, id, false)
	})
}

func (s *Service) UpdatePlayerLevel(ctx context.Context, cmd UpdatePlayerLevelCmd) error {
	if cmd.Level < 1 {
		return player.ErrValidation
	}
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.players.GetByID(ctx, cmd.ID); err != nil {
			return err
		}
		return s.players.UpdateLevel(ctx, cmd)
	})
}

func (s *Service) KickPlayers(ctx context.Context, cmd KickPlayersCmd) error {
	if len(cmd.PlayerIDs) == 0 {
		return player.ErrValidation
	}
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		return s.players.KickPlayers(ctx, cmd)
	})
}

func (s *Service) GetPlayerDocuments(ctx context.Context, playerID uuid.UUID) ([]*player.Document, error) {
	if _, err := s.players.GetByID(ctx, playerID); err != nil {
		return nil, err
	}
	return s.documents.GetByPlayerID(ctx, playerID)
}

func (s *Service) UpdateDocumentStatus(ctx context.Context, cmd UpdateDocumentStatusCmd) error {
	status, err := player.ParseDocumentStatus(cmd.Status)
	if err != nil {
		return err
	}
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.documents.GetByID(ctx, cmd.ID); err != nil {
			return err
		}
		return s.documents.UpdateStatus(ctx, cmd.ID, status, s.clock.Now())
	})
}

func (s *Service) GetPlayerRequisites(ctx context.Context, playerID uuid.UUID) (*player.Requisites, error) {
	if _, err := s.players.GetByID(ctx, playerID); err != nil {
		return nil, err
	}
	return s.requisites.GetByPlayerID(ctx, playerID)
}

func (s *Service) UpdatePlayerRequisites(ctx context.Context, cmd UpdatePlayerRequisitesCmd) error {
	now := s.clock.Now()
	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.players.GetByID(ctx, cmd.PlayerID); err != nil {
			return err
		}

		existing, err := s.requisites.GetByPlayerID(ctx, cmd.PlayerID)
		if err != nil && err != player.ErrNotFound {
			return err
		}

		var req *player.Requisites
		if existing != nil && existing.PaymentMethodID == cmd.PaymentMethodID {
			// Update existing
			existing.UpdateFormData(cmd.FormData, now)
			req = existing
		} else {
			// Create new
			req = player.NewRequisites(cmd.PlayerID, cmd.PaymentMethodID, cmd.FormData, now)
		}

		return s.requisites.Upsert(ctx, req, now)
	})
}

type CreatePlayerAdminCmd struct {
	Login     string
	Password  string
	Country   string
	Currency  string
	PromoCode string
}

func (s *Service) CreatePlayerAdmin(ctx context.Context, cmd CreatePlayerAdminCmd) error {
	now := s.clock.Now()

	addr, err := player.NewAddress(strings.ToUpper(cmd.Country), "en", "UTC")
	if err != nil {
		return err
	}

	// Create player with temporary email (login will be set separately)
	p, err := player.NewPlayer(player.CreateParams{
		Email:   cmd.Login + "@temp.local",
		Address: addr,
	}, now)
	if err != nil {
		return err
	}

	return s.uow.WithinTx(ctx, func(ctx context.Context) error {
		// Check login uniqueness (simple check - can be enhanced)
		// For now, create player first
		if err := s.players.Create(ctx, p); err != nil {
			return err
		}

		// Update profile to set login/currency/country
		if err := s.players.UpdateProfile(ctx, UpdatePlayerProfileCmd{
			ID:       p.ID,
			Login:    &cmd.Login,
			Currency: &cmd.Currency,
			Country:  &cmd.Country,
		}); err != nil {
			return err
		}

		// Set password (hashed in repository)
		return s.players.UpdatePassword(ctx, UpdatePlayerPasswordCmd{
			ID:          p.ID,
			NewPassword: cmd.Password,
		})
	})
}
