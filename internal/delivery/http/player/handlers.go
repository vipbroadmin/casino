package playerhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	Currency       string         `json:"currency"`
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
		Currency:       req.Currency,
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

// --- admin endpoints ---

// ListPlayers handles GET /users/players
func (h *HTTP) ListPlayers(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	country := r.URL.Query().Get("country")
	currency := r.URL.Query().Get("currency")
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	query := playeruc.ListPlayersQuery{
		Offset:   offset,
		Limit:    limit,
		Search:   search,
		Country:  country,
		Currency: currency,
		SortBy:   sortBy,
		Order:    order,
	}
	if search != "" {
		if id, err := uuid.Parse(search); err == nil {
			query.SearchPlayerID = &id
			query.Search = ""
		}
	}

	rows, total, err := h.uc.ListPlayers(r.Context(), query)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	items := make([]map[string]any, len(rows))
	for i, row := range rows {
		items[i] = toAdminPlayerDTO(row)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
	})
}

// CreatePlayerAdmin handles POST /users/players (admin version with login/password)
func (h *HTTP) CreatePlayerAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login     string `json:"login"`
		Password  string `json:"password"`
		Country   string `json:"country"`
		Currency  string `json:"currency"`
		PromoCode string `json:"promoCode,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	if req.Login == "" || req.Password == "" || req.Country == "" || req.Currency == "" {
		writeErr(w, http.StatusBadRequest, "validation")
		return
	}

	// Use CreatePlayerAdmin usecase method
	err := h.uc.CreatePlayerAdmin(r.Context(), playeruc.CreatePlayerAdminCmd{
		Login:     req.Login,
		Password:  req.Password,
		Country:   req.Country,
		Currency:  req.Currency,
		PromoCode: req.PromoCode,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// GetPlayerAdmin handles GET /users/players/{id}
func (h *HTTP) GetPlayerAdmin(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	// Fetch admin row
	rows, err := h.uc.GetPlayersInfo(r.Context(), []uuid.UUID{id})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	if len(rows) == 0 {
		writeErr(w, http.StatusNotFound, "not_found")
		return
	}

	writeJSON(w, http.StatusOK, toAdminPlayerDTO(rows[0]))
}

// GetPlayersInfo handles POST /users/players/getInfo
func (h *HTTP) GetPlayersInfo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad_id")
			return
		}
		ids = append(ids, id)
	}

	rows, err := h.uc.GetPlayersInfo(r.Context(), ids)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	items := make([]map[string]any, len(rows))
	for i, row := range rows {
		items[i] = toAdminPlayerDTO(row)
	}

	writeJSON(w, http.StatusOK, items)
}

// UpdatePlayer handles PUT /users/players/{id}/update
func (h *HTTP) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req struct {
		Login    *string `json:"login,omitempty"`
		Email    *string `json:"email,omitempty"`
		Phone    *string `json:"phone,omitempty"`
		Name     *string `json:"name,omitempty"`
		Surname  *string `json:"surname,omitempty"`
		Nickname *string `json:"nickname,omitempty"`
		Currency *string `json:"currency,omitempty"`
		Country  *string `json:"country,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	err = h.uc.UpdatePlayerProfile(r.Context(), playeruc.UpdatePlayerProfileCmd{
		ID:       id,
		Login:    req.Login,
		Email:    req.Email,
		Phone:    req.Phone,
		Name:     req.Name,
		Surname:  req.Surname,
		Nickname: req.Nickname,
		Currency: req.Currency,
		Country:  req.Country,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// UpdatePassword handles PUT /users/players/{id}/update/pass
func (h *HTTP) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	if req.Password == "" {
		writeErr(w, http.StatusBadRequest, "validation")
		return
	}

	err = h.uc.UpdatePlayerPassword(r.Context(), playeruc.UpdatePlayerPasswordCmd{
		ID:          id,
		NewPassword: req.Password,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// BanPlayer handles PUT /users/players/{id}/ban
func (h *HTTP) BanPlayer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	err = h.uc.BanPlayer(r.Context(), id)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// UnbanPlayer handles PUT /users/players/{id}/unban
func (h *HTTP) UnbanPlayer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	err = h.uc.UnbanPlayer(r.Context(), id)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// UpdateLevel handles PUT /users/players/{id}/update/level
func (h *HTTP) UpdateLevel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req struct {
		Level int `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	if req.Level < 1 {
		writeErr(w, http.StatusBadRequest, "validation")
		return
	}

	err = h.uc.UpdatePlayerLevel(r.Context(), playeruc.UpdatePlayerLevelCmd{
		ID:    id,
		Level: req.Level,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// KickPlayers handles POST /users/players/kick
func (h *HTTP) KickPlayers(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerIDs []string `json:"playerIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	if len(req.PlayerIDs) == 0 {
		writeErr(w, http.StatusBadRequest, "validation")
		return
	}

	ids := make([]uuid.UUID, len(req.PlayerIDs))
	for i, idStr := range req.PlayerIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "bad_id")
			return
		}
		ids[i] = id
	}

	err := h.uc.KickPlayers(r.Context(), playeruc.KickPlayersCmd{
		PlayerIDs: ids,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// GetPlayerDocuments handles GET /users/players/{id}/documents
func (h *HTTP) GetPlayerDocuments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	docs, err := h.uc.GetPlayerDocuments(r.Context(), id)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	items := make([]map[string]any, len(docs))
	for i, doc := range docs {
		items[i] = map[string]any{
			"id":        doc.ID.String(),
			"playerId":  doc.PlayerID.String(),
			"type":      doc.Type,
			"status":    doc.Status.String(),
			"fileUrl":   doc.FileURL,
			"metadata":  doc.Metadata,
			"createdAt": doc.CreatedAt.Format(time.RFC3339),
			"updatedAt": doc.UpdatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, items)
}

// UpdateDocumentStatus handles PATCH /users/players/document/{id}
func (h *HTTP) UpdateDocumentStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	if req.Status == "" {
		writeErr(w, http.StatusBadRequest, "validation")
		return
	}

	err = h.uc.UpdateDocumentStatus(r.Context(), playeruc.UpdateDocumentStatusCmd{
		ID:     id,
		Status: req.Status,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// GetPlayerRequisites handles GET /finances/player-requisites-v2/{id}
func (h *HTTP) GetPlayerRequisites(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	req, err := h.uc.GetPlayerRequisites(r.Context(), id)
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":              req.ID.String(),
		"playerId":        req.PlayerID.String(),
		"paymentMethodId": req.PaymentMethodID.String(),
		"formData":        req.FormData,
		"createdAt":       req.CreatedAt.Format(time.RFC3339),
		"updatedAt":       req.UpdatedAt.Format(time.RFC3339),
	})
}

// UpdatePlayerRequisites handles POST /finances/player-requisites-v2/{id}
func (h *HTTP) UpdatePlayerRequisites(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_id")
		return
	}

	var req struct {
		PaymentMethodID string         `json:"paymentMethodId"`
		FormData        map[string]any `json:"formData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad_json")
		return
	}

	paymentMethodID, err := uuid.Parse(req.PaymentMethodID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_payment_method_id")
		return
	}

	err = h.uc.UpdatePlayerRequisites(r.Context(), playeruc.UpdatePlayerRequisitesCmd{
		PlayerID:        id,
		PaymentMethodID: paymentMethodID,
		FormData:        req.FormData,
	})
	if err != nil {
		encodeDomainErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, true)
}

// toAdminPlayerDTO converts PlayerRow to admin API format
func toAdminPlayerDTO(row playeruc.PlayerRow) map[string]any {
	return map[string]any{
		"id":        row.ID.String(),
		"login":     row.Login,
		"email":     row.Email,
		"phone":     row.Phone,
		"name":      row.Name,
		"surname":   row.Surname,
		"nickname":  row.Nickname,
		"currency":  row.Currency,
		"country":   row.Country,
		"isBanned":  row.IsBanned,
		"level":     row.Level,
		"createdAt": row.CreatedAt.Format(time.RFC3339),
	}
}
