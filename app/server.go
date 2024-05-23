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

func getRequestBody(input string) string {
	components := strings.Split(input, "\r\n\r\n")
	return components[len(components)-1]
}

func respondWithBody(body, headers interface{}, conn net.Conn) error {
	var contentLength int
	switch body := body.(type) {
	case string:
		contentLength = len(body)
	case []byte:
		contentLength = len(body)
	}
	var output string
	switch headers := headers.(type) {
	case map[string]string:
		output = "HTTP/1.1 200 OK\r\n"
		for header, value := range headers {
			output += fmt.Sprintf("%s: %s\r\n", header, value)
		}
		output += fmt.Sprintf("Content-Length: %d\r\n\r\n%s", contentLength, body)
	case string: // assume to be Content-Type
		output = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", headers, contentLength, body)
	}
	_, err := conn.Write([]byte(output))
	return err
}

func compressedEchoRequest(compression, target string, conn net.Conn) error {
	content := target[6:]
	var err error
	if compression == "gzip" {
		err = respondWithBody(content, map[string]string{"Content-Type": "text/plain", "Content-Encoding": "gzip"}, conn)
	} else {
		err = respondWithBody(content, "text/plain", conn)
	}
	return err
}

func postFile(target, body string, conn net.Conn) error {
	directory := os.Args[2]
	fileName := target[7:]
	fileContents := strings.Trim(body, "\x00")
	err := os.WriteFile(directory+fileName, []byte(fileContents), 0644)
	if err != nil {
		fmt.Println("Error creating/writing to file.")
		return err
	}
	_, err = conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
	return err
}

func getFile(target string, conn net.Conn) error {
	fileName := target[7:]
	directory := os.Args[2]
	data, err := os.ReadFile(directory + fileName)
	if err != nil {
		fmt.Println("File not found, returning response 404.")
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return err
	}
	return respondWithBody(data, map[string]string{"Content-Type": "application/octet-stream"}, conn)
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
	target := strings.Split(input, " ")[1]
	requestType := strings.Split(input, " ")[0]
	headers := parseRequestHeaders(input)
	var err error
	if requestType == "GET" {
		switch {
		case target == "/":
			_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		case strings.Contains(target, "/echo/"):
			if headers["Accept-Encoding"] != "" {
				err = compressedEchoRequest(headers["Accept-Encoding"], target, conn)
			} else {
				err = echoRequest(target, conn)
			}
		case target == "/user-agent":
			err = getUserAgent(headers["User-Agent"], conn)
		case strings.Contains(target, "/files/"):
			err = getFile(target, conn)
		default:
			_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	} else if requestType == "POST" {
		err = postFile(target, getRequestBody(input), conn)
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
