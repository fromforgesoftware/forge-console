package app

import "context"

// UserSettings holds per-user UI preferences. Kept minimal for now.
type UserSettings struct {
	UserID string
	Theme  string
}

// DefaultUserSettings is returned for a user with no stored row yet.
func DefaultUserSettings(userID string) UserSettings {
	return UserSettings{UserID: userID, Theme: "system"}
}

// SettingsRepository persists user settings.
type SettingsRepository interface {
	// Get returns the user's settings, or defaults when no row exists.
	Get(ctx context.Context, userID string) (UserSettings, error)
	Upsert(ctx context.Context, s UserSettings) error
}
