package playerhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *HTTP) http.Handler {
	r := chi.NewRouter()

	r.Route("/players", func(r chi.Router) {
		r.Post("/", h.CreatePlayer)
		r.Post("/{id}/status", h.ChangeStatus)
		r.Get("/{id}", h.GetPlayer) // TODO: implement query usecase
	})

	return r
}
