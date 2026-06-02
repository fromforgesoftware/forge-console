// Package internal wires the Foundry control-plane backend into one fx module
// that cmd/server composes alongside the kit's defaults.
package internal

import (
	"context"
	"os"

	"go.uber.org/fx"
	"k8s.io/client-go/kubernetes"

	"github.com/fromforgesoftware/go-kit/auth/jwt"
	"github.com/fromforgesoftware/go-kit/auth/password"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/forge/server/internal/app"
	"github.com/fromforgesoftware/forge/server/internal/db"
	"github.com/fromforgesoftware/forge/server/internal/k8s"
	foundryhttp "github.com/fromforgesoftware/forge/server/internal/transport/http"
)

const Version = "0.1.0"

func FxModule() fx.Option {
	return fx.Module("foundry",
		repositoriesFxModule(),
		usecasesFxModule(),
		transportFxModule(),
		fx.Invoke(registerBootstrap),
	)
}

func repositoriesFxModule() fx.Option {
	return fx.Module("foundry:repositories",
		fx.Provide(
			fx.Annotate(db.NewUserRepository, fx.As(new(app.UserRepository))),
			fx.Annotate(db.NewCredentialRepository, fx.As(new(app.CredentialRepository))),
			fx.Annotate(db.NewSessionRepository, fx.As(new(app.SessionRepository))),
			fx.Annotate(db.NewAppRepository, fx.As(new(app.AppRepository))),
			fx.Annotate(db.NewRoleRepository, fx.As(new(app.RoleRepository))),
			fx.Annotate(db.NewPermissionRepository, fx.As(new(app.PermissionRepository))),
			fx.Annotate(db.NewSettingsRepository, fx.As(new(app.SettingsRepository))),
			fx.Annotate(db.NewServiceAccountRepository, fx.As(new(app.ServiceAccountRepository))),
			k8s.NewClientSet,
			fx.Annotate(k8s.NewTopologyRepository, fx.As(new(app.TopologyRepository))),
			fx.Annotate(k8s.NewActions, fx.As(new(app.ClusterActions))),
			fx.Annotate(newTopologyWatcher,
				fx.As(new(app.TopologyWatcher)),
				fx.As(new(app.LogStreamer)),
			),
		),
	)
}

// newTopologyWatcher builds the informer-driven watcher and binds its lifecycle
// to the app: informers start on boot and stop on shutdown. It is exposed both
// as the change notifier and the pod-log streamer.
func newTopologyWatcher(lc fx.Lifecycle, cs kubernetes.Interface) *k8s.Watcher {
	w := k8s.NewWatcher(cs)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { w.Start(); return nil },
		OnStop:  func(context.Context) error { w.Stop(); return nil },
	})
	return w
}

// newServiceAccountTokenIssuer reuses the gateway's HMAC secret (FOUNDRY_TOKEN_SECRET
// falls back to FORGE_GATEWAY_SECRET) so one secret governs Foundry-issued tokens.
// nil when no secret is set — SA token issuance then returns a clear error.
func newServiceAccountTokenIssuer() jwt.Issuer {
	secret := os.Getenv("FOUNDRY_TOKEN_SECRET")
	if secret == "" {
		secret = os.Getenv("FORGE_GATEWAY_SECRET")
	}
	if secret == "" {
		return nil
	}
	iss, err := jwt.NewHMACIssuer(secret)
	if err != nil {
		return nil
	}
	return iss
}

func usecasesFxModule() fx.Option {
	return fx.Module("foundry:usecases",
		fx.Provide(
			fx.Annotate(password.NewArgon2id, fx.As(new(password.Hasher))),
			fx.Annotate(app.NewAuthUsecase, fx.As(new(app.AuthUsecase))),
			fx.Annotate(app.NewAccountUsecase, fx.As(new(app.AccountUsecase))),
			fx.Annotate(app.NewAppUsecase, fx.As(new(app.AppUsecase))),
			fx.Annotate(app.NewAppAdminUsecase, fx.As(new(app.AppAdminUsecase))),
			fx.Annotate(app.NewAuthzUsecase, fx.As(new(app.AuthzUsecase))),
			fx.Annotate(app.NewUserUsecase, fx.As(new(app.UserUsecase))),
			fx.Annotate(app.NewRoleUsecase, fx.As(new(app.RoleUsecase))),
			fx.Annotate(app.NewPermissionUsecase, fx.As(new(app.PermissionUsecase))),
			newServiceAccountTokenIssuer,
			fx.Annotate(app.NewServiceAccountUsecase, fx.As(new(app.ServiceAccountUsecase))),
			app.NewLogActionAuditor,
			fx.Annotate(app.NewTopologyUsecase, fx.As(new(app.TopologyUsecase))),
			newOIDCProviders,
		),
	)
}

func transportFxModule() fx.Option {
	return fx.Module("foundry:transport",
		kitrest.NewFxController(foundryhttp.NewAuthController),
		kitrest.NewFxController(foundryhttp.NewAccountController),
		kitrest.NewFxController(foundryhttp.NewAdminController),
		kitrest.NewFxController(foundryhttp.NewServiceAccountController),
		kitrest.NewFxController(foundryhttp.NewOIDCController),
		kitrest.NewFxController(foundryhttp.NewAppController),
		kitrest.NewFxController(foundryhttp.NewGatewayController),
		kitrest.NewFxController(foundryhttp.NewTopologyController),
		kitrest.NewFxController(foundryhttp.NewPlatformStreamController),
		kitrest.NewFxController(foundryhttp.NewSPAController),
	)
}
