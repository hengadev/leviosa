package envcheck

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The fixture test of issue 024: every Agent template, rendered against fake
// Infisical folders (testdata/infisical.json) with no Infisical access,
// produces exactly the expected key set per file, keeps every value intact,
// assembles the database, RabbitMQ and Redis addresses in the formats the
// apps parse, and never lets a backend value into frontend.env.

var (
	backendCore = []string{
		// set by the template itself
		"ENVIRONMENT", "SERVER_PORT", "LOG_FORMAT", "REDIS_DB",
		// assembled from atoms
		"DATABASE_URL", "REDIS_ADDR", "RABBITMQ_URL",
		"S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY", "S3_REGION",
		// /redis minus its address atoms
		"REDIS_PASSWORD",
		// /app
		"FRONTEND_ORIGIN", "SESSION_SECRET", "BOOKING_TOKEN_SECRET",
		"SESSION_COOKIE_NAME", "CLIENT_IP_HEADER", "COMING_SOON",
		// /encx
		"ENCX_PEPPER",
		// /aws (Terraform's keys in staging and prod)
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "S3_BUCKET_NAME", "ENCX_KMS_KEY_ALIAS",
	}
	providers = []string{
		"STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET", "STRIPE_CONNECT_WEBHOOK_SECRET",
		"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD",
		"TWILIO_ACCOUNT_SID", "TWILIO_API_KEY_SID", "TWILIO_API_KEY_SECRET", "TWILIO_REGION", "TWILIO_EDGE", "TWILIO_SENDER",
	}
	frontendBase = []string{
		"NODE_ENV", "PORT", "API_URL", "PROTOCOL_HEADER", "HOST_HEADER", "USE_MOCK_DATA",
		"COMING_SOON", "SESSION_COOKIE_NAME", "CLIENT_IP_HEADER",
	}
	adminKeys = []string{
		"ADMIN_EMAIL", "ADMIN_PASSWORD", "ADMIN_FIRST_NAME", "ADMIN_LAST_NAME", "ADMIN_TELEPHONE",
		"ADMIN_POSTAL_CODE", "ADMIN_CITY", "ADMIN_ADDRESS1", "ADMIN_ADDRESS2",
	}
)

func keys(groups ...[]string) []string {
	var out []string
	for _, g := range groups {
		out = append(out, g...)
	}
	sort.Strings(out)
	return out
}

// wantKeys: environment -> file -> its exact key set for the fixture folders.
var wantKeys = map[string]map[string][]string{
	"prod": {
		"backend.env":  keys(backendCore, providers, []string{"ENCX_KMS_KEY_ARN", "SMTP_FROM"}),
		"frontend.env": keys(frontendBase),
		"admin.env":    keys(adminKeys),
		"redis.env":    {"REDIS_PASSWORD"},
		"postgres.env": keys([]string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}),
		"rabbitmq.env": keys([]string{"RABBITMQ_DEFAULT_USER", "RABBITMQ_DEFAULT_PASS", "RABBITMQ_DEFAULT_VHOST"}),
		"caddy.env":    keys([]string{"CLOUDFLARE_API_TOKEN", "STAGING_BASIC_AUTH_USER", "STAGING_BASIC_AUTH_HASH"}),
		"backup.env":   keys([]string{"BACKUP_AWS_ACCESS_KEY_ID", "BACKUP_AWS_SECRET_ACCESS_KEY", "BACKUP_S3_BUCKET", "BACKUP_AWS_REGION", "POSTGRES_USER", "POSTGRES_DB"}),
	},
	"staging": {
		"backend.env":  keys(backendCore, providers, []string{"ENCX_KMS_KEY_ARN", "PUBLIC_PLAUSIBLE_DOMAIN"}),
		"frontend.env": keys(frontendBase, []string{"PUBLIC_PLAUSIBLE_DOMAIN"}),
		"admin.env":    keys(adminKeys),
		"redis.env":    {"REDIS_PASSWORD"},
	},
	"dev": {
		// No /stripe, /smtp or /twilio in the dev fixture (integrations off);
		// /aws holds LocalStack's values, endpoint included.
		"backend.env":  keys(backendCore, []string{"AWS_ENDPOINT_URL"}),
		"frontend.env": keys(frontendBase),
		"admin.env":    keys(adminKeys),
		"redis.env":    {"REDIS_PASSWORD"},
		"postgres.env": keys([]string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}),
		"rabbitmq.env": keys([]string{"RABBITMQ_DEFAULT_USER", "RABBITMQ_DEFAULT_PASS", "RABBITMQ_DEFAULT_VHOST"}),
	},
}

// wholeFolders: file -> folders whose keys it prints whole (per source env).
var wholeFolders = map[string][]string{
	"backend.env": {"/app", "/encx", "/aws", "/stripe", "/smtp", "/twilio"},
	"admin.env":   {"/admin"},
	"backup.env":  {"/backup"},
}

// excludedFromFrontend: no key or value of these folders may reach
// frontend.env (issue 024; /redis, /rabbitmq, /caddy, /backup and /host
// added: the SvelteKit server needs none of them either).
var excludedFromFrontend = []string{"/aws", "/encx", "/stripe", "/twilio", "/smtp", "/db", "/admin", "/redis", "/rabbitmq", "/caddy", "/backup", "/host"}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestEveryFileHasExactlyTheExpectedKeys(t *testing.T) {
	data := loadFixtures(t)
	for _, env := range Environments {
		names, err := Files(env)
		if err != nil {
			t.Fatal(err)
		}
		if got := sortedKeys(toSet(names)); !reflect.DeepEqual(got, sortedKeys(toSet(mapKeys(wantKeys[env])))) {
			t.Errorf("%s: Files() = %v, want the files of wantKeys %v", env, got, sortedKeys(toSet(mapKeys(wantKeys[env]))))
		}
		for _, file := range names {
			t.Run(env+"/"+file, func(t *testing.T) {
				content, _ := (&fakeInfisical{data: data}).render(t, file, env)
				got := sortedKeys(values(t, content))
				if !reflect.DeepEqual(got, wantKeys[env][file]) {
					t.Errorf("keys:\n got  %v\n want %v", got, wantKeys[env][file])
				}
				if r := Check(env, file, []byte(content)); !r.OK() {
					t.Errorf("envcheck problems: %v", r.Problems)
				}
				if r := Check(env, file, []byte(content)); env != "dev" && len(r.Warnings) > 0 {
					t.Errorf("envcheck warnings with complete fixtures: %v", r.Warnings)
				}

				// The Agent's API gives no ordering guarantee: the same
				// folders listed backwards must render the same values.
				reversed, _ := (&fakeInfisical{data: data, reverse: true}).render(t, file, env)
				if !reflect.DeepEqual(values(t, content), values(t, reversed)) {
					t.Errorf("rendering depends on the order of the secrets")
				}
			})
		}
	}
}

func TestWholeFoldersKeepEveryValueVerbatim(t *testing.T) {
	data := loadFixtures(t)
	for _, env := range Environments {
		names, _ := Files(env)
		for _, file := range names {
			for _, folder := range wholeFolders[file] {
				content, _ := (&fakeInfisical{data: data}).render(t, file, env)
				got := values(t, content)
				for k, want := range data[env][folder] {
					if got[k] != want {
						t.Errorf("%s/%s: %s from %s was not kept verbatim", env, file, k, folder)
					}
				}
			}
		}
	}

	// caddy.env: staging's /caddy and prod's /caddy, whole.
	content, _ := (&fakeInfisical{data: data}).render(t, "caddy.env", "prod")
	got := values(t, content)
	for _, env := range []string{"staging", "prod"} {
		for k, want := range data[env]["/caddy"] {
			if got[k] != want {
				t.Errorf("caddy.env: %s from %s/caddy was not kept verbatim", k, env)
			}
		}
	}
}

func TestCompositeAndDerivedValues(t *testing.T) {
	data := loadFixtures(t)
	want := map[string]map[string]string{
		"prod": {
			"DATABASE_URL": "postgres://leviosa:FixtureProdPg1@postgres:5432/leviosa?sslmode=disable",
			"RABBITMQ_URL": "amqp://leviosa:FixtureProdMq1@rabbitmq:5672/leviosa",
			"REDIS_ADDR":   "redis:6379",
			"ENVIRONMENT":  "production",
			"LOG_FORMAT":   "json",
		},
		"staging": {
			"DATABASE_URL": "postgres://leviosa_staging:FixtureStagingPg2@postgres:5432/leviosa_staging?sslmode=disable",
			"RABBITMQ_URL": "amqp://leviosa_staging:FixtureStagingMq2@rabbitmq:5672/leviosa_staging",
			"REDIS_ADDR":   "staging_redis:6379",
			"ENVIRONMENT":  "staging",
			"LOG_FORMAT":   "json",
		},
		"dev": {
			"DATABASE_URL": "postgres://user:password@localhost:5432/leviosa_dev?sslmode=disable",
			// The default vhost "/" must be written %2F: a bare trailing
			// slash is an empty vhost name for amqp091-go.
			"RABBITMQ_URL": "amqp://guest:guest@localhost:5672/%2F",
			"REDIS_ADDR":   "localhost:6379",
			"ENVIRONMENT":  "development",
			"LOG_FORMAT":   "text",
		},
	}
	for env, w := range want {
		content, _ := (&fakeInfisical{data: data}).render(t, "backend.env", env)
		got := values(t, content)
		for k, v := range w {
			if got[k] != v {
				t.Errorf("%s backend.env: %s = %q, want %q", env, k, got[k], v)
			}
		}
		if got["SERVER_PORT"] != "3500" || got["REDIS_DB"] != "0" {
			t.Errorf("%s backend.env: SERVER_PORT/REDIS_DB not the stack's fixed values", env)
		}
		// The backend's S3 client reads S3_*; Terraform writes AWS_* once.
		aws := data[env]["/aws"]
		for s3, awsKey := range map[string]string{"S3_ACCESS_KEY_ID": "AWS_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY": "AWS_SECRET_ACCESS_KEY", "S3_REGION": "AWS_REGION"} {
			if got[s3] != aws[awsKey] {
				t.Errorf("%s backend.env: %s is not %s", env, s3, awsKey)
			}
		}
		if got["REDIS_PASSWORD"] != data[env]["/redis"]["REDIS_PASSWORD"] {
			t.Errorf("%s backend.env: REDIS_PASSWORD is not /redis's", env)
		}
	}

	infra := map[string]map[string]string{
		"postgres.env": {"POSTGRES_USER": "leviosa", "POSTGRES_PASSWORD": "FixtureProdPg1", "POSTGRES_DB": "leviosa"},
		"rabbitmq.env": {"RABBITMQ_DEFAULT_USER": "leviosa", "RABBITMQ_DEFAULT_PASS": "FixtureProdMq1", "RABBITMQ_DEFAULT_VHOST": "leviosa"},
		"redis.env":    {"REDIS_PASSWORD": "FixtureProdRedis1"},
		"backup.env": {
			"POSTGRES_USER": "leviosa", "POSTGRES_DB": "leviosa",
			"BACKUP_AWS_ACCESS_KEY_ID": "AKIAFIXTUREBACKUP001", "BACKUP_AWS_SECRET_ACCESS_KEY": "fixture-backup-secret-not-real",
			"BACKUP_S3_BUCKET": "leviosa-fixture-backups", "BACKUP_AWS_REGION": "eu-central-1",
		},
	}
	for file, w := range infra {
		content, _ := (&fakeInfisical{data: data}).render(t, file, "prod")
		if got := values(t, content); !reflect.DeepEqual(got, w) {
			t.Errorf("prod %s = %v, want %v", file, got, w)
		}
	}
	content, _ := (&fakeInfisical{data: data}).render(t, "redis.env", "staging")
	if got := values(t, content)["REDIS_PASSWORD"]; got != "FixtureStagingRedis2" {
		t.Errorf("staging redis.env does not hold staging's own password")
	}
}

func TestFrontendEnvHoldsNothingFromBackendFolders(t *testing.T) {
	data := loadFixtures(t)
	secretish := regexp.MustCompile(`PASSWORD|SECRET|TOKEN|KEY|PEPPER|HASH|SID|_PASS$`)
	wantAPI := map[string]string{"prod": "http://backend:3500", "staging": "http://staging_backend:3500", "dev": "http://localhost:3500"}
	for _, env := range Environments {
		content, _ := (&fakeInfisical{data: data}).render(t, "frontend.env", env)
		got := values(t, content)

		for _, folder := range excludedFromFrontend {
			for k, v := range data[env][folder] {
				if _, ok := got[k]; ok {
					t.Errorf("%s frontend.env contains %s from %s", env, k, folder)
				}
				for fk, fv := range got {
					if v != "" && fv == v {
						t.Errorf("%s frontend.env %s carries the value of %s%s", env, fk, folder, k)
					}
				}
				if secretish.MatchString(k) && v != "" && strings.Contains(content, v) {
					t.Errorf("%s frontend.env contains the value of %s%s", env, folder, k)
				}
			}
		}
		// /app is rendered through an allowlist: its backend-only secrets
		// stay out.
		for _, k := range []string{"SESSION_SECRET", "BOOKING_TOKEN_SECRET", "FRONTEND_ORIGIN"} {
			if _, ok := got[k]; ok {
				t.Errorf("%s frontend.env contains the backend-only /app key %s", env, k)
			}
			if v := data[env]["/app"][k]; strings.Contains(content, v) {
				t.Errorf("%s frontend.env contains the value of /app %s", env, k)
			}
		}
		allowed := toSet(FrontendKeys)
		for k := range got {
			if _, ok := allowed[k]; !ok {
				t.Errorf("%s frontend.env: %s is not on the allowlist", env, k)
			}
		}
		if _, ok := got["ORIGIN"]; ok {
			t.Errorf("%s frontend.env sets ORIGIN", env)
		}
		if got["API_URL"] != wantAPI[env] || got["PROTOCOL_HEADER"] != "x-forwarded-proto" || got["HOST_HEADER"] != "host" || got["USE_MOCK_DATA"] != "false" {
			t.Errorf("%s frontend.env: stack wiring wrong: %v", env, got)
		}
	}
}

// The Agent re-renders a template (and runs its command) only when the ETag
// of its LAST secret call changes. One call per template, so every change in
// a folder it reads is detected; caddy.env is the one documented exception
// (two environments, staging first, prod last: see ../README.md).
func TestSecretCallsPerTemplate(t *testing.T) {
	data := loadFixtures(t)
	type call = secretCall
	want := map[string][]call{
		"backend.env":  {{Fn: "listSecrets", Path: "/", Recursive: true}},
		"frontend.env": {{Fn: "listSecrets", Path: "/app"}},
		"admin.env":    {{Fn: "listSecrets", Path: "/admin"}},
		"redis.env":    {{Fn: "listSecrets", Path: "/redis"}},
		"postgres.env": {{Fn: "listSecrets", Path: "/db"}},
		"rabbitmq.env": {{Fn: "listSecrets", Path: "/rabbitmq"}},
		"backup.env":   {{Fn: "listSecrets", Path: "/", Recursive: true}},
		"caddy.env": {
			{Fn: "listSecrets", Env: "staging", Path: "/caddy"},
			{Fn: "listSecrets", Env: "prod", Path: "/caddy"},
		},
	}
	seen := map[string]bool{}
	for _, env := range Environments {
		names, _ := Files(env)
		for _, file := range names {
			seen[file] = true
			_, calls := (&fakeInfisical{data: data}).render(t, file, env)
			w := append([]call(nil), want[file]...)
			for i := range w {
				w[i].Project = fakeProjectID
				if w[i].Env == "" {
					w[i].Env = env
				}
			}
			if !reflect.DeepEqual(calls, w) {
				t.Errorf("%s %s: secret calls\n got  %+v\n want %+v", env, file, calls, w)
			}
		}
	}
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if name := strings.TrimSuffix(e.Name(), ".tmpl"); strings.HasSuffix(e.Name(), ".tmpl") && !seen[name] {
			t.Errorf("template %s is rendered for no environment", e.Name())
		}
	}
}

// Every key a stack's harness fixture (infra/compose/test/fixtures, what the
// compose files actually consume) holds must be in the rendering of the same
// file. AWS_ENDPOINT_URL and S3_ENDPOINT exist only in the harness
// (LocalStack); the real files never set them.
func TestComposeHarnessKeysAreAllRendered(t *testing.T) {
	data := loadFixtures(t)
	harnessOnly := map[string]bool{"AWS_ENDPOINT_URL": true, "S3_ENDPOINT": true}
	keyLine := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=`)
	for _, env := range []string{"prod", "staging"} {
		names, _ := Files(env)
		for _, file := range names {
			path := filepath.Join(composeFixture, env, "env", file)
			f, err := os.Open(path)
			if os.IsNotExist(err) {
				continue // no harness fixture for this file (backup.env)
			}
			if err != nil {
				t.Fatal(err)
			}
			content, _ := (&fakeInfisical{data: data}).render(t, file, env)
			got := values(t, content)
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				m := keyLine.FindStringSubmatch(sc.Text())
				if m == nil || harnessOnly[m[1]] {
					continue
				}
				if _, ok := got[m[1]]; !ok {
					t.Errorf("%s %s: the harness fixture has %s, the rendering does not", env, file, m[1])
				}
			}
			f.Close()
		}
	}
}

// Docker Compose is what loads these files on the server (env_file) and in
// local development. Compose expands $ and strips " #..." in unquoted values,
// so the templates single-quote every value; this proves Compose reads every
// rendered value back exactly. Skipped when docker compose is unavailable.
func TestRenderedFilesParseInCompose(t *testing.T) {
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		t.Skip("docker compose not available")
	}
	data := loadFixtures(t)
	dir := t.TempDir()
	type rendered struct{ env, file string }
	var all []rendered
	var services strings.Builder
	services.WriteString("services:\n")
	for _, env := range Environments {
		names, _ := Files(env)
		for _, file := range names {
			content, _ := (&fakeInfisical{data: data}).render(t, file, env)
			name := env + "-" + strings.TrimSuffix(file, ".env")
			if err := os.WriteFile(filepath.Join(dir, name+".env"), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			services.WriteString("  " + name + ":\n    image: scratch\n    env_file: [" + name + ".env]\n")
			all = append(all, rendered{env, file})
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte(services.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("docker", "compose", "--project-directory", dir, "-f", filepath.Join(dir, "compose.yaml"), "config", "--format", "json")
	cmd.Env = append(os.Environ(), "COMPOSE_PROJECT_NAME=leviosa-envcheck")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("docker compose config: %v", err)
	}
	var cfg struct {
		Services map[string]struct {
			Environment map[string]*string `json:"environment"`
		} `json:"services"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, r := range all {
		name := r.env + "-" + strings.TrimSuffix(r.file, ".env")
		content, _ := (&fakeInfisical{data: data}).render(t, r.file, r.env)
		want := values(t, content)
		got := map[string]string{}
		for k, v := range cfg.Services[name].Environment {
			if v != nil {
				// config output escapes a literal $ as $$
				got[k] = strings.ReplaceAll(*v, "$$", "$")
			} else {
				got[k] = ""
			}
		}
		if !reflect.DeepEqual(got, want) {
			for k := range want {
				if got[k] != want[k] {
					t.Errorf("%s %s: Compose reads %s differently from what was rendered", r.env, r.file, k)
				}
			}
			for k := range got {
				if _, ok := want[k]; !ok {
					t.Errorf("%s %s: Compose sees an extra key %s", r.env, r.file, k)
				}
			}
		}
	}
}

func toSet(list []string) map[string]string {
	m := map[string]string{}
	for _, k := range list {
		m[k] = ""
	}
	return m
}

func mapKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
