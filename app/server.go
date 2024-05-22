package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	if _, err = conn.Read(buffer); err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
		os.Exit(1)
	}
	input := string(buffer)
	fmt.Println("Received input: ", input) //debug
	targetEndIdx := strings.Index(input, "HTTP")
	if targetEndIdx == -1 {
		fmt.Println("Error parsing input")
		os.Exit(1)
	}
	target := input[4 : targetEndIdx-1]
	if target == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

}
