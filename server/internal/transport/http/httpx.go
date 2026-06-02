// Package http serves Foundry's control-plane API: user auth (plain JSON +
// cookie session, like Aegis's protocol controllers) and the app registry.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fromforgesoftware/go-kit/auth/jwt"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/jsonapi"
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/api"
	"github.com/fromforgesoftware/forge/server/internal/app"
)

const contentTypeJSONAPI = "application/vnd.api+json; charset=utf-8"

// writeOneJSONAPI emits a single resource DTO as a JSON:API document.
func writeOneJSONAPI(w http.ResponseWriter, status int, model any) {
	w.Header().Set("Content-Type", contentTypeJSONAPI)
	w.WriteHeader(status)
	_ = jsonapi.MarshalPayload(w, model)
}

// writeManyJSONAPI emits a slice of domain values as a JSON:API collection via
// their resource DTO mapper.
func writeManyJSONAPI[R any, DTO any](w http.ResponseWriter, items []R, toDTO func(R) DTO) {
	list := resource.NewListResponse(items, len(items))
	doc := resource.ListResponseToDTO(toDTO)(list)
	w.Header().Set("Content-Type", contentTypeJSONAPI)
	w.WriteHeader(http.StatusOK)
	_ = jsonapi.MarshalManyPayloads(w, doc)
}

// writeRolesJSONAPI emits a list of roles as a JSON:API collection — used by the
// non-CRUD subject-role listing endpoints so their wire shape matches the role
// resource handlers.
func writeRolesJSONAPI(w http.ResponseWriter, roles []app.Role) {
	writeManyJSONAPI(w, roles, api.RoleToDTO)
}

const sessionCookie = "foundry_session"

// secureRequest reports whether the request reached us over HTTPS (directly or
// via a TLS-terminating proxy). Cookies are marked Secure only then — on plain
// http (local dev) a Secure cookie isn't stored/sent, breaking session refresh.
func secureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	status := apierrors.GetHTTPStatus(err)
	if status == 0 {
		status = http.StatusInternalServerError
	}
	msg := "error"
	if status < http.StatusInternalServerError {
		msg = err.Error()
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

// resolveUser authenticates the request from its session cookie. On failure
// it writes the error response and returns ok=false so the caller can return.
func resolveUser(w http.ResponseWriter, r *http.Request, auth app.AuthUsecase) (app.User, bool) {
	ck, err := r.Cookie(sessionCookie)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "no session"})
		return nil, false
	}
	u, err := auth.Authenticate(r.Context(), ck.Value)
	if err != nil {
		writeErr(w, err)
		return nil, false
	}
	return u, true
}

// userRolesAndPermissions loads a user's bound role slugs and effective
// permission patterns (the SPA hydration payload).
func userRolesAndPermissions(ctx context.Context, authz app.AuthzUsecase, roles app.RoleRepository, userID string) (roleSlugs, permissions []string, err error) {
	permissions, err = authz.EffectivePermissions(ctx, app.SubjectTypeUser, userID)
	if err != nil {
		return nil, nil, err
	}
	roleRecords, err := roles.RolesForSubject(ctx, app.SubjectTypeUser, userID)
	if err != nil {
		return nil, nil, err
	}
	roleSlugs = make([]string, len(roleRecords))
	for i, role := range roleRecords {
		roleSlugs[i] = role.Slug()
	}
	return roleSlugs, permissions, nil
}

// guard wraps a JSON:API handler with the session + permission check, so the
// resource handlers stay pure kit generics while the gating stays uniform.
func guard(auth app.AuthUsecase, authz app.AuthzUsecase, action string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := requirePermission(w, r, auth, authz, action)
		if !ok {
			return
		}
		next.ServeHTTP(w, r.WithContext(app.WithActor(r.Context(), u.ID())))
	})
}

// requirePermission authenticates the request and ensures the user is granted
// action (a "<resourceType>.<verb>" permission). On failure it writes the
// response and returns ok=false so the caller can return.
func requirePermission(w http.ResponseWriter, r *http.Request, auth app.AuthUsecase, authz app.AuthzUsecase, action string) (app.User, bool) {
	u, ok := resolveUser(w, r, auth)
	if !ok {
		return nil, false
	}
	allowed, err := authz.Can(r.Context(), app.SubjectTypeUser, u.ID(), action)
	if err != nil {
		writeErr(w, err)
		return nil, false
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return nil, false
	}
	return u, true
}

// SubjectKind distinguishes a console user from a machine identity.
type SubjectKind string

const (
	SubjectUser           SubjectKind = "USER"
	SubjectServiceAccount SubjectKind = "SERVICE_ACCOUNT"
)

// Subject is the authenticated caller: either a console user (session cookie)
// or a service account (bearer JWT minted by the SA token endpoint).
type Subject struct {
	ID   string
	Kind SubjectKind
}

// resolveSubject authenticates the request. It prefers the session cookie
// (→ USER); otherwise it validates a Bearer JWT against the service-account
// issuer (→ SERVICE_ACCOUNT). On failure it writes the response and returns
// ok=false. saIssuer may be nil when no token secret is configured.
func resolveSubject(w http.ResponseWriter, r *http.Request, auth app.AuthUsecase, saIssuer jwt.Validator) (Subject, bool) {
	if ck, err := r.Cookie(sessionCookie); err == nil {
		u, err := auth.Authenticate(r.Context(), ck.Value)
		if err != nil {
			writeErr(w, err)
			return Subject{}, false
		}
		return Subject{ID: u.ID(), Kind: SubjectUser}, true
	}
	token := bearerToken(r)
	if token != "" && saIssuer != nil {
		claims, err := saIssuer.Validate(r.Context(), token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return Subject{}, false
		}
		return Subject{ID: claims.AccountID.String(), Kind: SubjectServiceAccount}, true
	}
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "no session"})
	return Subject{}, false
}

// bearerToken extracts the token from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// setSessionCookie sets/clears the httpOnly session cookie. maxAge<0 clears.
func setSessionCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}
