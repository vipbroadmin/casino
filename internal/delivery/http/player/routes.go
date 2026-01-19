package playerhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *HTTP) http.Handler {
	r := chi.NewRouter()

	// Admin endpoints under /users/players
	r.Route("/users/players", func(r chi.Router) {
		r.Get("/", h.ListPlayers)
		r.Post("/", h.CreatePlayerAdmin)
		r.Get("/{id}", h.GetPlayerAdmin)
		r.Post("/getInfo", h.GetPlayersInfo)
		r.Put("/{id}/update", h.UpdatePlayer)
		r.Put("/{id}/update/pass", h.UpdatePassword)
		r.Put("/{id}/update/level", h.UpdateLevel)
		r.Put("/{id}/ban", h.BanPlayer)
		r.Put("/{id}/unban", h.UnbanPlayer)
		r.Post("/kick", h.KickPlayers)
		r.Get("/{id}/documents", h.GetPlayerDocuments)
		r.Patch("/document/{id}", h.UpdateDocumentStatus)
	})

	// Finances endpoints
	r.Route("/finances/player-requisites-v2", func(r chi.Router) {
		r.Get("/{id}", h.GetPlayerRequisites)
		r.Post("/{id}", h.UpdatePlayerRequisites)
	})

	// Legacy endpoints under /players
	r.Route("/players", func(r chi.Router) {
		r.Post("/", h.CreatePlayer)
		r.Post("/{id}/status", h.ChangeStatus)
		r.Get("/{id}", h.GetPlayer)
	})

	return r
}
