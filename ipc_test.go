package main

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTryDelegate_NoSocket(t *testing.T) {
	os.Remove(sockPath())
	if tryDelegate("/tmp/test.md") {
		t.Error("expected false when no socket exists")
	}
}

func TestTryDelegate_WithSocket(t *testing.T) {
	sock := sockPath()
	os.Remove(sock)

	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sock)

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	if !tryDelegate("/tmp/test.md") {
		t.Fatal("expected true when socket exists")
	}

	select {
	case path := <-received:
		if path != "/tmp/test.md" {
			t.Errorf("got %q, want /tmp/test.md", path)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout: delegated path not received")
	}
}

func TestTryDelegate_EmptyPath(t *testing.T) {
	sock := sockPath()
	os.Remove(sock)

	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sock)

	// Empty path: should still connect (returns true) but send nothing
	lines := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(300 * time.Millisecond))
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			lines <- scanner.Text()
		} else {
			lines <- ""
		}
	}()

	if !tryDelegate("") {
		t.Fatal("expected true when socket exists, even with empty path")
	}

	select {
	case line := <-lines:
		if strings.TrimSpace(line) != "" {
			t.Errorf("expected nothing sent for empty path, got %q", line)
		}
	case <-time.After(time.Second):
		// Acceptable: connection closed without sending — expected for empty path
	}
}

func TestHeadlessMCPState_OpenFile_MissingFile(t *testing.T) {
	state := &headlessMCPState{app: NewApp()}
	err := state.OpenFile("/tmp/definitely_does_not_exist_vibemd_test.md")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestHeadlessMCPState_OpenFile_SetsFilePath(t *testing.T) {
	// Create a temp file so os.Stat succeeds
	f, err := os.CreateTemp("", "vibemd-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	// Remove socket so tryDelegate returns false (avoids real spawn)
	os.Remove(sockPath())

	state := &headlessMCPState{app: NewApp()}
	// OpenFile will try to spawn a window — in test env that's a no-op failure
	// but the filePath should be set before the spawn attempt
	_ = state.OpenFile(f.Name())

	if state.app.filePath != f.Name() {
		t.Errorf("filePath = %q, want %q", state.app.filePath, f.Name())
	}
}

func TestHeadlessMCPState_OpenFile_DelegatesToSocket(t *testing.T) {
	f, err := os.CreateTemp("", "vibemd-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	// Start a mock vibemd window IPC listener
	sock := sockPath()
	os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sock)

	received := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	state := &headlessMCPState{app: NewApp()}
	if err := state.OpenFile(f.Name()); err != nil {
		t.Fatalf("OpenFile returned error: %v", err)
	}

	select {
	case path := <-received:
		if path != f.Name() {
			t.Errorf("socket received %q, want %q", path, f.Name())
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout: expected file path on socket, got nothing")
	}
}
