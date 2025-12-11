package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	err := runServer(os.Args)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

const responseTemplate = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n"

func runServer(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: client [port]")
	}
	port, err := strconv.Atoi(args[1])
	if err != nil || port > 65535 || port < 1 {
		return fmt.Errorf("invalid port: %s", args[1])
	}
	fmt.Printf("port: %d", port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("error listening on ports %d: %v", port, err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Print("handling connection...\n")
	headers := []string{}
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
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
