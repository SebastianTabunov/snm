package main

import (
	"auth-user-service/internal/auth"
	"auth-user-service/internal/config"
	"auth-user-service/internal/order"
	"auth-user-service/internal/user"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	// Инициализация сервисов
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Роутинг
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)
		r.Get("/user/profile", userHandler.GetProfile)
		r.Get("/orders", orderHandler.GetUserOrders)
		r.Post("/orders", orderHandler.CreateOrder)
	})

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
