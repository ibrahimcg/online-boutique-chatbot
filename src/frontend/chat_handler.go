// chat_handler.go
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// ChatRequest represents the incoming chat message from the frontend
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse represents the response format
type ChatResponse struct {
	Response string `json:"response"`
}

const (
	CHATBOT_SERVICE_URL = "http://chatbot-service/api/chat"  // Direct to chatbot-service
	REQUEST_TIMEOUT     = 30 * time.Second
)

// chatHandler handles chat requests directly to chatbot-service
func (fe *frontendServer) chatHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	
	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}
	
	// Parse the request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}
	
	log.Printf("Received chat message: %s", req.Message)
	
	// Create client with timeout
	client := &http.Client{
		Timeout: REQUEST_TIMEOUT,
	}
	
	// Prepare request to chatbot service
	reqBody, err := json.Marshal(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to process request"})
		return
	}
	
	// Make request to chatbot service
	resp, err := client.Post(CHATBOT_SERVICE_URL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("Failed to connect to chatbot service: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to connect to chatbot service", "details": err.Error()})
		return
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read response"})
		return
	}
	
	// Forward the response
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}