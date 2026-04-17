// Package watch observes the Claude Code JSONL tree for changes and emits
// debounced events so the daemon can reload its cache and notify SSE clients.
package watch

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	root string
	fsw  *fsnotify.Watcher
	mu   sync.Mutex
}

// New creates a watcher rooted at path, recursively registering every
// subdirectory.
func New(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{root: root, fsw: fsw}
	if err := w.addRecursive(root); err != nil {
		_ = fsw.Close()
		return nil, err
	}
	return w, nil
}

func (w *Watcher) addRecursive(path string) error {
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		w.mu.Lock()
		defer w.mu.Unlock()
		return w.fsw.Add(p)
	})
}

// Events returns a channel that emits once file-system activity settles for
// 1 second (debounce). Only JSONL writes and new directories trigger. The
// channel closes when ctx is cancelled or the watcher errors terminally.
func (w *Watcher) Events(ctx context.Context) <-chan struct{} {
	out := make(chan struct{}, 1)
	go func() {
		defer close(out)
		var pending *time.Timer
		fire := func() {
			select {
			case out <- struct{}{}:
			default:
				// drop if consumer is slow — next debounce will retry
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-w.fsw.Events:
				if !ok {
					return
				}
				if ev.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
						w.mu.Lock()
						_ = w.fsw.Add(ev.Name)
						w.mu.Unlock()
					}
				}
				relevant := strings.HasSuffix(ev.Name, ".jsonl") ||
					ev.Op&fsnotify.Create != 0
				if !relevant {
					continue
				}
				if pending != nil {
					pending.Stop()
				}
				pending = time.AfterFunc(1*time.Second, fire)
			case err, ok := <-w.fsw.Errors:
				if !ok {
					return
				}
				log.Printf("watch error: %v", err)
			}
		}
	}()
	return out
}

func (w *Watcher) Close() error { return w.fsw.Close() }
