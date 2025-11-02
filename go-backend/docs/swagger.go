//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title EasiTradeCoins API
// @version 2.0
// @description 专业加密货币交易平台 API 文档
// @description
// @description ## 功能特性
// @description - 现货交易 (限价单、市价单)
// @description - 止损止盈订单 (Stop-Loss, Take-Profit)
// @description - 跟踪止损 (Trailing Stop)
// @description - 条件单 (Conditional Orders)
// @description - 资产管理 (充值、提现、划转)
// @description - 质押挖矿 (Staking)
// @description - 空投活动 (Airdrops)
// @description - Token工厂 (Token Factory)
// @description
// @description ## 认证方式
// @description 使用JWT Bearer Token进行认证
// @description 在请求头中添加: `Authorization: Bearer <your_token>`
//
// @contact.name API Support
// @contact.email support@easitradecoins.com
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme.
//
// @schemes http https
//
// @tag.name Auth
// @tag.description 用户认证相关接口
//
// @tag.name Orders
// @tag.description 订单管理接口 - 支持限价单、市价单、止损止盈、跟踪止损等
//
// @tag.name Trades
// @tag.description 交易记录查询接口
//
// @tag.name Assets
// @tag.description 资产管理接口 - 查询余额、充值、提现
//
// @tag.name Staking
// @tag.description 质押挖矿接口 - 质押、赎回、查询收益
//
// @tag.name Airdrop
// @tag.description 空投活动接口 - 领取空投、查询活动
//
// @tag.name TokenFactory
// @tag.description 代币工厂接口 - 创建ERC20代币
//
// @tag.name Market
// @tag.description 市场数据接口 - K线、深度、行情
//
// @tag.name User
// @tag.description 用户管理接口 - 个人信息、KYC认证

func SetupSwagger(router *gin.Engine) {
	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Common response models

// SuccessResponse 成功响应
// @Description 标准成功响应格式
type SuccessResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse 错误响应
// @Description 标准错误响应格式
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"error message"`
}

// PaginationResponse 分页响应
// @Description 分页数据响应格式
type PaginationResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total" example:"100"`
	Page    int         `json:"page" example:"1"`
	Limit   int         `json:"limit" example:"20"`
}

// Order request/response models

// CreateOrderRequest 创建订单请求
// @Description 创建订单的请求参数
type CreateOrderRequest struct {
	Symbol    string  `json:"symbol" binding:"required" example:"BTC_USDT"`
	Side      string  `json:"side" binding:"required,oneof=buy sell" example:"buy"`
	Type      string  `json:"type" binding:"required,oneof=limit market stop_loss take_profit stop_limit trailing_stop" example:"limit"`
	Price     float64 `json:"price" example:"50000.0"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0" example:"0.1"`

	// Stop-loss and Take-profit fields
	StopPrice       *float64 `json:"stop_price,omitempty" example:"49000.0"`
	TakeProfitPrice *float64 `json:"take_profit_price,omitempty" example:"51000.0"`
	TrailingDelta   *float64 `json:"trailing_delta,omitempty" example:"500.0"`

	TimeInForce string  `json:"time_in_force" example:"GTC"`
}

// OrderResponse 订单响应
// @Description 订单信息
type OrderResponse struct {
	ID              string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID          uint    `json:"user_id" example:"1"`
	Symbol          string  `json:"symbol" example:"BTC_USDT"`
	Side            string  `json:"side" example:"buy"`
	Type            string  `json:"type" example:"limit"`
	Price           float64 `json:"price" example:"50000.0"`
	Quantity        float64 `json:"quantity" example:"0.1"`
	FilledQty       float64 `json:"filled_qty" example:"0.05"`
	FilledAmount    float64 `json:"filled_amount" example:"2500.0"`
	AvgPrice        float64 `json:"avg_price" example:"50000.0"`
	Status          string  `json:"status" example:"partial"`
	StopPrice       *float64 `json:"stop_price,omitempty" example:"49000.0"`
	TakeProfitPrice *float64 `json:"take_profit_price,omitempty" example:"51000.0"`
	TrailingDelta   *float64 `json:"trailing_delta,omitempty" example:"500.0"`
	IsTriggered     bool    `json:"is_triggered" example:"false"`
	CreateTime      string  `json:"create_time" example:"2024-01-01T00:00:00Z"`
	UpdateTime      string  `json:"update_time" example:"2024-01-01T00:00:00Z"`
}

// AssetResponse 资产响应
// @Description 用户资产信息
type AssetResponse struct {
	Currency   string  `json:"currency" example:"BTC"`
	Chain      string  `json:"chain" example:"ERC20"`
	Available  float64 `json:"available" example:"1.5"`
	Frozen     float64 `json:"frozen" example:"0.1"`
	Total      float64 `json:"total" example:"1.6"`
	USDValue   float64 `json:"usd_value" example:"80000.0"`
	UpdateTime string  `json:"update_time" example:"2024-01-01T00:00:00Z"`
}

// DepthResponse 深度响应
// @Description 订单簿深度数据
type DepthResponse struct {
	Symbol    string          `json:"symbol" example:"BTC_USDT"`
	Timestamp int64           `json:"timestamp" example:"1704067200"`
	Bids      [][]float64     `json:"bids" example:"[[50000.0,1.5],[49999.0,2.0]]"`
	Asks      [][]float64     `json:"asks" example:"[[50001.0,1.5],[50002.0,2.0]]"`
}
