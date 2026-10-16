//go:build integration

// Package notes_test checks the notes context's behaviours the way a browser reaches them: over
// HTTP, against the real routes and Postgres (workspace docs/testing-standards.md).
package notes_test

import (
	"net/http"
	"testing"

	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/itest"
)

type note struct{ ID, Text string }

type notes struct{ Items []note }

type problem struct {
	Status int
	Detail string
}

func TestANoteIsStoredWithoutItsSurroundingSpaces(t *testing.T) {
	t.Parallel()
	svc := itest.Start(t)
	var created note
	svc.Call(t, "ada", http.MethodPost, "/api/v1/notes", map[string]string{"text": "  Buy milk \n"}, http.StatusCreated, &created)
	var listed notes
	svc.Call(t, "ada", http.MethodGet, "/api/v1/notes", nil, http.StatusOK, &listed)
	if created.Text != "Buy milk" || len(listed.Items) != 1 || listed.Items[0] != created {
		t.Fatalf("created %+v, listed %+v", created, listed.Items)
	}
}

func TestNotesAreListedNewestFirst(t *testing.T) {
	t.Parallel()
	svc := itest.Start(t)
	for _, text := range []string{"first", "second"} {
		svc.Call(t, "ada", http.MethodPost, "/api/v1/notes", map[string]string{"text": text}, http.StatusCreated, nil)
	}
	var listed notes
	svc.Call(t, "ada", http.MethodGet, "/api/v1/notes", nil, http.StatusOK, &listed)
	if len(listed.Items) != 2 || listed.Items[0].Text != "second" || listed.Items[1].Text != "first" {
		t.Fatalf("listed %+v, want second then first", listed.Items)
	}
}

func TestAListStopsAtItsLimit(t *testing.T) {
	t.Parallel()
	svc := itest.Start(t)
	for _, text := range []string{"first", "second"} {
		svc.Call(t, "ada", http.MethodPost, "/api/v1/notes", map[string]string{"text": text}, http.StatusCreated, nil)
	}
	var listed notes
	svc.Call(t, "ada", http.MethodGet, "/api/v1/notes?limit=1", nil, http.StatusOK, &listed)
	if len(listed.Items) != 1 || listed.Items[0].Text != "second" {
		t.Fatalf("listed %+v, want only the newest", listed.Items)
	}
}

func TestANoteOfOnlySpacesIsRefusedAndNotStored(t *testing.T) {
	t.Parallel()
	svc := itest.Start(t)
	var refused problem
	svc.Call(t, "ada", http.MethodPost, "/api/v1/notes", map[string]string{"text": "   "}, http.StatusUnprocessableEntity, &refused)
	var listed notes
	svc.Call(t, "ada", http.MethodGet, "/api/v1/notes", nil, http.StatusOK, &listed)
	if refused.Detail != "a note needs between 1 and 500 characters of text" || len(listed.Items) != 0 {
		t.Fatalf("refused %+v, listed %+v", refused, listed.Items)
	}
}

func TestTheNotesNeedAnIdentity(t *testing.T) {
	t.Parallel()
	itest.Start(t).Call(t, "", http.MethodGet, "/api/v1/notes", nil, http.StatusUnauthorized, nil)
}
