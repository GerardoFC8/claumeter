package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GerardoFC8/claumeter/internal/server"
	"github.com/GerardoFC8/claumeter/internal/usage"
	"github.com/GerardoFC8/claumeter/internal/watch"
)

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	defaultRoot, _ := usage.DefaultProjectsDir()
	root := fs.String("root", defaultRoot, "directory with Claude Code JSONL transcripts")
	port := fs.Int("port", 7777, "TCP port to listen on")
	host := fs.String("host", "127.0.0.1", "host interface to bind (use 0.0.0.0 to expose externally — requires --token)")
	token := fs.String("token", "", "optional bearer token; required if --host is not a loopback address")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if *host != "127.0.0.1" && *host != "localhost" && *token == "" {
		fmt.Fprintln(os.Stderr, "refusing to expose claumeter without --token (host is non-loopback)")
		os.Exit(2)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	srv, err := server.New(server.Options{
		Root:    *root,
		Addr:    addr,
		Token:   *token,
		Version: version,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// File-watch for live updates. Non-fatal: if the watcher fails we still
	// serve static snapshots — just no SSE push.
	watcher, werr := watch.New(*root)
	if werr != nil {
		log.Printf("warning: file-watch disabled (%v); /live will only keepalive", werr)
	} else {
		defer watcher.Close()
		go runLiveFeed(ctx, watcher, srv)
	}

	fmt.Printf("claumeter %s serving %s on http://%s\n", version, *root, addr)
	if *token != "" {
		fmt.Println("auth: Bearer token required on every request except /healthz")
	}
	if watcher != nil {
		fmt.Println("file-watch: enabled (SSE /live will push updates)")
	}
	fmt.Println("press ctrl+c to stop")

	if err := srv.ListenAndServe(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runLiveFeed reloads the store after every debounced filesystem event and
// publishes a fresh `today` payload to SSE subscribers.
func runLiveFeed(ctx context.Context, watcher *watch.Watcher, srv *server.Server) {
	for range watcher.Events(ctx) {
		if err := srv.Store().Reload(); err != nil {
			log.Printf("reload error: %v", err)
			continue
		}
		srv.PublishToday()
	}
}
