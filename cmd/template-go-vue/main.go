// Command template-go-vue serves the API and the embedded web app, migrates its database at
// startup, and drains on SIGTERM.
//
// Environment: DATABASE_URL (required), ADDR (default :8080), and DEV_USER, which stands in for
// the platform's forward-auth on a local run and must never be set in a deployment.
package main

import (
	"context"
	"os"
)

// version is set at build time with -ldflags "-X main.version=…".
var version = "dev"

// main is os.Exit around start, which the tests drive; the coverage gate leaves this file out.
func main() {
	os.Exit(start(context.Background(), os.Getenv))
}
