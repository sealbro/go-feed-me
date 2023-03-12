package storage

import (
	"context"
	"errors"
	"github.com/sealbro/go-feed-me/pkg/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Resource struct {
	Created   time.Time `json:"created"`
	Modified  time.Time `json:"modified"`
	Published time.Time `json:"published"`
	Url       string    `json:"url" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Active    bool      `json:"active"`
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

func (r *ResourceRepository) Get(ctx context.Context, id int) (*Resource, error) {
	dbModel := &Resource{}
	last := r.db.WithContext(ctx).Last(dbModel, "id = ?", id)
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
