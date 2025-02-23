package test

import (
	"path/filepath"
	"runtime"
)

func GetSQLiteMigrationPath() string {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../..") // Adjust the number of "../" based on your file structure
	return filepath.Join(projectRoot, "migrations/test")
}
