package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// TODO: add description
func echoRequest(target string, conn net.Conn) {
	content := target[6:]
	var contentLength int = len(content)
	contentType := "text/plain"
	output := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", contentType, contentLength, content)
	conn.Write([]byte(output))
}

// TODO: add description
func parseRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	if _, err := conn.Read(buffer); err != nil {
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
	} else if strings.Contains(target, "/echo/") {
		echoRequest(target, conn)
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
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
	defer func() {
		if conn.Close() != nil || l.Close() != nil {
			fmt.Println("Error closing connection to client")
		}
	}()

	parseRequest(conn)
}
