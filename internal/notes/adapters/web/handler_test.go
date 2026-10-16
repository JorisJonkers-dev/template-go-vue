package web_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/adapters/web"
	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/domain"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/oas"
)

// useCases answers with notes or fails with err, and records the limit it was asked for.
type useCases struct {
	notes []domain.Note
	err   error
	limit int
}

func (u *useCases) Create(_ context.Context, raw string) (domain.Note, error) {
	if u.err != nil {
		return domain.Note{}, u.err
	}
	text, err := domain.NewText(raw)
	if err != nil {
		return domain.Note{}, err
	}
	return domain.Note{ID: uuid.New(), Text: text, CreatedAt: time.Date(2026, 10, 2, 9, 30, 0, 0, time.UTC)}, nil
}

func (u *useCases) List(_ context.Context, limit int) ([]domain.Note, error) {
	u.limit = limit
	return u.notes, u.err
}

var errBoom = errors.New("boom")

func TestACreatedNoteIsAnsweredWithItsIDTextAndTime(t *testing.T) {
	res, err := web.New(&useCases{}).CreateNote(t.Context(), &oas.NewNote{Text: " Buy milk "})
	note, ok := res.(*oas.Note)
	if err != nil || !ok || note.Text != "Buy milk" || note.ID == uuid.Nil || note.CreatedAt.IsZero() {
		t.Fatalf("CreateNote = %#v, %v", res, err)
	}
}

func TestTextTheDomainRefusesIsAnswered422WithTheRule(t *testing.T) {
	res, err := web.New(&useCases{}).CreateNote(t.Context(), &oas.NewNote{Text: "   "})
	refused, ok := res.(*oas.CreateNoteUnprocessableEntity)
	if err != nil || !ok || refused.StatusCode != http.StatusUnprocessableEntity || refused.Response.Detail.Value != domain.ErrInvalidText.Error() {
		t.Fatalf("CreateNote = %#v, %v", res, err)
	}
}

func TestACreateThatFailsIsAnErrorForTheServerToMap(t *testing.T) {
	if _, err := web.New(&useCases{err: errBoom}).CreateNote(t.Context(), &oas.NewNote{Text: "x"}); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestListedNotesAreAnsweredInTheOrderTheUseCaseGives(t *testing.T) {
	first, _ := domain.NewText("first")
	second, _ := domain.NewText("second")
	uc := &useCases{notes: []domain.Note{{Text: second}, {Text: first}}}
	res, err := web.New(uc).ListNotes(t.Context(), oas.ListNotesParams{})
	list, ok := res.(*oas.NoteList)
	if err != nil || !ok || len(list.Items) != 2 || list.Items[0].Text != "second" || list.Items[1].Text != "first" {
		t.Fatalf("ListNotes = %#v, %v", res, err)
	}
}

func TestAListWithoutALimitLeavesTheDefaultToTheUseCase(t *testing.T) {
	uc := &useCases{}
	if _, err := web.New(uc).ListNotes(t.Context(), oas.ListNotesParams{}); err != nil || uc.limit != 0 {
		t.Fatalf("limit = %d, %v", uc.limit, err)
	}
}

func TestAListPassesItsLimitToTheUseCase(t *testing.T) {
	uc := &useCases{}
	if _, err := web.New(uc).ListNotes(t.Context(), oas.ListNotesParams{Limit: oas.NewOptInt32(7)}); err != nil || uc.limit != 7 {
		t.Fatalf("limit = %d, %v", uc.limit, err)
	}
}

func TestAListThatFailsIsAnErrorForTheServerToMap(t *testing.T) {
	if _, err := web.New(&useCases{err: errBoom}).ListNotes(t.Context(), oas.ListNotesParams{}); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}
