// Package app holds the Foundry control-plane usecases and domain types.
// Foundry's accounts are users only (console/app administrators),
// never application end-users — identity is global, with no realm.
package app

import (
	"context"
	"time"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
)

const ResourceTypeUser resource.Type = "users"

type UserStatus string

const (
	UserEnabled  UserStatus = "ENABLED"
	UserDisabled UserStatus = "DISABLED"
)

// User is a console administrator.
type User interface {
	resource.Resource
	Email() string
	DisplayName() string
	Status() UserStatus
}

type user struct {
	resource.Resource

	email       string
	displayName string
	status      UserStatus
}

type UserOption func(*user)

func WithUserID(id string) UserOption {
	return func(u *user) { u.Resource = resource.Update(u.Resource, resource.WithID(id)) }
}

func WithUserDisplayName(name string) UserOption {
	return func(u *user) { u.displayName = name }
}

func WithUserStatus(status UserStatus) UserOption {
	return func(u *user) { u.status = status }
}

// NewUser builds a User aggregate. email is the immutable login identity.
func NewUser(email string, opts ...UserOption) User {
	u := &user{
		Resource: resource.New(resource.WithType(ResourceTypeUser)),
		email:    email,
		status:   UserEnabled,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

func (u *user) Email() string       { return u.email }
func (u *user) DisplayName() string { return u.displayName }
func (u *user) Status() UserStatus  { return u.status }

// UserRepository persists users via the kit generic surface.
type UserRepository interface {
	repository.Getter[User]
	repository.Lister[User]
	repository.Creator[User]
	repository.Patcher[User]
}

// GetUserByEmail loads a single user by email through the generic Getter.
func GetUserByEmail(ctx context.Context, repo repository.Getter[User], email string) (User, error) {
	return repo.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "email", email)))
}

// getUserByID loads a single user by id through the generic Getter.
func getUserByID(ctx context.Context, repo repository.Getter[User], id string) (User, error) {
	return repo.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, "id", id)))
}

// CredentialRepository persists the local argon2 password hash per user.
type CredentialRepository interface {
	HashFor(ctx context.Context, userID string) (string, error)
	Set(ctx context.Context, userID, hash string) error
}

// Session is a server-side browser session.
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

// SessionRepository persists sessions.
type SessionRepository interface {
	Create(ctx context.Context, s Session) (Session, error)
	Get(ctx context.Context, id string) (Session, error)
	Delete(ctx context.Context, id string) error
}
