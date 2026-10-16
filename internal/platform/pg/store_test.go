package pg_test

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
)

const unreachable = "postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=1"

func TestOpeningAnUnreachableOrMalformedDatabaseFails(t *testing.T) {
	for _, url := range []string{unreachable, "::not a url::"} {
		t.Run(url, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			if _, err := pg.Open(ctx, url); err == nil {
				t.Fatal("Open succeeded")
			}
		})
	}
}

func TestMigratingAnUnreachableDatabaseFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	if err := pg.Migrate(ctx, unreachable); err == nil {
		t.Fatal("Migrate against an unreachable database succeeded")
	}
}

func TestMigratingWithoutAMigrationsDirectoryFails(t *testing.T) {
	if err := pg.MigrateFS(t.Context(), unreachable, fstest.MapFS{}); err == nil {
		t.Fatal("a filesystem without migrations/ was accepted")
	}
}
