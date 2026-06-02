package app

import (
	"context"
	"strings"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

const ResourceTypeApp resource.Type = "apps"

// App is a managed forge app (Aegis, Hallmark, Herald) the console
// administers. AdminBaseURL is the in-cluster admin API the gateway proxies to.
// The slug is the resource id.
type App interface {
	resource.Resource
	Slug() string
	Name() string
	Kind() string
	AdminBaseURL() string
	Enabled() bool
}

type appRes struct {
	resource.Resource

	name         string
	kind         string
	adminBaseURL string
	enabled      bool
}

type AppOption func(*appRes)

func WithAppName(name string) AppOption        { return func(a *appRes) { a.name = name } }
func WithAppKind(kind string) AppOption        { return func(a *appRes) { a.kind = kind } }
func WithAppAdminBaseURL(url string) AppOption { return func(a *appRes) { a.adminBaseURL = url } }
func WithAppEnabled(enabled bool) AppOption    { return func(a *appRes) { a.enabled = enabled } }

// NewApp builds an App aggregate keyed by slug.
func NewApp(slug string, opts ...AppOption) App {
	a := &appRes{
		Resource: resource.New(resource.WithType(ResourceTypeApp), resource.WithID(slug)),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *appRes) Slug() string         { return a.ID() }
func (a *appRes) Name() string         { return a.name }
func (a *appRes) Kind() string         { return a.kind }
func (a *appRes) AdminBaseURL() string { return a.adminBaseURL }
func (a *appRes) Enabled() bool        { return a.enabled }

// AppRepository persists the app registry via the kit generic surface plus an
// idempotent slug-keyed Upsert used by bootstrap and the admin registry editor.
type AppRepository interface {
	repository.Getter[App]
	repository.Lister[App]
	Upsert(ctx context.Context, a App) error
}

// AppUsecase is the read surface for the registry: the SPA nav (ListEnabled)
// and the gateway's per-request app lookup (Get).
type AppUsecase interface {
	ListEnabled(ctx context.Context) ([]App, error)
	Get(ctx context.Context, slug string) (App, error)
}

type appUsecase struct {
	apps AppRepository
}

func NewAppUsecase(apps AppRepository) AppUsecase {
	return &appUsecase{apps: apps}
}

// Get returns an enabled app by slug. A disabled app is treated as
// absent so the gateway can't proxy to it.
func (uc *appUsecase) Get(ctx context.Context, slug string) (App, error) {
	a, err := uc.apps.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", slug)))
	if err != nil {
		return nil, err
	}
	if a == nil || !a.Enabled() {
		return nil, apierrors.NotFound("app", slug)
	}
	return a, nil
}

func (uc *appUsecase) ListEnabled(ctx context.Context) ([]App, error) {
	res, err := uc.apps.List(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "enabled", true)))
	if err != nil {
		return nil, err
	}
	return res.Results(), nil
}

// UpsertAppCommand creates or updates a registry entry keyed by slug.
type UpsertAppCommand struct {
	Slug         string
	Name         string
	Kind         string
	AdminBaseURL string
	Enabled      bool
}

// AppAdminUsecase is the admin registry read/write surface.
type AppAdminUsecase interface {
	repository.Getter[App]
	repository.Lister[App]

	Upsert(ctx context.Context, cmd UpsertAppCommand) (App, error)
}

type appAdminUsecase struct {
	repository.Getter[App]
	repository.Lister[App]

	apps AppRepository
}

func NewAppAdminUsecase(apps AppRepository) AppAdminUsecase {
	return &appAdminUsecase{
		Getter: usecase.NewGetter[App](apps, ResourceTypeApp),
		Lister: usecase.NewLister[App](apps),
		apps:   apps,
	}
}

func (uc *appAdminUsecase) Upsert(ctx context.Context, cmd UpsertAppCommand) (App, error) {
	if cmd.Slug == "" {
		return nil, apierrors.InvalidArgument("slug is required")
	}
	a := NewApp(cmd.Slug,
		WithAppName(cmd.Name),
		WithAppKind(cmd.Kind),
		WithAppAdminBaseURL(cmd.AdminBaseURL),
		WithAppEnabled(cmd.Enabled),
	)
	if err := uc.apps.Upsert(ctx, a); err != nil {
		return nil, err
	}
	return uc.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", cmd.Slug)))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
