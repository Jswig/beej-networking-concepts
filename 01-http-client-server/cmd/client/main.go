package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"simple-client-server/internal/http"
	"strconv"
)

const noBody = ""

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type clientParams struct {
	host    string
	port    int
	payload string
}

func run(args []string) error {
	params, err := parseArgs(args)
	if err != nil {
		return err
	}

	conn, err := makeConnection(params)
	if err != nil {
		return err
	}
	defer conn.Close()

	request := buildRequest(params)

	err = writeRequest(conn, request)
	if err != nil {
		return err
	}

	response, err := readResponse(conn)
	if err != nil {
		return err
	}
	fmt.Print(response)

	return nil
}

const defaultPort = 80

const hostNotProvided = ""

const payloadNotProvided = ""

func parseArgs(args []string) (*clientParams, error) {
	f := flag.NewFlagSet("client", flag.ContinueOnError)
	invalidArguments := errors.New("usage: client --host=[host] (--port=[port]) (--payload=[payload])")

	port := f.Int("port", defaultPort, "port to connect to")
	host := f.String("host", hostNotProvided, "host to connect to")
	payload := f.String("payload", payloadNotProvided, "HTTP payload")
	err := f.Parse(args)
	if err != nil {
		return nil, err
	}

	if *host == hostNotProvided {
		return nil, invalidArguments
	}

	return &clientParams{*host, *port, *payload}, nil
}

func makeConnection(p *clientParams) (net.Conn, error) {
	address := p.host + ":" + strconv.Itoa(p.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", address, err)
	}
	return conn, nil
}

func buildRequest(p *clientParams) *http.Request {
	var method http.Method
	var headers = http.Headers{}

	hasBody := p.payload != noBody
	if hasBody {
		method = http.Post
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = fmt.Sprintf("%d", len(p.payload))
	} else {
		method = http.Get
	}
	headers["Host"] = p.host
	headers["Connection"] = "close"

	return &http.Request{
		Method:  method,
		Headers: headers,
		Body:    []byte(p.payload),
		HasBody: hasBody,
	}
}

func writeRequest(conn net.Conn, request *http.Request) error {
	err := request.Encode(conn)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %s", err)
	}
	return nil
}

func readResponse(conn net.Conn) (string, error) {
	results := [][]byte{}
	for responseComplete := false; !responseComplete; {
		b := make([]byte, 4096)
		count, err := conn.Read(b)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("error reading response: %s", err)
		}
		if count > 0 {
			results = append(results, b[:count])
		}
		if err == io.EOF {
			responseComplete = true
		}
	}
	response := string(bytes.Join(results, []byte{}))
	return response, nil
}
