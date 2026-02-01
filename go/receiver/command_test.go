package receiver

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"rofi-chrome-tab/protocol"
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
			gotPath := GetSocketPath(tt.processID, tt.debugMode)
			if gotPath != tt.wantPath {
				t.Errorf("GetSocketPath(%d, %v) = %v, want %v",
					tt.processID, tt.debugMode, gotPath, tt.wantPath)
			}
		})
	}
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

func TestStartCommandReceiver(t *testing.T) {
	// Set up test environment
	pid := 12345
	debug := false
	testCmdCh := make(chan CommandWithConn, 1)

	// Test socket path
	socketPath := GetSocketPath(pid, debug)
	defer os.RemoveAll(socketPath)

	// Start the command receiver
	StartCommandReceiver(socketPath, testCmdCh)

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
			if _, ok := cmdWithConn.Cmd.(protocol.ListCommand); !ok {
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
			selectCmd, ok := cmdWithConn.Cmd.(protocol.SelectCommand)
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
}
