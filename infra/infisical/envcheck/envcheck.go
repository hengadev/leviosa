// Package envcheck checks the env files the Infisical Agent templates render
// (../templates, issue 024): which files each Infisical environment produces,
// the key set each consumer must get, what it must never get, and the formats
// of the values the apps parse (database, RabbitMQ and Redis addresses).
//
// It reports key names only. No message ever contains a value: the same
// checks run on real renderings of staging and prod (../check-env.sh), so a
// value in a message would leak a secret into the operator's terminal.
package envcheck

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Environments are the Infisical environments of the `leviosa` project.
var Environments = []string{"dev", "staging", "prod"}

// files lists, per Infisical environment, the env files rendered from it.
// prod and staging are what each stack loads (infra/compose); dev is what
// `make env` writes (backend + admin -> backend/development.env, frontend ->
// frontend/.env) plus the three infrastructure files, rendered only so the
// check proves those templates against real folders before the server does.
// caddy.env and backup.env exist for prod only: Caddy belongs to production's
// stack (it reads staging's /caddy too) and /backup is prod-only.
var files = map[string][]string{
	"prod":    {"backend.env", "frontend.env", "admin.env", "redis.env", "postgres.env", "rabbitmq.env", "caddy.env", "backup.env"},
	"staging": {"backend.env", "frontend.env", "admin.env", "redis.env"},
	"dev":     {"backend.env", "frontend.env", "admin.env", "redis.env", "postgres.env", "rabbitmq.env"},
}

// Files returns the env files rendered from env, in a fixed order.
func Files(env string) ([]string, error) {
	f, ok := files[env]
	if !ok {
		return nil, fmt.Errorf("unknown environment %q (expected dev, staging or prod)", env)
	}
	return append([]string(nil), f...), nil
}

// Report is the result of checking one rendered file. Keys, Problems and
// Warnings never hold a value.
type Report struct {
	File     string
	Keys     []string // sorted key names
	Problems []string // the file is wrong: fix Infisical or the template
	Warnings []string // expected later in the rollout (e.g. provider keys)
}

// OK reports whether the file has no problem.
func (r Report) OK() bool { return len(r.Problems) == 0 }

type entry struct {
	key, value string
	line       int
}

var (
	lineRE    = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)='([^']*)'$`)
	keyRE     = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=`)
	alnumRE   = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	identRE   = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	vhostRE   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	bcryptRE  = regexp.MustCompile(`^\$2[aby]?\$[0-9]{2}\$[./A-Za-z0-9]{53}$`)
	digitsRE  = regexp.MustCompile(`^[0-9]+$`)
	commentRE = regexp.MustCompile(`^\s*(#.*)?$`)
)

// parse reads a rendered env file. Every line is blank, a # comment, or
// KEY='value' with no single quote and no line break inside the value — the
// only form Compose, systemd, sh, godotenv and dotenv all read literally
// (an unquoted $ is expanded by Compose, an unquoted " #" starts a comment).
func parse(content string) ([]entry, []string) {
	var entries []entry
	var problems []string
	for i, line := range strings.Split(content, "\n") {
		n := i + 1
		if commentRE.MatchString(line) {
			continue
		}
		if m := lineRE.FindStringSubmatch(line); m != nil {
			entries = append(entries, entry{key: m[1], value: m[2], line: n})
			continue
		}
		if m := keyRE.FindStringSubmatch(line); m != nil {
			problems = append(problems, fmt.Sprintf("%s (line %d): the value is not single-quoted, or contains a single quote or a line break; such a value cannot be written safely, change it in Infisical", m[1], n))
			continue
		}
		problems = append(problems, fmt.Sprintf("line %d is neither a comment nor KEY='value'", n))
	}
	return entries, problems
}

// rule describes one env file.
type rule struct {
	// required keys must be present with a non-empty value.
	required []string
	// later keys are expected in staging and prod once their provider is set
	// up (issues 034-036); their absence is a warning, not a problem.
	later []string
	// only, when set, is the complete list of allowed keys.
	only []string
	// forbidden returns why key must not be in the file, or "".
	forbidden func(env, key string) string
	// values checks the formats of present values.
	values func(env string, v map[string]string) []string
}

var backendRequired = []string{
	"ENVIRONMENT", "SERVER_PORT", "LOG_FORMAT",
	"DATABASE_URL", "REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB", "RABBITMQ_URL",
	"ENCX_KMS_KEY_ALIAS", "ENCX_PEPPER",
	"AWS_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
	"S3_REGION", "S3_BUCKET_NAME", "S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY",
	"SESSION_SECRET", "BOOKING_TOKEN_SECRET", "FRONTEND_ORIGIN",
}

var backendLater = []string{
	"STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET", "STRIPE_CONNECT_WEBHOOK_SECRET",
	"TWILIO_ACCOUNT_SID", "TWILIO_API_KEY_SID", "TWILIO_API_KEY_SECRET", "TWILIO_REGION", "TWILIO_EDGE", "TWILIO_SENDER",
	"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD",
}

// FrontendKeys is the complete allowlist of frontend.env: what the SvelteKit
// server (and adapter-node) reads, nothing else.
var FrontendKeys = []string{
	"NODE_ENV", "PORT", "API_URL", "PROTOCOL_HEADER", "HOST_HEADER", "USE_MOCK_DATA",
	"COMING_SOON", "SESSION_COOKIE_NAME", "CLIENT_IP_HEADER",
	"PUBLIC_PLAUSIBLE_DOMAIN", "PUBLIC_PLAUSIBLE_SRC",
}

var adminRequired = []string{
	"ADMIN_EMAIL", "ADMIN_PASSWORD", "ADMIN_FIRST_NAME", "ADMIN_LAST_NAME",
	"ADMIN_TELEPHONE", "ADMIN_POSTAL_CODE", "ADMIN_CITY", "ADMIN_ADDRESS1",
}

// Hosts each stack reaches its services at (service names on
// leviosa_network, infra/compose). dev has none: it uses whatever the
// developer's machine exposes.
var expectedHosts = map[string]struct{ postgres, rabbitmq, redis string }{
	"prod":    {"postgres", "rabbitmq", "redis"},
	"staging": {"postgres", "rabbitmq", "staging_redis"},
}

var expectedEnvironment = map[string]string{"prod": "production", "staging": "staging", "dev": "development"}

var expectedAPIURL = map[string]string{
	"prod":    "http://backend:3500",
	"staging": "http://staging_backend:3500",
	"dev":     "http://localhost:3500",
}

var expectedNodeEnv = map[string]string{"prod": "production", "staging": "production", "dev": "development"}

var rules = map[string]rule{
	"backend.env": {
		required:  backendRequired,
		later:     backendLater,
		forbidden: backendForbidden,
		values:    backendValues,
	},
	"frontend.env": {
		required: []string{"NODE_ENV", "PORT", "API_URL", "PROTOCOL_HEADER", "HOST_HEADER", "USE_MOCK_DATA", "COMING_SOON", "CLIENT_IP_HEADER"},
		only:     FrontendKeys,
		values:   frontendValues,
	},
	"admin.env": {
		required: adminRequired,
		forbidden: func(_, key string) string {
			if !strings.HasPrefix(key, "ADMIN_") {
				return "admin.env holds the /admin folder only"
			}
			return ""
		},
	},
	"redis.env": {
		required: []string{"REDIS_PASSWORD"},
		only:     []string{"REDIS_PASSWORD"},
	},
	"postgres.env": {
		required: []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"},
		only:     []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"},
		values: func(_ string, v map[string]string) []string {
			return credentialRules("POSTGRES_USER", v["POSTGRES_USER"], "POSTGRES_PASSWORD", v["POSTGRES_PASSWORD"])
		},
	},
	"rabbitmq.env": {
		required: []string{"RABBITMQ_DEFAULT_USER", "RABBITMQ_DEFAULT_PASS", "RABBITMQ_DEFAULT_VHOST"},
		only:     []string{"RABBITMQ_DEFAULT_USER", "RABBITMQ_DEFAULT_PASS", "RABBITMQ_DEFAULT_VHOST"},
		values: func(_ string, v map[string]string) []string {
			return credentialRules("RABBITMQ_DEFAULT_USER", v["RABBITMQ_DEFAULT_USER"], "RABBITMQ_DEFAULT_PASS", v["RABBITMQ_DEFAULT_PASS"])
		},
	},
	"caddy.env": {
		required: []string{"CLOUDFLARE_API_TOKEN", "STAGING_BASIC_AUTH_USER", "STAGING_BASIC_AUTH_HASH"},
		values: func(_ string, v map[string]string) []string {
			if h, ok := v["STAGING_BASIC_AUTH_HASH"]; ok && h != "" && !bcryptRE.MatchString(h) {
				return []string{"STAGING_BASIC_AUTH_HASH is not a bcrypt hash ($2a$<cost>$<53 characters>, e.g. from `caddy hash-password`)"}
			}
			return nil
		},
	},
	"backup.env": {
		required: []string{"BACKUP_AWS_ACCESS_KEY_ID", "BACKUP_AWS_SECRET_ACCESS_KEY", "BACKUP_S3_BUCKET", "BACKUP_AWS_REGION", "POSTGRES_USER", "POSTGRES_DB"},
		forbidden: func(_, key string) string {
			if key == "POSTGRES_PASSWORD" {
				return "the dump runs through docker exec and needs no password"
			}
			return ""
		},
	},
}

// Check checks one rendered file of environment env. file is its name
// (backend.env, frontend.env, ...).
func Check(env, file string, content []byte) Report {
	r := Report{File: file}
	if _, ok := files[env]; !ok {
		r.Problems = append(r.Problems, fmt.Sprintf("unknown environment %q", env))
		return r
	}
	ru, ok := rules[file]
	if !ok {
		r.Problems = append(r.Problems, fmt.Sprintf("no rule for %s", file))
		return r
	}

	entries, problems := parse(string(content))
	r.Problems = append(r.Problems, problems...)

	values := map[string]string{}
	count := map[string]int{}
	for _, e := range entries {
		count[e.key]++
		values[e.key] = e.value
	}
	for key := range count {
		r.Keys = append(r.Keys, key)
	}
	sort.Strings(r.Keys)

	for _, key := range r.Keys {
		if count[key] > 1 {
			r.Problems = append(r.Problems, fmt.Sprintf("%s is set %d times (Compose keeps the last one): it exists in two folders or in Infisical and the template; keep one", key, count[key]))
		}
	}

	for _, key := range ru.required {
		v, present := values[key]
		switch {
		case !present:
			r.Problems = append(r.Problems, key+" is missing")
		case strings.TrimSpace(v) == "":
			r.Problems = append(r.Problems, key+" is empty")
		}
	}
	if env != "dev" {
		for _, key := range ru.later {
			if v, present := values[key]; !present || v == "" {
				r.Warnings = append(r.Warnings, key+" is not set yet (needed before that provider is used)")
			}
		}
	}

	if ru.only != nil {
		allowed := map[string]bool{}
		for _, k := range ru.only {
			allowed[k] = true
		}
		for _, key := range r.Keys {
			if allowed[key] {
				continue
			}
			if file == "frontend.env" && key == "ORIGIN" {
				r.Problems = append(r.Problems, "ORIGIN must never be set: adapter-node derives the origin per host from PROTOCOL_HEADER and HOST_HEADER")
				continue
			}
			r.Problems = append(r.Problems, fmt.Sprintf("%s must not be in %s (allowed: %s)", key, file, strings.Join(ru.only, ", ")))
		}
	}
	if ru.forbidden != nil {
		for _, key := range r.Keys {
			if why := ru.forbidden(env, key); why != "" {
				r.Problems = append(r.Problems, fmt.Sprintf("%s must not be in %s: %s", key, file, why))
			}
		}
	}
	if ru.values != nil {
		r.Problems = append(r.Problems, ru.values(env, values)...)
	}
	return r
}

func backendForbidden(env, key string) string {
	for _, p := range []string{"ADMIN_", "BACKUP_", "CLOUDFLARE_", "STAGING_BASIC_AUTH_", "POSTGRES_", "RABBITMQ_DEFAULT_"} {
		if strings.HasPrefix(key, p) {
			return "it belongs to another consumer's folder (/admin, /backup, /caddy) or is a raw /db atom"
		}
	}
	switch key {
	case "RABBITMQ_USER", "RABBITMQ_PASSWORD", "RABBITMQ_VHOST", "RABBITMQ_HOST", "RABBITMQ_PORT":
		return "the /rabbitmq atoms reach the backend only inside RABBITMQ_URL"
	case "REDIS_HOST", "REDIS_PORT":
		return "the /redis address atoms reach the backend only inside REDIS_ADDR"
	case "AWS_ENDPOINT_URL", "S3_ENDPOINT":
		if env != "dev" {
			return "LocalStack only; it would point KMS or S3 away from AWS"
		}
	}
	return ""
}

func backendValues(env string, v map[string]string) []string {
	var p []string
	if got, ok := v["ENVIRONMENT"]; ok && got != expectedEnvironment[env] {
		p = append(p, fmt.Sprintf("ENVIRONMENT must be %s for Infisical %s", expectedEnvironment[env], env))
	}
	if got, ok := v["SERVER_PORT"]; ok && got != "3500" {
		p = append(p, "SERVER_PORT must be 3500 (the compose healthchecks and API_URL use it)")
	}
	if got, ok := v["REDIS_DB"]; ok && !digitsRE.MatchString(got) {
		p = append(p, "REDIS_DB must be a number")
	}
	hosts, pinned := expectedHosts[env]
	if raw, ok := v["DATABASE_URL"]; ok && raw != "" {
		p = append(p, databaseURLProblems(raw, hosts.postgres, pinned)...)
	}
	if raw, ok := v["RABBITMQ_URL"]; ok && raw != "" {
		p = append(p, rabbitmqURLProblems(raw, hosts.rabbitmq, pinned)...)
	}
	if raw, ok := v["REDIS_ADDR"]; ok && raw != "" {
		p = append(p, redisAddrProblems(raw, hosts.redis, pinned)...)
	}
	if alias, ok := v["ENCX_KMS_KEY_ALIAS"]; ok && alias != "" && (!strings.HasPrefix(alias, "alias/") || len(alias) == len("alias/")) {
		p = append(p, "ENCX_KMS_KEY_ALIAS must be a KMS alias name (alias/<name>), not a key ID or ARN")
	}
	if pepper, ok := v["ENCX_PEPPER"]; ok && pepper != "" {
		if b, err := base64.StdEncoding.DecodeString(pepper); err != nil || len(b) != 32 {
			p = append(p, "ENCX_PEPPER must be standard base64 of exactly 32 bytes")
		}
	}
	return p
}

func frontendValues(env string, v map[string]string) []string {
	var p []string
	want := map[string]string{
		"API_URL":         expectedAPIURL[env],
		"NODE_ENV":        expectedNodeEnv[env],
		"PORT":            "3000",
		"PROTOCOL_HEADER": "x-forwarded-proto",
		"HOST_HEADER":     "host",
		"USE_MOCK_DATA":   "false",
	}
	for _, key := range []string{"API_URL", "NODE_ENV", "PORT", "PROTOCOL_HEADER", "HOST_HEADER", "USE_MOCK_DATA"} {
		if got, ok := v[key]; ok && got != want[key] {
			p = append(p, fmt.Sprintf("%s must be %s for Infisical %s", key, want[key], env))
		}
	}
	if got, ok := v["COMING_SOON"]; ok && got != "true" && got != "false" {
		p = append(p, "COMING_SOON must be true or false")
	}
	return p
}

// credentialRules: users are identifiers, passwords letters and digits only
// (PRD: passwords that end up in a URL; Go's urlquery would turn a space
// into "+", which is wrong inside userinfo).
func credentialRules(userKey, user, passKey, pass string) []string {
	var p []string
	if user != "" && !identRE.MatchString(user) {
		p = append(p, userKey+" must be letters, digits and underscores only")
	}
	if pass != "" && !alnumRE.MatchString(pass) {
		p = append(p, passKey+" must be letters and digits only (it is embedded in a URL)")
	}
	return p
}

func validPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && digitsRE.MatchString(s) && n > 0 && n < 65536
}

// databaseURLProblems: postgres://<user>:<password>@<host>:<port>/<db>?sslmode=disable,
// the form pgxpool.ParseConfig reads (the Postgres container has no TLS).
func databaseURLProblems(raw, wantHost string, pinned bool) []string {
	const k = "DATABASE_URL"
	u, err := url.Parse(raw)
	if err != nil {
		return []string{k + " does not parse as a URL"}
	}
	var p []string
	if u.Scheme != "postgres" {
		p = append(p, k+" must start with postgres://")
	}
	p = append(p, userinfoProblems(k, u)...)
	p = append(p, hostPortProblems(k, u.Host, wantHost, pinned)...)
	if db := strings.TrimPrefix(u.Path, "/"); !identRE.MatchString(db) {
		p = append(p, k+" must name a database (/<letters, digits, underscores>)")
	}
	if u.Query().Get("sslmode") != "disable" || len(u.Query()) != 1 {
		p = append(p, k+" must end with ?sslmode=disable and nothing else")
	}
	return p
}

// rabbitmqURLProblems: amqp://<user>:<password>@<host>:<port>/<vhost>, the
// vhost always explicit ("/" written as %2F), as amqp091-go reads it.
func rabbitmqURLProblems(raw, wantHost string, pinned bool) []string {
	const k = "RABBITMQ_URL"
	u, err := url.Parse(raw)
	if err != nil {
		return []string{k + " does not parse as a URL"}
	}
	var p []string
	if u.Scheme != "amqp" {
		p = append(p, k+" must start with amqp://")
	}
	p = append(p, userinfoProblems(k, u)...)
	p = append(p, hostPortProblems(k, u.Host, wantHost, pinned)...)
	vhost := strings.TrimPrefix(u.Path, "/")
	if vhost != "/" && !vhostRE.MatchString(vhost) {
		p = append(p, k+" must end with an explicit vhost (/<name>, or /%2F for the default vhost)")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		p = append(p, k+" must not carry a query or fragment")
	}
	return p
}

func redisAddrProblems(raw, wantHost string, pinned bool) []string {
	return hostPortProblems("REDIS_ADDR", raw, wantHost, pinned)
}

func userinfoProblems(k string, u *url.URL) []string {
	if u.User == nil {
		return []string{k + " has no user and password"}
	}
	var p []string
	if !identRE.MatchString(u.User.Username()) {
		p = append(p, k+" user must be letters, digits and underscores only")
	}
	pass, ok := u.User.Password()
	if !ok || pass == "" {
		p = append(p, k+" has no password")
	} else if !alnumRE.MatchString(pass) {
		p = append(p, k+" password must be letters and digits only")
	}
	return p
}

func hostPortProblems(k, hostport, wantHost string, pinned bool) []string {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil || host == "" {
		return []string{k + " must contain <host>:<port>"}
	}
	var p []string
	if !validPort(port) {
		p = append(p, k+" port must be a number between 1 and 65535")
	}
	if pinned && host != wantHost {
		p = append(p, fmt.Sprintf("%s host must be %s (this stack's service name)", k, wantHost))
	}
	return p
}
