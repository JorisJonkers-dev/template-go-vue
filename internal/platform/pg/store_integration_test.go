//go:build integration

package pg_test

import (
	"testing"
	"testing/fstest"

	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg/pgtest"
)

func TestMigratingAnUpToDateDatabaseAgainChangesNothing(t *testing.T) {
	t.Parallel()
	if err := pg.Migrate(t.Context(), pgtest.URL(t)); err != nil {
		t.Fatal(err)
	}
}

func TestAnOpenStoreAnswersAndHasTheNotesTable(t *testing.T) {
	t.Parallel()
	store, err := pg.Open(t.Context(), pgtest.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Queries().ListNotes(t.Context(), 1); err != nil {
		t.Fatalf("the migrated schema has no notes table: %v", err)
	}
}

func TestAnEmptyMigrationSetIsRefused(t *testing.T) {
	t.Parallel()
	if err := pg.MigrateFS(t.Context(), pgtest.URL(t), fstest.MapFS{"migrations/.keep": {}}); err == nil {
		t.Fatal("an empty migration set was accepted")
	}
}
