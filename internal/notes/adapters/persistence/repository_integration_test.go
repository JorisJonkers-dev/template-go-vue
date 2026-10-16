//go:build integration

package persistence_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/adapters/persistence"
	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/domain"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg/pgtest"
)

func open(t *testing.T, url string) *pg.Store {
	t.Helper()
	store, err := pg.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	return store
}

func text(t *testing.T, raw string) domain.Text {
	t.Helper()
	v, err := domain.NewText(raw)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestACreatedNoteComesBackWithItsTextAndATimeInUTC(t *testing.T) {
	t.Parallel()
	created, err := persistence.New(open(t, pgtest.URL(t)).Queries()).Create(t.Context(), text(t, "first"))
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(created.CreatedAt) > time.Minute || created.CreatedAt.Location() != time.UTC || created.Text.String() != "first" {
		t.Fatalf("created %+v", created)
	}
}

func TestStoredNotesAreListedNewestFirstUpToTheLimit(t *testing.T) {
	t.Parallel()
	repo := persistence.New(open(t, pgtest.URL(t)).Queries())
	for _, raw := range []string{"first", "second", "third"} {
		if _, err := repo.Create(t.Context(), text(t, raw)); err != nil {
			t.Fatal(err)
		}
	}
	notes, err := repo.List(t.Context(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 || notes[0].Text.String() != "third" || notes[1].Text.String() != "second" {
		t.Fatalf("listed %+v, want third then second", notes)
	}
}

func TestAStoredRowTheDomainWouldRefuseFailsTheList(t *testing.T) {
	t.Parallel()
	url := pgtest.URL(t)
	conn, err := pgx.Connect(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close(t.Context()) }()
	if _, err := conn.Exec(t.Context(), "INSERT INTO notes (text) VALUES ('   ')"); err != nil {
		t.Fatal(err)
	}
	if _, err := persistence.New(open(t, url).Queries()).List(t.Context(), 10); !errors.Is(err, domain.ErrInvalidText) {
		t.Fatalf("List error = %v, want ErrInvalidText", err)
	}
}

// closed is a repository on a store that has been closed, so every query fails.
func closed(t *testing.T) *persistence.Repository {
	t.Helper()
	store, err := pg.Open(t.Context(), pgtest.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	store.Close()
	return persistence.New(store.Queries())
}

func TestCreatingANoteFailsWhenTheDatabaseDoes(t *testing.T) {
	t.Parallel()
	if _, err := closed(t).Create(t.Context(), text(t, "x")); err == nil {
		t.Fatal("Create on a closed pool succeeded")
	}
}

func TestListingNotesFailsWhenTheDatabaseDoes(t *testing.T) {
	t.Parallel()
	if _, err := closed(t).List(t.Context(), 1); err == nil {
		t.Fatal("List on a closed pool succeeded")
	}
}
