package envcheck

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// render-template.sh: the only way a template becomes a concrete
// per-environment copy (make env, check-env.sh, the Agent role).

func runRender(t *testing.T, projectOverride string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command("sh", append([]string{renderScript}, args...)...)
	env := []string{}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "LEVIOSA_INFISICAL_PROJECT_ID=") {
			env = append(env, kv)
		}
	}
	if projectOverride != "" {
		env = append(env, "LEVIOSA_INFISICAL_PROJECT_ID="+projectOverride)
	}
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func templateFiles(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(templatesDir, "*.env.tmpl"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("no templates found: %v", err)
	}
	return matches
}

func TestRenderTemplateSubstitutesEveryPlaceholder(t *testing.T) {
	for _, tmpl := range templateFiles(t) {
		src, err := os.ReadFile(tmpl)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), "__LEVIOSA_PROJECT_ID__") {
			t.Errorf("%s does not use the project placeholder", tmpl)
		}
		if regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-`).Match(src) {
			t.Errorf("%s hard-codes a project ID", tmpl)
		}
		for _, env := range Environments {
			out, stderr, err := runRender(t, fakeProjectID, tmpl, env)
			if err != nil {
				t.Fatalf("%s %s: %v: %s", tmpl, env, err, stderr)
			}
			if strings.Contains(out, "__LEVIOSA_") {
				t.Errorf("%s %s: a placeholder survived", tmpl, env)
			}
			if !strings.Contains(out, `"`+fakeProjectID+`"`) {
				t.Errorf("%s %s: project ID not substituted", tmpl, env)
			}
		}
		after, _ := os.ReadFile(tmpl)
		if !bytes.Equal(src, after) {
			t.Errorf("%s was modified by rendering", tmpl)
		}
	}
}

func TestRenderTemplateSubstitutesTheEnvironment(t *testing.T) {
	tmpl := filepath.Join(templatesDir, "backend.env.tmpl")
	for _, env := range Environments {
		out, stderr, err := runRender(t, fakeProjectID, tmpl, env)
		if err != nil {
			t.Fatalf("%v: %s", err, stderr)
		}
		if !strings.Contains(out, `"`+fakeProjectID+`" "`+env+`" "/"`) {
			t.Errorf("%s: listSecrets does not target environment %s", env, env)
		}
	}
}

func TestRenderTemplateRefusesBadInput(t *testing.T) {
	tmpl := filepath.Join(templatesDir, "backend.env.tmpl")
	cases := []struct {
		name    string
		project string
		args    []string
		stderr  string
	}{
		{"unknown environment", fakeProjectID, []string{tmpl, "production"}, "invalid environment"},
		{"uppercase environment", fakeProjectID, []string{tmpl, "DEV"}, "invalid environment"},
		{"environment with a slash", fakeProjectID, []string{tmpl, "dev/x"}, "invalid environment"},
		{"environment with shell metacharacters", fakeProjectID, []string{tmpl, "dev;id"}, "invalid environment"},
		{"empty environment", fakeProjectID, []string{tmpl, ""}, "invalid environment"},
		{"missing template", fakeProjectID, []string{filepath.Join(templatesDir, "nope.env.tmpl"), "dev"}, "template not found"},
		{"no arguments", fakeProjectID, nil, "usage"},
		{"extra argument", fakeProjectID, []string{tmpl, "dev", "x"}, "usage"},
		{"project ID not a UUID", "REPLACE_WITH_LEVIOSA_PROJECT_ID", []string{tmpl, "dev"}, "project ID"},
		{"project ID with sed metacharacters", "0/&;\\", []string{tmpl, "dev"}, "project ID"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, stderr, err := runRender(t, c.project, c.args...)
			if err == nil {
				t.Fatalf("expected failure, got output %q", out)
			}
			if !strings.Contains(stderr, c.stderr) {
				t.Errorf("stderr %q does not mention %q", stderr, c.stderr)
			}
			if out != "" {
				t.Errorf("printed a partial template on failure")
			}
		})
	}
}

// Without the override, the project ID comes from terraform.tfvars
// (infisical_project_id), the one place the operator sets it.
func TestRenderTemplateReadsTheProjectIDFromTerraform(t *testing.T) {
	tfvars, err := os.ReadFile("../../terraform/terraform.tfvars")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^\s*infisical_project_id\s*=\s*"([^"]*)"`).FindSubmatch(tfvars)
	if m == nil {
		t.Fatal("terraform.tfvars has no infisical_project_id")
	}
	id := string(m[1])
	out, stderr, err := runRender(t, "", filepath.Join(templatesDir, "admin.env.tmpl"), "dev")
	if regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(id) {
		if err != nil || !strings.Contains(out, `"`+id+`"`) {
			t.Errorf("did not render with the tfvars project ID: %v %s", err, stderr)
		}
		return
	}
	// Still the committed placeholder (issue 008 has not created the project).
	if err == nil || !strings.Contains(stderr, "terraform.tfvars") {
		t.Errorf("with the tfvars placeholder: err=%v stderr=%q, want a refusal naming terraform.tfvars", err, stderr)
	}
}
