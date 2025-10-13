package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": token,
		"email": user.Email,
		"id":    user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": token,
		"email": user.Email,
		"id":    user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// Refresh - обновление JWT токена
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Используем AuthMiddleware для проверки токена
	// Затем генерируем новый токен для того же пользователя
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	newToken, err := h.service.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": newToken,
		"email": user.Email,
		"id":    user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// Logout - выход из системы (базовая реализация)
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// В будущем можно добавить blacklist токенов в Redis
	// Сейчас просто возвращаем успешный ответ
	response := map[string]string{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔐 AuthMiddleware: Checking authorization...")

		tokenString := r.Header.Get("Authorization")
		log.Printf("🔐 AuthMiddleware: Token header: %s", tokenString)

		if tokenString == "" {
			log.Println("🔐 AuthMiddleware: No Authorization header")
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Убираем "Bearer " префикс
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		log.Printf("🔐 AuthMiddleware: Validating token: %s...", tokenString[:10])

		userID, email, err := h.service.ValidateToken(tokenString)
		if err != nil {
			log.Printf("🔐 AuthMiddleware: Token validation failed: %v", err)
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		log.Printf("🔐 AuthMiddleware: Token valid - UserID: %d, Email: %s", userID, email)

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userID)
		ctx = context.WithValue(ctx, "userEmail", email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
