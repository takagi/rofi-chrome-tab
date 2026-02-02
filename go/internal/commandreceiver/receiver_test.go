package commandreceiver

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"rofi-chrome-tab/internal/command"
)

func TestStartCommandReceiver(t *testing.T) {
	pid := 12345
	debugMode := false
	testCmdCh := make(chan CommandWithConn, 1)

	socketPath := Start(pid, debugMode, testCmdCh)
	defer os.RemoveAll(socketPath)

	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	t.Run("list command", func(t *testing.T) {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			t.Fatalf("Failed to connect to socket: %v", err)
		}
		defer conn.Close()

		if _, err := conn.Write([]byte("list\n")); err != nil {
			t.Fatalf("Failed to write command: %v", err)
		}

		select {
		case cmdWithConn := <-testCmdCh:
			if _, ok := cmdWithConn.Cmd.(command.ListCommand); !ok {
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

	t.Run("select command", func(t *testing.T) {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			t.Fatalf("Failed to connect to socket: %v", err)
		}
		defer conn.Close()

		if _, err := conn.Write([]byte("select 123\n")); err != nil {
			t.Fatalf("Failed to write command: %v", err)
		}

		select {
		case cmdWithConn := <-testCmdCh:
			selectCmd, ok := cmdWithConn.Cmd.(command.SelectCommand)
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

				if _, err := conn.Write([]byte("list\n")); err != nil {
					t.Errorf("Connection %d write failed: %v", id, err)
					done <- struct{}{}
					return
				}
				done <- struct{}{}
			}(i)
		}

		for i := 0; i < numConnections; i++ {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for concurrent connection")
			}

			select {
			case cmdWithConn := <-testCmdCh:
				if _, ok := cmdWithConn.Cmd.(command.ListCommand); !ok {
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
		if _, err := os.Stat(socketPath); err == nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket %s", socketPath)
}

func TestStartCommandReceiverDebugMode(t *testing.T) {
	pid := 12345
	debugMode := true
	testCmdCh := make(chan CommandWithConn, 1)

	socketPath := Start(pid, debugMode, testCmdCh)
	defer os.RemoveAll(socketPath)

	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("list\n")); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	select {
	case cmdWithConn := <-testCmdCh:
		if _, ok := cmdWithConn.Cmd.(command.ListCommand); !ok {
			t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
		}
		cmdWithConn.Conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command on channel")
	}
}

func TestStartCommandReceiverInvalidCommand(t *testing.T) {
	pid := 12346
	debugMode := false
	testCmdCh := make(chan CommandWithConn, 1)

	socketPath := Start(pid, debugMode, testCmdCh)
	defer os.RemoveAll(socketPath)

	if err := waitForSocket(socketPath, 2*time.Second); err != nil {
		t.Fatalf("Socket not ready: %v", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to socket: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("invalid\n")); err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}

	select {
	case cmdWithConn := <-testCmdCh:
		t.Errorf("Expected no command on channel for invalid input, got %T", cmdWithConn.Cmd)
		cmdWithConn.Conn.Close()
	case <-time.After(500 * time.Millisecond):
	}
}
