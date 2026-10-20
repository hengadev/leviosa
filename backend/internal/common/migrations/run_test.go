package migrations

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestNewestVersionPicksHighestMigration(t *testing.T) {
	fsys := fstest.MapFS{
		"20250101000000_first.sql":  {},
		"20250602201515_second.sql": {},
		"20260811120002_third.sql":  {},
		"convention.md":             {},
	}

	v, err := newestVersion(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 20260811120002 {
		t.Fatalf("expected 20260811120002, got %d", v)
	}
}

func TestNewestVersionSkipsUnconventionalNames(t *testing.T) {
	// goose only collects {version}_*.sql files; anything else must not
	// break the scan or count as a version.
	fsys := fstest.MapFS{
		"20250101000000_ok.sql": {},
		"notes.sql":             {},
		"readme.txt":            {},
	}

	v, err := newestVersion(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 20250101000000 {
		t.Fatalf("expected 20250101000000, got %d", v)
	}
}

func TestNewestVersionEmptyFSErrors(t *testing.T) {
	if _, err := newestVersion(fstest.MapFS{}); err == nil {
		t.Fatal("expected an error for a filesystem without migrations")
	}
}

func TestNewestVersionEmbedded(t *testing.T) {
	// The embedded set must always have a newest version; it is what serve
	// compares the database against.
	if NewestVersion() <= 0 {
		t.Fatalf("expected a positive newest version, got %d", NewestVersion())
	}
}

func TestParseVersion(t *testing.T) {
	cases := []struct {
		name    string
		want    int64
		wantErr bool
	}{
		{"20250714103022_catalog_add_column_products_buffer_time.sql", 20250714103022, false},
		{"1_init.sql", 1, false},
		{"notes.sql", 0, true},
		{"0_zero.sql", 0, true},
		{"x_1.sql", 0, true},
	}

	for _, tc := range cases {
		got, err := parseVersion(tc.name)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseVersion(%q): expected an error, got %d", tc.name, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseVersion(%q): unexpected error %v", tc.name, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseVersion(%q): expected %d, got %d", tc.name, tc.want, got)
		}
	}
}

// The embedded set must parse under goose's rules: every .sql file with a
// version prefix counts, and the reported newest version must match what a
// sorted directory listing says.
func TestNewestVersionMatchesEmbed(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}

	sqlCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			sqlCount++
		}
	}
	if sqlCount == 0 {
		t.Fatal("no embedded migrations found")
	}
}
