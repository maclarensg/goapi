package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	docs "github.com/maclarensg/goapi/docs"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var maxQueryHistory int
var redisClient *redis.Client
var ctx = context.Background()

const RedisKey = "queries"
const defaultMaxQueryHistory = 20

type Address struct {
	IP string `json:"ip"`
}

type Query struct {
	Addresses []Address `json:"addresses"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt int64     `json:"created_at"`
	Domain    string    `json:"domain"`
}

type QueryLog struct {
	Domain    string `json:"domain"`
	ClientIP  string `json:"client_ip"`
	CreatedAt int64  `json:"created_at"`
	Addresses string `json:"addresses"`
}

type HTTPError struct {
	Message string `json:"message"`
}

type ValidateIPRequest struct {
	IP string `json:"ip"`
}

type ValidateIPResponse struct {
	Status bool `json:"status"`
}

func initRedis(host string, port int, password string, db int) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

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

// @BasePath /v1

// PingExample godoc
// @Summary self implemenetation ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func Helloworld(g *gin.Context) {
	g.JSON(http.StatusOK, "helloworld")
}

type RootResponse struct {
	Version    string `json:"version"`
	Date       int64  `json:"date"`
	Kubernetes bool   `json:"kubernetes"`
}

// Check if running in k8s
// @Summary Get root information
// @Description Get application version, current date (UNIX epoch), and Kubernetes status.
// @ID get-root
// @Produce json
// @Success 200 {object} RootResponse
// @Router / [get]
func isRunningInKubernetes() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return err == nil
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

func logQuery(q Query) {
	queryJSON, err := json.Marshal(q)
	if err != nil {
		fmt.Println("Error marshaling query:", err)
		return
	}

	err = redisClient.LPush(ctx, RedisKey, queryJSON).Err()

	if err != nil {
		log.Printf("Failed to log query: %v", err)
		return
	}
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

func init() {

	maxQueryHistory = defaultMaxQueryHistory

	host := os.Getenv("REDIS_HOST")
	port := 6379
	portStr := os.Getenv("REDIS_PORT")
	password := ""
	db := 0
	if host == "" {
		host = "localhost"
	}
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Errorf("invalid REDIS_PORT value: %v", err))
		}
	}

	if password == "" {
		password = os.Getenv("REDIS_PASSWORD")
	}

	if v := os.Getenv("MAX_QUERY_HISTORY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxQueryHistory = n
		}
	}
	dbStr := os.Getenv("REDIS_DB")
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			panic(fmt.Errorf("invalid REDIS_DB value: %v", err))
		}
	}
	initRedis(host, port, password, db)
}

func main() {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = "/"

	v1 := r.Group("/v1")
	{
		eg := v1.Group("/example")
		{
			eg.GET("/helloworld", Helloworld)
		}
	}

	history := v1.Group("/history")
	{
		history.GET("", historyHandler)
	}

	tools := v1.Group("/tools")
	{
		tools.GET("/lookup", lookupHandler)
		tools.POST("/validate", validateHandler)
	}

	r.GET("/", func(c *gin.Context) {
		resp := RootResponse{
			Version:    "0.1.0",
			Date:       time.Now().Unix(),
			Kubernetes: isRunningInKubernetes(),
		}
		c.JSON(http.StatusOK, resp)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	r.GET("/health", healthHandler)

	// create http.Server object
	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	// start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v\n", err)
		}
	}()

	// wait for an interrupt signal (e.g. SIGINT or SIGTERM) to gracefully shutdown the server
	quit := make(chan os.Signal, 1) // add buffer to the channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// shut down the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}
	log.Println("Server shutdown successful")
}
