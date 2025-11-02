//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package main

import (
	"log"
	"os"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/handlers"
	"github.com/easitradecoins/backend/internal/matching"
	"github.com/easitradecoins/backend/internal/middleware"
	"github.com/easitradecoins/backend/internal/services"
	"github.com/easitradecoins/backend/internal/websocket"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	// Load configuration
	loadConfig()

	// Initialize database
	dbConfig := &database.Config{
		PostgresDSN: viper.GetString("DATABASE_URL"),
		RedisURL:    viper.GetString("REDIS_URL"),
	}

	if err := database.InitDatabase(dbConfig); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize services
	matchingEngine := matching.NewMatchingEngine()
	assetService := services.NewAssetService()
	userService := services.NewUserService()
	orderService := services.NewOrderService(matchingEngine, assetService)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Start trade processor
	go processTrades(matchingEngine, hub)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, assetService)
	orderHandler := handlers.NewOrderHandler(orderService)
	marketHandler := handlers.NewMarketHandler(orderService)

	// Setup router
	router := setupRouter(userHandler, orderHandler, marketHandler, hub)

	// Start server
	port := viper.GetString("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	// Set defaults
	viper.SetDefault("API_PORT", "8080")
	viper.SetDefault("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/easitradecoins")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
}

func setupRouter(
	userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler,
	marketHandler *handlers.MarketHandler,
	hub *websocket.Hub,
) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Public endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Market data (public)
		market := v1.Group("/market")
		{
			market.GET("/depth/:symbol", marketHandler.GetDepth)
			market.GET("/trades/:symbol", marketHandler.GetTrades)
		}

		// Protected endpoints
		authMiddleware := middleware.AuthMiddleware(viper.GetString("JWT_SECRET"))

		// Order endpoints
		order := v1.Group("/order").Use(authMiddleware)
		{
			order.POST("/create", orderHandler.CreateOrder)
			order.DELETE("/:orderId", orderHandler.CancelOrder)
			order.GET("/:orderId", orderHandler.GetOrder)
			order.GET("/open", orderHandler.GetOpenOrders)
			order.GET("/history", orderHandler.GetOrderHistory)
		}

		// Account endpoints
		account := v1.Group("/account").Use(authMiddleware)
		{
			account.GET("/balance", userHandler.GetBalance)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, hub)
	})

	return router
}

import (
	"net/http"
	"github.com/google/uuid"
)

func handleWebSocket(c *gin.Context, hub *websocket.Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &websocket.Client{
		ID:            uuid.New().String(),
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Hub:           hub,
		Subscriptions: make(map[string]bool),
	}

	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func processTrades(engine *matching.MatchingEngine, hub *websocket.Hub) {
	tradeChan := engine.GetTradeChan()

	for trade := range tradeChan {
		// Broadcast trade via WebSocket
		hub.BroadcastTrade(trade)

		// You can add more processing here, like:
		// - Sending notifications
		// - Updating statistics
		// - Logging to external systems
	}
}
