package ports

import (
	"context"
	"io"

	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
)

type SettingsRepository interface {
	// non encrypted
	GetString(ctx context.Context, key string) (*domain.Setting[string], error)
	GetInt(ctx context.Context, key string) (*domain.Setting[int], error)
	SetString(ctx context.Context, setting *domain.Setting[string]) error
	SetInt(ctx context.Context, setting *domain.Setting[int]) error
	// encrypted - using generated ENCX structs
	GetEncryptedSetting(ctx context.Context, key string) (*domain.SettingEncryptedEncx, error)
	SetEncryptedSetting(ctx context.Context, settingEncx *domain.SettingEncryptedEncx) error
}

type SettingsMedia interface {
	GetLogo(ctx context.Context) ([]byte, error)
	UploadLogo(ctx context.Context, key string, file io.Reader, size int64, contentType string) (string, error)
	// DeleteLogo(ctx context.Context)

	// Uploaded media (the practice page video). PutObject needs a seekable
	// body for the request signature. GetObject reads the byte range of an
	// HTTP Range header value, or the whole object when it is empty; a key
	// that does not exist is a not-found error, a range outside the object
	// domain.ErrMediaRangeNotSatisfiable.
	PutObject(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error
	GetObject(ctx context.Context, key, rangeHeader string) (*domain.MediaObject, error)
	DeleteObject(ctx context.Context, key string) error
}
