package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"

	"players_service/internal/integration"
	playeruc "players_service/internal/usecase/player"
)

type OutboxRepository interface {
	FetchUnpublishedByType(ctx context.Context, typ string, limit int) ([]playeruc.OutboxMessage, error)
	MarkPublished(ctx context.Context, id uuid.UUID, at time.Time) error
}

type WalletsOutboxConfig struct {
	PollInterval time.Duration
	BatchSize    int
}

type WalletsOutboxPublisher struct {
	outbox  OutboxRepository
	wallets *integration.WalletsClient
	cfg     WalletsOutboxConfig
}

func NewWalletsOutboxPublisher(outbox OutboxRepository, wallets *integration.WalletsClient, cfg WalletsOutboxConfig) *WalletsOutboxPublisher {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	return &WalletsOutboxPublisher{
		outbox:  outbox,
		wallets: wallets,
		cfg:     cfg,
	}
}

func (p *WalletsOutboxPublisher) Run(ctx context.Context) {
	p.publish(ctx)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.publish(ctx)
		}
	}
}

func (p *WalletsOutboxPublisher) publish(ctx context.Context) {
	msgs, err := p.outbox.FetchUnpublishedByType(ctx, playeruc.WalletCreateOutboxType, p.cfg.BatchSize)
	if err != nil {
		log.Printf("wallets outbox fetch error: %v", err)
		return
	}
	for _, msg := range msgs {
		var payload playeruc.WalletCreatePayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("wallets outbox decode error (id=%s): %v", msg.ID, err)
			continue
		}
		if payload.PlayerID == "" || payload.Currency == "" {
			log.Printf("wallets outbox invalid payload (id=%s)", msg.ID)
			continue
		}
		if payload.Type == "" {
			payload.Type = "real"
		}

		err := p.wallets.CreateWallet(ctx, integration.CreateWalletRequest{
			PlayerID: payload.PlayerID,
			Currency: payload.Currency,
			Type:     payload.Type,
		})
		if err != nil && !errors.Is(err, integration.ErrWalletConflict) {
			log.Printf("wallets outbox publish error (id=%s): %v", msg.ID, err)
			continue
		}

		if err := p.outbox.MarkPublished(ctx, msg.ID, time.Now()); err != nil {
			log.Printf("wallets outbox mark published error (id=%s): %v", msg.ID, err)
		}
	}
}
