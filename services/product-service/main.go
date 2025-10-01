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

type ProductService struct {
	products map[int64]*models.Product
	nextID   int64
	mu       sync.RWMutex
}

func NewProductService() *ProductService {
	service := &ProductService{
		products: make(map[int64]*models.Product),
		nextID:   1,
	}
	// Add some sample products
	service.addSampleProducts()
	return service
}

func (s *ProductService) addSampleProducts() {
	sampleProducts := []models.CreateProductRequest{
		{
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999.99,
			Stock:       10,
			Category:    "Electronics",
		},
		{
			Name:        "Smartphone",
			Description: "Latest model smartphone",
			Price:       699.99,
			Stock:       20,
			Category:    "Electronics",
		},
		{
			Name:        "Headphones",
			Description: "Wireless noise-cancelling headphones",
			Price:       199.99,
			Stock:       50,
			Category:    "Electronics",
		},
	}

	for _, req := range sampleProducts {
		product := &models.Product{
			ID:          s.nextID,
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			Stock:       req.Stock,
			Category:    req.Category,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		s.products[s.nextID] = product
		s.nextID++
	}
}

func (s *ProductService) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Price <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "Invalid product data")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	product := &models.Product{
		ID:          s.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.products[s.nextID] = product
	s.nextID++

	utils.WriteJSON(w, http.StatusCreated, models.SuccessResponse(product))
}

func (s *ProductService) GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	productID := r.URL.Query().Get("id")
	if productID == "" {
		utils.WriteError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var id int64
	fmt.Sscanf(productID, "%d", &id)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if product, exists := s.products[id]; exists {
		utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(product))
		return
	}

	utils.WriteError(w, http.StatusNotFound, "Product not found")
}

func (s *ProductService) ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]*models.Product, 0, len(s.products))
	for _, product := range s.products {
		products = append(products, product)
	}

	utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(products))
}

func (s *ProductService) UpdateStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		ProductID int64 `json:"product_id"`
		Stock     int   `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if product, exists := s.products[req.ProductID]; exists {
		product.Stock = req.Stock
		product.UpdatedAt = time.Now()
		utils.WriteJSON(w, http.StatusOK, models.SuccessResponse(product))
		return
	}

	utils.WriteError(w, http.StatusNotFound, "Product not found")
}

func main() {
	cfg := config.Load()
	service := NewProductService()

	http.HandleFunc("/products", service.ListProducts)
	http.HandleFunc("/product", service.GetProduct)
	http.HandleFunc("/product/create", service.CreateProduct)
	http.HandleFunc("/product/stock", service.UpdateStock)

	port := cfg.Port
	if port == "" {
		port = "8082"
	}

	log.Printf("Product Service starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
