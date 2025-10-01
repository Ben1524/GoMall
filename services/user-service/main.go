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

type UserService struct {
	users  map[int64]*models.User
	nextID int64
	mu     sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		users:  make(map[int64]*models.User),
		nextID: 1,
	}
}

func (s *UserService) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username already exists
	for _, user := range s.users {
		if user.Username == req.Username {
			utils.WriteError(w, http.StatusConflict, "Username already exists")
			return
		}
	}

	user := &models.User{
		ID:        s.nextID,
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // In production, hash this!
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.users[s.nextID] = user
	s.nextID++

	utils.WriteJSON(w, http.StatusCreated, models.SuccessResponse(user))
}

func (s *UserService) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == req.Username && user.Password == req.Password {
			utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(user))
			return
		}
	}

	utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
}

func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user ID from query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		utils.WriteError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var id int64
	fmt.Sscanf(userID, "%d", &id)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if user, exists := s.users[id]; exists {
		utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(user))
		return
	}

	utils.WriteError(w, http.StatusNotFound, "User not found")
}

func (s *UserService) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(users))
}

func main() {
	cfg := config.Load()
	service := NewUserService()

	http.HandleFunc("/register", service.Register)
	http.HandleFunc("/login", service.Login)
	http.HandleFunc("/user", service.GetUser)
	http.HandleFunc("/users", service.ListUsers)

	port := cfg.Port
	if port == "" {
		port = "8081"
	}

	log.Printf("User Service starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
