package db

import (
	"context"
	"errors"
	"time"

	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/postgres"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	"gorm.io/gorm"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

var settingsFieldMapping = map[string]string{
	"userId": "user_id",
	"theme":  "theme",
}

type settingsEntity struct {
	EUserID    string    `gorm:"column:user_id;type:uuid;primaryKey"`
	ECreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	EUpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`
	ETheme     string    `gorm:"column:theme"`
}

func (settingsEntity) TableName() string { return "foundry.user_settings" }

type settingsRepo struct{ *postgres.Repo }

func NewSettingsRepository(db *gormdb.DBClient) (*settingsRepo, error) {
	r, err := postgres.NewRepo(db, settingsFieldMapping)
	if err != nil {
		return nil, err
	}
	return &settingsRepo{Repo: r}, nil
}

// Get returns the user's settings, falling back to defaults when no row
// exists yet (settings are created lazily on first write).
func (r *settingsRepo) Get(ctx context.Context, userID string) (app.UserSettings, error) {
	s := search.New(search.WithQueryOpts(query.FilterBy(filter.OpEq, "userId", userID)))
	var e settingsEntity
	if err := r.QueryApply(ctx, s.Query()).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app.DefaultUserSettings(userID), nil
		}
		return app.UserSettings{}, postgres.NewErrUnknown(err)
	}
	return app.UserSettings{UserID: e.EUserID, Theme: e.ETheme}, nil
}

func (r *settingsRepo) Upsert(ctx context.Context, s app.UserSettings) error {
	res := r.DB.WithContext(ctx).Exec(
		`INSERT INTO foundry.user_settings (user_id, theme) VALUES (?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET theme = EXCLUDED.theme, updated_at = now()`,
		s.UserID, s.Theme,
	)
	if res.Error != nil {
		return postgres.NewErrUnknown(res.Error)
	}
	return nil
}
