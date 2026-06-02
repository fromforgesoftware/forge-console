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

var permissionFieldMapping = map[string]string{
	"id":           "id",
	"resourceType": "resource_type",
	"verb":         "verb",
	"description":  "description",
}

type permissionEntity struct {
	EID           string `gorm:"column:id;primaryKey"`
	EResourceType string `gorm:"column:resource_type"`
	EVerb         string `gorm:"column:verb"`
	EDescription  string `gorm:"column:description"`
}

func (*permissionEntity) TableName() string       { return "foundry.permission" }
func (e *permissionEntity) ID() string            { return e.EID }
func (e *permissionEntity) LID() string           { return "" }
func (e *permissionEntity) Type() resource.Type   { return app.ResourceTypePermission }
func (e *permissionEntity) CreatedAt() time.Time  { return time.Time{} }
func (e *permissionEntity) UpdatedAt() time.Time  { return time.Time{} }
func (e *permissionEntity) DeletedAt() *time.Time { return nil }
func (e *permissionEntity) ResourceType() string  { return e.EResourceType }
func (e *permissionEntity) Verb() string          { return e.EVerb }
func (e *permissionEntity) Description() string   { return e.EDescription }

type permissionRepo struct{ *postgres.Repo }

func NewPermissionRepository(db *gormdb.DBClient) (*permissionRepo, error) {
	r, err := postgres.NewRepo(db, permissionFieldMapping)
	if err != nil {
		return nil, err
	}
	return &permissionRepo{Repo: r}, nil
}

func (r *permissionRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[app.Permission], error) {
	s := search.New(append([]search.Option{search.WithQueryOpts(query.SortBy("id", query.SortAsc))}, opts...)...)
	var found []*permissionEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(permissionEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *permissionEntity) app.Permission { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *permissionRepo) Upsert(ctx context.Context, p app.Permission) error {
	res := r.DB.WithContext(ctx).Exec(
		`INSERT INTO foundry.permission (id, resource_type, verb, description)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (id) DO UPDATE SET
		   resource_type = EXCLUDED.resource_type, verb = EXCLUDED.verb,
		   description = EXCLUDED.description`,
		p.ID(), p.ResourceType(), p.Verb(), p.Description(),
	)
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	return nil
}
