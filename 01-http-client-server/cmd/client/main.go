package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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
	invalidArguments := errors.New("usage: client --host=[host] (--port=[port]) ([payload])")

	port := f.Int("port", defaultPort, "port to connect to")
	host := f.String("host", hostNotProvided, "host to connect to")
	payload := f.String("payload", payloadNotProvided, "HTTP payload")
	err := f.Parse(args)
	if err != nil {
		return &clientParams{}, err
	}

	if *host == hostNotProvided {
		return &clientParams{}, invalidArguments
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

func buildRequest(p *clientParams) string {
	var httpMethod string
	if p.payload != noBody {
		httpMethod = "GET"
	} else {
		httpMethod = "POST"
	}
	methodHeader := fmt.Sprintf("%s / HTTP/1.1", httpMethod)
	hostHeader := fmt.Sprintf("Host: %s", p.host)
	closeHeader := "Connection: close"
	blankLine := "\r\n"

	var lines []string
	if p.payload != noBody {
		contentTypeHeader := "Content-Type: text/plain"
		contentLengthHeader := fmt.Sprintf("Content-Length: %d", len(p.payload))
		lines = []string{
			methodHeader,
			hostHeader,
			contentTypeHeader,
			contentLengthHeader,
			closeHeader,
			blankLine,
			p.payload,
		}
	} else {
		lines = []string{
			methodHeader,
			hostHeader,
			closeHeader,
			blankLine,
		}
	}
	return strings.Join(lines, "\r\n")
}

func writeRequest(conn net.Conn, request string) error {
	_, err := conn.Write([]byte(request))
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
