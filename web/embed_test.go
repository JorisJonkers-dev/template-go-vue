package web_test

import (
	"io/fs"
	"testing"

	"github.com/JorisJonkers-dev/template-go-vue/web"
)

func TestTheEmbeddedBuildIsServedFromItsOwnRoot(t *testing.T) {
	if _, err := fs.Stat(web.Dist(), ".gitkeep"); err != nil {
		t.Fatalf("dist/.gitkeep is not at the root: %v", err)
	}
}
