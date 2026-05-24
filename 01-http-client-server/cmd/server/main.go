package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"simple-client-server/internal/http"
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	ipAddress := getIpAddress(conn)
	log.Printf("connection from IP address: %s\n", ipAddress)

	request, err := ParseRequest(conn)
	if err != nil {
		log.Printf("error parsing HTTP request: %s", err)
		return
	}
	log.Printf("HTTP request method: %v", request.Method)
	if request.HasBody {
		log.Printf("HTTP request body: %s", request.Body)
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

const noRequestBody = ""

func ParseRequest(r io.Reader) (*http.Request, error) {
	buf := bufio.NewReader(r)
	method, headers, err := parseHead(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTTP request head: %s", err)
	}

	numBytes := numBodyBytes(headers)
	var hasBody bool
	var body string
	if numBytes > 0 {
		body, err = readBody(buf, numBytes)
		if err != nil {
			return nil, err
		}
		hasBody = true
	} else {
		body = noRequestBody
		hasBody = false
	}

	return &http.Request{method, headers, body, hasBody}, nil
}

func parseHead(buf *bufio.Reader) (http.Method, http.Headers, error) {
	headers := make(http.Headers)
	var method http.Method

	line, err := buf.ReadString('\n')
	if err != nil {
		return "", nil, fmt.Errorf("error reading HTTP request line: %s", err)
	}
	line = strings.TrimRight(line, "\r\n")
	method, err = parseRequestLine(line)
	if err != nil {
		return "", nil, fmt.Errorf("error parsing HTTP request line: %s", err)
	}

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return "", nil, fmt.Errorf("error reading HTTP header line: %s", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		key, value, err := parseHeader(line)
		if err != nil {
			return "", nil, fmt.Errorf("error parsing HTTP header: %s", err)
		}
		headers[key] = value
	}

	return method, headers, nil
}

func parseMethod(s string) (http.Method, error) {
	if slices.Contains(http.ValidMethods, http.Method(s)) {
		return http.Method(s), nil
	} else {
		return "", fmt.Errorf("%s is not a valid HTTP method", s)
	}
}

func parseRequestLine(line string) (http.Method, error) {
	words := strings.Split(line, " ")
	if len(words) < 1 {
		return "", fmt.Errorf("no HTTP method found")
	}
	return parseMethod(words[0])
}

func parseHeader(line string) (h string, v string, err error) {
	elements := strings.Split(line, ": ")
	if len(elements) != 2 {
		err = fmt.Errorf("%s is not a valid HTTP header", line)
	} else {
		// HTTP headers are case-insensitive
		h = strings.ToLower(elements[0])
		v = elements[1]
	}
	return h, v, err
}

func numBodyBytes(h http.Headers) int {
	_, hasContentType := h["content-type"]
	length, hasContentLength := h["content-length"]

	if hasContentType && hasContentLength {
		numBytes, _ := strconv.Atoi(length)
		return numBytes
	}
	return 0
}

func readBody(buf *bufio.Reader, numBytes int) (string, error) {
	b := make([]byte, numBytes)
	numBytesRead, err := io.ReadFull(buf, b)
	if err != nil {
		return "", fmt.Errorf("error reading request body: %s", err)
	}
	if numBytesRead != numBytes {
		return "", fmt.Errorf("expected %d bytes, got %d bytes", numBytes, numBytesRead)
	}
	return string(b), nil
}
