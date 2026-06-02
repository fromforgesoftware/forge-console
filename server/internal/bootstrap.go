package internal

import (
	"context"
	"os"
	"strings"

	"go.uber.org/fx"

	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/monitoring/logger"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// bootstrapConfig seeds the first user + the managed app registry from
// the environment (Keycloak KEYCLOAK_ADMIN-style), idempotently every boot.
type bootstrapConfig struct {
	adminEmail    string
	adminPassword string
	adminName     string
	apps          string // "slug=Name=adminBaseURL[=moduleUri],slug2=Name2=url2"
}

func newBootstrapConfig() bootstrapConfig {
	return bootstrapConfig{
		adminEmail:    os.Getenv("FOUNDRY_BOOTSTRAP_ADMIN_EMAIL"),
		adminPassword: os.Getenv("FOUNDRY_BOOTSTRAP_ADMIN_PASSWORD"),
		adminName:     envOr("FOUNDRY_BOOTSTRAP_ADMIN_NAME", "Administrator"),
		apps:          envOr("FOUNDRY_APPS", os.Getenv("FOUNDRY_PRODUCTS")),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func registerBootstrap(lc fx.Lifecycle, users app.UserRepository, creds app.CredentialRepository, roles app.RoleRepository, perms app.PermissionRepository, apps app.AppRepository, hasher password.Hasher) {
	cfg := newBootstrapConfig()
	log := logger.New()
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := ensureBootstrap(context.Background(), cfg, users, creds, roles, perms, apps, hasher, log); err != nil {
				log.Error("bootstrap seed failed", "error", err)
			}
			return nil
		},
	})
}

func ensureBootstrap(ctx context.Context, cfg bootstrapConfig, users app.UserRepository, creds app.CredentialRepository, roles app.RoleRepository, perms app.PermissionRepository, apps app.AppRepository, hasher password.Hasher, log logger.Logger) error {
	if err := ensurePermissionCatalog(ctx, perms); err != nil {
		return err
	}
	if cfg.adminEmail != "" && cfg.adminPassword != "" {
		userID, err := ensureAdmin(ctx, cfg, users, creds, hasher, log)
		if err != nil {
			return err
		}
		if err := ensureSuperadmin(ctx, roles, userID, log); err != nil {
			return err
		}
	}
	return ensureApps(ctx, cfg, apps, perms, log)
}

// controlPlaneResources name the control-plane resource types whose read/write
// permissions seed the catalog (the RoleForm picker lists them).
var controlPlaneResources = []string{"users", "roles", "apps", "service_accounts"}

// ensurePermissionCatalog seeds read+write permissions for each control-plane
// resource type (idempotent).
func ensurePermissionCatalog(ctx context.Context, perms app.PermissionRepository) error {
	for _, rt := range controlPlaneResources {
		for _, verb := range []string{"read", "write"} {
			p := app.NewPermission(rt+"."+verb,
				app.WithPermissionResourceType(rt),
				app.WithPermissionVerb(verb),
				app.WithPermissionDescription(verb+" "+rt),
			)
			if err := perms.Upsert(ctx, p); err != nil {
				return err
			}
		}
	}
	for _, p := range platformPermissions {
		perm := app.NewPermission(p.id,
			app.WithPermissionResourceType(p.resourceType),
			app.WithPermissionVerb(p.verb),
			app.WithPermissionDescription(p.description),
		)
		if err := perms.Upsert(ctx, perm); err != nil {
			return err
		}
	}
	return nil
}

// platformPermissions are the fine-grained permissions gating the platform
// topology surface: read the graph, manage workloads, and the stronger
// node-level/destructive actions.
var platformPermissions = []struct {
	id, resourceType, verb, description string
}{
	{"platform.read", "platform", "read", "view the platform topology"},
	{"platform.manage", "platform", "manage", "restart, scale, pause and resume workloads"},
	{"platform:workload.delete", "platform:workload", "delete", "delete workloads"},
	{"platform:cluster.manage", "platform:cluster", "manage", "cordon, uncordon and drain nodes"},
}

// ensureAdmin find-or-creates the bootstrap user and returns its id.
func ensureAdmin(ctx context.Context, cfg bootstrapConfig, users app.UserRepository, creds app.CredentialRepository, hasher password.Hasher, log logger.Logger) (string, error) {
	email := strings.ToLower(strings.TrimSpace(cfg.adminEmail))
	if u, err := app.GetUserByEmail(ctx, users, email); err == nil && u != nil {
		return u.ID(), nil
	} else if err != nil && !apierrors.Is(err, apierrors.CodeNotFound) {
		return "", err
	}
	u, err := users.Create(ctx, app.NewUser(email,
		app.WithUserDisplayName(cfg.adminName),
		app.WithUserStatus(app.UserEnabled),
	))
	if err != nil {
		return "", err
	}
	hashed, err := hasher.Hash(cfg.adminPassword)
	if err != nil {
		return "", err
	}
	if err := creds.Set(ctx, u.ID(), hashed.Encoded); err != nil {
		return "", err
	}
	log.Info("bootstrap created admin user", "email", email)
	return u.ID(), nil
}

// ensureAdminRole seeds the SYSTEM roles (admin, viewer, editor) and binds the
// bootstrap user to admin (idempotent).
func ensureSuperadmin(ctx context.Context, roles app.RoleRepository, userID string, log logger.Logger) error {
	systemRoles := []app.Role{
		app.NewRole("admin", app.WithRoleName("Admin"), app.WithRoleKind(app.RoleSystem), app.WithRolePermissions([]string{"*.*"})),
		app.NewRole("viewer", app.WithRoleName("Viewer"), app.WithRoleKind(app.RoleSystem), app.WithRolePermissions([]string{"*.read"})),
		app.NewRole("editor", app.WithRoleName("Editor"), app.WithRoleKind(app.RoleSystem), app.WithRolePermissions([]string{"*.read", "*.write"})),
		app.NewRole("platform-operator", app.WithRoleName("Platform Operator"), app.WithRoleKind(app.RoleSystem), app.WithRolePermissions([]string{"platform.read", "platform.manage"})),
	}
	for _, role := range systemRoles {
		if _, err := roles.Create(ctx, role); err != nil {
			return err
		}
	}
	if err := roles.BindSubject(ctx, app.SubjectTypeUser, userID, "admin"); err != nil {
		return err
	}
	log.Info("bootstrap bound user to admin", "userId", userID)
	return nil
}

func ensureApps(ctx context.Context, cfg bootstrapConfig, apps app.AppRepository, perms app.PermissionRepository, log logger.Logger) error {
	for _, entry := range strings.Split(cfg.apps, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Entry is slug=Name=adminBaseURL[=moduleUri]. The 4th field
		// (moduleUri, the browser-reachable Module-Federation remote) is
		// optional — 3-field entries without a console remote stay valid.
		parts := strings.SplitN(entry, "=", 4)
		if len(parts) < 3 {
			log.Warn("bootstrap skipping malformed FOUNDRY_APPS entry", "entry", entry)
			continue
		}
		moduleURI := ""
		if len(parts) == 4 {
			moduleURI = parts[3]
		}
		a := app.NewApp(parts[0],
			app.WithAppName(parts[1]),
			app.WithAppKind(parts[0]),
			app.WithAppAdminBaseURL(parts[2]),
			app.WithAppModuleURI(moduleURI),
			app.WithAppEnabled(true),
		)
		if err := apps.Upsert(ctx, a); err != nil {
			return err
		}
		if err := ensureAppPermissions(ctx, perms, a.Slug()); err != nil {
			return err
		}
		log.Info("bootstrap registered app", "slug", a.Slug(), "adminBaseURL", a.AdminBaseURL(), "moduleUri", a.ModuleURI())
	}
	return nil
}

// ensureAppPermissions seeds the per-app read/write permissions so the role
// picker can list them (idempotent).
func ensureAppPermissions(ctx context.Context, perms app.PermissionRepository, slug string) error {
	rt := "app:" + slug
	for _, verb := range []string{"read", "write"} {
		p := app.NewPermission(rt+"."+verb,
			app.WithPermissionResourceType(rt),
			app.WithPermissionVerb(verb),
			app.WithPermissionDescription(verb+" app "+slug),
		)
		if err := perms.Upsert(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
