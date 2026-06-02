package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/fx"
	"sigs.k8s.io/yaml"

	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/monitoring/logger"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

const (
	// defaultAppsConfigPath is where the forge-apps ConfigMap's apps.yaml is
	// mounted by the Helm subchart.
	defaultAppsConfigPath = "/etc/forge/apps.yaml"
)

// appsConfig is the parsed forge-apps ConfigMap (apps.yaml). It is the sole
// app registry source — install[] lists the console plugin bundles fetched and
// served by the init-container, enable[] drives GET /apps (visibility +
// apiBase). moduleUri is DERIVED (not configured): an app that is both enabled
// and installed serves its SystemJS module at /public/plugins/<id>/module.js.
type appsConfig struct {
	Install []installEntry `json:"install"`
	Enable  []enableEntry  `json:"enable"`
}

// installEntry is a console plugin bundle the init-container pulls (oci://…).
type installEntry struct {
	ID     string `json:"id"`
	Bundle string `json:"bundle"`
}

// enableEntry makes an app visible in GET /apps. apiBase is the gateway-internal
// admin API; name is optional and defaults from the id.
type enableEntry struct {
	ID      string `json:"id"`
	APIBase string `json:"apiBase"`
	Name    string `json:"name"`
}

// bootstrapConfig seeds the first user + the managed app registry,
// idempotently every boot. The admin user comes from the environment
// (Keycloak KEYCLOAK_ADMIN-style); the app registry comes from the mounted
// forge-apps ConfigMap (apps.yaml).
type bootstrapConfig struct {
	adminEmail    string
	adminPassword string
	adminName     string
	apps          appsConfig
}

func newBootstrapConfig() bootstrapConfig {
	return bootstrapConfig{
		adminEmail:    os.Getenv("FOUNDRY_BOOTSTRAP_ADMIN_EMAIL"),
		adminPassword: os.Getenv("FOUNDRY_BOOTSTRAP_ADMIN_PASSWORD"),
		adminName:     envOr("FOUNDRY_BOOTSTRAP_ADMIN_NAME", "Administrator"),
		apps:          loadAppsConfig(envOr("FORGE_APPS_CONFIG", defaultAppsConfigPath)),
	}
}

// loadAppsConfig reads + parses the forge-apps ConfigMap apps.yaml. A missing
// or empty file is not an error — the server still starts, just with no apps
// (standalone subchart / no console.install).
func loadAppsConfig(path string) appsConfig {
	raw, err := os.ReadFile(path)
	if err != nil {
		return appsConfig{}
	}
	var cfg appsConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return appsConfig{}
	}
	return cfg
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

// ensureApps seeds the registry from the forge-apps ConfigMap's enable[] list.
// Each enabled app's moduleUri is DERIVED: an app that is also installed (has
// an install[] bundle) serves its SystemJS module at
// /public/plugins/<id>/module.js (the init-container unpacks it there);
// enabled-but-not-installed apps get an empty moduleUri so the console falls
// back to a bundled module.
func ensureApps(ctx context.Context, cfg bootstrapConfig, apps app.AppRepository, perms app.PermissionRepository, log logger.Logger) error {
	installed := map[string]bool{}
	for _, in := range cfg.apps.Install {
		if id := strings.TrimSpace(in.ID); id != "" {
			installed[id] = true
		}
	}
	for _, en := range cfg.apps.Enable {
		id := strings.TrimSpace(en.ID)
		if id == "" {
			log.Warn("bootstrap skipping enable entry with empty id")
			continue
		}
		name := strings.TrimSpace(en.Name)
		if name == "" {
			name = id
		}
		a := app.NewApp(id,
			app.WithAppName(name),
			app.WithAppKind(id),
			app.WithAppAdminBaseURL(strings.TrimSpace(en.APIBase)),
			app.WithAppModuleURI(deriveModuleURI(id, installed[id])),
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

// deriveModuleURI returns the browser-reachable SystemJS module path for an app
// that is both enabled and installed — the forge server serves the unpacked
// bundle there (Grafana-style). Not installed → empty (host falls back to a
// bundled module).
func deriveModuleURI(id string, installed bool) string {
	if !installed {
		return ""
	}
	return fmt.Sprintf("/public/plugins/%s/module.js", id)
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
