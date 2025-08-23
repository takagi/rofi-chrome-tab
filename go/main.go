package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// Main

type CommandWithConn struct {
	Cmd  Command
	Conn net.Conn
}

var (
	tabs  []Tab
	evCh  = make(chan Event, 1)
	cmdCh = make(chan CommandWithConn, 1)

	pid = os.Getpid()
)

func main() {
	// Set up log file
	logFile, err := os.OpenFile("/tmp/rofi_chrome_tab.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open log file:", err)
		os.Exit(1)
	} else {
		log.SetOutput(logFile)
	}
	defer logFile.Close()

	if debug {
		log.Println("Debug mode")
	}

	// Receive events from stdin
	go func() {
		for {
			ev, err := RecvEvent(os.Stdin)
			if err == io.EOF {
				log.Println("stdin closed")
				return
			}
			if err != nil {
				log.Println("Error receiving message:", err)
				continue
			}
			log.Printf("Received event: %T", ev)
			evCh <- ev
		}
	}()

	// Set up a socket file
	var socketPath string
	if !debug {
		socketPath = fmt.Sprintf("/tmp/native-app.%d.sock", pid)
	} else {
		socketPath = fmt.Sprintf("/tmp/native-app.sock")
	}
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

				cmd, err := ParseCommand(line)
				if err != nil {
					log.Println("Parse error:", err, "line:", line)
				}
				log.Printf("Received command: %T", cmd)

				cmdCh <- CommandWithConn{Cmd: cmd, Conn: c}
			}(conn)
		}
	}()

	for {
		select {
		case ev := <-evCh:
			ev.Handle()
		case cw := <-cmdCh:
			cw.Cmd.Execute(cw.Conn)
			cw.Conn.Close()
		}
	}
}
