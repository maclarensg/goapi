package main

import (
	"net/http"
	"os"
	"time"

	docs "github.com/maclarensg/goapi/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
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

func isRunningInKubernetes() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return err == nil
}

func main() {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		eg := v1.Group("/example")
		{
			eg.GET("/helloworld", Helloworld)
		}
	}

	// @Summary Get root information
	// @Description Get application version, current date (UNIX epoch), and Kubernetes status.
	// @ID get-root
	// @Produce json
	// @Success 200 {object} RootResponse
	// @Router / [get]
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
	r.Run(":8080")
}
