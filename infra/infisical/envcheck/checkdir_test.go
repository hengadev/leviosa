package envcheck

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CheckDir is what the operator runs on real renderings (check-env.sh): it
// lists key names and must never print a value.
func TestCheckDirListsKeyNamesOnly(t *testing.T) {
	data := loadFixtures(t)
	for _, env := range Environments {
		dir := t.TempDir()
		names, _ := Files(env)
		for _, name := range names {
			content, _ := (&fakeInfisical{data: data}).render(t, name, env)
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		var out bytes.Buffer
		ok, err := CheckDir(env, dir, &out)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v\n%s", env, ok, err, out.String())
		}
		for _, name := range names {
			for _, k := range wantKeys[env][name] {
				if !strings.Contains(out.String(), "  "+k+"\n") {
					t.Errorf("%s: %s not listed", env, k)
				}
			}
		}
		assertNoValue(t, data, out.String())
	}
}

func TestCheckDirReportsMissingAndBrokenFiles(t *testing.T) {
	data := loadFixtures(t)
	dir := t.TempDir()
	content, _ := (&fakeInfisical{data: data}).render(t, "frontend.env", "staging")
	content += "AWS_SECRET_ACCESS_KEY='fixture/staging+aws-secret'\n"
	if err := os.WriteFile(filepath.Join(dir, "frontend.env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ok, err := CheckDir("staging", dir, &out)
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	for _, want := range []string{"backend.env (0 keys)\n  PROBLEM: not rendered", "PROBLEM: AWS_SECRET_ACCESS_KEY must not be in frontend.env", "problem(s)"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
	assertNoValue(t, data, out.String())
	if _, err := CheckDir("production", dir, &out); err == nil {
		t.Error("unknown environment accepted")
	}
}

// assertNoValue: no fixture value longer than a few characters appears in
// out (short ones like "0" or "587" would match key names or counts; host
// names are service names like "postgres", also a file name).
func assertNoValue(t *testing.T, data folders, out string) {
	t.Helper()
	for env, byFolder := range data {
		for folder, kv := range byFolder {
			for k, v := range kv {
				if len(v) >= 8 && !strings.HasSuffix(k, "_HOST") && strings.Contains(out, v) {
					t.Errorf("output contains the value of %s%s %s", env, folder, k)
				}
			}
		}
	}
	if strings.Contains(out, "='") {
		t.Errorf("output contains a KEY='value' line")
	}
}
