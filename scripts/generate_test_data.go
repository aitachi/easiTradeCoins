package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

// 生成测试用户和行为数据的脚本
func main() {
	// 初始化数据库连接
	config := &database.Config{
		PostgresDSN: "postgresql://postgres:postgres@localhost:5432/easitradecoins",
	}

	if err := database.InitDatabase(config); err != nil {
		panic(err)
	}
	defer database.Close()

	fmt.Println("=== 开始生成测试数据 ===")

	// 生成测试用户
	users := generateTestUsers(20)
	fmt.Printf("✅ 生成 %d 个测试用户\n", len(users))

	// 生成资产数据
	generateUserAssets(users)
	fmt.Println("✅ 生成用户资产数据")

	// 生成订单数据
	generateOrders(users, 200)
	fmt.Println("✅ 生成 200 个测试订单")

	// 生成成交数据
	generateTrades(users, 150)
	fmt.Println("✅ 生成 150 个测试成交")

	// 生成充值数据
	generateDeposits(users, 50)
	fmt.Println("✅ 生成 50 个充值记录")

	// 生成提现数据
	generateWithdrawals(users, 30)
	fmt.Println("✅ 生成 30 个提现记录")

	// 生成风险事件数据
	generateRiskEvents(users, 100)
	fmt.Println("✅ 生成 100 个风险事件")

	// 生成违规记录
	generateViolations(users, 15)
	fmt.Println("✅ 生成 15 个违规记录")

	// 生成提现白名单
	generateWithdrawalWhitelists(users, 40)
	fmt.Println("✅ 生成 40 个白名单地址")

	fmt.Println("\n=== 测试数据生成完成 ===")
	printDataStatistics()
}

// generateTestUsers 生成测试用户
func generateTestUsers(count int) []models.User {
	users := make([]models.User, 0, count)
	ips := []string{
		"192.168.1.100", "192.168.1.101", "192.168.1.102",
		"10.0.0.50", "10.0.0.51", "172.16.0.100",
		"203.0.113.1", "203.0.113.2", // 同一个IP的关联账户
	}

	for i := 0; i < count; i++ {
		salt := fmt.Sprintf("salt%d", i)
		password := fmt.Sprintf("password%d", i)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)

		user := models.User{
			Email:         fmt.Sprintf("test_user_%d@example.com", i+1),
			Phone:         fmt.Sprintf("+8613800138%03d", i),
			PasswordHash:  string(hashedPassword),
			Salt:          salt,
			KYCLevel:      []int{0, 1, 2}[rand.Intn(3)], // 随机KYC等级
			Status:        []int{1, 1, 1, 2}[rand.Intn(4)], // 大部分正常,少数冻结
			RegisterIP:    ips[rand.Intn(len(ips))],
			RegisterTime:  time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
			LastLoginIP:   ips[rand.Intn(len(ips))],
			LastLoginTime: time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
		}

		if err := database.DB.Create(&user).Error; err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
			continue
		}

		users = append(users, user)
	}

	return users
}

// generateUserAssets 生成用户资产
func generateUserAssets(users []models.User) {
	currencies := []string{"BTC", "ETH", "USDT", "BNB"}

	for _, user := range users {
		for _, currency := range currencies {
			available := decimal.NewFromFloat(rand.Float64() * 10000)
			frozen := decimal.NewFromFloat(rand.Float64() * 100)

			asset := models.UserAsset{
				UserID:     user.ID,
				Currency:   currency,
				Chain:      "ERC20",
				Available:  available,
				Frozen:     frozen,
				UpdateTime: time.Now(),
			}

			database.DB.Create(&asset)
		}
	}
}

// generateOrders 生成测试订单
func generateOrders(users []models.User, count int) {
	symbols := []string{"BTC_USDT", "ETH_USDT", "BNB_USDT"}
	sides := []models.OrderSide{models.OrderSideBuy, models.OrderSideSell}
	types := []models.OrderType{models.OrderTypeLimit, models.OrderTypeMarket}
	statuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPartial,
		models.OrderStatusFilled,
		models.OrderStatusCancelled,
	}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		symbol := symbols[rand.Intn(len(symbols))]

		var price decimal.Decimal
		if symbol == "BTC_USDT" {
			price = decimal.NewFromFloat(40000 + rand.Float64()*10000)
		} else if symbol == "ETH_USDT" {
			price = decimal.NewFromFloat(2000 + rand.Float64()*1000)
		} else {
			price = decimal.NewFromFloat(200 + rand.Float64()*100)
		}

		quantity := decimal.NewFromFloat(rand.Float64() * 2)
		filledQty := quantity.Mul(decimal.NewFromFloat(rand.Float64()))

		order := models.Order{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			Symbol:       symbol,
			Side:         sides[rand.Intn(len(sides))],
			Type:         types[rand.Intn(len(types))],
			Price:        price,
			Quantity:     quantity,
			FilledQty:    filledQty,
			FilledAmount: filledQty.Mul(price),
			AvgPrice:     price,
			Fee:          decimal.Zero,
			Status:       statuses[rand.Intn(len(statuses))],
			TimeInForce:  models.TimeInForceGTC,
			CreateTime:   time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
			UpdateTime:   time.Now(),
		}

		database.DB.Create(&order)
	}
}

// generateTrades 生成成交数据
func generateTrades(users []models.User, count int) {
	symbols := []string{"BTC_USDT", "ETH_USDT", "BNB_USDT"}

	for i := 0; i < count; i++ {
		buyer := users[rand.Intn(len(users))]
		seller := users[rand.Intn(len(users))]

		// 避免直接自成交
		for buyer.ID == seller.ID {
			seller = users[rand.Intn(len(users))]
		}

		symbol := symbols[rand.Intn(len(symbols))]
		var price decimal.Decimal
		if symbol == "BTC_USDT" {
			price = decimal.NewFromFloat(40000 + rand.Float64()*10000)
		} else if symbol == "ETH_USDT" {
			price = decimal.NewFromFloat(2000 + rand.Float64()*1000)
		} else {
			price = decimal.NewFromFloat(200 + rand.Float64()*100)
		}

		quantity := decimal.NewFromFloat(rand.Float64() * 2)
		amount := quantity.Mul(price)
		feeRate := decimal.NewFromFloat(0.001)

		trade := models.Trade{
			ID:          uuid.New().String(),
			Symbol:      symbol,
			BuyOrderID:  uuid.New().String(),
			SellOrderID: uuid.New().String(),
			BuyerID:     buyer.ID,
			SellerID:    seller.ID,
			Price:       price,
			Quantity:    quantity,
			Amount:      amount,
			BuyerFee:    amount.Mul(feeRate),
			SellerFee:   amount.Mul(feeRate),
			TradeTime:   time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		database.DB.Create(&trade)
	}
}

// generateDeposits 生成充值数据
func generateDeposits(users []models.User, count int) {
	currencies := []string{"BTC", "ETH", "USDT"}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		currency := currencies[rand.Intn(len(currencies))]

		var amount decimal.Decimal
		if currency == "BTC" {
			amount = decimal.NewFromFloat(rand.Float64() * 2)
		} else if currency == "ETH" {
			amount = decimal.NewFromFloat(rand.Float64() * 20)
		} else {
			amount = decimal.NewFromFloat(rand.Float64() * 10000)
		}

		deposit := models.Deposit{
			UserID:                user.ID,
			Currency:              currency,
			Chain:                 "ERC20",
			Amount:                amount,
			Address:               fmt.Sprintf("0x%s", uuid.New().String()[:40]),
			TxID:                  fmt.Sprintf("0x%s", uuid.New().String()),
			Confirmations:         12,
			RequiredConfirmations: 12,
			Status:                1, // 已到账
			CreateTime:            time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		now := time.Now()
		deposit.ConfirmTime = &now

		database.DB.Create(&deposit)
	}
}

// generateWithdrawals 生成提现数据
func generateWithdrawals(users []models.User, count int) {
	currencies := []string{"BTC", "ETH", "USDT"}
	statuses := []int{0, 1, 2, 3, 4} // 待审核, 审核通过, 处理中, 已完成, 拒绝

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		currency := currencies[rand.Intn(len(currencies))]

		var amount decimal.Decimal
		if currency == "BTC" {
			amount = decimal.NewFromFloat(rand.Float64() * 1)
		} else if currency == "ETH" {
			amount = decimal.NewFromFloat(rand.Float64() * 10)
		} else {
			amount = decimal.NewFromFloat(rand.Float64() * 5000)
		}

		withdrawal := models.Withdrawal{
			UserID:     user.ID,
			Currency:   currency,
			Chain:      "ERC20",
			Amount:     amount,
			Fee:        amount.Mul(decimal.NewFromFloat(0.001)),
			Address:    fmt.Sprintf("0x%s", uuid.New().String()[:40]),
			Status:     statuses[rand.Intn(len(statuses))],
			CreateTime: time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		if withdrawal.Status >= 3 {
			txid := fmt.Sprintf("0x%s", uuid.New().String())
			withdrawal.TxID = txid
			now := time.Now()
			withdrawal.CompleteTime = &now
		}

		database.DB.Create(&withdrawal)
	}
}

// generateRiskEvents 生成风险事件
func generateRiskEvents(users []models.User, count int) {
	eventTypes := []string{
		"order_validation", "withdrawal_validation", "rate_limit_exceeded",
		"price_deviation", "abnormal_pattern", "first_time_address",
		"quick_in_out", "high_risk_address", "daily_limit_exceeded",
	}
	severities := []string{"low", "medium", "high", "critical"}
	actions := []string{"allowed", "blocked", "flagged"}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]

		event := models.RiskEvent{
			UserID:      user.ID,
			EventType:   eventTypes[rand.Intn(len(eventTypes))],
			Severity:    severities[rand.Intn(len(severities))],
			Description: "测试风险事件",
			Details:     fmt.Sprintf("测试详情 #%d", i+1),
			Action:      actions[rand.Intn(len(actions))],
			CreateTime:  time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		database.DB.Create(&event)
	}
}

// generateViolations 生成违规记录
func generateViolations(users []models.User, count int) {
	violationTypes := []string{
		"self_trading", "wash_trading", "suspicious_withdrawal",
		"rapid_trading", "abnormal_price", "account_sharing",
	}
	statuses := []string{"active", "resolved"}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]

		violation := models.Violation{
			UserID:      user.ID,
			Type:        violationTypes[rand.Intn(len(violationTypes))],
			Status:      statuses[rand.Intn(len(statuses))],
			Severity:    rand.Intn(10) + 1,
			Description: fmt.Sprintf("测试违规记录 #%d", i+1),
			CreateTime:  time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		if violation.Status == "resolved" {
			resolveTime := time.Now().Add(-time.Duration(rand.Intn(360)) * time.Hour)
			violation.ResolveTime = &resolveTime
		}

		database.DB.Create(&violation)
	}
}

// generateWithdrawalWhitelists 生成提现白名单
func generateWithdrawalWhitelists(users []models.User, count int) {
	currencies := []string{"BTC", "ETH", "USDT"}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]

		whitelist := models.WithdrawalWhitelist{
			UserID:     user.ID,
			Currency:   currencies[rand.Intn(len(currencies))],
			Address:    fmt.Sprintf("0x%s", uuid.New().String()[:40]),
			Label:      fmt.Sprintf("我的钱包 #%d", i+1),
			IsActive:   rand.Float32() > 0.1, // 90%激活
			CreateTime: time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		database.DB.Create(&whitelist)
	}
}

// printDataStatistics 打印数据统计
func printDataStatistics() {
	var stats struct {
		Users               int64
		Orders              int64
		Trades              int64
		Deposits            int64
		Withdrawals         int64
		RiskEvents          int64
		Violations          int64
		WithdrawalWhitelist int64
	}

	database.DB.Model(&models.User{}).Count(&stats.Users)
	database.DB.Model(&models.Order{}).Count(&stats.Orders)
	database.DB.Model(&models.Trade{}).Count(&stats.Trades)
	database.DB.Model(&models.Deposit{}).Count(&stats.Deposits)
	database.DB.Model(&models.Withdrawal{}).Count(&stats.Withdrawals)
	database.DB.Model(&models.RiskEvent{}).Count(&stats.RiskEvents)
	database.DB.Model(&models.Violation{}).Count(&stats.Violations)
	database.DB.Model(&models.WithdrawalWhitelist{}).Count(&stats.WithdrawalWhitelist)

	fmt.Println("\n=== 数据统计 ===")
	fmt.Printf("用户: %d\n", stats.Users)
	fmt.Printf("订单: %d\n", stats.Orders)
	fmt.Printf("成交: %d\n", stats.Trades)
	fmt.Printf("充值: %d\n", stats.Deposits)
	fmt.Printf("提现: %d\n", stats.Withdrawals)
	fmt.Printf("风险事件: %d\n", stats.RiskEvents)
	fmt.Printf("违规记录: %d\n", stats.Violations)
	fmt.Printf("白名单地址: %d\n", stats.WithdrawalWhitelist)
}

// models.go中需要添加的新模型(已在前面定义):
// type RiskEvent struct {...}
// type Violation struct {...}
// type WithdrawalWhitelist struct {...}
