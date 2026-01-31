package main

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func TestGetSocketPath(t *testing.T) {
	tests := []struct {
		name      string
		processID int
		debugMode bool
		wantPath  string
	}{
		{
			name:      "normal mode with pid 12345",
			processID: 12345,
			debugMode: false,
			wantPath:  "/tmp/native-app.12345.sock",
		},
		{
			name:      "debug mode with pid 12345",
			processID: 12345,
			debugMode: true,
			wantPath:  "/tmp/native-app.sock",
		},
		{
			name:      "normal mode with pid 99999",
			processID: 99999,
			debugMode: false,
			wantPath:  "/tmp/native-app.99999.sock",
		},
		{
			name:      "debug mode with pid 99999",
			processID: 99999,
			debugMode: true,
			wantPath:  "/tmp/native-app.sock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath := getSocketPath(tt.processID, tt.debugMode)
			if gotPath != tt.wantPath {
				t.Errorf("getSocketPath(%d, %v) = %v, want %v",
					tt.processID, tt.debugMode, gotPath, tt.wantPath)
			}
		})
	}
}

func TestStartCommandReceiver(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	defer func() {
		pid = originalPid
		debug = originalDebug
	}()

	// Set up test environment
	pid = 12345
	debug = false
	testCmdCh := make(chan CommandWithConn, 1)

	// Test socket path
	socketPath := getSocketPath(pid, debug)
	defer os.RemoveAll(socketPath)

	// Start the command receiver
	startCommandReceiver(socketPath, testCmdCh)

	// Wait for socket to be ready with retry logic
	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	// Test 1: Connect to the socket and send a list command
	t.Run("list command", func(t *testing.T) {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			t.Fatalf("Failed to connect to socket: %v", err)
		}
		defer conn.Close()

		// Send a list command
		_, err = conn.Write([]byte("list\n"))
		if err != nil {
			t.Fatalf("Failed to write command: %v", err)
		}

		// Wait for command to be received on channel
		select {
		case cmdWithConn := <-testCmdCh:
			if _, ok := cmdWithConn.Cmd.(ListCommand); !ok {
				t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
			}
			if cmdWithConn.Conn == nil {
				t.Error("Expected non-nil connection")
			}
			cmdWithConn.Conn.Close()
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for command on channel")
		}
	})

	// Test 2: Send a select command
	t.Run("select command", func(t *testing.T) {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			t.Fatalf("Failed to connect to socket: %v", err)
		}
		defer conn.Close()

		// Send a select command
		_, err = conn.Write([]byte("select 123\n"))
		if err != nil {
			t.Fatalf("Failed to write command: %v", err)
		}

		// Wait for command to be received on channel
		select {
		case cmdWithConn := <-testCmdCh:
			selectCmd, ok := cmdWithConn.Cmd.(SelectCommand)
			if !ok {
				t.Errorf("Expected SelectCommand, got %T", cmdWithConn.Cmd)
			}
			if selectCmd.TabID != 123 {
				t.Errorf("Expected TabID 123, got %d", selectCmd.TabID)
			}
			if cmdWithConn.Conn == nil {
				t.Error("Expected non-nil connection")
			}
			cmdWithConn.Conn.Close()
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for command on channel")
		}
	})

	// Test 3: Multiple concurrent connections
	t.Run("concurrent connections", func(t *testing.T) {
		numConnections := 3
		done := make(chan struct{}, numConnections)

		for i := 0; i < numConnections; i++ {
			go func(id int) {
				conn, err := net.Dial("unix", socketPath)
				if err != nil {
					t.Errorf("Connection %d failed: %v", id, err)
					done <- struct{}{}
					return
				}
				defer conn.Close()

				_, err = conn.Write([]byte("list\n"))
				if err != nil {
					t.Errorf("Connection %d write failed: %v", id, err)
					done <- struct{}{}
					return
				}
				done <- struct{}{}
			}(i)
		}

		// Collect all commands
		for i := 0; i < numConnections; i++ {
			select {
			case <-done:
				// Connection completed
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for concurrent connection")
			}

			select {
			case cmdWithConn := <-testCmdCh:
				if _, ok := cmdWithConn.Cmd.(ListCommand); !ok {
					t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
				}
				cmdWithConn.Conn.Close()
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for command on channel")
			}
		}
	})
}

func waitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := os.Stat(socketPath)
		if err == nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket %s", socketPath)
}

func TestStartCommandReceiverDebugMode(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	defer func() {
		pid = originalPid
		debug = originalDebug
	}()

	// Set up test environment in debug mode
	pid = 12345
	debug = true
	testCmdCh := make(chan CommandWithConn, 1)

	// In debug mode, socket path should be fixed
	socketPath := getSocketPath(pid, debug)
	defer os.RemoveAll(socketPath)

	// Start the command receiver
	startCommandReceiver(socketPath, testCmdCh)

	// Wait for socket to be ready with retry logic
	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	// Connect and test
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Send a list command
	_, err = conn.Write([]byte("list\n"))
	if err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	// Wait for command to be received on channel
	select {
	case cmdWithConn := <-testCmdCh:
		if _, ok := cmdWithConn.Cmd.(ListCommand); !ok {
			t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
		}
		cmdWithConn.Conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command on channel")
	}
}

func TestStartCommandReceiverInvalidCommand(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	defer func() {
		pid = originalPid
		debug = originalDebug
	}()

	// Set up test environment
	pid = 12346
	debug = false
	testCmdCh := make(chan CommandWithConn, 1)

	socketPath := getSocketPath(pid, debug)
	defer os.RemoveAll(socketPath)

	// Start the command receiver
	startCommandReceiver(socketPath, testCmdCh)

	// Wait for socket to be ready with retry logic
	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	// Connect and send invalid command
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	// Send an invalid command
	_, err = conn.Write([]byte("invalid\n"))
	if err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	// The receiver should still send the command (even if nil) on the channel
	// based on the code logic
	select {
	case cmdWithConn := <-testCmdCh:
		// Invalid commands result in nil cmd
		if cmdWithConn.Cmd != nil {
			t.Errorf("Expected nil command for invalid input, got %T", cmdWithConn.Cmd)
		}
		cmdWithConn.Conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command on channel")
	}
}
