package test

import (
	"context"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"

	userService "github.com/GaryHY/leviosa/internal/domain/user"
	userRepository "github.com/GaryHY/leviosa/internal/repository/sqlite/user"
	"github.com/GaryHY/leviosa/internal/server/app"
	userHandler "github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/pkg/config"

	testdb "github.com/GaryHY/leviosa/pkg/sqliteutil/testdatabase"
)

// TODO: handle the different ways to import the different domain
// - use the repoconstructor thing
// - return the repository interface that implements the GetDB(0) function
func Setup(
	t testing.TB,
	ctx context.Context,
	version int64,
	config *config.Config,
) *userHandler.AppInstance {
	t.Helper()
	sqlitedb, err := testdb.NewDatabase(ctx)
	if err != nil {
		t.Error(err)
	}
	if err := testdb.Setup(ctx, sqlitedb, version); err != nil {
		t.Error(err)
	}
	userRepo := userRepository.New(ctx, sqlitedb)
	userService := userService.New(userRepo, config.GetSecurity())
	appsvc := app.Services{User: userService}
	// apprepo := handler.Repos{User: readerRepo}
	apprepo := app.Repos{User: userRepo}
	h := app.New(&appsvc, &apprepo)
	return userHandler.New(h)
}

// NOTE: link for the number generator : https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func GetSQLiteMigrationPath() string {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../..") // Adjust the number of "../" based on your file structure
	return filepath.Join(projectRoot, "migrations/test")
}
