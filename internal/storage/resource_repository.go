package storage

import (
	"context"
	"errors"
	"github.com/sealbro/go-feed-me/internal/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Resource struct {
	Active    bool      `json:"active"`
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
	Published time.Time `json:"published"`
	Title     string    `json:"title"`
	Url       string    `json:"url" gorm:"primaryKey"`
}

type ResourceRepository struct {
	db *db.DB
}

func NewResourceRepository(db *db.DB) (*ResourceRepository, error) {
	err := db.AutoMigrate(&Resource{})
	if err != nil {
		return nil, err
	}
	return &ResourceRepository{db: db}, nil
}

func (r *ResourceRepository) Get(ctx context.Context, url string) (*Resource, error) {
	dbModel := &Resource{}
	last := r.db.WithContext(ctx).Last(dbModel, "url = ?", url)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return dbModel, last.Error
}

func (r *ResourceRepository) List(ctx context.Context, active bool) ([]*Resource, error) {
	resources := make([]*Resource, 0)
	last := r.db.WithContext(ctx).Find(&resources, "active = ?", active)
	if errors.Is(last.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return resources, last.Error
}

func (r *ResourceRepository) Upsert(ctx context.Context, repoInfo *Resource) error {
	repoInfo.Modified = time.Now()

	columns := []string{"title", "published", "modified", "active"}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(repoInfo)

	return tx.Error
}

func (r *ResourceRepository) Delete(ctx context.Context, urls []string) error {
	tx := r.db.WithContext(ctx).Delete(&Resource{}, "url IN ?", urls)

	return tx.Error
}

func (r *ResourceRepository) Activate(ctx context.Context, urls []string, active bool) error {
	modified := time.Now()

	tx := r.db.WithContext(ctx).Model(&Resource{}).Where("url IN ?", urls).Updates(map[string]interface{}{
		"active":   active,
		"modified": modified,
	})

	return tx.Error
}
