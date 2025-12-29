package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DevServer handles hot-reload development server
type DevServer struct {
	port       int
	target     string
	watch      bool
	binaryPath string
	cmd        *exec.Cmd
	mu         sync.Mutex
	watcher    *fsnotify.Watcher
	debounce   *time.Timer
}

// NewDevServer creates a new development server
func NewDevServer(port int, target string, watch bool) *DevServer {
	return &DevServer{
		port:       port,
		target:     target,
		watch:      watch,
		binaryPath: filepath.Join(os.TempDir(), fmt.Sprintf("echonext-dev-%s-%d", target, os.Getpid())),
	}
}

// Run starts the development server
func (s *DevServer) Run() error {
	// Verify we're in an EchoNext project
	if !isEchoNextProject() {
		return fmt.Errorf("not in an EchoNext project. Run from project root")
	}

	// Verify target exists
	targetPath := filepath.Join("cmd", s.target)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Also check if there's a main.go in the root (for simple projects)
		if _, err := os.Stat("main.go"); os.IsNotExist(err) {
			return fmt.Errorf("target '%s' not found at %s (and no main.go in root)", s.target, targetPath)
		}
		// Use root main.go
		targetPath = "."
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Initial build and start
	if err := s.buildAndStart(); err != nil {
		return err
	}

	// Set up file watcher if enabled
	if s.watch {
		var err error
		s.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create file watcher: %w", err)
		}
		defer s.watcher.Close()

		// Add directories to watch
		if err := s.addWatchDirs(); err != nil {
			return fmt.Errorf("failed to set up file watching: %w", err)
		}

		// Start watching in a goroutine
		go s.watchFiles(ctx)
	}

	// Wait for signal
	<-sigCh
	fmt.Println("\nShutting down...")
	s.stop()
	return nil
}

// buildAndStart builds the target and starts it
func (s *DevServer) buildAndStart() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing process if running
	s.stopProcess()

	fmt.Printf("Building %s...\n", s.target)

	// Determine target path - must use "./" prefix for local paths
	targetPath := "./" + filepath.Join("cmd", s.target)
	if _, err := os.Stat(filepath.Join("cmd", s.target)); os.IsNotExist(err) {
		targetPath = "."
	}

	// Build the target
	buildCmd := exec.Command("go", "build", "-o", s.binaryPath, targetPath)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("Starting %s on port %d...\n", s.target, s.port)

	// Start the built binary
	s.cmd = exec.Command(s.binaryPath)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	s.cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", s.port))

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	fmt.Printf("Server running at http://localhost:%d\n", s.port)
	if s.watch {
		fmt.Println("Watching for file changes... (Ctrl+C to stop)")
	}

	return nil
}

// addWatchDirs adds directories to the file watcher
func (s *DevServer) addWatchDirs() error {
	// Directories to watch
	watchDirs := []string{
		"cmd",
		"domain",
		"internal",
		"pkg",
	}

	for _, dir := range watchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		// Walk directory and add all subdirectories
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip vendor and hidden directories
				if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return s.watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Also watch root go files
	if err := s.watcher.Add("."); err != nil {
		return err
	}

	return nil
}

// watchFiles watches for file changes and triggers rebuilds
func (s *DevServer) watchFiles(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Only watch .go files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			// Skip test files
			if strings.HasSuffix(event.Name, "_test.go") {
				continue
			}

			// Ignore events that aren't writes or creates
			if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
				continue
			}

			// Debounce rapid changes (e.g., save file triggers multiple events)
			s.mu.Lock()
			if s.debounce != nil {
				s.debounce.Stop()
			}
			s.debounce = time.AfterFunc(100*time.Millisecond, func() {
				fmt.Printf("\nFile changed: %s\n", event.Name)
				fmt.Println("Rebuilding...")
				if err := s.buildAndStart(); err != nil {
					log.Printf("Rebuild failed: %v", err)
					fmt.Println("Waiting for file changes...")
				}
			})
			s.mu.Unlock()

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// stopProcess stops the running process
func (s *DevServer) stopProcess() {
	if s.cmd != nil && s.cmd.Process != nil {
		// Send SIGTERM first for graceful shutdown
		s.cmd.Process.Signal(syscall.SIGTERM)

		// Wait a bit, then force kill if still running
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(3 * time.Second):
			// Force kill
			s.cmd.Process.Kill()
			<-done
		}
		s.cmd = nil
	}
}

// stop cleans up the dev server
func (s *DevServer) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stopProcess()

	// Clean up binary
	os.Remove(s.binaryPath)
}
