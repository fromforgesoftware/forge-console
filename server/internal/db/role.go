package db

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"github.com/fromforgesoftware/go-kit/slicesx"
	"gorm.io/gorm"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

var roleFieldMapping = map[string]string{
	"id":   "slug",
	"slug": "slug",
	"name": "name",
	"kind": "kind",
}

type roleEntity struct {
	ESlug        string    `gorm:"column:slug;primaryKey"`
	ECreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	EName        string    `gorm:"column:name"`
	EKind        string    `gorm:"column:kind"`
	EPermissions []string  `gorm:"-"`
}

func (*roleEntity) TableName() string       { return "foundry.role" }
func (e *roleEntity) ID() string            { return e.ESlug }
func (e *roleEntity) LID() string           { return "" }
func (e *roleEntity) Type() resource.Type   { return app.ResourceTypeRole }
func (e *roleEntity) CreatedAt() time.Time  { return e.ECreatedAt }
func (e *roleEntity) UpdatedAt() time.Time  { return e.EUpdatedAt }
func (e *roleEntity) DeletedAt() *time.Time { return nil }
func (e *roleEntity) Slug() string          { return e.ESlug }
func (e *roleEntity) Name() string          { return e.EName }
func (e *roleEntity) Kind() app.RoleKind    { return app.RoleKind(e.EKind) }
func (e *roleEntity) Permissions() []string {
	if e.EPermissions == nil {
		return []string{}
	}
	return e.EPermissions
}

type roleRepo struct{ *postgres.Repo }

func NewRoleRepository(db *gormdb.DBClient) (*roleRepo, error) {
	r, err := postgres.NewRepo(db, roleFieldMapping)
	if err != nil {
		return nil, err
	}
	return &roleRepo{Repo: r}, nil
}

func (r *roleRepo) Get(ctx context.Context, opts ...search.Option) (app.Role, error) {
	s := search.New(opts...)
	var e roleEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	perms, err := r.permissionsFor(ctx, e.ESlug)
	if err != nil {
		return nil, err
	}
	e.EPermissions = perms
	return &e, nil
}

func (r *roleRepo) List(ctx context.Context, opts ...search.Option) (resource.ListResponse[app.Role], error) {
	s := search.New(append([]search.Option{search.WithQueryOpts(query.SortBy("name", query.SortAsc))}, opts...)...)
	var found []*roleEntity
	if err := r.QueryApply(ctx, s.Query()).Find(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	for _, e := range found {
		perms, err := r.permissionsFor(ctx, e.ESlug)
		if err != nil {
			return nil, err
		}
		e.EPermissions = perms
	}
	var total int64
	if err := r.CountApply(ctx, new(roleEntity), s.Query()).Count(&total).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	out := slicesx.Map(found, func(e *roleEntity) app.Role { return e })
	return resource.NewListResponse(out, int(total)), nil
}

// Create upserts a role and replaces its permission set. Idempotent so bootstrap
// can reuse it for seeding the SYSTEM roles.
func (r *roleRepo) Create(ctx context.Context, role app.Role) (app.Role, error) {
	kind := role.Kind()
	if kind == "" {
		kind = app.RoleCustom
	}
	if err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`INSERT INTO foundry.role (slug, name, kind) VALUES (?, ?, ?)
			 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, updated_at = now()`,
			role.Slug(), role.Name(), string(kind),
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM foundry.role_permission WHERE role_slug = ?`, role.Slug()).Error; err != nil {
			return err
		}
		for _, p := range role.Permissions() {
			if err := tx.Exec(
				`INSERT INTO foundry.role_permission (role_slug, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
				role.Slug(), p,
			).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return r.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", role.Slug())))
}

func (r *roleRepo) Delete(ctx context.Context, delType repository.DeleteType, opts ...search.Option) error {
	s := search.New(opts...)
	op := r.QueryApply(ctx, s.Query())
	if delType == repository.DeleteTypeHard {
		op = op.Unscoped()
	}
	if err := op.Delete(&roleEntity{}).Error; err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}

func (r *roleRepo) BindSubject(ctx context.Context, subjectType app.SubjectType, subjectID, roleSlug string) error {
	if err := r.DB.WithContext(ctx).Exec(
		`INSERT INTO foundry.subject_role (subject_type, subject_id, role_slug) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`,
		string(subjectType), subjectID, roleSlug,
	).Error; err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}

func (r *roleRepo) SetSubjectRoles(ctx context.Context, subjectType app.SubjectType, subjectID string, roleSlugs []string) error {
	if err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`DELETE FROM foundry.subject_role WHERE subject_type = ? AND subject_id = ?`,
			string(subjectType), subjectID,
		).Error; err != nil {
			return err
		}
		for _, slug := range roleSlugs {
			if err := tx.Exec(
				`INSERT INTO foundry.subject_role (subject_type, subject_id, role_slug) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`,
				string(subjectType), subjectID, slug,
			).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return postgres.NewErrUnknown(err)
	}
	return nil
}

func (r *roleRepo) RolesForSubject(ctx context.Context, subjectType app.SubjectType, subjectID string) ([]app.Role, error) {
	var found []*roleEntity
	if err := r.DB.WithContext(ctx).Raw(
		`SELECT ro.slug, ro.name, ro.kind FROM foundry.role ro
		 JOIN foundry.subject_role sr ON sr.role_slug = ro.slug
		 WHERE sr.subject_type = ? AND sr.subject_id = ?`, string(subjectType), subjectID,
	).Scan(&found).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	for _, e := range found {
		perms, err := r.permissionsFor(ctx, e.ESlug)
		if err != nil {
			return nil, err
		}
		e.EPermissions = perms
	}
	return slicesx.Map(found, func(e *roleEntity) app.Role { return e }), nil
}

// CountEnabledUsersWithPermission counts distinct enabled users holding a
// permission, joining subject bindings to roles to the user table.
func (r *roleRepo) CountEnabledUsersWithPermission(ctx context.Context, permission string) (int, error) {
	var count int64
	if err := r.DB.WithContext(ctx).Raw(
		`SELECT COUNT(DISTINCT sr.subject_id)
		 FROM foundry.subject_role sr
		 JOIN foundry.role_permission rp ON rp.role_slug = sr.role_slug
		 JOIN foundry.app_user u ON u.id = sr.subject_id
		 WHERE sr.subject_type = 'USER' AND u.status = 'ENABLED' AND rp.permission_id = ?`,
		permission,
	).Scan(&count).Error; err != nil {
		return 0, postgres.NewErrUnknown(err)
	}
	return int(count), nil
}

func (r *roleRepo) permissionsFor(ctx context.Context, roleSlug string) ([]string, error) {
	var perms []string
	if err := r.DB.WithContext(ctx).Raw(
		`SELECT permission_id FROM foundry.role_permission WHERE role_slug = ? ORDER BY permission_id`, roleSlug,
	).Scan(&perms).Error; err != nil {
		return nil, postgres.NewErrUnknown(err)
	}
	return perms, nil
}
