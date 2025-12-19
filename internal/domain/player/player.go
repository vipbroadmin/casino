package player

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID             uuid.UUID
	Email          string
	Phone          string
	Status         Status
	StatusReason   string
	Address        Address
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Gender         Gender
	RegistrationIP net.IP
	RegisteredAt   time.Time
	LastLoginAt    time.Time
	Metadata       map[string]any

	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	reEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	rePhone = regexp.MustCompile(`^\+?[0-9]{6,20}$`)
)

type CreateParams struct {
	Email          string
	Phone          string
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Gender         Gender
	Address        Address
	RegistrationIP net.IP
	Metadata       map[string]any
	RegisteredAt   time.Time
}

func NewPlayer(p CreateParams, now time.Time) (*Player, error) {
	email := strings.TrimSpace(strings.ToLower(p.Email))
	if email == "" || !reEmail.MatchString(email) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEmail, p.Email)
	}
	if p.Phone != "" && !rePhone.MatchString(p.Phone) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPhone, p.Phone)
	}
	if err := p.Address.Validate(); err != nil {
		return nil, err
	}

	pl := &Player{
		ID:             uuid.New(),
		Email:          email,
		Phone:          p.Phone,
		Status:         StatusActive,
		StatusReason:   "",
		Address:        p.Address,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		BirthDate:      p.BirthDate,
		Gender:         p.Gender,
		RegistrationIP: p.RegistrationIP,
		RegisteredAt:   p.RegisteredAt,
		LastLoginAt:    time.Time{},
		Metadata:       p.Metadata,

		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := pl.Validate(); err != nil {
		return nil, err
	}
	return pl, nil
}

func (p *Player) Validate() error {
	if p.Email == "" || !reEmail.MatchString(p.Email) {
		return fmt.Errorf("%w: %s", ErrInvalidEmail, p.Email)
	}
	if p.Phone != "" && !rePhone.MatchString(p.Phone) {
		return fmt.Errorf("%w: %s", ErrInvalidPhone, p.Phone)
	}
	if p.Status == StatusUnknown {
		return ErrInvalidStatus
	}
	if err := p.Address.Validate(); err != nil {
		return err
	}
	return nil
}

// Domain rule: status change must include non-empty reason.
func (p *Player) ChangeStatus(to Status, reason string, actor ActorType, now time.Time) (PlayerStatusEvent, error) {
	if to == StatusUnknown {
		return PlayerStatusEvent{}, ErrInvalidStatus
	}
	if to == p.Status {
		return PlayerStatusEvent{}, fmt.Errorf("%w: status already %s", ErrValidation, to.String())
	}
	if strings.TrimSpace(reason) == "" {
		return PlayerStatusEvent{}, fmt.Errorf("%w: status_reason required", ErrValidation)
	}

	from := p.Status
	p.Status = to
	p.StatusReason = reason
	p.Version++
	p.UpdatedAt = now

	ev := NewPlayerStatusEvent(p.ID, from, to, reason, actor, now)
	return ev, nil
}

func (p *Player) MarkLogin(at time.Time) {
	p.LastLoginAt = at
	p.Version++
	p.UpdatedAt = at
}
