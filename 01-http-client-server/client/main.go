package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	runServer(os.Args)
}

func runServer(args []string) error {
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
	fmt.Printf("host=%s,port=%v", host, port)
	return nil
}
