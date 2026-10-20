package envcheck

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// make-env.sh with a stub `infisical` (never the real CLI: PATH holds only
// the stub and the few tools the script needs), run from a copy of the repo
// layout so the real backend/development.env and frontend/.env are never
// touched.

const stubCLI = `#!/bin/sh
if [ "$1 $2 $3" = "user get token" ]; then
	[ -n "${STUB_LOGGED_IN:-}" ] && echo stub-session-token
	exit 0
fi
if [ "$1" = export ]; then
	t=
	for a; do case $a in --template=*) t=${a#--template=} ;; esac; done
	grep -q __LEVIOSA_ "$t" && { echo "stub: unsubstituted template" >&2; exit 3; }
	for f in "$STUB_DIR"/*.env; do
		name=${f##*/}
		if grep -q "^# leviosa $name - rendered from Infisical" "$t"; then
			grep -q "\"${STUB_ENV:-dev}\" \"/" "$t" || { echo "stub: not rendered for ${STUB_ENV:-dev}" >&2; exit 6; }
			[ "${STUB_FAIL:-}" = "$name" ] && { echo "stub: export failed" >&2; exit 1; }
			cat "$f"
			exit 0
		fi
	done
	echo "stub: unexpected template" >&2
	exit 4
fi
exit 5
`

type makeEnvSandbox struct {
	root, bin, stubDir string
}

func newMakeEnvSandbox(t *testing.T, withCLI bool) makeEnvSandbox {
	t.Helper()
	root := t.TempDir()
	s := makeEnvSandbox{root: root, bin: filepath.Join(root, "bin"), stubDir: filepath.Join(root, "stub")}
	for _, d := range []string{"infra/infisical/templates", "infra/infisical/envcheck/cmd/envcheck", "backend", "frontend", "bin", "stub"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	copyFile := func(src, dst string, mode os.FileMode) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, b, mode); err != nil {
			t.Fatal(err)
		}
	}
	inf := filepath.Join(root, "infra/infisical")
	copyFile("../make-env.sh", filepath.Join(inf, "make-env.sh"), 0o755)
	copyFile("../check-env.sh", filepath.Join(inf, "check-env.sh"), 0o755)
	copyFile(renderScript, filepath.Join(inf, "render-template.sh"), 0o755)
	for _, tmpl := range templateFiles(t) {
		copyFile(tmpl, filepath.Join(inf, "templates", filepath.Base(tmpl)), 0o644)
	}
	// The checker's sources (check-env.sh runs it with go run).
	sources, _ := filepath.Glob("*.go")
	for _, src := range append(sources, "go.mod", "cmd/envcheck/main.go") {
		if !strings.HasSuffix(src, "_test.go") {
			copyFile(src, filepath.Join(inf, "envcheck", src), 0o644)
		}
	}
	for _, tool := range []string{"sh", "sed", "grep", "mktemp", "cat", "mv", "rm", "dirname", "head"} {
		p, err := exec.LookPath(tool)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(p, filepath.Join(s.bin, tool)); err != nil {
			t.Fatal(err)
		}
	}
	if withCLI {
		if err := os.WriteFile(filepath.Join(s.bin, "infisical"), []byte(stubCLI), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	s.stub(t, "dev", []string{"backend.env", "admin.env", "frontend.env"})
	// Pre-existing local files, to prove what is (not) overwritten.
	for _, f := range []string{"backend/development.env", "frontend/.env"} {
		if err := os.WriteFile(filepath.Join(root, f), []byte("OLD=1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// stub makes the fake CLI answer with the fixture renderings of names.
func (s makeEnvSandbox) stub(t *testing.T, env string, names []string) {
	t.Helper()
	old, _ := filepath.Glob(filepath.Join(s.stubDir, "*.env"))
	for _, f := range old {
		os.Remove(f)
	}
	data := loadFixtures(t)
	for _, name := range names {
		content, _ := (&fakeInfisical{data: data}).render(t, name, env)
		if err := os.WriteFile(filepath.Join(s.stubDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func (s makeEnvSandbox) run(t *testing.T, extraEnv ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(filepath.Join(s.bin, "sh"), filepath.Join(s.root, "infra/infisical/make-env.sh"))
	cmd.Env = append([]string{
		"PATH=" + s.bin,
		"HOME=" + s.root,
		"TMPDIR=" + s.root,
		"STUB_DIR=" + s.stubDir,
		"LEVIOSA_INFISICAL_PROJECT_ID=" + fakeProjectID,
	}, extraEnv...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.String(), err
}

func (s makeEnvSandbox) read(t *testing.T, f string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(s.root, f))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func (s makeEnvSandbox) untouched(t *testing.T) {
	t.Helper()
	for _, f := range []string{"backend/development.env", "frontend/.env"} {
		if got := s.read(t, f); got != "OLD=1\n" {
			t.Errorf("%s was changed on failure", f)
		}
	}
}

func TestMakeEnvWritesBothFilesFromDev(t *testing.T) {
	s := newMakeEnvSandbox(t, true)
	out, err := s.run(t, "STUB_LOGGED_IN=1")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "OVERWROTE backend/development.env and frontend/.env") {
		t.Errorf("does not say it overwrote the files:\n%s", out)
	}
	dev := s.read(t, "backend/development.env")
	if dev != s.readStub(t, "backend.env")+s.readStub(t, "admin.env") {
		t.Errorf("backend/development.env is not backend.env + admin.env")
	}
	if fe := s.read(t, "frontend/.env"); fe != s.readStub(t, "frontend.env") {
		t.Errorf("frontend/.env is not frontend.env")
	}
	for _, f := range []string{"backend/development.env", "frontend/.env"} {
		info, err := os.Stat(filepath.Join(s.root, f))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode %v, want 0600", f, info.Mode().Perm())
		}
	}
	// The combined backend file is still a valid backend + admin env file.
	if r := Check("dev", "backend.env", []byte(s.readStub(t, "backend.env"))); !r.OK() {
		t.Errorf("%v", r.Problems)
	}
	if entries, _ := os.ReadDir(s.root); len(entries) != 5 { // infra backend frontend bin stub
		t.Errorf("left temporary files behind: %v", entries)
	}
}

func (s makeEnvSandbox) readStub(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(s.stubDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestMakeEnvFailsClearlyAndWritesNothing(t *testing.T) {
	cases := []struct {
		name    string
		withCLI bool
		env     []string
		want    string
	}{
		{"CLI missing", false, nil, "the Infisical CLI is not installed"},
		{"not logged in", true, nil, "not logged in to Infisical. Run: infisical login"},
		{"render fails", true, []string{"STUB_LOGGED_IN=1", "STUB_FAIL=admin.env"}, "rendering admin.env from Infisical 'dev' failed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newMakeEnvSandbox(t, c.withCLI)
			out, err := s.run(t, c.env...)
			if err == nil {
				t.Fatalf("succeeded:\n%s", out)
			}
			if !strings.Contains(out, c.want) {
				t.Errorf("output lacks %q:\n%s", c.want, out)
			}
			if strings.Contains(out, "stub-session-token") {
				t.Errorf("printed the session token")
			}
			s.untouched(t)
		})
	}

	t.Run("incomplete dev folders", func(t *testing.T) {
		s := newMakeEnvSandbox(t, true)
		backend := strings.ReplaceAll(s.readStub(t, "backend.env"), "ENCX_PEPPER=", "XENCX_PEPPER=")
		if err := os.WriteFile(filepath.Join(s.stubDir, "backend.env"), []byte(backend), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := s.run(t, "STUB_LOGGED_IN=1")
		if err == nil || !strings.Contains(out, "has no ENCX_PEPPER") {
			t.Errorf("err=%v\n%s", err, out)
		}
		s.untouched(t)
	})
}
