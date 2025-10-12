package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"auth-user-service/internal/auth"
	"auth-user-service/internal/database"
	"auth-user-service/internal/order"
	"auth-user-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Конфигурация БД
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "pass"),
		DBName:   getEnv("DB_NAME", "auth_service"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Подключаемся к PostgreSQL
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	log.Println("✅ Database connected successfully")

	// Инициализация сервисов с реальной БД
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, getEnv("JWT_SECRET", "fallback-secret-key"))
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)

		r.Get("/orders", orderHandler.GetUserOrders)
		r.Get("/orders/{id}", orderHandler.GetOrder)
		r.Post("/orders", orderHandler.CreateOrder)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err2 := w.Write([]byte("✅ OK"))
		if err2 != nil {
			return
		}
	})

	port := getEnv("PORT", "8080")
	log.Printf("🚀 Server starting on :%s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
