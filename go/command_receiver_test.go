package main

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func TestStartCommandReceiver(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	originalCmdCh := cmdCh
	defer func() {
		pid = originalPid
		debug = originalDebug
		cmdCh = originalCmdCh
	}()

	// Set up test environment
	pid = 12345
	debug = false
	cmdCh = make(chan CommandWithConn, 1)

	// Start the command receiver
	startCommandReceiver()

	// Give it time to start listening
	time.Sleep(100 * time.Millisecond)

	// Test socket path
	socketPath := fmt.Sprintf("/tmp/native-app.%d.sock", pid)

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
		case cmdWithConn := <-cmdCh:
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
		case cmdWithConn := <-cmdCh:
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
		done := make(chan bool, numConnections)

		for i := 0; i < numConnections; i++ {
			go func(id int) {
				conn, err := net.Dial("unix", socketPath)
				if err != nil {
					t.Errorf("Connection %d failed: %v", id, err)
					done <- false
					return
				}
				defer conn.Close()

				_, err = conn.Write([]byte("list\n"))
				if err != nil {
					t.Errorf("Connection %d write failed: %v", id, err)
					done <- false
					return
				}
				done <- true
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
			case cmdWithConn := <-cmdCh:
				if _, ok := cmdWithConn.Cmd.(ListCommand); !ok {
					t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
				}
				cmdWithConn.Conn.Close()
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for command on channel")
			}
		}
	})

	// Clean up socket
	os.RemoveAll(socketPath)
}

func TestStartCommandReceiverDebugMode(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	originalCmdCh := cmdCh
	defer func() {
		pid = originalPid
		debug = originalDebug
		cmdCh = originalCmdCh
	}()

	// Set up test environment in debug mode
	pid = 12345
	debug = true
	cmdCh = make(chan CommandWithConn, 1)

	// Start the command receiver
	startCommandReceiver()

	// Give it time to start listening
	time.Sleep(100 * time.Millisecond)

	// In debug mode, socket path should be fixed
	socketPath := "/tmp/native-app.sock"

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
	case cmdWithConn := <-cmdCh:
		if _, ok := cmdWithConn.Cmd.(ListCommand); !ok {
			t.Errorf("Expected ListCommand, got %T", cmdWithConn.Cmd)
		}
		cmdWithConn.Conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command on channel")
	}

	// Clean up socket
	os.RemoveAll(socketPath)
}

func TestStartCommandReceiverInvalidCommand(t *testing.T) {
	// Save original values and restore after test
	originalPid := pid
	originalDebug := debug
	originalCmdCh := cmdCh
	defer func() {
		pid = originalPid
		debug = originalDebug
		cmdCh = originalCmdCh
	}()

	// Set up test environment
	pid = 12346
	debug = false
	cmdCh = make(chan CommandWithConn, 1)

	// Start the command receiver
	startCommandReceiver()

	// Give it time to start listening
	time.Sleep(100 * time.Millisecond)

	socketPath := fmt.Sprintf("/tmp/native-app.%d.sock", pid)

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
	case cmdWithConn := <-cmdCh:
		// Invalid commands result in nil cmd
		if cmdWithConn.Cmd != nil {
			t.Errorf("Expected nil command for invalid input, got %T", cmdWithConn.Cmd)
		}
		cmdWithConn.Conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command on channel")
	}

	// Clean up socket
	os.RemoveAll(socketPath)
}
