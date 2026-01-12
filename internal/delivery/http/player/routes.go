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
		r.Put("/{id}/ban", h.BanPlayer)
		r.Put("/{id}/unban", h.UnbanPlayer)
	})

	// Legacy endpoints under /players
	r.Route("/players", func(r chi.Router) {
		r.Post("/", h.CreatePlayer)
		r.Post("/{id}/status", h.ChangeStatus)
		r.Get("/{id}", h.GetPlayer)
	})

	return r
}
