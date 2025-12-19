package playerpg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"

	"players_service/internal/domain/player"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error) {
	ex := pickExecutor(ctx, r.db)

	const q = `
SELECT id, email, phone, status, status_reason,
       country_code, locale, time_zone,
       first_name, last_name, birth_date, gender,
       registration_ip, registered_at, last_login_at,
       metadata, version, created_at, updated_at
  FROM players
 WHERE id = $1
`
	var (
		p                         player.Player
		status, gender            int16
		country, locale, tz       sql.NullString
		phone, reason             sql.NullString
		first, last               sql.NullString
		regIP                     sql.NullString
		birth                     sql.NullTime
		registeredAt, lastLoginAt sql.NullTime
		metadataRaw               []byte
		version                   int64
		createdAt, updatedAt      time.Time
	)

	row := ex.QueryRowContext(ctx, q, id)
	err := row.Scan(
		&p.ID, &p.Email, &phone, &status, &reason,
		&country, &locale, &tz,
		&first, &last, &birth, &gender,
		&regIP, &registeredAt, &lastLoginAt,
		&metadataRaw, &version, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, player.ErrNotFound
		}
		return nil, err
	}

	p.Phone = phone.String
	p.Status = player.Status(status)
	p.StatusReason = reason.String
	p.Address = player.Address{CountryCode: country.String, Locale: locale.String, TimeZone: tz.String}
	p.FirstName = first.String
	p.LastName = last.String
	if birth.Valid {
		p.BirthDate = birth.Time
	}
	p.Gender = player.Gender(gender)

	if regIP.Valid && regIP.String != "" {
		p.RegistrationIP = net.ParseIP(regIP.String)
	}
	if registeredAt.Valid {
		p.RegisteredAt = registeredAt.Time
	}
	if lastLoginAt.Valid {
		p.LastLoginAt = lastLoginAt.Time
	}
	if len(metadataRaw) > 0 {
		_ = json.Unmarshal(metadataRaw, &p.Metadata)
	} else {
		p.Metadata = map[string]any{}
	}

	p.Version = version
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return &p, nil
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (*player.Player, error) {
	ex := pickExecutor(ctx, r.db)

	const q = `SELECT id FROM players WHERE email = $1`
	var id uuid.UUID
	err := ex.QueryRowContext(ctx, q, email).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, player.ErrNotFound
		}
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repo) Create(ctx context.Context, p *player.Player) error {
	ex := pickExecutor(ctx, r.db)

	meta, _ := json.Marshal(p.Metadata)

	const q = `
INSERT INTO players (
  id, email, phone, status, status_reason,
  country_code, locale, time_zone,
  first_name, last_name, birth_date, gender,
  registration_ip, registered_at, last_login_at,
  metadata, version, created_at, updated_at
) VALUES (
  $1,$2,$3,$4,$5,
  $6,$7,$8,
  $9,$10,$11,$12,
  $13,$14,$15,
  $16,$17,$18,$19
)
`
	_, err := ex.ExecContext(ctx, q,
		p.ID, p.Email, nullStr(p.Phone), int16(p.Status), nullStr(p.StatusReason),
		nullStr(p.Address.CountryCode), nullStr(p.Address.Locale), nullStr(p.Address.TimeZone),
		nullStr(p.FirstName), nullStr(p.LastName), nullTime(p.BirthDate), int16(p.Gender),
		nullIP(p.RegistrationIP), nullTime(p.RegisteredAt), nullTime(p.LastLoginAt),
		meta, p.Version, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *Repo) Update(ctx context.Context, p *player.Player) error {
	ex := pickExecutor(ctx, r.db)

	meta, _ := json.Marshal(p.Metadata)

	// optimistic lock by version
	const q = `
UPDATE players
   SET phone=$2,
       status=$3,
       status_reason=$4,
       country_code=$5, locale=$6, time_zone=$7,
       first_name=$8, last_name=$9, birth_date=$10, gender=$11,
       registration_ip=$12,
       registered_at=$13, last_login_at=$14,
       metadata=$15,
       version=$16,
       updated_at=$17
 WHERE id=$1 AND version=$18
`
	res, err := ex.ExecContext(ctx, q,
		p.ID,
		nullStr(p.Phone),
		int16(p.Status),
		nullStr(p.StatusReason),
		nullStr(p.Address.CountryCode), nullStr(p.Address.Locale), nullStr(p.Address.TimeZone),
		nullStr(p.FirstName), nullStr(p.LastName), nullTime(p.BirthDate), int16(p.Gender),
		nullIP(p.RegistrationIP),
		nullTime(p.RegisteredAt), nullTime(p.LastLoginAt),
		meta,
		p.Version,
		p.UpdatedAt,
		p.Version-1,
	)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return player.ErrConflict
	}
	return nil
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullIP(ip net.IP) any {
	if ip == nil || len(ip) == 0 {
		return nil
	}
	return ip.String()
}
