package app

import (
	"context"
	"sort"
	"strings"
)

// AuthzUsecase answers authorization questions from a subject's role bindings,
// matching a requested action against the granted permission patterns.
type AuthzUsecase interface {
	// Can reports whether the subject holds a permission pattern matching action.
	Can(ctx context.Context, subjectType SubjectType, subjectID, action string) (bool, error)
	// EffectivePermissions returns the sorted unique union of granted patterns.
	EffectivePermissions(ctx context.Context, subjectType SubjectType, subjectID string) ([]string, error)

	// IsAdmin reports whether the user is granted everything ("*.*").
	IsAdmin(ctx context.Context, userID string) (bool, error)
	// CanAccessApp reports whether the user may read the app (app:<slug>.read).
	CanAccessApp(ctx context.Context, userID, appSlug string) (bool, error)
	// CanServiceAccountAccessApp mirrors CanAccessApp for a service account.
	CanServiceAccountAccessApp(ctx context.Context, saID, appSlug string) (bool, error)
}

type authzUsecase struct {
	roles RoleRepository
}

func NewAuthzUsecase(roles RoleRepository) AuthzUsecase {
	return &authzUsecase{roles: roles}
}

func (uc *authzUsecase) granted(ctx context.Context, subjectType SubjectType, subjectID string) ([]string, error) {
	roles, err := uc.roles.RolesForSubject(ctx, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range roles {
		out = append(out, r.Permissions()...)
	}
	return out, nil
}

func (uc *authzUsecase) Can(ctx context.Context, subjectType SubjectType, subjectID, action string) (bool, error) {
	granted, err := uc.granted(ctx, subjectType, subjectID)
	if err != nil {
		return false, err
	}
	for _, p := range granted {
		if permissionMatches(p, action) {
			return true, nil
		}
	}
	return false, nil
}

func (uc *authzUsecase) EffectivePermissions(ctx context.Context, subjectType SubjectType, subjectID string) ([]string, error) {
	granted, err := uc.granted(ctx, subjectType, subjectID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(granted))
	for _, p := range granted {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (uc *authzUsecase) IsAdmin(ctx context.Context, userID string) (bool, error) {
	return uc.Can(ctx, SubjectTypeUser, userID, "*.*")
}

func (uc *authzUsecase) CanAccessApp(ctx context.Context, userID, appSlug string) (bool, error) {
	return uc.Can(ctx, SubjectTypeUser, userID, "app:"+appSlug+".read")
}

func (uc *authzUsecase) CanServiceAccountAccessApp(ctx context.Context, saID, appSlug string) (bool, error) {
	return uc.Can(ctx, SubjectTypeServiceAccount, saID, "app:"+appSlug+".read")
}

// permissionMatches reports whether a granted pattern authorizes action. Both
// are split on the LAST dot into resourceType + verb (resourceType may itself
// contain a colon, e.g. "app:aegis.read"). A "*" in either segment is a
// wildcard for that segment; a bare "*" pattern matches everything.
func permissionMatches(pattern, action string) bool {
	if pattern == "*" {
		return true
	}
	pRT, pVerb := splitPermission(pattern)
	aRT, aVerb := splitPermission(action)
	return segmentMatches(pRT, aRT) && segmentMatches(pVerb, aVerb)
}

func splitPermission(s string) (resourceType, verb string) {
	i := strings.LastIndex(s, ".")
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func segmentMatches(pattern, value string) bool {
	return pattern == "*" || pattern == value
}
