package watcher

import (
	"context"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

// OnChange is called when the watched file changes.
type OnChange func(path string, data []byte)

// Watch monitors path and calls onChange on each write event.
// Blocks until ctx is cancelled.
func Watch(ctx context.Context, path string, onChange OnChange) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	if err := w.Add(path); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				data, err := os.ReadFile(path)
				if err != nil {
					log.Printf("watcher: read error: %v", err)
					continue
				}
				onChange(path, data)
			}
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			log.Printf("watcher: error: %v", err)
		}
	}
}
