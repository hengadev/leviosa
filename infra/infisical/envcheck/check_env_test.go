package envcheck

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// check-env.sh (the operator's read-only check) end to end, with the stub
// CLI of make_env_test.go standing in for Infisical and the real checker
// built by go run from a copy of this module.

func (s makeEnvSandbox) checkEnv(t *testing.T, env string, extraEnv ...string) (string, error) {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	link := filepath.Join(s.bin, "go")
	if _, err := os.Lstat(link); err != nil {
		if err := os.Symlink(goBin, link); err != nil {
			t.Fatal(err)
		}
	}
	cache, err := exec.Command(goBin, "env", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(s.bin, "sh"), filepath.Join(s.root, "infra/infisical/check-env.sh"), env)
	cmd.Env = append([]string{
		"PATH=" + s.bin,
		"HOME=" + s.root,
		"TMPDIR=" + s.root,
		"STUB_DIR=" + s.stubDir,
		"STUB_ENV=" + env,
		"LEVIOSA_INFISICAL_PROJECT_ID=" + fakeProjectID,
		"GOCACHE=" + strings.TrimSpace(string(cache)),
		"GOPATH=" + filepath.Join(s.root, "gopath"),
		"GOENV=off",
		"GOTOOLCHAIN=local",
		"GOFLAGS=",
		"CGO_ENABLED=0",
	}, extraEnv...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err = cmd.Run()
	return out.String(), err
}

func TestCheckEnvListsEveryFileOfEachEnvironment(t *testing.T) {
	data := loadFixtures(t)
	for _, env := range Environments {
		t.Run(env, func(t *testing.T) {
			s := newMakeEnvSandbox(t, true)
			names, _ := Files(env)
			s.stub(t, env, names)
			out, err := s.checkEnv(t, env, "STUB_LOGGED_IN=1")
			if err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			for _, name := range names {
				if !strings.Contains(out, "\n"+name+" (") {
					t.Errorf("%s not reported", name)
				}
			}
			if !strings.Contains(out, "envcheck: "+env+": OK") {
				t.Errorf("no OK line:\n%s", out)
			}
			assertNoValue(t, data, out)
			if leftovers, _ := filepath.Glob(filepath.Join(s.root, "leviosa-env-check.*")); len(leftovers) != 0 {
				t.Errorf("renderings left behind: %v", leftovers)
			}
		})
	}
}

func TestCheckEnvFailures(t *testing.T) {
	s := newMakeEnvSandbox(t, true)
	names, _ := Files("staging")
	s.stub(t, "staging", names)

	out, err := s.checkEnv(t, "staging")
	if err == nil || !strings.Contains(out, "no Infisical session. Run: infisical login") {
		t.Errorf("not logged in: err=%v\n%s", err, out)
	}

	out, err = s.checkEnv(t, "staging", "STUB_LOGGED_IN=1", "STUB_FAIL=frontend.env")
	if err == nil || !strings.Contains(out, "frontend.env (0 keys)\n  PROBLEM: not rendered") {
		t.Errorf("failed rendering: err=%v\n%s", err, out)
	}
	if strings.Contains(out, "stub-session-token") {
		t.Errorf("printed the session token")
	}

	out, err = s.checkEnv(t, "production", "STUB_LOGGED_IN=1")
	if err == nil || !strings.Contains(out, "invalid environment") {
		t.Errorf("bad environment: err=%v\n%s", err, out)
	}
}
