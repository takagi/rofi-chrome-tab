package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func startCommandReceiver(pid int, debugMode bool, cmdCh chan CommandWithConn) string {
	// Determine socket path based on process ID and debug mode
	socketPath := "/tmp/native-app.sock"
	if !debugMode {
		socketPath = fmt.Sprintf("/tmp/native-app.%d.sock", pid)
	}
	
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
	
	return socketPath
}
