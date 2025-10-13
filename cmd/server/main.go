package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"auth-user-service/internal/auth"
	"auth-user-service/internal/database"
	"auth-user-service/internal/order"
	"auth-user-service/internal/redis"
	"auth-user-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Конфигурация БД
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "password"),
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

	// Подключаемся к Redis
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	var redisClient *redis.Client
	redisClient, err = redis.NewClient(redisURL)
	if err != nil {
		log.Printf("⚠️ Failed to connect to Redis: %v", err)
		log.Println("⚠️ Continuing without Redis...")
		redisClient = nil
	} else {
		defer func(redisClient *redis.Client) {
			err := redisClient.Close()
			if err != nil {

			}
		}(redisClient)
		log.Println("✅ Redis connected successfully")
	}

	// Инициализация сервисов
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, getEnv("JWT_SECRET", "fallback-secret-key"))
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, redisClient)
	userHandler := user.NewHandler(userService)

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Роутер
	r := chi.NewRouter()

	// CORS middleware - оптимизировано для Tilda
	r.Use(cors.Handler(cors.Options{
		// Разрешаем основные домены Tilda + локальная разработка
		AllowedOrigins:   getCORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "Origin", "Cache-Control"},
		ExposedHeaders:   []string{"Link", "Content-Length", "X-Total-Count"},
		AllowCredentials: true, // Важно для работы с куками/сессиями
		MaxAge:           300,
	}))

	// Базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected auth routes (требуют AuthMiddleware)
	r.With(authHandler.AuthMiddleware).Post("/auth/refresh", authHandler.Refresh)
	r.With(authHandler.AuthMiddleware).Post("/auth/logout", authHandler.Logout)

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)

		r.Get("/orders", orderHandler.GetUserOrders)
		r.Get("/orders/{id}", orderHandler.GetOrder)
		r.Post("/orders", orderHandler.CreateOrder)
	})

	// Специальные эндпоинты для Tilda
	r.Route("/tilda", func(r chi.Router) {
		// Webhook для Tilda
		r.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
			// Обработка вебхуков от Tilda
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"status":"ok"}`))
			if err != nil {
				return
			}
		})

		// Health check для Tilda
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err2 := w.Write([]byte(`{"status":"ok","service":"auth-user-service"}`))
			if err2 != nil {
				return
			}
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем PostgreSQL
		if err := db.Ping(); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}

		// Проверяем Redis если подключен
		if redisClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			testKey := "health_check"
			if err := redisClient.Set(ctx, testKey, "test", 1*time.Second); err != nil {
				http.Error(w, "Redis unavailable", http.StatusServiceUnavailable)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, err2 := w.Write([]byte(`{"status":"ok","database":"connected","redis":"connected"}`))
		if err2 != nil {
			return
		}
	})

	// Preflight handler для всех OPTIONS запросов
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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

// getCORSAllowedOrigins возвращает список разрешенных доменов для CORS
func getCORSAllowedOrigins() []string {
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if corsOrigins == "" {
		// По умолчанию разрешаем локальную разработку и основные домены Tilda
		return []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://localhost:3000",
			"https://localhost:8080",
			"https://*.tilda.ws",
			"https://*.tilda.com",
			"https://tilda.ws",
			"https://tilda.com",
		}
	}

	// Разбиваем строку из переменной окружения
	return strings.Split(corsOrigins, ",")
}
