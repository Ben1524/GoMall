package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/models"
	"github.com/Ben1524/GoMall/common/utils"
)

type OrderService struct {
	orders map[int64]*models.Order
	nextID int64
	mu     sync.RWMutex
}

func NewOrderService() *OrderService {
	return &OrderService{
		orders: make(map[int64]*models.Order),
		nextID: 1,
	}
}

func (s *OrderService) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 || len(req.Items) == 0 {
		utils.WriteError(w, http.StatusBadRequest, "Invalid order data")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate total price (in production, this would fetch actual product prices)
	var totalPrice float64
	orderItems := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		// Mock price calculation
		price := float64(item.Quantity) * 100.0
		totalPrice += price

		orderItems[i] = models.OrderItem{
			ID:        int64(i + 1),
			OrderID:   s.nextID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		}
	}

	order := &models.Order{
		ID:         s.nextID,
		UserID:     req.UserID,
		TotalPrice: totalPrice,
		Status:     models.OrderStatusPending,
		Items:      orderItems,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	s.orders[s.nextID] = order
	s.nextID++

	utils.WriteJSON(w, http.StatusCreated, models.SuccessResponse(order))
}

func (s *OrderService) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	var id int64
	fmt.Sscanf(orderID, "%d", &id)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if order, exists := s.orders[id]; exists {
		utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(order))
		return
	}

	utils.WriteError(w, http.StatusNotFound, "Order not found")
}

func (s *OrderService) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	userID := r.URL.Query().Get("user_id")
	orders := make([]*models.Order, 0)

	if userID != "" {
		var uid int64
		fmt.Sscanf(userID, "%d", &uid)

		for _, order := range s.orders {
			if order.UserID == uid {
				orders = append(orders, order)
			}
		}
	} else {
		for _, order := range s.orders {
			orders = append(orders, order)
		}
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(orders))
}

func (s *OrderService) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		OrderID int64              `json:"order_id"`
		Status  models.OrderStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if order, exists := s.orders[req.OrderID]; exists {
		order.Status = req.Status
		order.UpdatedAt = time.Now()
		utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(order))
		return
	}

	utils.WriteError(w, http.StatusNotFound, "Order not found")
}

func main() {
	cfg := config.Load()
	service := NewOrderService()

	http.HandleFunc("/orders", service.ListOrders)
	http.HandleFunc("/order", service.GetOrder)
	http.HandleFunc("/order/create", service.CreateOrder)
	http.HandleFunc("/order/status", service.UpdateOrderStatus)

	port := cfg.Port
	if port == "" {
		port = "8083"
	}

	log.Printf("Order Service starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
