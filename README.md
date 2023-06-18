# Vhoster

Vhoster is a library for creating virtual hosts in a Go server. It is intended
to be used as a library, but it also includes a gateway that can be used to
create and delete vhosts dynamically.

The gateway can also be used as a central point of entry for all requests, and
it can be used to load balance requests to multiple servers.


## Quick Start

1. Start the server listening on port 8082
    ```sh
    go run cmd/server/main.go
    ```

2. Start the gateway listening on port 8083
    ```sh
    go run cmd/gateway/main.go
    ```

3. Create the vhost in the gateway
    ```sh
    curl -X POST -H "Content-Type: application/json" -d '{"host_prefix": "test", "target": "http://localhost:8082"}' http://localhost:8083/vhost/
    ```

    then, to test, run `curl test.localhost:8081`, and you'll get the "Hello, World!"

4. Delete the vhost in the gateway
    ```sh
    curl -X DELETE localhost:8083/vhost/test
    ```
