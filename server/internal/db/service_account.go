package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/application/repository"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/fromforgesoftware/go-kit/slicesx"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

var serviceAccountFieldMapping = map[string]string{
	"id":         "id",
	"name":       "name",
	"clientId":   "client_id",
	"status":     "status",
	"lastUsedAt": "last_used_at",
}

type serviceAccountEntity struct {
	EID         string     `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt  time.Time  `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt  time.Time  `gorm:"column:updated_at;type:timestamptz;default:now()"`
	EName       string     `gorm:"column:name"`
	EClientID   string     `gorm:"column:client_id"`
	ESecretHash string     `gorm:"column:secret_hash"`
	EStatus     string     `gorm:"column:status"`
	ELastUsedAt *time.Time `gorm:"column:last_used_at"`
}

func (*serviceAccountEntity) TableName() string       { return "foundry.service_account" }
func (e *serviceAccountEntity) ID() string            { return e.EID }
func (e *serviceAccountEntity) LID() string           { return "" }
func (e *serviceAccountEntity) Type() resource.Type   { return app.ResourceTypeServiceAccount }
func (e *serviceAccountEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *serviceAccountEntity) UpdatedAt() time.Time  { return e.EUpdatedAt }
func (e *serviceAccountEntity) DeletedAt() *time.Time { return nil }
func (e *serviceAccountEntity) Name() string          { return e.EName }
func (e *serviceAccountEntity) ClientID() string      { return e.EClientID }
func (e *serviceAccountEntity) Status() app.ServiceAccountStatus {
	return app.ServiceAccountStatus(e.EStatus)
}
func (e *serviceAccountEntity) LastUsedAt() *time.Time { return e.ELastUsedAt }

type serviceAccountRepo struct{ *postgres.Repo }

func NewServiceAccountRepository(db *gormdb.DBClient) (*serviceAccountRepo, error) {
	r, err := postgres.NewRepo(db, serviceAccountFieldMapping)
	if err != nil {
		return nil, err
	}
	return &serviceAccountRepo{Repo: r}, nil
}

func (r *serviceAccountRepo) Get(ctx context.Context, opts ...search.Option) (app.ServiceAccount, error) {
	s := search.New(opts...)
	var e serviceAccountEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *serviceAccountRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[app.ServiceAccount], error) {
	s := search.New(append([]search.Option{search.WithQueryOpts(query.SortBy("name", query.SortAsc))}, opts...)...)
	var found []*serviceAccountEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(serviceAccountEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *serviceAccountEntity) app.ServiceAccount { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *serviceAccountRepo) GetByClientID(ctx context.Context, clientID string) (app.ServiceAccount, string, error) {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "clientId", clientID)))
	var e serviceAccountEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return nil, "", postgres.NewErrUnknown(err)
	}
	return &e, e.ESecretHash, nil
}

func (r *serviceAccountRepo) CreateWithSecret(ctx context.Context, sa app.ServiceAccount, secretHash string) (app.ServiceAccount, error) {
	status := string(sa.Status())
	if status == "" {
		status = string(app.ServiceAccountEnabled)
	}
	e := serviceAccountEntity{
		EName:       sa.Name(),
		EClientID:   sa.ClientID(),
		ESecretHash: secretHash,
		EStatus:     status,
	}
	if err := r.DB.WithContext(ctx).Omit("id", "created_at", "updated_at", "last_used_at").Create(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *serviceAccountRepo) Delete(ctx context.Context, delType repository.DeleteType, opts ...search.Option) error {
	s := search.New(opts...)
	op := r.QueryApply(ctx, s.Query())
	if delType == repository.DeleteTypeHard {
		op = op.Unscoped()
	}
	if err := op.Delete(&serviceAccountEntity{}).Error; err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}

func (r *serviceAccountRepo) TouchLastUsed(ctx context.Context, id string, at time.Time) error {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", id)))
	res := r.PatchApply(ctx, s.Query(), &serviceAccountEntity{}, map[string]any{
		"lastUsedAt": at,
		"updated_at": time.Now().UTC(),
	})
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	if res.RowsAffected == 0 {
		return apierrors.NotFound("service_account", id)
	}
	return nil
}
