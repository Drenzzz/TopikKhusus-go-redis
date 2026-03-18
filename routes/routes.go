package routes

import (
	"net/http"
	"strings"

	"topikkhusus-methodtracker/internal/handlers"
	"topikkhusus-methodtracker/internal/middleware"
)

func Register(userHandler *handlers.UserHandler, middlewares ...middleware.Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetAllUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetUserByID(w, r, id)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", userHandler.Health)

	return middleware.Chain(mux, middlewares...)
}
