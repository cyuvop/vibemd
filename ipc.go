package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const sockName = "vibemd.sock"

func sockPath() string {
	return filepath.Join(os.TempDir(), sockName)
}

// tryDelegate attempts to send filePath to an already-running vibemd instance.
// Returns true if a running instance was found (caller should exit).
func tryDelegate(filePath string) bool {
	conn, err := net.Dial("unix", sockPath())
	if err != nil {
		return false
	}
	defer conn.Close()
	if filePath != "" {
		_, _ = conn.Write([]byte(filePath + "\n"))
	}
	return true
}

// listenForFiles starts a Unix socket server so subsequent vibemd invocations
// can hand file paths to this running instance instead of silently failing.
func listenForFiles(app *App) {
	sock := sockPath()
	os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		log.Printf("ipc: could not listen on %s: %v", sock, err)
		return
	}
	go func() {
		defer os.Remove(sock)
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					path := strings.TrimSpace(scanner.Text())
					if path != "" {
						_ = app.OpenFile(path)
					}
				}
			}(conn)
		}
	}()
}
