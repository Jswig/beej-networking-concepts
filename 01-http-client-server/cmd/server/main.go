package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"simple-client-server/internal/http"
	"strings"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

const responseTemplate = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n"

func run(args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}

	listener, err := setupListener(cfg)
	if err != nil {
		return err
	}
	defer listener.Close()

	err = acceptConnections(listener)
	if err != nil {
		return err
	}

	return nil
}

type serverConfig struct {
	port int
}

const defaultPort = 80

func parseArgs(args []string) (*serverConfig, error) {
	f := flag.NewFlagSet("server", flag.ContinueOnError)
	port := f.Int("port", defaultPort, "port to listen on")
	err := f.Parse(args)
	if err != nil {
		return nil, err
	}
	return &serverConfig{*port}, nil
}

func setupListener(cfg *serverConfig) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.port))
	if err != nil {
		return listener, fmt.Errorf("error listening on ports %d: %v", cfg.port, err)
	}
	return listener, nil
}

func acceptConnections(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connections: %v", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	ipAddress := getIpAddress(conn)
	log.Printf("connection from IP address: %s\n", ipAddress)

	request, err := http.Decode(conn)
	if err != nil {
		log.Printf("error parsing HTTP request: %s", err)
		return
	}
	log.Printf("HTTP request method: %v", request.Method)
	if request.HasBody {
		log.Printf("HTTP request body: %s", string(request.Body))
	}

	err = writeDefaultResponse(conn)
	if err != nil {
		log.Printf("error writing HTTP response: %s", err)
		return
	}
}

func getIpAddress(conn net.Conn) string {
	address := conn.RemoteAddr()
	return strings.Split(address.String(), ":")[0]
}

func writeDefaultResponse(w io.Writer) error {
	responseText := "Hello!"
	_, err := fmt.Fprintf(w, responseTemplate, len(responseText), responseText)
	if err != nil {
		log.Printf("error writing response to request: %s", err)
	}
	return nil
}
