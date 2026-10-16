package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JorisJonkers-dev/template-go-vue/internal/assembly"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
	"github.com/JorisJonkers-dev/template-go-vue/internal/server"
)

const (
	defaultAddr = ":8080"
	// shutdownGrace is how long in-flight requests get to finish after SIGTERM.
	shutdownGrace = 10 * time.Second
)

// start is main minus os.Exit, so tests can drive it.
func start(parent context.Context, getenv func(string) string) int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, logger, getenv); err != nil {
		logger.Error("template-go-vue stopped", "error", err)
		return 1
	}
	return 0
}

// run reads the environment, migrates and opens the database, and serves what assembly builds.
func run(ctx context.Context, logger *slog.Logger, getenv func(string) string) error {
	dbURL := getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("DATABASE_URL is not set")
	}
	if err := pg.Migrate(ctx, dbURL); err != nil {
		return err
	}
	store, err := pg.Open(ctx, dbURL)
	if err != nil {
		return err
	}
	defer store.Close()

	routes, err := assembly.Routes(assembly.Config{Logger: logger, Store: store, DevUser: getenv("DEV_USER")})
	if err != nil {
		return err
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr(getenv))
	if err != nil {
		return err
	}
	return server.New(logger, version, shutdownGrace, routes).Serve(ctx, ln)
}

// addr is where to listen: ADDR, or the default.
func addr(getenv func(string) string) string {
	if a := getenv("ADDR"); a != "" {
		return a
	}
	return defaultAddr
}
