package envcheck

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"text/template"
)

// The weekly Infisical snapshot template (roles/backup, issue 028): not an
// env template, so it lives with the backup role, but it runs in the same
// engine (`infisical export --template`), so it is checked here against the
// same fake.
const snapshotTemplate = "../../ansible/roles/backup/files/leviosa-snapshot.json.tmpl"

type snapshot struct {
	Format      string `json:"format"`
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Secrets     []struct {
		Path    string `json:"path"`
		Key     string `json:"key"`
		Value   string `json:"value"`
		Comment string `json:"comment"`
	} `json:"secrets"`
}

func renderSnapshot(t *testing.T, f *fakeInfisical, env string) (string, []secretCall) {
	t.Helper()
	src := substitute(t, snapshotTemplate, env)
	tmpl, err := template.New("snapshot").Funcs(f.funcs()).Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	f.calls = nil
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		t.Fatalf("execute: %v", err)
	}
	return buf.String(), f.calls
}

func parseSnapshot(t *testing.T, out string) snapshot {
	t.Helper()
	var s snapshot
	dec := json.NewDecoder(strings.NewReader(out))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		t.Fatalf("the snapshot is not valid JSON: %v\n%s", err, out)
	}
	return s
}

// Every secret of every prod folder is in the snapshot, with its folder,
// verbatim, through ONE recursive call on /.
func TestSnapshotHoldsEveryProdSecretVerbatim(t *testing.T) {
	data := loadFixtures(t)
	for _, reverse := range []bool{false, true} {
		f := &fakeInfisical{data: data, reverse: reverse}
		out, calls := renderSnapshot(t, f, "prod")
		s := parseSnapshot(t, out)

		if s.Format != "leviosa-infisical-snapshot/1" || s.Project != fakeProjectID || s.Environment != "prod" {
			t.Errorf("header: %+v", s)
		}
		if len(calls) != 1 || calls[0].Fn != "listSecrets" || calls[0].Path != "/" || !calls[0].Recursive || calls[0].Env != "prod" {
			t.Errorf("want one recursive listSecrets on prod /, got %+v", calls)
		}

		got := map[string]string{}
		for _, sec := range s.Secrets {
			got[sec.Path+"|"+sec.Key] = sec.Value
		}
		want := 0
		for folder, kv := range data["prod"] {
			for k, v := range kv {
				want++
				if g, ok := got[folder+"|"+k]; !ok {
					t.Errorf("%s %s missing from the snapshot", folder, k)
				} else if g != v {
					t.Errorf("%s %s not kept verbatim", folder, k)
				}
			}
		}
		if len(s.Secrets) != want {
			t.Errorf("snapshot holds %d secrets, prod has %d", len(s.Secrets), want)
		}
		if got["/encx|ENCX_PEPPER"] == "" {
			t.Error("the pepper is not in the snapshot")
		}
	}
}

// Values that are hard to quote survive the round trip, and nested folders
// keep their full path.
func TestSnapshotQuotesHardValues(t *testing.T) {
	hard := map[string]string{
		"QUOTES":    `a "double" and 'single' quote`,
		"BACKSLASH": `C:\path\to\{{ not a template }}`,
		"DOLLAR":    `$2a$14$hash#not-a-comment`,
		"PEM":       "-----BEGIN KEY-----\nline1\nline2\n-----END KEY-----\n",
		"TAB":       "a\tb\r\n",
		"UNICODE":   "été — 日本 \u2028 ok",
		"EMPTY":     "",
	}
	f := &fakeInfisical{data: folders{"prod": {"/odd": hard, "/odd/nested": {"X": "y"}}}}
	out, _ := renderSnapshot(t, f, "prod")
	s := parseSnapshot(t, out)
	seen := 0
	for _, sec := range s.Secrets {
		switch sec.Path {
		case "/odd":
			seen++
			if sec.Value != hard[sec.Key] {
				t.Errorf("%s: got %q, want %q", sec.Key, sec.Value, hard[sec.Key])
			}
		case "/odd/nested":
			seen++
		default:
			t.Errorf("unexpected path %q", sec.Path)
		}
	}
	if seen != len(hard)+1 {
		t.Errorf("saw %d secrets, want %d", seen, len(hard)+1)
	}
}

// A control character Go quotes as \x.. is not valid JSON: the snapshot job's
// jq check then refuses the export (nothing uploaded) instead of storing an
// unreadable snapshot. This pins that documented behaviour.
func TestSnapshotWithControlCharacterIsRejectedAsJSON(t *testing.T) {
	f := &fakeInfisical{data: folders{"prod": {"/odd": {"BELL": "a\x01b"}}}}
	out, _ := renderSnapshot(t, f, "prod")
	if json.Valid([]byte(out)) {
		t.Fatalf("expected invalid JSON for a control character, got %s", out)
	}
}

// No folder at all renders a valid, empty snapshot (leviosa-snapshot then
// refuses it: "the export holds no secret").
func TestSnapshotOfAnEmptyEnvironmentIsValidJSON(t *testing.T) {
	f := &fakeInfisical{data: folders{"prod": {}}}
	out, _ := renderSnapshot(t, f, "prod")
	if s := parseSnapshot(t, out); len(s.Secrets) != 0 {
		t.Fatalf("want no secret, got %d", len(s.Secrets))
	}
}
