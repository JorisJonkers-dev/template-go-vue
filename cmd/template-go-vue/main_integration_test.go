//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg/pgtest"
)

func TestAnAddressThatCannotBeListenedOnExitsWithOne(t *testing.T) {
	if code := start(t.Context(), env(map[string]string{"DATABASE_URL": pgtest.URL(t), "ADDR": "not-an-address"})); code != 1 {
		t.Fatalf("start = %d, want 1", code)
	}
}

func TestAServiceStoppedByItsContextExitsWithZero(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	vars := map[string]string{"DATABASE_URL": pgtest.URL(t), "ADDR": "127.0.0.1:0", "DEV_USER": "dev"}
	if code := start(ctx, env(vars)); code != 0 {
		t.Fatalf("start = %d, want 0", code)
	}
}
