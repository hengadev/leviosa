package envcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/template"
)

// A stand-in for the Infisical Agent's template engine (and `infisical
// export --template`, which uses the same code): Go text/template with the
// Agent's secret functions, answering from testdata/infisical.json instead of
// the API. Mirrors CLI v0.43.138, packages/cmd/agent.go:
//
//   listSecrets <projectID> <env> <path> [<json options>]  -> []secret
//   getSecretByName <projectID> <env> <path> <name>          -> secret
//
// Every call is recorded: the Agent change-detects a template through the
// ETag of its LAST secret call only, so the number and order of calls per
// template are part of the contract (see ../README.md).
//
// Sprig is deliberately not registered. Recent Agents add it (minus env and
// expandenv), but the templates use Go built-ins only, so they do not depend
// on the Agent version; a template that calls a sprig function fails here.

const (
	// fakeProjectID stands in for the leviosa project's ID (render-template.sh
	// reads the real one from infra/terraform/terraform.tfvars).
	fakeProjectID  = "00000000-0000-4000-8000-00000000f1c5"
	templatesDir   = "../templates"
	renderScript   = "../render-template.sh"
	fixturesFile   = "testdata/infisical.json"
	composeFixture = "../../compose/test/fixtures"
)

// secret mirrors models.SingleEnvironmentVariable (CLI v0.43.138), so a
// template that uses a field the Agent does not have fails here too.
type secret struct {
	Key                   string
	WorkspaceId           string
	Value                 string
	Type                  string
	ID                    string
	SecretPath            string
	Tags                  []struct{ ID, Name, Slug, Color string }
	Comment               string
	Etag                  string
	SkipMultilineEncoding bool
}

type secretCall struct {
	Fn        string
	Project   string
	Env       string
	Path      string
	Recursive bool
}

// folders: environment -> folder path -> key -> value.
type folders map[string]map[string]map[string]string

func loadFixtures(t *testing.T) folders {
	t.Helper()
	raw, err := os.ReadFile(fixturesFile)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	f := folders{}
	for env, msg := range top {
		if strings.HasPrefix(env, "_") {
			continue
		}
		var byFolder map[string]map[string]string
		if err := json.Unmarshal(msg, &byFolder); err != nil {
			t.Fatalf("%s: %v", env, err)
		}
		f[env] = byFolder
	}
	return f
}

type fakeInfisical struct {
	data    folders
	reverse bool // return secrets in reverse order: templates must not depend on order
	calls   []secretCall
}

func (f *fakeInfisical) funcs() template.FuncMap {
	unsupported := func(name string) func(...string) (any, error) {
		return func(...string) (any, error) {
			return nil, fmt.Errorf("%s is not used by Leviosa's templates", name)
		}
	}
	return template.FuncMap{
		"listSecrets":              f.listSecrets,
		"secret":                   f.listSecrets, // deprecated alias in the Agent
		"getSecretByName":          f.getSecretByName,
		"listSecretsByProjectSlug": unsupported("listSecretsByProjectSlug"),
		"dynamicSecret":            unsupported("dynamicSecret"),
		"dynamic_secret":           unsupported("dynamic_secret"),
	}
}

func (f *fakeInfisical) env(project, env string) (map[string]map[string]string, error) {
	if project != fakeProjectID {
		return nil, fmt.Errorf("unknown project %q", project)
	}
	byFolder, ok := f.data[env]
	if !ok {
		return nil, fmt.Errorf("unknown environment %q", env)
	}
	return byFolder, nil
}

func (f *fakeInfisical) listSecrets(project, env, path string, args ...string) ([]secret, error) {
	var opts struct {
		Recursive bool  `json:"recursive"`
		Expand    *bool `json:"expandSecretReferences,omitempty"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal([]byte(args[0]), &opts); err != nil {
			return nil, err
		}
	}
	f.calls = append(f.calls, secretCall{Fn: "listSecrets", Project: project, Env: env, Path: path, Recursive: opts.Recursive})

	byFolder, err := f.env(project, env)
	if err != nil {
		return nil, err
	}
	var paths []string
	for folder := range byFolder {
		switch {
		case folder == path:
		case opts.Recursive && (path == "/" || strings.HasPrefix(folder, path+"/")):
		default:
			continue
		}
		paths = append(paths, folder)
	}
	if len(paths) == 0 && !opts.Recursive {
		return nil, fmt.Errorf("folder %s not found in %s", path, env)
	}
	sort.Strings(paths)

	var out []secret
	for _, folder := range paths {
		keys := make([]string, 0, len(byFolder[folder]))
		for k := range byFolder[folder] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = append(out, secret{Key: k, Value: byFolder[folder][k], SecretPath: folder, Type: "shared", WorkspaceId: project})
		}
	}
	if f.reverse {
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return out, nil
}

func (f *fakeInfisical) getSecretByName(project, env, path, name string) (secret, error) {
	f.calls = append(f.calls, secretCall{Fn: "getSecretByName", Project: project, Env: env, Path: path})
	byFolder, err := f.env(project, env)
	if err != nil {
		return secret{}, err
	}
	v, ok := byFolder[path][name]
	if !ok {
		return secret{}, fmt.Errorf("secret %s not found in %s%s", name, env, path)
	}
	return secret{Key: name, Value: v, SecretPath: path, Type: "shared", WorkspaceId: project}, nil
}

// substitute runs render-template.sh, exactly as make env, check-env.sh and
// the Agent role do, with the fake project ID.
func substitute(t *testing.T, templateFile, env string) string {
	t.Helper()
	cmd := exec.Command("sh", renderScript, templateFile, env)
	cmd.Env = append(os.Environ(), "LEVIOSA_INFISICAL_PROJECT_ID="+fakeProjectID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("render-template.sh %s %s: %v\n%s", templateFile, env, err, stderr.String())
	}
	return string(out)
}

// render renders templates/<file>.tmpl for env against the fake.
func (f *fakeInfisical) render(t *testing.T, file, env string) (string, []secretCall) {
	t.Helper()
	src := substitute(t, filepath.Join(templatesDir, file+".tmpl"), env)
	tmpl, err := template.New(file + ".tmpl").Funcs(f.funcs()).Parse(src)
	if err != nil {
		t.Fatalf("%s: parse: %v", file, err)
	}
	f.calls = nil
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("%s (%s): execute: %v", file, env, err)
	}
	return buf.String(), append([]secretCall(nil), f.calls...)
}

// values parses a rendered file into key -> value (the test-side reading of
// the KEY='value' form; TestRenderedFilesParseInCompose checks Compose agrees).
func values(t *testing.T, content string) map[string]string {
	t.Helper()
	entries, problems := parse(content)
	if len(problems) > 0 {
		t.Fatalf("unparseable rendering: %v", problems)
	}
	m := map[string]string{}
	for _, e := range entries {
		if _, dup := m[e.key]; dup {
			t.Fatalf("duplicate key %s", e.key)
		}
		m[e.key] = e.value
	}
	return m
}
