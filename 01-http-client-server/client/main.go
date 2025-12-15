package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	if err := runServer(os.Args); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func runServer(args []string) error {
	var host string
	var port int
	var err error
	if len(args) == 3 {
		host = args[1]
		port, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[2])
		}
	} else if len(args) == 2 {
		host = args[1]
		port = 80
	} else {
		return fmt.Errorf("usage: client [host] ([port])")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("error creating connection: %v", err)
	}
	fmt.Printf("network address=%v", conn.RemoteAddr())
	httpGetHeaderTemplate := "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n"
	_, err = fmt.Fprintf(conn, httpGetHeaderTemplate, host)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %v", err)
	}
	response, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("error reading HTTP response %v", err)
	}
	fmt.Print(string(response))

	return nil
}
