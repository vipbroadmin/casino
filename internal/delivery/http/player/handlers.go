package playerhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"players_service/internal/domain/player"
	playeruc "players_service/internal/usecase/player"
)

type HTTP struct {
	uc *playeruc.Service
}

func New(uc *playeruc.Service) *HTTP {
	return &HTTP{uc: uc}
}

type createReq struct {
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	BirthDate      string         `json:"birth_date"` // YYYY-MM-DD
	Gender         string         `json:"gender"`     // none|male|female|other
	CountryCode    string         `json:"country_code"`
	Locale         string         `json:"locale"`
	TimeZone       string         `json:"time_zone"`
	RegistrationIP string         `json:"registration_ip"`
	Metadata       map[string]any `json:"metadata"`
	RegisteredAt   string         `json:"registered_at"` // RFC3339 optional
}

func (h *HTTP) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	var birth time.Time
	if strings.TrimSpace(req.BirthDate) != "" {
		t, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad_birth_date")
			return
		}
		birth = t
	}

	var regAt time.Time
	if strings.TrimSpace(req.RegisteredAt) != "" {
		t, err := time.Parse(time.RFC3339, req.RegisteredAt)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad_registered_at")
			return
		}
		regAt = t
	}

	p, err := h.uc.CreatePlayer(r.Context(), playeruc.CreatePlayerCmd{
		Email:          req.Email,
		Phone:          req.Phone,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		BirthDate:      birth,
		Gender:         req.Gender,
		CountryCode:    req.CountryCode,
		Locale:         req.Locale,
		TimeZone:       req.TimeZone,
		RegistrationIP: req.RegistrationIP,
		Metadata:       req.Metadata,
		RegisteredAt:   regAt,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPlayerDTO(p))
}

type changeStatusReq struct {
	ToStatus string `json:"to_status"` // active|blocked|frozen|closed
	Reason   string `json:"reason"`
	Actor    string `json:"actor"` // player|administrator|system
}

func (h *HTTP) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req changeStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	actor := parseActor(req.Actor)

	p, ev, err := h.uc.ChangeStatus(r.Context(), playeruc.ChangeStatusCmd{
		PlayerID: id,
		ToStatus: req.ToStatus,
		Reason:   req.Reason,
		Actor:    actor,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"player": toPlayerDTO(p),
		"event":  toEventDTO(ev),
	})
}

func (h *HTTP) GetPlayer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	p, err := h.uc.GetPlayer(r.Context(), id)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPlayerDTO(p))
}

// --- dto/mapping ---

func toPlayerDTO(p *player.Player) map[string]any {
	return map[string]any{
		"id":            p.ID.String(),
		"email":         p.Email,
		"phone":         p.Phone,
		"status":        p.Status.String(),
		"status_reason": p.StatusReason,
		"address": map[string]any{
			"country_code": p.Address.CountryCode,
			"locale":       p.Address.Locale,
			"time_zone":    p.Address.TimeZone,
		},
		"first_name":      p.FirstName,
		"last_name":       p.LastName,
		"birth_date":      fmtDate(p.BirthDate),
		"gender":          p.Gender.String(),
		"registration_ip": fmtIP(p.RegistrationIP),
		"registered_at":   fmtTime(p.RegisteredAt),
		"last_login_at":   fmtTime(p.LastLoginAt),
		"metadata":        p.Metadata,
		"version":         p.Version,
		"created_at":      fmtTime(p.CreatedAt),
		"updated_at":      fmtTime(p.UpdatedAt),
	}
}

func toEventDTO(ev player.PlayerStatusEvent) map[string]any {
	return map[string]any{
		"id":          ev.ID.String(),
		"player_id":   ev.PlayerID.String(),
		"from_status": ev.From.String(),
		"to_status":   ev.To.String(),
		"reason":      ev.Reason,
		"actor_type":  ev.ActorType.String(),
		"created_at":  fmtTime(ev.CreatedAt),
	}
}

func fmtTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Format(time.RFC3339)
}

func fmtDate(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.Format("2006-01-02")
}

func fmtIP(ip any) any {
	// ip is net.IP in entity; keep as string or nil
	if ip == nil {
		return nil
	}
	// net.IP implements fmt.Stringer
	return fmt.Sprintf("%v", ip)
}

func parseActor(s string) player.ActorType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "player":
		return player.ActorPlayer
	case "administrator", "administator", "admin":
		return player.ActorAdmin
	default:
		return player.ActorSystem
	}
}

// --- errors ---

func encodeDomainErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, player.ErrNotFound):
		writeErr(w, http.StatusNotFound, "not_found")
	case errors.Is(err, player.ErrConflict):
		writeErr(w, http.StatusConflict, "conflict")
	case errors.Is(err, player.ErrValidation),
		errors.Is(err, player.ErrInvalidEmail),
		errors.Is(err, player.ErrInvalidPhone),
		errors.Is(err, player.ErrInvalidStatus),
		errors.Is(err, player.ErrInvalidGender),
		errors.Is(err, player.ErrInvalidCountryCode),
		errors.Is(err, player.ErrInvalidLocale),
		errors.Is(err, player.ErrInvalidTimeZone):
		writeErr(w, http.StatusBadRequest, "validation")
	default:
		writeErr(w, http.StatusInternalServerError, "internal")
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, kind string) {
	writeJSON(w, code, map[string]any{"error": kind})
}
