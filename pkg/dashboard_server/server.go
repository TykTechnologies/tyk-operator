package dashboard_server

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi"
)

type Server struct {
}

func (s Server) Start() *httptest.Server {
	r := chi.NewMux()

	r.Route("/api", func(r chi.Router) {
		r.Route("/apis", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			})
		})

		r.Route("/policies", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("world"))
			})
		})
	})

	return httptest.NewServer(r)
}
