package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/application/usecase"
	"github.com/fromforgesoftware/go-kit/auth/jwt"
	"github.com/fromforgesoftware/go-kit/auth/password"
	apierrors "github.com/fromforgesoftware/go-kit/errors"
	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypeServiceAccount resource.Type = "service-accounts"

type ServiceAccountStatus string

const (
	ServiceAccountEnabled  ServiceAccountStatus = "ENABLED"
	ServiceAccountDisabled ServiceAccountStatus = "DISABLED"
)

// ServiceAccount is a machine identity that authenticates via
// client-credentials and is granted app access through roles.
type ServiceAccount interface {
	resource.Resource
	Name() string
	ClientID() string
	Status() ServiceAccountStatus
	LastUsedAt() *time.Time
}

type serviceAccount struct {
	resource.Resource

	name       string
	clientID   string
	status     ServiceAccountStatus
	lastUsedAt *time.Time
}

type ServiceAccountOption func(*serviceAccount)

func WithServiceAccountID(id string) ServiceAccountOption {
	return func(sa *serviceAccount) { sa.Resource = resource.Update(sa.Resource, resource.WithID(id)) }
}

func WithServiceAccountClientID(clientID string) ServiceAccountOption {
	return func(sa *serviceAccount) { sa.clientID = clientID }
}

func WithServiceAccountStatus(status ServiceAccountStatus) ServiceAccountOption {
	return func(sa *serviceAccount) { sa.status = status }
}

func WithServiceAccountLastUsedAt(at *time.Time) ServiceAccountOption {
	return func(sa *serviceAccount) { sa.lastUsedAt = at }
}

// NewServiceAccount builds a ServiceAccount aggregate.
func NewServiceAccount(name string, opts ...ServiceAccountOption) ServiceAccount {
	sa := &serviceAccount{
		Resource: resource.New(resource.WithType(ResourceTypeServiceAccount)),
		name:     name,
		status:   ServiceAccountEnabled,
	}
	for _, opt := range opts {
		opt(sa)
	}
	return sa
}

func (sa *serviceAccount) Name() string                 { return sa.name }
func (sa *serviceAccount) ClientID() string             { return sa.clientID }
func (sa *serviceAccount) Status() ServiceAccountStatus { return sa.status }
func (sa *serviceAccount) LastUsedAt() *time.Time       { return sa.lastUsedAt }

// ServiceAccountCredentials is returned once at creation; the plaintext secret
// is shown to the caller exactly once and never persisted.
type ServiceAccountCredentials struct {
	ServiceAccount ServiceAccount
	ClientID       string
	ClientSecret   string
}

// ServiceAccountRepository persists service accounts via the kit generic
// surface plus the credential-specific lookups token exchange needs.
type ServiceAccountRepository interface {
	repository.Getter[ServiceAccount]
	repository.Lister[ServiceAccount]
	repository.Deleter

	// CreateWithSecret persists a new service account together with its secret hash.
	CreateWithSecret(ctx context.Context, sa ServiceAccount, secretHash string) (ServiceAccount, error)
	// GetByClientID returns the account plus its stored secret hash, for auth.
	GetByClientID(ctx context.Context, clientID string) (sa ServiceAccount, secretHash string, err error)
	TouchLastUsed(ctx context.Context, id string, at time.Time) error
}

// ServiceAccountUsecase manages service accounts and their client-credentials
// token exchange.
type ServiceAccountUsecase interface {
	repository.Getter[ServiceAccount]
	repository.Lister[ServiceAccount]
	repository.Deleter

	Create(ctx context.Context, name string) (ServiceAccountCredentials, error)
	// IssueToken exchanges client-credentials for a bearer JWT subject'd to the
	// service account id. expiresIn is in seconds.
	IssueToken(ctx context.Context, clientID, clientSecret string) (token string, expiresIn int, err error)
}

// serviceAccountTokenTTL mirrors the kit HMAC issuer's 24h token lifetime.
const serviceAccountTokenTTL = 24 * time.Hour

type serviceAccountUsecase struct {
	repository.Getter[ServiceAccount]
	repository.Lister[ServiceAccount]
	repository.Deleter

	repo   ServiceAccountRepository
	hasher password.Hasher
	issuer jwt.Issuer // nil when no token secret is configured
	now    func() time.Time
}

func NewServiceAccountUsecase(repo ServiceAccountRepository, hasher password.Hasher, issuer jwt.Issuer) ServiceAccountUsecase {
	return &serviceAccountUsecase{
		Getter:  usecase.NewGetter[ServiceAccount](repo, ResourceTypeServiceAccount),
		Lister:  usecase.NewLister[ServiceAccount](repo),
		Deleter: usecase.NewDeleter(repo),
		repo:    repo,
		hasher:  hasher,
		issuer:  issuer,
		now:     time.Now,
	}
}

func (uc *serviceAccountUsecase) Create(ctx context.Context, name string) (ServiceAccountCredentials, error) {
	clientID := "sa_" + uuid.NewString()
	secret, err := randomSecret()
	if err != nil {
		return ServiceAccountCredentials{}, apierrors.InternalError("could not generate secret")
	}
	hashed, err := uc.hasher.Hash(secret)
	if err != nil {
		return ServiceAccountCredentials{}, err
	}
	sa, err := uc.repo.CreateWithSecret(ctx, NewServiceAccount(name,
		WithServiceAccountClientID(clientID),
		WithServiceAccountStatus(ServiceAccountEnabled),
	), hashed.Encoded)
	if err != nil {
		return ServiceAccountCredentials{}, err
	}
	return ServiceAccountCredentials{ServiceAccount: sa, ClientID: clientID, ClientSecret: secret}, nil
}

func (uc *serviceAccountUsecase) IssueToken(ctx context.Context, clientID, clientSecret string) (string, int, error) {
	if uc.issuer == nil {
		return "", 0, apierrors.InternalError("token signing not configured")
	}
	sa, hash, err := uc.repo.GetByClientID(ctx, clientID)
	if err != nil {
		if apierrors.Is(err, apierrors.CodeNotFound) {
			return "", 0, invalidCredentials()
		}
		return "", 0, err
	}
	if sa == nil || sa.Status() != ServiceAccountEnabled {
		return "", 0, invalidCredentials()
	}
	ok, err := uc.hasher.Verify(clientSecret, hash)
	if err != nil || !ok {
		return "", 0, invalidCredentials()
	}
	id, err := uuid.Parse(sa.ID())
	if err != nil {
		return "", 0, apierrors.InternalError("invalid service account id")
	}
	token, err := uc.issuer.Issue(ctx, id, sa.ClientID())
	if err != nil {
		return "", 0, apierrors.InternalError("could not issue token")
	}
	if err := uc.repo.TouchLastUsed(ctx, sa.ID(), uc.now().UTC()); err != nil {
		return "", 0, err
	}
	return token, int(serviceAccountTokenTTL.Seconds()), nil
}

// randomSecret returns a 32-byte URL-safe random secret.
func randomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
