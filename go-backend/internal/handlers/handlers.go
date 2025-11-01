package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/easitradecoins/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
)

// OrderHandler handles order-related requests
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrderRequest represents a create order request
type CreateOrderRequest struct {
	Symbol      string  `json:"symbol" binding:"required"`
	Side        string  `json:"side" binding:"required,oneof=buy sell"`
	Type        string  `json:"type" binding:"required,oneof=limit market"`
	Price       string  `json:"price"`
	Quantity    string  `json:"quantity" binding:"required"`
	TimeInForce string  `json:"timeInForce" binding:"omitempty,oneof=GTC IOC FOK"`
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT
	userID := getUserIDFromContext(c)

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity"})
		return
	}

	// Parse price for limit orders
	var price decimal.Decimal
	if req.Type == "limit" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
			return
		}
	}

	// Set default time in force
	timeInForce := req.TimeInForce
	if timeInForce == "" {
		timeInForce = "GTC"
	}

	// Create order
	order, trades, err := h.orderService.CreateOrder(&models.Order{
		UserID:      userID,
		Symbol:      req.Symbol,
		Side:        models.OrderSide(req.Side),
		Type:        models.OrderType(req.Type),
		Price:       price,
		Quantity:    quantity,
		TimeInForce: models.TimeInForce(timeInForce),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":  order,
		"trades": trades,
	})
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	userID := getUserIDFromContext(c)

	if err := h.orderService.CancelOrder(orderID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// GetOrder gets order details
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("orderId")
	userID := getUserIDFromContext(c)

	order, err := h.orderService.GetOrder(orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOpenOrders gets all open orders for a user
func (h *OrderHandler) GetOpenOrders(c *gin.Context) {
	userID := getUserIDFromContext(c)
	symbol := c.Query("symbol")

	orders, err := h.orderService.GetOpenOrders(userID, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrderHistory gets order history
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	userID := getUserIDFromContext(c)
	symbol := c.Query("symbol")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderService.GetOrderHistory(userID, symbol, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UserHandler handles user-related requests
type UserHandler struct {
	userService  *services.UserService
	assetService *services.AssetService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, assetService *services.AssetService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		assetService: assetService,
	}
}

// RegisterRequest represents a register request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register registers a new user
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	user, err := h.userService.Register(req.Email, req.Phone, req.Password, ip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	user, err := h.userService.Login(req.Email, req.Password, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetBalance gets user balance
func (h *UserHandler) GetBalance(c *gin.Context) {
	userID := getUserIDFromContext(c)

	assets, err := h.assetService.GetAllUserAssets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assets)
}

// MarketHandler handles market data requests
type MarketHandler struct {
	orderService *services.OrderService
}

// NewMarketHandler creates a new market handler
func NewMarketHandler(orderService *services.OrderService) *MarketHandler {
	return &MarketHandler{
		orderService: orderService,
	}
}

// GetDepth gets order book depth
func (h *MarketHandler) GetDepth(c *gin.Context) {
	symbol := c.Param("symbol")
	depth, _ := strconv.Atoi(c.DefaultQuery("depth", "20"))

	bids, asks := h.orderService.GetOrderBookDepth(symbol, depth)

	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"bids":   bids,
		"asks":   asks,
	})
}

// GetTrades gets recent trades
func (h *MarketHandler) GetTrades(c *gin.Context) {
	symbol := c.Param("symbol")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	trades, err := h.orderService.GetRecentTrades(symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// Helper functions

// generateJWT generates a JWT token
func generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key")) // TODO: Move to config
}

// getUserIDFromContext gets user ID from context
func getUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}

	return userID.(uint)
}
