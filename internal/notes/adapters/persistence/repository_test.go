package persistence_test

import (
	"testing"

	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/adapters/persistence"
)

func TestANegativeLimitIsRefusedBeforeTheDatabase(t *testing.T) {
	if _, err := persistence.New(nil).List(t.Context(), -1); err == nil {
		t.Fatal("List(-1) succeeded")
	}
}
