package http

import (
	"fmt"
	"io"
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

func Decode(r io.Reader) *Request {
	return &Request{}
}
