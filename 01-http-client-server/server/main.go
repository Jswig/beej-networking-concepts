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
	log.Printf("HTTP request method: %v", request.method)
	if request.hasBody {
		log.Printf("HTTP request body: %s", request.body)
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

type Request struct {
	Head
	body    string
	hasBody bool
}

const noRequestBody = ""

func ParseRequest(r io.Reader) (*Request, error) {
	head, err := parseHead(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTTP request head: %s", err)
	}

	numBytes := numBodyBytes(head)
	var hasBody bool
	var body string
	if numBytes > 0 {
		body, err = readBody(r, numBytes)
		if err != nil {
			return nil, err
		}
		log.Printf("HTTP request body: %s", body)
		hasBody = true
	} else {
		body = noRequestBody
		hasBody = false
	}

	return &Request{*head, body, hasBody}, nil
}

type Head struct {
	method  method
	headers map[header]string
}

func parseHead(r io.Reader) (*Head, error) {
	headers := make(map[header]string)
	var meth method

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	scanner.Scan()
	line := scanner.Text()
	meth, err := parseRequestLine(line)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTTP request line: %s", err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		key, value, err := parseHeader(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing HTTP header: %s", err)
		}
		headers[key] = value
	}

	return &Head{meth, headers}, nil
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

func numBodyBytes(h *Head) int {
	_, hasContentType := h.headers[contentType]
	length, hasContentLength := h.headers[contentLength]

	var numBytes int
	if hasContentType && hasContentLength {
		numBytes, _ = strconv.Atoi(length)
	}
	return numBytes
}

func readBody(r io.Reader, numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return "", fmt.Errorf("error reading request body: %s", err)
	}
	return string(buf), nil
}
