package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateHandler(t *testing.T) {
	// Create a new Gin router and add the validateHandler function as a route
	router := gin.Default()
	router.POST("/validate", validateHandler)

	// Create a sample request body
	reqBody := map[string]string{
		"ip": "192.168.1.1",
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create a new HTTP request with the sample request body
	req, err := http.NewRequest("POST", "/validate", bytes.NewBuffer(reqBytes))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP recorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the validateHandler function with the sample request and recorder
	router.ServeHTTP(recorder, req)

	// Check that the response status code is 200 OK
	if recorder.Code != http.StatusOK {
		t.Errorf("Unexpected status code: got %v, want %v", recorder.Code, http.StatusOK)
	}

	// Check that the response body contains the expected JSON
	expected := ValidateIPResponse{Status: true}
	var actual ValidateIPResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	if actual != expected {
		t.Errorf("Unexpected response body: got %v, want %v", actual, expected)
	}
}

func TestLookupHandler(t *testing.T) {
	// Create a new Gin router and add the lookupHandler function as a route
	router := gin.Default()
	router.GET("/lookup", lookupHandler)

	// Create a sample query parameter
	params := url.Values{}
	params.Set("domain", "example.com")

	// Create a new HTTP request with the sample query parameter
	req, err := http.NewRequest("GET", "/lookup?"+params.Encode(), nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Create a new HTTP recorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the lookupHandler function with the sample request and recorder
	router.ServeHTTP(recorder, req)

	// Check that the response status code is 200 OK
	if recorder.Code != http.StatusOK {
		t.Errorf("Unexpected status code: got %v, want %v", recorder.Code, http.StatusOK)
	}

	// Check that the response body contains the expected JSON
	var actual Query
	if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	if actual.Domain != "example.com" {
		t.Errorf("Unexpected domain name: got %v, want %v", actual.Domain, "example.com")
	}
	if len(actual.Addresses) == 0 {
		t.Errorf("No IP addresses found for domain: %v", actual.Domain)
	}
}

func TestRootHandler(t *testing.T) {
	// Create a new Gin router and add the rootHandler function as a route
	router := gin.Default()
	router.GET("/", rootHandler)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Create a new HTTP recorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the rootHandler function with the sample request and recorder
	router.ServeHTTP(recorder, req)

	// Check that the response status code is 200 OK
	if recorder.Code != http.StatusOK {
		t.Errorf("Unexpected status code: got %v, want %v", recorder.Code, http.StatusOK)
	}

	// Check that the response body contains the expected JSON
	var actual RootResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	if actual.Version != "0.1.0" {
		t.Errorf("Unexpected version: got %v, want %v", actual.Version, "0.1.0")
	}
	if actual.Date == 0 {
		t.Error("Expected date to be set, but it was not")
	}
}
