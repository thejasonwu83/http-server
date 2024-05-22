package main

// TODO: give each function a description

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func respondWithBody(body, contentType string, conn net.Conn) error {
	contentLength := len(body)
	output := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", contentType, contentLength, body)
	_, err := conn.Write([]byte(output))
	return err
}

func echoRequest(target string, conn net.Conn) error {
	content := target[6:]
	return respondWithBody(content, "text/plain", conn)
}

func getUserAgent(userAgent string, conn net.Conn) error {
	if userAgent == "" {
		return errors.New("request does not contain User-Agent header")
	}
	return respondWithBody(userAgent, "text/plain", conn)
}

func parseRequestHeaders(input string) map[string]string {
	headersStart := strings.Index(input, "\r\n")
	headersEnd := strings.Index(input, "\r\n\r\n")
	headersContent := input[headersStart+4 : headersEnd]
	fields := strings.Split(headersContent, "\r\n")
	headers := make(map[string]string)
	for _, field := range fields {
		divider := strings.Index(field, ":")
		headers[field[:divider]] = field[divider+2:]
	}
	return headers
}

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
	target := input[4 : targetEndIdx-1] // GET <target> HTTP/1.1
	headers := parseRequestHeaders(input)
	var err error
	switch {
	case target == "/":
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case strings.Contains(target, "/echo/"):
		err = echoRequest(target, conn)
	case target == "/user-agent":
		err = getUserAgent(headers["User-Agent"], conn)
	default:
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	if err != nil {
		fmt.Println("Error providing response :", err.Error())
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

	for {
		go parseRequest(conn)
		conn, err = l.Accept()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("Reached EOF for client input, terminating.")
				break
			}
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
	}
}
