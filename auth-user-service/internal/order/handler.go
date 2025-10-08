package order

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type CreateOrderRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(orderID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to get order"}`, http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		return
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, `{"error": "Title is required"}`, http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		http.Error(w, `{"error": "Price must be positive"}`, http.StatusBadRequest)
		return
	}

	order, err := h.service.CreateOrder(userID, req.Title, req.Description, req.Price)
	if err != nil {
		http.Error(w, `{"error": "Failed to create order"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		return
	}
}

func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := 1 // Заглушка

	orders, err := h.service.GetUserOrders(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to get orders"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		return
	}
}
