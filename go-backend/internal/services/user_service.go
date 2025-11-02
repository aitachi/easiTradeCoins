//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package services

import (
	"context"
	"errors"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService handles user-related operations
type UserService struct{}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{}
}

// Register registers a new user
func (s *UserService) Register(email, phone, password, ip string) (*models.User, error) {
	// Check if email already exists
	var existingUser models.User
	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        email,
		Phone:        phone,
		PasswordHash: string(passwordHash),
		Salt:         "", // Salt is handled by bcrypt
		KYCLevel:     0,
		Status:       1,
		RegisterIP:   ip,
		RegisterTime: time.Now(),
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	// Initialize user assets for common currencies
	currencies := []string{"BTC", "ETH", "USDT"}
	for _, currency := range currencies {
		asset := &models.UserAsset{
			UserID:     user.ID,
			Currency:   currency,
			Chain:      "ERC20",
			Available:  decimal.Zero,
			Frozen:     decimal.Zero,
			UpdateTime: time.Now(),
		}
		database.DB.Create(asset)
	}

	return user, nil
}

// Login authenticates a user
func (s *UserService) Login(email, password, ip string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Update last login info
	user.LastLoginIP = ip
	user.LastLoginTime = time.Now()
	database.DB.Save(&user)

	return &user, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateKYCLevel updates user KYC level
func (s *UserService) UpdateKYCLevel(userID uint, level int) error {
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("kyc_level", level).Error
}

// AssetService handles asset-related operations
type AssetService struct{}

// NewAssetService creates a new asset service
func NewAssetService() *AssetService {
	return &AssetService{}
}

// GetUserAsset gets user asset
func (s *AssetService) GetUserAsset(userID uint, currency, chain string) (*models.UserAsset, error) {
	var asset models.UserAsset
	if err := database.DB.Where("user_id = ? AND currency = ? AND chain = ?", userID, currency, chain).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create asset if not exists
			asset = models.UserAsset{
				UserID:     userID,
				Currency:   currency,
				Chain:      chain,
				Available:  decimal.Zero,
				Frozen:     decimal.Zero,
				UpdateTime: time.Now(),
			}
			if err := database.DB.Create(&asset).Error; err != nil {
				return nil, err
			}
			return &asset, nil
		}
		return nil, err
	}
	return &asset, nil
}

// GetAllUserAssets gets all user assets
func (s *AssetService) GetAllUserAssets(userID uint) ([]models.UserAsset, error) {
	var assets []models.UserAsset
	if err := database.DB.Where("user_id = ?", userID).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

// FreezeAsset freezes asset for trading
func (s *AssetService) FreezeAsset(ctx context.Context, userID uint, currency, chain string, amount decimal.Decimal) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var asset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?", userID, currency, chain).
			First(&asset).Error; err != nil {
			return err
		}

		if asset.Available.LessThan(amount) {
			return errors.New("insufficient available balance")
		}

		asset.Available = asset.Available.Sub(amount)
		asset.Frozen = asset.Frozen.Add(amount)
		asset.UpdateTime = time.Now()

		return tx.Save(&asset).Error
	})
}

// UnfreezeAsset unfreezes asset
func (s *AssetService) UnfreezeAsset(ctx context.Context, userID uint, currency, chain string, amount decimal.Decimal) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var asset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?", userID, currency, chain).
			First(&asset).Error; err != nil {
			return err
		}

		if asset.Frozen.LessThan(amount) {
			return errors.New("insufficient frozen balance")
		}

		asset.Frozen = asset.Frozen.Sub(amount)
		asset.Available = asset.Available.Add(amount)
		asset.UpdateTime = time.Now()

		return tx.Save(&asset).Error
	})
}

// TransferAsset transfers asset between users
func (s *AssetService) TransferAsset(ctx context.Context, fromUserID, toUserID uint, currency, chain string, amount decimal.Decimal) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Deduct from sender
		var fromAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?", fromUserID, currency, chain).
			First(&fromAsset).Error; err != nil {
			return err
		}

		if fromAsset.Available.LessThan(amount) {
			return errors.New("insufficient balance")
		}

		fromAsset.Available = fromAsset.Available.Sub(amount)
		fromAsset.UpdateTime = time.Now()

		if err := tx.Save(&fromAsset).Error; err != nil {
			return err
		}

		// Add to recipient
		var toAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?", toUserID, currency, chain).
			First(&toAsset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				toAsset = models.UserAsset{
					UserID:     toUserID,
					Currency:   currency,
					Chain:      chain,
					Available:  amount,
					Frozen:     decimal.Zero,
					UpdateTime: time.Now(),
				}
				return tx.Create(&toAsset).Error
			}
			return err
		}

		toAsset.Available = toAsset.Available.Add(amount)
		toAsset.UpdateTime = time.Now()

		return tx.Save(&toAsset).Error
	})
}

// CreateDeposit creates a deposit record
func (s *AssetService) CreateDeposit(deposit *models.Deposit) error {
	return database.DB.Create(deposit).Error
}

// CreateWithdrawal creates a withdrawal record
func (s *AssetService) CreateWithdrawal(withdrawal *models.Withdrawal) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Freeze the withdrawal amount
		var asset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
			withdrawal.UserID, withdrawal.Currency, withdrawal.Chain).
			First(&asset).Error; err != nil {
			return err
		}

		totalAmount := withdrawal.Amount.Add(withdrawal.Fee)
		if asset.Available.LessThan(totalAmount) {
			return errors.New("insufficient balance")
		}

		asset.Available = asset.Available.Sub(totalAmount)
		asset.Frozen = asset.Frozen.Add(totalAmount)
		asset.UpdateTime = time.Now()

		if err := tx.Save(&asset).Error; err != nil {
			return err
		}

		return tx.Create(withdrawal).Error
	})
}

// FreezeAssetWithTx freezes asset within an existing transaction
func (s *AssetService) FreezeAssetWithTx(tx *gorm.DB, userID uint, currency, chain string, amount decimal.Decimal) error {
	var asset models.UserAsset
	if err := tx.Where("user_id = ? AND currency = ? AND chain = ?", userID, currency, chain).
		First(&asset).Error; err != nil {
		return err
	}

	if asset.Available.LessThan(amount) {
		return errors.New("insufficient available balance")
	}

	asset.Available = asset.Available.Sub(amount)
	asset.Frozen = asset.Frozen.Add(amount)
	asset.UpdateTime = time.Now()

	return tx.Save(&asset).Error
}
