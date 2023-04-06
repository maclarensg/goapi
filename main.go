package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	docs "github.com/maclarensg/goapi/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var maxQueryHistory int
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

type HTTPError struct {
	Message string `json:"message"`
}

type ValidateIPRequest struct {
	IP string `json:"ip"`
}

type ValidateIPResponse struct {
	Status bool `json:"status"`
}

type RootResponse struct {
	Version    string `json:"version"`
	Date       int64  `json:"date"`
	Kubernetes bool   `json:"kubernetes"`
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

	// Routes
	docs.SwaggerInfo.BasePath = "/"

	v1 := r.Group("/v1")

	history := v1.Group("/history")
	{
		history.GET("", historyHandler)
	}

	tools := v1.Group("/tools")
	{
		tools.GET("/lookup", lookupHandler)
		tools.POST("/validate", validateHandler)
	}

	r.GET("/", rootHandler)

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
	bgctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(bgctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}
	log.Println("Server shutdown successful")
}
