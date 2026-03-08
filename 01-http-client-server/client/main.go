package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type clientParams struct {
	host string
	port int
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

	err = writeRequest(conn, params.host)
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

func parseArgs(args []string) (clientParams, error) {
	var host string
	var port int
	var err error
	if len(args) == 3 {
		host = args[1]
		port, err = strconv.Atoi(args[2])
		if err != nil {
			return clientParams{}, fmt.Errorf("invalid port: %s", args[2])
		}
	} else if len(args) == 2 {
		host = args[1]
		port = 80
	} else {
		return clientParams{}, fmt.Errorf("usage: client [host] ([port])")
	}
	return clientParams{host, port}, nil
}

func makeConnection(p clientParams) (net.Conn, error) {
	address := p.host + ":" + strconv.Itoa(p.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %v", address, err)
	}
	return conn, nil
}

func writeRequest(conn net.Conn, host string) error {
	request := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", host)
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
