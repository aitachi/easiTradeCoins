package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TradingCommunity represents a trading community
type TradingCommunity struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	OwnerID     uint      `json:"owner_id" gorm:"index"`
	MemberCount int       `json:"member_count" gorm:"default:0"`
	PostCount   int       `json:"post_count" gorm:"default:0"`
	IsPublic    bool      `json:"is_public" gorm:"default:true"`
	Category    string    `json:"category"` // general/signals/education/analysis
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// CommunityMember represents a community member
type CommunityMember struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CommunityID uint      `json:"community_id" gorm:"index"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Role        string    `json:"role" gorm:"default:member"` // owner/moderator/member
	JoinTime    time.Time `json:"join_time"`
}

// Post represents a community post
type Post struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	CommunityID uint            `json:"community_id" gorm:"index"`
	AuthorID    uint            `json:"author_id" gorm:"index"`
	Title       string          `json:"title"`
	Content     string          `json:"content" gorm:"type:text"`
	Images      string          `json:"images" gorm:"type:text"` // JSON array of image URLs
	Likes       int             `json:"likes" gorm:"default:0"`
	Comments    int             `json:"comments" gorm:"default:0"`
	Views       int             `json:"views" gorm:"default:0"`
	IsPinned    bool            `json:"is_pinned" gorm:"default:false"`
	Tags        string          `json:"tags"` // Comma-separated tags
	CreateTime  time.Time       `json:"create_time"`
	UpdateTime  time.Time       `json:"update_time"`
}

// Comment represents a comment on a post
type Comment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PostID     uint      `json:"post_id" gorm:"index"`
	AuthorID   uint      `json:"author_id" gorm:"index"`
	ParentID   *uint     `json:"parent_id,omitempty"` // For nested comments
	Content    string    `json:"content" gorm:"type:text"`
	Likes      int       `json:"likes" gorm:"default:0"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// Like represents a like on a post or comment
type Like struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	TargetType string    `json:"target_type"` // post/comment
	TargetID   uint      `json:"target_id" gorm:"index"`
	CreateTime time.Time `json:"create_time"`
}

// TradingSignal represents a trading signal shared in community
type TradingSignal struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	AuthorID    uint            `json:"author_id" gorm:"index"`
	CommunityID uint            `json:"community_id" gorm:"index"`
	Symbol      string          `json:"symbol"`
	Type        string          `json:"type"` // long/short
	EntryPrice  decimal.Decimal `json:"entry_price" gorm:"type:decimal(36,18)"`
	StopLoss    decimal.Decimal `json:"stop_loss" gorm:"type:decimal(36,18)"`
	TakeProfit1 decimal.Decimal `json:"take_profit1" gorm:"type:decimal(36,18)"`
	TakeProfit2 *decimal.Decimal `json:"take_profit2,omitempty" gorm:"type:decimal(36,18)"`
	TakeProfit3 *decimal.Decimal `json:"take_profit3,omitempty" gorm:"type:decimal(36,18)"`
	Status      string          `json:"status" gorm:"default:active"` // active/hit_tp1/hit_tp2/hit_tp3/hit_sl/closed
	Result      *decimal.Decimal `json:"result,omitempty" gorm:"type:decimal(10,4)"` // P&L percentage
	CreateTime  time.Time       `json:"create_time"`
	UpdateTime  time.Time       `json:"update_time"`
}

// CommunityService manages trading communities
type CommunityService struct {
	mutex sync.RWMutex
	db    *gorm.DB
}

// NewCommunityService creates a new community service
func NewCommunityService(db *gorm.DB) *CommunityService {
	return &CommunityService{
		db: db,
	}
}

// CreateCommunity creates a new trading community
func (s *CommunityService) CreateCommunity(
	ctx context.Context,
	ownerID uint,
	name, description, category string,
	isPublic bool,
) (*TradingCommunity, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	community := &TradingCommunity{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		MemberCount: 1, // Owner is first member
		PostCount:   0,
		IsPublic:    isPublic,
		Category:    category,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	return community, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return err
		}

		// Add owner as member
		member := &CommunityMember{
			CommunityID: community.ID,
			UserID:      ownerID,
			Role:        "owner",
			JoinTime:    time.Now(),
		}

		return tx.Create(member).Error
	})
}

// JoinCommunity joins a community
func (s *CommunityService) JoinCommunity(ctx context.Context, communityID, userID uint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if already member
	var existing CommunityMember
	err := s.db.Where("community_id = ? AND user_id = ?", communityID, userID).First(&existing).Error
	if err == nil {
		return errors.New("already a member")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		member := &CommunityMember{
			CommunityID: communityID,
			UserID:      userID,
			Role:        "member",
			JoinTime:    time.Now(),
		}

		if err := tx.Create(member).Error; err != nil {
			return err
		}

		// Update community member count
		return tx.Model(&TradingCommunity{}).
			Where("id = ?", communityID).
			UpdateColumn("member_count", gorm.Expr("member_count + ?", 1)).Error
	})
}

// LeaveCommunity leaves a community
func (s *CommunityService) LeaveCommunity(ctx context.Context, communityID, userID uint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("community_id = ? AND user_id = ? AND role != ?", communityID, userID, "owner").
			Delete(&CommunityMember{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("not a member or cannot leave as owner")
		}

		// Update community member count
		return tx.Model(&TradingCommunity{}).
			Where("id = ?", communityID).
			UpdateColumn("member_count", gorm.Expr("member_count - ?", 1)).Error
	})
}

// CreatePost creates a new post in a community
func (s *CommunityService) CreatePost(
	ctx context.Context,
	communityID, authorID uint,
	title, content, tags string,
	images []string,
) (*Post, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Verify membership
	var member CommunityMember
	if err := s.db.Where("community_id = ? AND user_id = ?", communityID, authorID).First(&member).Error; err != nil {
		return nil, errors.New("not a member of this community")
	}

	post := &Post{
		CommunityID: communityID,
		AuthorID:    authorID,
		Title:       title,
		Content:     content,
		Tags:        tags,
		Likes:       0,
		Comments:    0,
		Views:       0,
		IsPinned:    false,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	return post, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// Update community post count
		return tx.Model(&TradingCommunity{}).
			Where("id = ?", communityID).
			UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).Error
	})
}

// AddComment adds a comment to a post
func (s *CommunityService) AddComment(
	ctx context.Context,
	postID, authorID uint,
	content string,
	parentID *uint,
) (*Comment, error) {
	comment := &Comment{
		PostID:     postID,
		AuthorID:   authorID,
		ParentID:   parentID,
		Content:    content,
		Likes:      0,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	return comment, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// Update post comment count
		return tx.Model(&Post{}).
			Where("id = ?", postID).
			UpdateColumn("comments", gorm.Expr("comments + ?", 1)).Error
	})
}

// LikePost likes a post
func (s *CommunityService) LikePost(ctx context.Context, postID, userID uint) error {
	// Check if already liked
	var existing Like
	err := s.db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, "post", postID).
		First(&existing).Error

	if err == nil {
		return errors.New("already liked")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		like := &Like{
			UserID:     userID,
			TargetType: "post",
			TargetID:   postID,
			CreateTime: time.Now(),
		}

		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// Update post like count
		return tx.Model(&Post{}).
			Where("id = ?", postID).
			UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error
	})
}

// PublishSignal publishes a trading signal
func (s *CommunityService) PublishSignal(
	ctx context.Context,
	authorID, communityID uint,
	symbol, signalType string,
	entryPrice, stopLoss, takeProfit1 decimal.Decimal,
	takeProfit2, takeProfit3 *decimal.Decimal,
) (*TradingSignal, error) {
	signal := &TradingSignal{
		AuthorID:    authorID,
		CommunityID: communityID,
		Symbol:      symbol,
		Type:        signalType,
		EntryPrice:  entryPrice,
		StopLoss:    stopLoss,
		TakeProfit1: takeProfit1,
		TakeProfit2: takeProfit2,
		TakeProfit3: takeProfit3,
		Status:      "active",
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	if err := s.db.Create(signal).Error; err != nil {
		return nil, err
	}

	return signal, nil
}

// GetCommunities gets all communities
func (s *CommunityService) GetCommunities(ctx context.Context, category string, limit, offset int) ([]TradingCommunity, int64, error) {
	var communities []TradingCommunity
	var total int64

	query := s.db.Model(&TradingCommunity{}).Where("is_public = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("member_count DESC").
		Limit(limit).
		Offset(offset).
		Find(&communities).Error; err != nil {
		return nil, 0, err
	}

	return communities, total, nil
}

// GetCommunityPosts gets posts in a community
func (s *CommunityService) GetCommunityPosts(ctx context.Context, communityID uint, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var total int64

	if err := s.db.Model(&Post{}).Where("community_id = ?", communityID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("community_id = ?", communityID).
		Order("is_pinned DESC, create_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostComments gets comments for a post
func (s *CommunityService) GetPostComments(ctx context.Context, postID uint) ([]Comment, error) {
	var comments []Comment
	if err := s.db.Where("post_id = ?", postID).
		Order("create_time ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	return comments, nil
}

// GetCommunitySignals gets trading signals in a community
func (s *CommunityService) GetCommunitySignals(ctx context.Context, communityID uint, status string) ([]TradingSignal, error) {
	var signals []TradingSignal
	query := s.db.Where("community_id = ?", communityID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("create_time DESC").Find(&signals).Error; err != nil {
		return nil, err
	}

	return signals, nil
}
