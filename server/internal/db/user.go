// Package db holds Foundry's Postgres repositories, built on the kit's
// persistence/postgres Repo (QueryApply/CountApply/PatchApply over a field map).
// Foundry's transport stays plain JSON, but the query mechanics use the kit.
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

var userFieldMapping = map[string]string{
	"id":          "id",
	"email":       "email",
	"displayName": "display_name",
	"status":      "status",
}

type userEntity struct {
	EID          string    `gorm:"column:id;type:uuid;default:uuid_generate_v4();primaryKey"`
	ECreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	EEmail       string    `gorm:"column:email"`
	EDisplayName string    `gorm:"column:display_name"`
	EStatus      string    `gorm:"column:status"`
}

func (*userEntity) TableName() string        { return "foundry.app_user" }
func (e *userEntity) ID() string             { return e.EID }
func (e *userEntity) LID() string            { return "" }
func (e *userEntity) Type() resource.Type    { return app.ResourceTypeUser }
func (e *userEntity) CreatedAt() time.Time   { return e.ECreatedAt }
func (e *userEntity) UpdatedAt() time.Time   { return e.EUpdatedAt }
func (e *userEntity) DeletedAt() *time.Time  { return nil }
func (e *userEntity) Email() string          { return e.EEmail }
func (e *userEntity) DisplayName() string    { return e.EDisplayName }
func (e *userEntity) Status() app.UserStatus { return app.UserStatus(e.EStatus) }

type userRepo struct{ *postgres.Repo }

func NewUserRepository(db *gormdb.DBClient) (*userRepo, error) {
	r, err := postgres.NewRepo(db, userFieldMapping)
	if err != nil {
		return nil, err
	}
	return &userRepo{Repo: r}, nil
}

func (r *userRepo) Get(ctx context.Context, opts ...search.Option) (app.User, error) {
	s := search.New(opts...)
	var e userEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *userRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[app.User], error) {
	s := search.New(append([]search.Option{search.WithQueryOpts(query.SortBy("email", query.SortAsc))}, opts...)...)
	var found []*userEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	var total int64
	if err := r.CountApply(ctx, new(userEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *userEntity) app.User { return e })
	return resource.NewListResponse(out, int(total)), nil
}

func (r *userRepo) Create(ctx context.Context, u app.User) (app.User, error) {
	status := string(u.Status())
	if status == "" {
		status = string(app.UserEnabled)
	}
	e := userEntity{EEmail: u.Email(), EDisplayName: u.DisplayName(), EStatus: status}
	if err := r.DB.WithContext(ctx).Omit("id", "created_at", "updated_at").Create(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return &e, nil
}

func (r *userRepo) Patch(ctx context.Context, opts ...repository.PatchOption) ([]app.User, error) {
	p := repository.NewPatchQuery(opts...)
	s := search.New(p.SearchOpts()...)
	fields := p.PatchFields()
	fields["updated_at"] = time.Now().UTC()
	res := r.PatchApply(ctx, s.Query(), &userEntity{}, fields)
	if res.Error != nil {
		return nil, postgres.NewErrUnknown(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, apierrors.NotFound("user", nil)
	}
	var found []*userEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return slicesx.Map(found, func(e *userEntity) app.User { return e }), nil
}

type credentialEntity struct {
	EUserID    string    `gorm:"column:user_id;type:uuid;primaryKey"`
	ECreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	EHash      string    `gorm:"column:hash"`
}

func (credentialEntity) TableName() string { return "foundry.user_credential" }

var credentialFieldMapping = map[string]string{
	"userId": "user_id",
	"hash":   "hash",
}

type credentialRepo struct{ *postgres.Repo }

func NewCredentialRepository(db *gormdb.DBClient) (*credentialRepo, error) {
	r, err := postgres.NewRepo(db, credentialFieldMapping)
	if err != nil {
		return nil, err
	}
	return &credentialRepo{Repo: r}, nil
}

func (r *credentialRepo) HashFor(ctx context.Context, userID string) (string, error) {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "userId", userID)))
	var e credentialEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return "", postgres.NewErrUnknown(err)
	}
	return e.EHash, nil
}

func (r *credentialRepo) Set(ctx context.Context, userID, hash string) error {
	res := r.DB.WithContext(ctx).Exec(
		`INSERT INTO foundry.user_credential (user_id, hash) VALUES (?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET hash = EXCLUDED.hash, updated_at = now()`,
		userID, hash,
	)
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	return nil
}
