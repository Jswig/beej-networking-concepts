package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
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

	lineCount := 0
	headers := make(map[header]string)

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if line == "" {
			break
		} else if lineCount == 1 {
			httpMethod, err := parseRequestLine(line)
			if err == nil {
				log.Printf("HTTP method: %v\n", httpMethod)
			} else {
				log.Printf("error parsing HTTP methods: %s\n", err)
			}
		} else {
			key, value, err := parseHeader(line)
			if err != nil {
				log.Printf("error parsing HTTP header: %s", err)
			} else {
				headers[key] = value
			}
		}
	}

	_, hasContentType := headers[contentType]
	length, hasContentLength := headers[contentLength]
	if hasContentType && hasContentLength {
		bytesCount, _ := strconv.Atoi(length)
		buf := make([]byte, bytesCount)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			log.Printf("error reading HTTP request body: %s", err)
		} else {
			body := string(buf)
			log.Printf("HTTP request body: %s", body)
		}
	}

	responseText := "Hello!"
	_, err := fmt.Fprintf(conn, responseTemplate, len(responseText), responseText)
	if err != nil {
		log.Printf("error writing response to request: %s", err)
	}
}

type header string

const (
	connection    header = "connection"
	contentLength header = "content-length"
	contentType   header = "content-type"
	host          header = "host"
)

// constant for legal HTTP headers
var headers = []header{
	connection,
	contentLength,
	contentType,
	host,
}

func parseHeader(line string) (h header, v string, err error) {
	elements := strings.Split(line, ": ")
	if len(elements) != 2 {
		print("failed at length check")
		err = fmt.Errorf("%s is not a valid HTTP header", line)
	} else {
		// HTTP headers are case-insensitive
		h = header(strings.ToLower(elements[0]))
		v = elements[1]
		if !slices.Contains(headers, h) {
			err = fmt.Errorf("%s is not a known HTTP header", h)
		}
	}
	return h, v, err
}

type method string

const (
	get    method = "GET"
	delete method = "DELETE"
	patch  method = "PATCH"
	post   method = "POST"
	put    method = "PUT"
)

// constant for legal HTTP methods
var methods = []method{
	get,
	delete,
	patch,
	post,
	put,
}

func parseMethod(s string) (method, error) {
	if slices.Contains(methods, method(s)) {
		return method(s), nil
	} else {
		return "", fmt.Errorf("%s is not a valid HTTP method", s)
	}
}

func parseRequestLine(line string) (method, error) {
	words := strings.Split(line, " ")
	if len(words) < 1 {
		return "", fmt.Errorf("no HTTP method found")
	}
	return parseMethod(words[0])
}
