# Lightweight HTTP Server Clone
Designed completely using Go, following this [Codecrafters Demo](shttps://app.codecrafters.io/courses/http-server/overview)

## Core Features
- Connect to TCP ports
- Parses requested URL path
- Supports requests from concurrent connections
- Supports `GET` requests
   - Supports `echo` requests
   - Supports requests for fetching `User-Agent`
   - Supports requests for *file retrieval* from server
- Supports `POST` requests
   - Supports requests for saving a file to a server

## Additional Features
- Supports content-encoding via `gzip` compression

## See it in action!
> Please don't use this for commercial deployment.
1. Download app/server.go, build and run `server`
2. In a separate terminal instance, try the following commands!
- To connect to a port:
`nc -vz 127.0.0.1 4221`
- To send a basic HTTP `GET` request
`curl -i http://localhost:4221`
- To send a `GET` request for a URL
`curl -i GET http://localhost:4221/index.html`
