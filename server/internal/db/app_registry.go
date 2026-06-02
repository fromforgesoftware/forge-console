package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/fromforgesoftware/go-kit/slicesx"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

var appFieldMapping = map[string]string{
	"id":           "slug",
	"slug":         "slug",
	"name":         "name",
	"kind":         "kind",
	"adminBaseURL": "admin_base_url",
	"enabled":      "enabled",
}

type appEntity struct {
	ESlug         string    `gorm:"column:slug;primaryKey"`
	ECreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	EName         string    `gorm:"column:name"`
	EKind         string    `gorm:"column:kind"`
	EAdminBaseURL string    `gorm:"column:admin_base_url"`
	EEnabled      bool      `gorm:"column:enabled"`
}

func (*appEntity) TableName() string       { return "foundry.app" }
func (e *appEntity) ID() string            { return e.ESlug }
func (e *appEntity) LID() string           { return "" }
func (e *appEntity) Type() resource.Type   { return app.ResourceTypeApp }
func (e *appEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *appEntity) UpdatedAt() time.Time  { return e.EUpdatedAt }
func (e *appEntity) DeletedAt() *time.Time { return nil }
func (e *appEntity) Slug() string          { return e.ESlug }
func (e *appEntity) Name() string          { return e.EName }
func (e *appEntity) Kind() string          { return e.EKind }
func (e *appEntity) AdminBaseURL() string  { return e.EAdminBaseURL }
func (e *appEntity) Enabled() bool         { return e.EEnabled }

type appRepo struct{ *postgres.Repo }

func NewAppRepository(db *gormdb.DBClient) (*appRepo, error) {
	r, err := postgres.NewRepo(db, appFieldMapping)
	if err != nil {
		return nil, err
	}
	return &appRepo{Repo: r}, nil
}

func (r *appRepo) Get(ctx context.Context, opts ...search.Option) (app.App, error) {
	s := search.New(opts...)
	var e appEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *appRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[app.App], error) {
	s := search.New(append([]search.Option{search.WithQueryOpts(query.SortBy("name", query.SortAsc))}, opts...)...)
	var found []*appEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(appEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *appEntity) app.App { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *appRepo) Upsert(ctx context.Context, a app.App) error {
	res := r.DB.WithContext(ctx).Exec(
		`INSERT INTO foundry.app (slug, name, kind, admin_base_url, enabled)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (slug) DO UPDATE SET
		   name = EXCLUDED.name, kind = EXCLUDED.kind,
		   admin_base_url = EXCLUDED.admin_base_url, enabled = EXCLUDED.enabled,
		   updated_at = now()`,
		a.Slug(), a.Name(), a.Kind(), a.AdminBaseURL(), a.Enabled(),
	)
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	return nil
}
