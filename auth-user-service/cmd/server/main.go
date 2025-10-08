package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"auth-user-service/internal/auth"
	"auth-user-service/internal/order"
	"auth-user-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Получение переменных окружения
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/auth_service?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-2024")
	port := getEnv("PORT", "8080")

	// Подключение к БД
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	if err := db.Ping(); err != nil {
		log.Printf("⚠️ Database connection failed: %v", err)
		log.Println("⚠️ Starting without database...")
		// Продолжаем без БД для тестирования
		db = nil
	} else {
		log.Println("✅ Database connected successfully")
	}

	// Инициализация сервисов
	var authHandler *auth.Handler
	var userHandler *user.Handler
	var orderHandler *order.Handler

	if db != nil {
		// Реальные реализации с БД
		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, jwtSecret)
		authHandler = auth.NewHandler(authService)

		userRepo := user.NewRepository(db)
		userService := user.NewService(userRepo)
		userHandler = user.NewHandler(userService)

		orderRepo := order.NewRepository(db)
		orderService := order.NewService(orderRepo)
		orderHandler = order.NewHandler(orderService)
	} else {
		// Заглушки для тестирования без БД
		authHandler = &auth.Handler{}
		userHandler = &user.Handler{}
		orderHandler = &order.Handler{}
	}

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		if db != nil {
			r.Use(authHandler.AuthMiddleware)
		}

		// User routes
		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)

		// Order routes
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
