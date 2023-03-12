package storage

import (
	"context"
	"errors"
	"github.com/sealbro/go-feed-me/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Article struct {
	Created     time.Time `json:"created"`
	Published   time.Time `json:"published"`
	ResourceId  string    `json:"resource_id"`
	Link        string    `json:"link" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	Image       string    `json:"image"`
}

type ArticleRepository struct {
	db *db.DB
}

func NewArticleRepository(db *db.DB) (*ArticleRepository, error) {
	err := db.AutoMigrate(&Article{})
	if err != nil {
		return nil, err
	}
	return &ArticleRepository{db: db}, nil
}

func (r *ArticleRepository) Upsert(ctx context.Context, article *Article) error {
	columns := []string{"title", "published", "description", "content", "author", "image"}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(article)

	return tx.Error
}

func (r *ArticleRepository) List(ctx context.Context, after time.Time) ([]*Article, error) {
	articles := make([]*Article, 0)
	last := r.db.WithContext(ctx).Order("published desc").Find(&articles, "published > ?", after)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return articles, last.Error
}
