package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

var sessionFieldMapping = map[string]string{
	"id":        "id",
	"userId":    "user_id",
	"expiresAt": "expires_at",
}

type sessionEntity struct {
	EID        string    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUserID    string    `gorm:"column:user_id;type:uuid"`
	EExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz"`
}

func (sessionEntity) TableName() string { return "foundry.session" }

type sessionRepo struct{ *postgres.Repo }

func NewSessionRepository(db *gormdb.DBClient) (*sessionRepo, error) {
	r, err := postgres.NewRepo(db, sessionFieldMapping)
	if err != nil {
		return nil, err
	}
	return &sessionRepo{Repo: r}, nil
}

func (r *sessionRepo) Create(ctx context.Context, s app.Session) (app.Session, error) {
	e := sessionEntity{EUserID: s.UserID, EExpiresAt: s.ExpiresAt}
	if err := r.DB.WithContext(ctx).Omit("id", "created_at").Create(&e).Error; err != nil {
		return app.Session{}, postgres.NewErrUnknown(err)
	}
	return app.Session{ID: e.EID, UserID: e.EUserID, ExpiresAt: e.EExpiresAt}, nil
}

func (r *sessionRepo) Get(ctx context.Context, id string) (app.Session, error) {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", id)))
	var e sessionEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return app.Session{}, postgres.NewErrUnknown(err)
	}
	return app.Session{ID: e.EID, UserID: e.EUserID, ExpiresAt: e.EExpiresAt}, nil
}

func (r *sessionRepo) Delete(ctx context.Context, id string) error {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", id)))
	if err := r.QueryApply(ctx, s.Query()).Delete(&sessionEntity{}).Error; err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}
