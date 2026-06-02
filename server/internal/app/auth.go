package app

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
)

// sessionTTL is how long a Foundry user session stays valid.
const sessionTTL = 12 * time.Hour

// AuthUsecase authenticates users (local password for now) and resolves
// sessions. OIDC providers plug in later via the same session issuance.
type AuthUsecase interface {
	Login(ctx context.Context, email, plaintext string) (Session, error)
	Logout(ctx context.Context, sessionID string) error
	Authenticate(ctx context.Context, sessionID string) (User, error)
	// StartSession mints a session for an already-identified user (used by
	// the OIDC callback after the external IdP verifies the identity).
	StartSession(ctx context.Context, userID string) (Session, error)
}

type authUsecase struct {
	users    UserRepository
	creds    CredentialRepository
	sessions SessionRepository
	hasher   password.Hasher
	now      func() time.Time
}

func NewAuthUsecase(users UserRepository, creds CredentialRepository, sessions SessionRepository, hasher password.Hasher) AuthUsecase {
	return &authUsecase{users: users, creds: creds, sessions: sessions, hasher: hasher, now: time.Now}
}

// invalidCredentials is returned uniformly for unknown email and wrong password
// so login can't be used to enumerate users.
func invalidCredentials() error { return apierrors.Unauthenticated("invalid credentials") }

func (uc *authUsecase) Login(ctx context.Context, email, plaintext string) (Session, error) {
	u, err := GetUserByEmail(ctx, uc.users, normalizeEmail(email))
	if err != nil {
		if apierrors.Is(err, apierrors.CodeNotFound) {
			return Session{}, invalidCredentials()
		}
		return Session{}, err
	}
	if u == nil || u.Status() != UserEnabled {
		return Session{}, invalidCredentials()
	}
	hash, err := uc.creds.HashFor(ctx, u.ID())
	if err != nil {
		if apierrors.Is(err, apierrors.CodeNotFound) {
			return Session{}, invalidCredentials()
		}
		return Session{}, err
	}
	ok, err := uc.hasher.Verify(plaintext, hash)
	if err != nil || !ok {
		return Session{}, invalidCredentials()
	}
	return uc.sessions.Create(ctx, Session{
		UserID:    u.ID(),
		ExpiresAt: uc.now().UTC().Add(sessionTTL),
	})
}

func (uc *authUsecase) StartSession(ctx context.Context, userID string) (Session, error) {
	return uc.sessions.Create(ctx, Session{
		UserID:    userID,
		ExpiresAt: uc.now().UTC().Add(sessionTTL),
	})
}

func (uc *authUsecase) Logout(ctx context.Context, sessionID string) error {
	return uc.sessions.Delete(ctx, sessionID)
}

func (uc *authUsecase) Authenticate(ctx context.Context, sessionID string) (User, error) {
	s, err := uc.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, apierrors.Unauthenticated("no active session")
	}
	if !uc.now().UTC().Before(s.ExpiresAt) {
		return nil, apierrors.Unauthenticated("session expired")
	}
	return getUserByID(ctx, uc.users, s.UserID)
}
