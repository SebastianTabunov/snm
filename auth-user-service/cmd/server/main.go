package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes - FIXED VERSION
	r.Post("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"message": "register endpoint"}`))
		if err != nil {
			return
		}
	})

	r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"message": "login endpoint"}`))
		if err != nil {
			return
		}
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		// Пока без middleware - просто тестируем
		r.Get("/user/profile", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"message": "user profile"}`))
			if err != nil {
				return
			}
		})

		r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"message": "user orders"}`))
			if err != nil {
				return
			}
		})
	})

	// Health check - FIXED PATH
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("✅ OK"))
		if err != nil {
			return
		}
	})

	log.Println("🚀 Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
