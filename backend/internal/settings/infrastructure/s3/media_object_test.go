package media_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaObject(t *testing.T) {
	ctx := context.Background()
	td.ClearBucket(t, ctx, s3Client)

	data := []byte("0123456789abcdefghij")
	key := domain.NewPracticePageVideoKey("mp4")
	require.NoError(t, repo.PutObject(ctx, key, bytes.NewReader(data), int64(len(data)), "video/mp4"))

	t.Run("reads the whole object", func(t *testing.T) {
		object, err := repo.GetObject(ctx, key, "")
		require.NoError(t, err)
		defer object.Body.Close()
		body, _ := io.ReadAll(object.Body)
		assert.Equal(t, data, body)
		assert.Equal(t, "video/mp4", object.ContentType)
		assert.Equal(t, int64(len(data)), object.ContentLength)
		assert.Empty(t, object.ContentRange)
		assert.NotEmpty(t, object.ETag)
	})

	t.Run("reads a byte range", func(t *testing.T) {
		object, err := repo.GetObject(ctx, key, "bytes=2-5")
		require.NoError(t, err)
		defer object.Body.Close()
		body, _ := io.ReadAll(object.Body)
		assert.Equal(t, []byte("2345"), body)
		assert.Equal(t, int64(4), object.ContentLength)
		assert.Equal(t, "bytes 2-5/20", object.ContentRange)
	})

	t.Run("a range past the end is not satisfiable", func(t *testing.T) {
		_, err := repo.GetObject(ctx, key, "bytes=100-200")
		assert.ErrorIs(t, err, domain.ErrMediaRangeNotSatisfiable)
	})

	t.Run("a deleted object is not found", func(t *testing.T) {
		require.NoError(t, repo.DeleteObject(ctx, key))
		_, err := repo.GetObject(ctx, key, "")
		assert.ErrorIs(t, err, errs.ErrRepositoryNotFound)
	})
}
