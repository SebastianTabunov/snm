package user

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// В реальном приложении userID берется из контекста после аутентификации
	// Сейчас используем заглушку для тестирования
	userID := 1

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to get profile"}`, http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, `{"error": "Profile not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(profile)
	if err != nil {
		return
	}
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	profile := &Profile{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Address:   req.Address,
	}

	err := h.service.UpdateProfile(userID, profile)
	if err != nil {
		http.Error(w, `{"error": "Failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"status": "profile updated"})
	if err != nil {
		return
	}
}
