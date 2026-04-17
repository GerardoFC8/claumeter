// Package server exposes claumeter's parsed usage over a small HTTP API.
// Widgets (waybar, eww, sketchybar, …) and dashboards consume these endpoints
// instead of re-parsing the JSONL transcripts.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GerardoFC8/claumeter/internal/export"
	"github.com/GerardoFC8/claumeter/internal/stats"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

// Store caches parsed usage data and lets handlers query it without touching
// the filesystem on every request. Reload() refreshes it.
type Store struct {
	mu   sync.RWMutex
	data usage.Data
	root string
}

func NewStore(root string) (*Store, error) {
	s := &Store{root: root}
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Reload() error {
	d, err := usage.ParseAll(s.root, nil)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.data = d
	s.mu.Unlock()
	return nil
}

func (s *Store) Data() usage.Data {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *Store) Root() string { return s.root }

// Options configures a Server.
type Options struct {
	Root    string
	Addr    string // e.g. "127.0.0.1:7777"
	Token   string // optional bearer; empty = open
	Version string // passed into /healthz
}

type Server struct {
	store *Server_Store // alias below for external reuse of *Store
	opts  Options
	http  *http.Server
}

// Server_Store is a type alias so callers can pre-build the store (e.g. to
// share with a file-watch loop).
type Server_Store = Store

func New(opts Options) (*Server, error) {
	store, err := NewStore(opts.Root)
	if err != nil {
		return nil, err
	}
	return NewWithStore(opts, store), nil
}

func NewWithStore(opts Options, store *Store) *Server {
	s := &Server{store: store, opts: opts}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /today", s.today)
	mux.HandleFunc("GET /stats", s.stats)
	mux.HandleFunc("GET /range", s.rangeHandler)
	mux.HandleFunc("GET /session/{id}", s.session)
	s.http = &http.Server{
		Addr:              opts.Addr,
		Handler:           s.auth(withLogging(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Store() *Store { return s.store }

// ListenAndServe runs until the context is cancelled or the server errors.
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() { errCh <- s.http.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) auth(next http.Handler) http.Handler {
	if s.opts.Token == "" {
		return next
	}
	want := "Bearer " + s.opts.Token
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("Authorization") != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.RequestURI(), time.Since(start))
	})
}

// --- handlers ---

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": s.opts.Version,
		"root":    s.opts.Root,
		"events":  len(s.store.Data().Events),
	})
}

func (s *Server) today(w http.ResponseWriter, r *http.Request) {
	filtered := stats.FilterToday.Apply(s.store.Data())
	report := stats.Build(filtered)
	from, to := stats.FilterToday.Range(time.Now())
	payload := export.NewCompact(stats.FilterToday.Label(), from, to, report)
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	rangeParam := r.URL.Query().Get("range")
	label, from, to, filtered, err := s.resolveAndApply(rangeParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	report := stats.Build(filtered)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = export.ToJSON(w, label, from, to, report)
}

func (s *Server) rangeHandler(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	if fromStr == "" {
		http.Error(w, "from is required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	arg := fromStr
	if to := r.URL.Query().Get("to"); to != "" {
		arg = fromStr + ":" + to
	}
	from, to, err := stats.ParseRange(arg, time.Local)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filtered := stats.ApplyRange(s.store.Data(), from, to)
	report := stats.Build(filtered)
	label := fmt.Sprintf("%s → %s", from.Format("2006-01-02"), to.AddDate(0, 0, -1).Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = export.ToJSON(w, label, from, to, report)
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}
	report := stats.Build(s.store.Data())
	for _, sess := range report.BySession {
		if sess.SessionID == id || strings.HasPrefix(sess.SessionID, id) {
			writeJSON(w, http.StatusOK, map[string]any{
				"session_id":  sess.SessionID,
				"cwd":         sess.Cwd,
				"first_seen":  sess.FirstSeen,
				"last_seen":   sess.LastSeen,
				"duration_s":  sess.LastSeen.Sub(sess.FirstSeen).Seconds(),
				"prompts":     sess.Totals.Prompts,
				"turns":       sess.Totals.Turns,
				"tokens":      sess.Totals.GrandTotal(),
				"cost_usd":    round2(sess.Totals.Cost),
				"models":      keys(sess.Models),
			})
			return
		}
	}
	http.Error(w, "session not found", http.StatusNotFound)
}

// resolveAndApply handles the `range` query param — accepts presets
// ("today", "last-7d", ...) or raw `YYYY-MM-DD[:YYYY-MM-DD]`.
func (s *Server) resolveAndApply(rangeParam string) (label string, from, to time.Time, filtered usage.Data, err error) {
	if rangeParam == "" {
		rangeParam = "all"
	}
	if p, ok := stats.ResolvePreset(rangeParam); ok {
		filtered = p.Apply(s.store.Data())
		if p != stats.FilterAll {
			from, to = p.Range(time.Now())
		}
		return p.Label(), from, to, filtered, nil
	}
	from, to, err = stats.ParseRange(rangeParam, time.Local)
	if err != nil {
		return "", time.Time{}, time.Time{}, usage.Data{}, err
	}
	filtered = stats.ApplyRange(s.store.Data(), from, to)
	label = fmt.Sprintf("%s → %s", from.Format("2006-01-02"), to.AddDate(0, 0, -1).Format("2006-01-02"))
	return label, from, to, filtered, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
