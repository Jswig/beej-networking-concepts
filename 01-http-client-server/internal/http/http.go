package http

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

type Request struct {
	Method  Method
	Headers Headers
	Body    []byte
	HasBody bool
}

type Method string

type Headers map[string]string

const (
	Get    Method = "GET"
	Delete Method = "DELETE"
	Patch  Method = "PATCH"
	Post   Method = "POST"
	Put    Method = "PUT"
)

// constant for legal HTTP methods
var ValidMethods = []Method{
	Get,
	Delete,
	Patch,
	Post,
	Put,
}

const startLineTemplate = "%s / HTTP/1.1\r\n"
const headerLineTemplate = "%s: %s\r\n"
const emptyLine = "\r\n"

func (request *Request) Encode(w io.Writer) error {
	// 1 start line, 1 line per header and 1 empty line
	lines := make([]string, 0, 2+len(request.Headers))

	startLine := fmt.Sprintf(startLineTemplate, request.Method)
	lines = append(lines, startLine)

	var headerLine string
	for name, value := range request.Headers {
		headerLine = fmt.Sprintf(headerLineTemplate, name, value)
		lines = append(lines, headerLine)
	}

	lines = append(lines, emptyLine)

	_, err := io.WriteString(w, strings.Join(lines, ""))
	if err != nil {
		return fmt.Errorf("error writing HTTP request: %v", err)
	}

	if request.HasBody {
		w.Write(request.Body)
	}

	return nil
}

const noRequestBody = ""

func Decode(r io.Reader) (*Request, error) {
	buf := bufio.NewReader(r)

	method, err := parseStartLine(buf)
	if err != nil {
		return nil, fmt.Errorf("error decoding HTTP request: %s", err)
	}

	headers, err := parseHeaders(buf)
	if err != nil {
		return nil, fmt.Errorf("error decoding HTTP request: %s", err)
	}

	numBytes := numBodyBytes(headers)
	var hasBody bool
	var body []byte
	if numBytes > 0 {
		body, err = readBody(buf, numBytes)
		if err != nil {
			return nil, fmt.Errorf("error decoding HTTP request: %s", err)
		}
		hasBody = true
	} else {
		body = []byte(noRequestBody)
		hasBody = false
	}

	return &Request{
		Method:  method,
		Headers: headers,
		Body:    body,
		HasBody: hasBody,
	}, nil
}

func parseStartLine(buf *bufio.Reader) (Method, error) {
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading HTTP request line: %s", err)
	}

	words := strings.Split(line, " ")
	if len(words) < 1 {
		return "", fmt.Errorf("no HTTP method found")
	}

	methodToken := words[0]
	if slices.Contains(ValidMethods, Method(methodToken)) {
		return Method(methodToken), nil
	} else {
		return "", fmt.Errorf("%s is not a valid HTTP method", methodToken)
	}
}

func parseHeaders(buf *bufio.Reader) (Headers, error) {
	headers := make(Headers)

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading header line: %s", err)
		}
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		key, value, err := parseHeaderLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing header line: %s", err)
		}
		headers[key] = value
	}

	return headers, nil
}

func parseHeaderLine(line string) (h string, v string, err error) {
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

func numBodyBytes(h Headers) int {
	_, hasContentType := h["content-type"]
	length, hasContentLength := h["content-length"]

	if hasContentType && hasContentLength {
		numBytes, _ := strconv.Atoi(length)
		return numBytes
	}
	return 0
}

func readBody(buf *bufio.Reader, numBytes int) ([]byte, error) {
	b := make([]byte, numBytes)
	numBytesRead, err := io.ReadFull(buf, b)
	if err != nil {
		return []byte(""), fmt.Errorf("error reading request body: %s", err)
	}
	if numBytesRead != numBytes {
		return []byte(""), fmt.Errorf("error reading request body, expected %d bytes, got %d bytes", numBytes, numBytesRead)
	}
	return b, nil
}
