package domain

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

// The practice page video, uploaded by an administrator (instead of a link to
// a video hosted elsewhere). It lives in the private media bucket under
// PracticePageVideoKeyPrefix, one object per upload, and visitors fetch it
// through the backend's public media route, which streams it from the bucket.
// The practice page video setting holds its public path (MediaPathPrefix +
// key), so the page itself does not change.
const (
	// PracticePageVideoMaxBytes caps an upload: enough for a 20–40 s 1080p
	// loop, small enough to load on a phone. There is no transcoding.
	PracticePageVideoMaxBytes int64 = 50 << 20

	PracticePageVideoKeyPrefix = "practice-page/video/"

	// MediaPathPrefix is the public path media objects are served under.
	MediaPathPrefix = "/media/"
)

// ErrMediaRangeNotSatisfiable is returned when a requested byte range lies
// outside the media object.
var ErrMediaRangeNotSatisfiable = errors.New("requested range not satisfiable")

// practicePageVideoTypes are the formats every browser plays in a muted,
// looping <video>, with the extension their objects get.
var practicePageVideoTypes = map[string]string{
	"video/mp4":  "mp4",
	"video/webm": "webm",
}

// PracticePageVideoExtension returns the object extension of an accepted
// video content type, or false for any other type.
func PracticePageVideoExtension(contentType string) (string, bool) {
	ext, ok := practicePageVideoTypes[contentType]
	return ext, ok
}

// NewPracticePageVideoKey returns a fresh object key: each upload gets its
// own, so its URL can be cached forever.
func NewPracticePageVideoKey(ext string) string {
	return PracticePageVideoKeyPrefix + uuid.NewString() + "." + ext
}

// IsPracticePageVideoKey reports whether key names an uploaded practice page
// video: the prefix, a UUID and an accepted extension, nothing else. The
// public media route serves only such keys.
func IsPracticePageVideoKey(key string) bool {
	name, ok := strings.CutPrefix(key, PracticePageVideoKeyPrefix)
	if !ok {
		return false
	}
	id, ext, ok := strings.Cut(name, ".")
	if !ok {
		return false
	}
	if _, err := uuid.Parse(id); err != nil || len(id) != 36 {
		return false
	}
	for _, accepted := range practicePageVideoTypes {
		if ext == accepted {
			return true
		}
	}
	return false
}

// PracticePageVideoURL is the public path of an uploaded video.
func PracticePageVideoURL(key string) string {
	return MediaPathPrefix + key
}

// PracticePageVideoKeyFromURL returns the object key behind a practice page
// video setting value, when that value is an uploaded video (and not empty or
// a link to a video hosted elsewhere).
func PracticePageVideoKeyFromURL(value string) (string, bool) {
	key, ok := strings.CutPrefix(value, MediaPathPrefix)
	if !ok || !IsPracticePageVideoKey(key) {
		return "", false
	}
	return key, true
}

// MediaObject is a media file, or the requested byte range of it, read from
// the bucket. The caller closes Body.
type MediaObject struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	// ContentRange is set when only a range was requested ("bytes 0-99/1234").
	ContentRange string
	ETag         string
	LastModified *time.Time
}

type UploadPracticePageVideoResponse struct {
	VideoURL string `json:"video_url"`
}
