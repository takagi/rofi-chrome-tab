package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// getSocketPath returns the Unix domain socket path based on the process ID and debug mode.
// This function is placed in command_receiver.go because:
// 1. It's specifically used for creating the socket path for the command receiver
// 2. It's logically grouped with startCommandReceiver which uses the socket path
// 3. Creating a receiver struct just for this utility would be over-engineering
// 4. Placing it in main.go would put implementation details in the wrong architectural layer
func getSocketPath(processID int, debugMode bool) string {
	if !debugMode {
		return fmt.Sprintf("/tmp/native-app.%d.sock", processID)
	}
	return "/tmp/native-app.sock"
}

func startCommandReceiver(socketPath string, cmdCh chan CommandWithConn) {
	// Remove existing socket file
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	// Receive commands from an Unix domain socket
	go func() {
		lis, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Fatal("listen error:", err)
		}
		defer lis.Close()

		log.Printf("Listening on socket: %s", socketPath)

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Println("Accept error:", err)
				continue
			}

			go func(c net.Conn) {
				scanner := bufio.NewScanner(c)

				scanner.Scan()
				if err := scanner.Err(); err != nil {
					log.Println("Read error:", err)
					c.Close()
					return
				}

				line := strings.TrimSpace(scanner.Text())

				cmd, err := ParseCommand(line)
				if err != nil {
					log.Println("Parse error:", err, "line:", line)
					c.Close()
					return
				}
				log.Printf("Received command: %T", cmd)

				cmdCh <- CommandWithConn{Cmd: cmd, Conn: c}
			}(conn)
		}
	}()
}
