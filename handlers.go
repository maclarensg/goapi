package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @BasePath /v1

// Get queries history
// @Summary Returns a list of all queries made to the API
// @Description Returns a list of all queries made to the API
// @ID queries_history
// @Produce json
// @Success 200 {array} Query
// @Failure 400 {object} HTTPError
// @Router /v1/history [get]
func historyHandler(c *gin.Context) {
	// Retrieve the latest 20 saved queries from Redis
	queriesJSON, err := redisClient.LRange(ctx, RedisKey, 0, int64(maxQueryHistory)-1).Result()
	if err != nil {
		c.JSON(500, HTTPError{"Error retrieving query history from Redis"})
		return
	}

	queries := make([]Query, len(queriesJSON))
	for i, queryJSON := range queriesJSON {
		err = json.Unmarshal([]byte(queryJSON), &queries[i])
		if err != nil {
			fmt.Println("Error unmarshaling query JSON:", err)
			continue
		}
	}

	// Return the queries as a JSON response
	c.JSON(200, queries)
}

// Check the health of the server
// @Summary Returns information about the health of the server
// @Description Returns information about the health of the server, including the current time, uptime, and database connection status.
// @ID check_health
// @Produce json
// @Success 200
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Simple IP validation
// @Summary Validates an IP address
// @Description Validates an IP address (IPv4 or IPv6)
// @ID validate_ip
// @Accept json
// @Produce json
// @Param request body ValidateIPRequest true "IP to validate"
// @Success 200 {object} ValidateIPResponse
// @Failure 400 {object} HTTPError
// @Router /v1/tools/validate [post]
func validateHandler(c *gin.Context) {
	var req ValidateIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, HTTPError{"Invalid request: " + err.Error()})
		return
	}

	ip := net.ParseIP(req.IP)
	if ip == nil {
		c.JSON(400, HTTPError{"Invalid IP address"})
		return
	}

	if ip.To4() == nil && ip.To16() == nil {
		c.JSON(400, HTTPError{"Invalid IP address"})
		return
	}

	c.JSON(200, ValidateIPResponse{Status: true})
}

// Lookup domain
// @Summary Performs a DNS lookup for the specified domain and returns all IPv4 addresses
// @Description Performs a DNS lookup for the specified domain and returns all IPv4 addresses
// @ID lookup_domain
// @Param domain query string true "Domain name"
// @Produce json
// @Success 200 {object} Query
// @Failure 400 {object} HTTPError
// @Failure 404 {object} HTTPError
// @Router /v1/tools/lookup [get]
func lookupHandler(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(400, HTTPError{"Missing required parameter: domain"})
		return
	}

	ips, err := net.LookupIP(domain)
	if err != nil {
		c.JSON(404, HTTPError{"Unable to find IP addresses for domain: " + domain})
		return
	}

	var addresses []Address
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			addresses = append(addresses, Address{IP: ipv4.String()})
		}
	}

	q := Query{
		Addresses: addresses,
		ClientIP:  c.ClientIP(),
		CreatedAt: time.Now().Unix(),
		Domain:    domain,
	}

	logQuery(q)

	c.JSON(200, q)
}

// rootHandler handles requests to the root endpoint and returns information about the application.
//
// Returns an object containing the version of the application, the current date and time, and a boolean value indicating whether the application is running in a Kubernetes cluster.
//
// @Summary Get information about the application
// @Description Get information about the application, including its version, date, and Kubernetes status
// @Produce json
// @Success 200 {object} RootResponse
// @Router / [get]
func rootHandler(c *gin.Context) {
	resp := RootResponse{
		Version:    "0.1.0",
		Date:       time.Now().Unix(),
		Kubernetes: isRunningInKubernetes(),
	}
	c.JSON(http.StatusOK, resp)
}
