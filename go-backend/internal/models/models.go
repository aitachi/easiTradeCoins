package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderSide represents buy or sell
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeLimit       OrderType = "limit"        // 限价单
	OrderTypeMarket      OrderType = "market"       // 市价单
	OrderTypeStopLoss    OrderType = "stop_loss"    // 止损单
	OrderTypeTakeProfit  OrderType = "take_profit"  // 止盈单
	OrderTypeStopLimit   OrderType = "stop_limit"   // 止损限价单
	OrderTypeTrailingStop OrderType = "trailing_stop" // 跟踪止损
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPartial   OrderStatus = "partial"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// TimeInForce represents order time in force
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or Kill
)

// Order represents a trading order
type Order struct {
	ID            string          `json:"id" gorm:"primaryKey"`
	UserID        uint            `json:"user_id" gorm:"index"`
	Symbol        string          `json:"symbol" gorm:"index"`
	Side          OrderSide       `json:"side"`
	Type          OrderType       `json:"type"`
	Price         decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"`
	Quantity      decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	FilledQty     decimal.Decimal `json:"filled_qty" gorm:"type:decimal(36,18)"`
	FilledAmount  decimal.Decimal `json:"filled_amount" gorm:"type:decimal(36,18)"`
	AvgPrice      decimal.Decimal `json:"avg_price" gorm:"type:decimal(36,18)"`
	Fee           decimal.Decimal `json:"fee" gorm:"type:decimal(36,18)"`
	FeeCurrency   string          `json:"fee_currency"`
	Status        OrderStatus     `json:"status" gorm:"index"`
	TimeInForce   TimeInForce     `json:"time_in_force"`

	// Stop-loss and Take-profit fields
	StopPrice     *decimal.Decimal `json:"stop_price,omitempty" gorm:"type:decimal(36,18)"` // 触发价格
	TakeProfitPrice *decimal.Decimal `json:"take_profit_price,omitempty" gorm:"type:decimal(36,18)"` // 止盈价
	TrailingDelta *decimal.Decimal `json:"trailing_delta,omitempty" gorm:"type:decimal(36,18)"` // 跟踪止损价差
	TriggerCondition string        `json:"trigger_condition,omitempty"` // 触发条件: >=, <=
	IsTriggered   bool             `json:"is_triggered" gorm:"default:false"` // 是否已触发
	TriggerTime   *time.Time       `json:"trigger_time,omitempty"` // 触发时间

	CreateTime    time.Time       `json:"create_time" gorm:"index"`
	UpdateTime    time.Time       `json:"update_time"`
}

// Trade represents a trade execution
type Trade struct {
	ID           string          `json:"id" gorm:"primaryKey"`
	Symbol       string          `json:"symbol" gorm:"index"`
	BuyOrderID   string          `json:"buy_order_id" gorm:"index"`
	SellOrderID  string          `json:"sell_order_id" gorm:"index"`
	BuyerID      uint            `json:"buyer_id" gorm:"index"`
	SellerID     uint            `json:"seller_id" gorm:"index"`
	Price        decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"`
	Quantity     decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	Amount       decimal.Decimal `json:"amount" gorm:"type:decimal(36,18)"`
	BuyerFee     decimal.Decimal `json:"buyer_fee" gorm:"type:decimal(36,18)"`
	SellerFee    decimal.Decimal `json:"seller_fee" gorm:"type:decimal(36,18)"`
	TradeTime    time.Time       `json:"trade_time" gorm:"index"`
}

// User represents a platform user
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"unique;not null"`
	Phone        string    `json:"phone" gorm:"unique"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Salt         string    `json:"-" gorm:"not null"`
	KYCLevel     int       `json:"kyc_level" gorm:"default:0"`
	Status       int       `json:"status" gorm:"default:1"`
	RegisterIP   string    `json:"register_ip"`
	RegisterTime time.Time `json:"register_time"`
	LastLoginIP  string    `json:"last_login_ip"`
	LastLoginTime time.Time `json:"last_login_time"`
}

// UserAsset represents user asset balance
type UserAsset struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	UserID     uint            `json:"user_id" gorm:"index"`
	Currency   string          `json:"currency" gorm:"not null"`
	Chain      string          `json:"chain"`
	Available  decimal.Decimal `json:"available" gorm:"type:decimal(36,18);default:0"`
	Frozen     decimal.Decimal `json:"frozen" gorm:"type:decimal(36,18);default:0"`
	UpdateTime time.Time       `json:"update_time"`
}

// Deposit represents a deposit record
type Deposit struct {
	ID                     uint            `json:"id" gorm:"primaryKey"`
	UserID                 uint            `json:"user_id" gorm:"index"`
	Currency               string          `json:"currency" gorm:"not null"`
	Chain                  string          `json:"chain"`
	Amount                 decimal.Decimal `json:"amount" gorm:"type:decimal(36,18)"`
	Address                string          `json:"address"`
	TxID                   string          `json:"txid" gorm:"index"`
	Confirmations          int             `json:"confirmations" gorm:"default:0"`
	RequiredConfirmations  int             `json:"required_confirmations"`
	Status                 int             `json:"status" gorm:"index"`
	CreateTime             time.Time       `json:"create_time" gorm:"index"`
	ConfirmTime            *time.Time      `json:"confirm_time"`
}

// Withdrawal represents a withdrawal record
type Withdrawal struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	UserID        uint            `json:"user_id" gorm:"index"`
	Currency      string          `json:"currency" gorm:"not null"`
	Chain         string          `json:"chain"`
	Amount        decimal.Decimal `json:"amount" gorm:"type:decimal(36,18)"`
	Fee           decimal.Decimal `json:"fee" gorm:"type:decimal(36,18)"`
	Address       string          `json:"address" gorm:"not null"`
	TxID          string          `json:"txid"`
	Status        int             `json:"status" gorm:"index"`
	AuditUserID   *uint           `json:"audit_user_id"`
	AuditTime     *time.Time      `json:"audit_time"`
	CompleteTime  *time.Time      `json:"complete_time"`
	CreateTime    time.Time       `json:"create_time" gorm:"index"`
	Remark        string          `json:"remark"`
}

// TradingPair represents a trading pair configuration
type TradingPair struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	Symbol            string          `json:"symbol" gorm:"unique;not null"`
	BaseCurrency      string          `json:"base_currency" gorm:"not null"`
	QuoteCurrency     string          `json:"quote_currency" gorm:"not null"`
	PricePrecision    int             `json:"price_precision" gorm:"default:8"`
	QuantityPrecision int             `json:"quantity_precision" gorm:"default:8"`
	MinQuantity       decimal.Decimal `json:"min_quantity" gorm:"type:decimal(36,18)"`
	MaxQuantity       decimal.Decimal `json:"max_quantity" gorm:"type:decimal(36,18)"`
	MinAmount         decimal.Decimal `json:"min_amount" gorm:"type:decimal(36,18)"`
	TakerFeeRate      decimal.Decimal `json:"taker_fee_rate" gorm:"type:decimal(10,8);default:0.001"`
	MakerFeeRate      decimal.Decimal `json:"maker_fee_rate" gorm:"type:decimal(10,8);default:0.001"`
	IsActive          bool            `json:"is_active" gorm:"default:true"`
	CreateTime        time.Time       `json:"create_time"`
}

func (Order) TableName() string {
	return "orders"
}

func (Trade) TableName() string {
	return "trades"
}

func (User) TableName() string {
	return "users"
}

func (UserAsset) TableName() string {
	return "user_assets"
}

func (Deposit) TableName() string {
	return "deposits"
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}

func (TradingPair) TableName() string {
	return "trading_pairs"
}

// RiskEvent represents a risk event for logging and monitoring
type RiskEvent struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	EventType   string    `json:"event_type" gorm:"index"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Details     string    `json:"details" gorm:"type:text"`
	Action      string    `json:"action"`
	CreateTime  time.Time `json:"create_time" gorm:"index"`
}

func (RiskEvent) TableName() string {
	return "risk_events"
}

// Violation represents a user violation record
type Violation struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"index"`
	Type        string     `json:"type" gorm:"index"`
	Status      string     `json:"status"`
	Severity    int        `json:"severity"`
	Description string     `json:"description"`
	CreateTime  time.Time  `json:"create_time" gorm:"index"`
	ResolveTime *time.Time `json:"resolve_time"`
}

func (Violation) TableName() string {
	return "violations"
}

// WithdrawalWhitelist represents whitelisted withdrawal addresses
type WithdrawalWhitelist struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Currency   string    `json:"currency"`
	Address    string    `json:"address" gorm:"index"`
	Label      string    `json:"label"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreateTime time.Time `json:"create_time"`
}

func (WithdrawalWhitelist) TableName() string {
	return "withdrawal_whitelists"
}

// BeforeCreate hook for Order
func (o *Order) BeforeCreate() error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.CreateTime.IsZero() {
		o.CreateTime = time.Now()
	}
	o.UpdateTime = time.Now()
	return nil
}

// BeforeCreate hook for Trade
func (t *Trade) BeforeCreate() error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.TradeTime.IsZero() {
		t.TradeTime = time.Now()
	}
	return nil
}
