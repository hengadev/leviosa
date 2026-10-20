package envcheck

import (
	"strings"
	"testing"
)

// The checker itself: it must catch what a wrong Infisical value or a wrong
// template would produce, and never print a value while doing so.

// render a valid fixture file, then let mutate change it.
func fixtureFile(t *testing.T, env, file string) string {
	t.Helper()
	content, _ := (&fakeInfisical{data: loadFixtures(t)}).render(t, file, env)
	return content
}

func replaceLine(content, key, line string) string {
	var out []string
	for _, l := range strings.Split(content, "\n") {
		if strings.HasPrefix(l, key+"=") {
			if line == "" {
				continue
			}
			l = line
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func problems(t *testing.T, env, file, content string) []string {
	t.Helper()
	r := Check(env, file, []byte(content))
	return r.Problems
}

func expectProblem(t *testing.T, env, file, content, want string, secrets ...string) {
	t.Helper()
	ps := problems(t, env, file, content)
	found := false
	for _, p := range ps {
		if strings.Contains(p, want) {
			found = true
		}
		for _, s := range secrets {
			if strings.Contains(p, s) {
				t.Errorf("problem %q leaks a value", p)
			}
		}
	}
	if !found {
		t.Errorf("%s %s: no problem mentioning %q in %v", env, file, want, ps)
	}
}

func TestCheckFrontendExclusions(t *testing.T) {
	base := fixtureFile(t, "prod", "frontend.env")
	expectProblem(t, "prod", "frontend.env", base+"ORIGIN='https://leviosa.care'\n", "ORIGIN must never be set")
	for _, key := range []string{"AWS_SECRET_ACCESS_KEY", "ENCX_PEPPER", "STRIPE_SECRET_KEY", "TWILIO_API_KEY_SECRET", "SMTP_PASSWORD", "DATABASE_URL", "ADMIN_PASSWORD", "SESSION_SECRET", "BOOKING_TOKEN_SECRET"} {
		expectProblem(t, "prod", "frontend.env", base+key+"='s3cr3tvalue'\n", key+" must not be in frontend.env", "s3cr3tvalue")
	}
	expectProblem(t, "prod", "frontend.env", replaceLine(base, "USE_MOCK_DATA", "USE_MOCK_DATA='true'"), "USE_MOCK_DATA must be false")
	expectProblem(t, "staging", "frontend.env", replaceLine(fixtureFile(t, "staging", "frontend.env"), "API_URL", "API_URL='http://backend:3500'"), "API_URL must be http://staging_backend:3500")
	expectProblem(t, "prod", "frontend.env", replaceLine(base, "COMING_SOON", ""), "COMING_SOON is missing")
}

func TestCheckURLFormats(t *testing.T) {
	base := fixtureFile(t, "prod", "backend.env")
	cases := map[string]string{
		"DATABASE_URL='postgresql://leviosa:Sup3rSecret@postgres:5432/leviosa?sslmode=disable'": "must start with postgres://",
		"DATABASE_URL='postgres://leviosa:Sup3rSecret@postgres:5432/leviosa'":                   "sslmode=disable",
		"DATABASE_URL='postgres://leviosa:Sup%2BSecret@postgres:5432/leviosa?sslmode=disable'":  "password must be letters and digits only",
		"DATABASE_URL='postgres://leviosa:Sup3rSecret@postgres:/leviosa?sslmode=disable'":       "port must be a number",
		"DATABASE_URL='postgres://leviosa:Sup3rSecret@postgres:5432/?sslmode=disable'":          "must name a database",
		"DATABASE_URL='postgres://leviosa:Sup3rSecret@db.example:5432/leviosa?sslmode=disable'": "host must be postgres",
		"DATABASE_URL='postgres://leviosa@postgres:5432/leviosa?sslmode=disable'":               "has no password",
		"RABBITMQ_URL='amqp://leviosa:Sup3rSecret@rabbitmq:5672/'":                              "explicit vhost",
		"RABBITMQ_URL='amqps://leviosa:Sup3rSecret@rabbitmq:5672/leviosa'":                      "must start with amqp://",
		"RABBITMQ_URL='amqp://leviosa:Sup3rSecret@rabbitmq/leviosa'":                            "<host>:<port>",
		"REDIS_ADDR='redis'":                         "<host>:<port>",
		"REDIS_ADDR='redis:http'":                    "port must be a number",
		"ENCX_KMS_KEY_ALIAS='arn:aws:kms:x:y:key/z'": "KMS alias name",
		"ENCX_PEPPER='c2hvcnQ='":                     "exactly 32 bytes",
		"ENVIRONMENT='development'":                  "ENVIRONMENT must be production",
	}
	for line, want := range cases {
		key := line[:strings.Index(line, "=")]
		expectProblem(t, "prod", "backend.env", replaceLine(base, key, line), want, "Sup3rSecret", "Sup%2BSecret")
	}

	// Staging must reach its own Redis, never production's.
	staging := fixtureFile(t, "staging", "backend.env")
	expectProblem(t, "staging", "backend.env", replaceLine(staging, "REDIS_ADDR", "REDIS_ADDR='redis:6379'"), "REDIS_ADDR host must be staging_redis")
	// dev pins no host.
	dev := fixtureFile(t, "dev", "backend.env")
	if ps := problems(t, "dev", "backend.env", replaceLine(dev, "REDIS_ADDR", "REDIS_ADDR='127.0.0.1:6379'")); len(ps) != 0 {
		t.Errorf("dev: %v", ps)
	}
}

func TestCheckBackendBoundaries(t *testing.T) {
	base := fixtureFile(t, "prod", "backend.env")
	expectProblem(t, "prod", "backend.env", base+"ADMIN_PASSWORD='Adm1nSecret'\n", "ADMIN_PASSWORD must not be in backend.env", "Adm1nSecret")
	expectProblem(t, "prod", "backend.env", base+"POSTGRES_PASSWORD='Pg5ecret'\n", "POSTGRES_PASSWORD must not be in backend.env", "Pg5ecret")
	expectProblem(t, "prod", "backend.env", base+"CLOUDFLARE_API_TOKEN='cf'\n", "CLOUDFLARE_API_TOKEN must not be in backend.env")
	expectProblem(t, "prod", "backend.env", base+"AWS_ENDPOINT_URL='http://localstack:4566'\n", "AWS_ENDPOINT_URL must not be in backend.env")
	expectProblem(t, "prod", "backend.env", replaceLine(base, "ENCX_PEPPER", ""), "ENCX_PEPPER is missing")
	expectProblem(t, "prod", "backend.env", replaceLine(base, "SESSION_SECRET", "SESSION_SECRET=''"), "SESSION_SECRET is empty")
	expectProblem(t, "prod", "backend.env", base+"ENVIRONMENT='production'\n", "ENVIRONMENT is set 2 times")

	// Providers arrive late in the rollout: a warning, not a problem.
	r := Check("staging", "backend.env", []byte(replaceLine(fixtureFile(t, "staging", "backend.env"), "TWILIO_SENDER", "")))
	if !r.OK() || len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "TWILIO_SENDER") {
		t.Errorf("missing provider key: problems %v warnings %v", r.Problems, r.Warnings)
	}
}

func TestCheckFileFormatNeverLeaksValues(t *testing.T) {
	base := fixtureFile(t, "prod", "admin.env")
	for _, line := range []string{
		"ADMIN_PASSWORD='it's-a-Secret1'", // a single quote cannot be written
		"ADMIN_PASSWORD=Unquoted$Secret1", // Compose would expand the $
		"ADMIN_PASSWORD='multi\nSecret1'", // a line break cannot be written
	} {
		expectProblem(t, "prod", "admin.env", replaceLine(base, "ADMIN_PASSWORD", line), "ADMIN_PASSWORD (line", "Secret1", "it's", "Unquoted")
	}
	expectProblem(t, "prod", "admin.env", base+"garbage line\n", "neither a comment nor")
	expectProblem(t, "prod", "admin.env", base+"OTHER='x'\n", "admin.env holds the /admin folder only")

	caddy := fixtureFile(t, "prod", "caddy.env")
	expectProblem(t, "prod", "caddy.env", replaceLine(caddy, "STAGING_BASIC_AUTH_HASH", "STAGING_BASIC_AUTH_HASH='plaintext-password'"), "not a bcrypt hash", "plaintext-password")
	expectProblem(t, "prod", "caddy.env", replaceLine(caddy, "CLOUDFLARE_API_TOKEN", ""), "CLOUDFLARE_API_TOKEN is missing")

	expectProblem(t, "prod", "postgres.env", replaceLine(fixtureFile(t, "prod", "postgres.env"), "POSTGRES_PASSWORD", "POSTGRES_PASSWORD='with space1'"), "letters and digits only", "with space1")
	expectProblem(t, "prod", "redis.env", fixtureFile(t, "prod", "redis.env")+"REDIS_HOST='redis'\n", "REDIS_HOST must not be in redis.env")
	expectProblem(t, "prod", "backup.env", fixtureFile(t, "prod", "backup.env")+"POSTGRES_PASSWORD='p'\n", "POSTGRES_PASSWORD must not be in backup.env")
}

func TestFilesRejectsUnknownEnvironment(t *testing.T) {
	if _, err := Files("production"); err == nil {
		t.Error("Files(production) succeeded")
	}
	if r := Check("production", "backend.env", nil); r.OK() {
		t.Error("Check accepted an unknown environment")
	}
}
