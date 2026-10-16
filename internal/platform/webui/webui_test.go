package webui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/webui"
)

var app = fstest.MapFS{
	"index.html":        {Data: []byte("<!doctype html><title>app</title>")},
	"assets/app-abc.js": {Data: []byte("console.log(1)")},
	"favicon.svg":       {Data: []byte("<svg></svg>")},
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
	return rec
}

func TestTheRootAndEveryClientRouteGetTheAppUncached(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"/", "/index.html", "/notes/42", "/settings", "/assets"} {
		t.Run(p, func(t *testing.T) {
			rec := get(webui.Handler(app), p)
			if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-cache" || rec.Body.String() != "<!doctype html><title>app</title>" {
				t.Fatalf("%d %q %q", rec.Code, rec.Header().Get("Cache-Control"), rec.Body.String())
			}
		})
	}
}

func TestAHashedAssetIsCachedForever(t *testing.T) {
	t.Parallel()
	rec := get(webui.Handler(app), "/assets/app-abc.js")
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("asset: %d %v", rec.Code, rec.Header())
	}
}

func TestAFileOutsideTheAssetsIsServedWithoutACachePolicy(t *testing.T) {
	t.Parallel()
	if rec := get(webui.Handler(app), "/favicon.svg"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "" {
		t.Fatalf("root file: %d %v", rec.Code, rec.Header())
	}
}

func TestAMissingFileIsNotFound(t *testing.T) {
	t.Parallel()
	if rec := get(webui.Handler(app), "/assets/missing.js"); rec.Code != 404 {
		t.Fatalf("missing asset: %d", rec.Code)
	}
}

func TestWithoutABuildEveryPageIsNotFound(t *testing.T) {
	t.Parallel()
	rec := get(webui.Handler(fstest.MapFS{}), "/")
	if rec.Code != 404 {
		t.Fatalf("unbuilt: %d", rec.Code)
	}
}
