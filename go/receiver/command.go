package receiver

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"rofi-chrome-tab/protocol"
)

type CommandWithConn struct {
	Cmd  protocol.Command
	Conn net.Conn
}

// GetSocketPath returns the Unix domain socket path based on the process ID and debug mode.
func GetSocketPath(processID int, debugMode bool) string {
	if !debugMode {
		return fmt.Sprintf("/tmp/native-app.%d.sock", processID)
	}
	return "/tmp/native-app.sock"
}

func StartCommandReceiver(socketPath string, cmdCh chan CommandWithConn) {
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
				}

				line := strings.TrimSpace(scanner.Text())

				cmd, err := protocol.ParseCommand(line)
				if err != nil {
					log.Println("Parse error:", err, "line:", line)
				}
				log.Printf("Received command: %T", cmd)

				cmdCh <- CommandWithConn{Cmd: cmd, Conn: c}
			}(conn)
		}
	}()
}
