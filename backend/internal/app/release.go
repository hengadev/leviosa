package app

// ReleaseSHA is the commit this binary was built from. It is injected at build
// time with:
//
//	go build -ldflags "-X github.com/Leviosa-care/leviosa/backend/internal/app.ReleaseSHA=<sha>"
//
// A binary built without it (go run, air, ad-hoc) reports "dev" instead of a
// misleading version. /health exposes it so deploys and Promote can verify
// what is actually serving.
var ReleaseSHA = "dev"
