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
	err := runClient(os.Args)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runClient(args []string) error {
	var host string
	var port int
	var err error
	if len(args) == 3 {
		host = args[1]
		port, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[2])
		}
	} else if len(args) == 2 {
		host = args[1]
		port = 80
	} else {
		return fmt.Errorf("usage: client [host] ([port])")
	}

	address := host + ":" + strconv.Itoa(port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to %s: %v", address, err)
	}
	defer conn.Close()

	request := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", host)
	_, err = conn.Write([]byte(request))
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %s", err)
	}

	results := [][]byte{}
	for responseComplete := false; !responseComplete; {
		b := make([]byte, 4096)
		count, err := conn.Read(b)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading response: %s", err)
		}
		if count > 0 {
			results = append(results, b[:count])
		}
		if err == io.EOF {
			responseComplete = true
		}
	}
	response := string(bytes.Join(results, []byte{}))
	fmt.Print(response)

	return nil
}
