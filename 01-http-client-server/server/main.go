package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
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
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	address := conn.RemoteAddr()
	ipAddress := strings.Split(address.String(), ":")[0]
	log.Printf("connection from IP address: %s\n", ipAddress)
	headers := []string{}
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineCount == 0 {
			httpMethod, err := parseHTTPMethod(line)
			if err == nil {
				log.Printf("HTTP method: %v\n", httpMethod)
			} else {
				log.Printf("error parsing HTTP methods: %s\n", err)
			}
		}
		lineCount++
		if line == "" {
			break
		}
		headers = append(headers, line)
	}
	responseText := "Hello!"
	_, err := fmt.Fprintf(conn, responseTemplate, len(responseText), responseText)
	if err != nil {
		log.Fatalf("error writing reponse: %s", err)
	}
}

type httpMethod string

func newHttpMethod(s string) (httpMethod, error) {
	methods := []string{
		"DELETE",
		"GET",
		"PATCH",
		"POST",
		"PUT",
	}
	if slices.Contains(methods, s) {
		return httpMethod(s), nil
	} else {
		return "", fmt.Errorf("%s is not a valid HTTP method", s)
	}
}

func parseHTTPMethod(line string) (httpMethod, error) {
	words := strings.Split(line, " ")
	if len(words) < 1 {
		return "", fmt.Errorf("no HTTP method found")
	}
	return newHttpMethod(words[0])
}
