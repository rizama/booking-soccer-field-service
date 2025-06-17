package cmd

import (
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	"field-service/controllers"
	"field-service/domain/models"
	"field-service/middlewares"
	"field-service/repositories"
	"field-service/routes"
	"field-service/services"
	"fmt"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()
		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic(err)
		}
		time.Local = loc

		// Migration & Seeding
		err = db.AutoMigrate(&models.Role{}, &models.User{})
		if err != nil {
			panic(err)
		}

		// Build Dependencies
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository)
		controller := controllers.NewControllerRegistry(service)

		// Set http Router
		router := gin.Default()
		router.Use(middlewares.HandlePanic())
		router.NoRoute(func(ctx *gin.Context) {
			ctx.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
			})
		})

		// Default Route
		router.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to User Service",
			})
		})

		// CORS
		router.Use(func(ctx *gin.Context) {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-api-key, x-request-at")
			ctx.Next()
		})

		// Rate Limiter
		limimter := tollbooth.NewLimiter(
			float64(config.Config.RateLimiterRequest),
			&limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Duration(config.Config.RateLimiterTimeSecond) * time.Second,
			},
		)
		router.Use(middlewares.RateLimit(limimter))

		// Setup Router
		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(controller, group)
		route.Serve()

		// Start Server
		port := fmt.Sprintf(":%d", config.Config.Port)
		router.Run(port)
	},
}

func Run() {
	command.Execute()
}
