package main

import (
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

const unreachable = "postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=1"

func TestWithoutADatabaseURLTheServiceExitsWithOne(t *testing.T) {
	if code := start(t.Context(), env(nil)); code != 1 {
		t.Fatalf("start = %d, want 1", code)
	}
}

func TestAnUnreachableDatabaseExitsWithOne(t *testing.T) {
	if code := start(t.Context(), env(map[string]string{"DATABASE_URL": unreachable})); code != 1 {
		t.Fatalf("start = %d, want 1", code)
	}
}

func TestWithoutAnAddressTheServiceListensOnTheDefault(t *testing.T) {
	if got := addr(env(nil)); got != defaultAddr {
		t.Fatalf("addr = %q, want %q", got, defaultAddr)
	}
}
