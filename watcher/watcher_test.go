package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatch_DetectsWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("# init"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	got := make(chan []byte, 1)
	go func() {
		_ = Watch(ctx, path, func(_ string, data []byte) {
			select {
			case got <- data:
			default:
			}
		})
	}()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	updated := []byte("# updated")
	if err := os.WriteFile(path, updated, 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case data := <-got:
		if string(data) != string(updated) {
			t.Errorf("got %q, want %q", data, updated)
		}
	case <-ctx.Done():
		t.Error("timeout: file change not detected")
	}
}
