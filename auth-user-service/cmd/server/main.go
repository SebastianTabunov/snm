package main

import (
	"database/sql"
	"log"
	"net/http"

	"auth-user-service/internal/auth"
	"auth-user-service/internal/order"
	"auth-user-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Подключение к БД
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/auth_service?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	// Инициализация auth
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, "your-secret-jwt-key")
	authHandler := auth.NewHandler(authService)

	// Инициализация user
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Инициализация order
	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные маршруты (аутентификация)
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Приватные маршруты
	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		// Профиль пользователя
		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)

		// Заказы
		r.Get("/orders", orderHandler.GetUserOrders)
		r.Get("/orders/{id}", orderHandler.GetOrder)
		r.Post("/orders", orderHandler.CreateOrder)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err2 := w.Write([]byte("OK"))
		if err2 != nil {
			return
		}
	})

	// Запуск сервера
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
