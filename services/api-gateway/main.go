package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/utils"
)

type APIGateway struct {
	userServiceURL    string
	productServiceURL string
	orderServiceURL   string
}

func NewAPIGateway() *APIGateway {
	return &APIGateway{
		userServiceURL:    "http://localhost:8081",
		productServiceURL: "http://localhost:8082",
		orderServiceURL:   "http://localhost:8083",
	}
}

func (g *APIGateway) proxyRequest(targetURL string, w http.ResponseWriter, r *http.Request) {
	// Read body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Create new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		utils.WriteError(w, http.StatusBadGateway, fmt.Sprintf("Service unavailable: %v", err))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *APIGateway) handleUserRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/register":
		g.proxyRequest(g.userServiceURL+"/register", w, r)
	case "/api/login":
		g.proxyRequest(g.userServiceURL+"/login", w, r)
	case "/api/user":
		g.proxyRequest(g.userServiceURL+"/user"+r.URL.RawQuery, w, r)
	case "/api/users":
		g.proxyRequest(g.userServiceURL+"/users", w, r)
	default:
		utils.WriteError(w, http.StatusNotFound, "Route not found")
	}
}

func (g *APIGateway) handleProductRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/products":
		g.proxyRequest(g.productServiceURL+"/products", w, r)
	case "/api/product":
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		g.proxyRequest(g.productServiceURL+"/product"+query, w, r)
	case "/api/product/create":
		g.proxyRequest(g.productServiceURL+"/product/create", w, r)
	case "/api/product/stock":
		g.proxyRequest(g.productServiceURL+"/product/stock", w, r)
	default:
		utils.WriteError(w, http.StatusNotFound, "Route not found")
	}
}

func (g *APIGateway) handleOrderRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/orders":
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		g.proxyRequest(g.orderServiceURL+"/orders"+query, w, r)
	case "/api/order":
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		g.proxyRequest(g.orderServiceURL+"/order"+query, w, r)
	case "/api/order/create":
		g.proxyRequest(g.orderServiceURL+"/order/create", w, r)
	case "/api/order/status":
		g.proxyRequest(g.orderServiceURL+"/order/status", w, r)
	default:
		utils.WriteError(w, http.StatusNotFound, "Route not found")
	}
}

func (g *APIGateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"services": map[string]string{
			"user":    g.userServiceURL,
			"product": g.productServiceURL,
			"order":   g.orderServiceURL,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (g *APIGateway) router(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case r.URL.Path == "/health":
		g.handleHealth(w, r)
	case r.URL.Path == "/api/register" || r.URL.Path == "/api/login" || 
		 r.URL.Path == "/api/user" || r.URL.Path == "/api/users":
		g.handleUserRoutes(w, r)
	case r.URL.Path == "/api/products" || r.URL.Path == "/api/product" || 
		 r.URL.Path == "/api/product/create" || r.URL.Path == "/api/product/stock":
		g.handleProductRoutes(w, r)
	case r.URL.Path == "/api/orders" || r.URL.Path == "/api/order" || 
		 r.URL.Path == "/api/order/create" || r.URL.Path == "/api/order/status":
		g.handleOrderRoutes(w, r)
	default:
		utils.WriteError(w, http.StatusNotFound, "Route not found")
	}
}

func main() {
	cfg := config.Load()
	gateway := NewAPIGateway()

	http.HandleFunc("/", gateway.router)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s\n", port)
	log.Printf("Routing to:\n")
	log.Printf("  - User Service: %s\n", gateway.userServiceURL)
	log.Printf("  - Product Service: %s\n", gateway.productServiceURL)
	log.Printf("  - Order Service: %s\n", gateway.orderServiceURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
