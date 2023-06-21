# Vhoster

Vhoster is a library for creating virtual hosts in a Go server. It is intended
to be used as a library, but it also includes a gateway that can be used to
create and delete vhosts dynamically.

The gateway can also be used as a central point of entry for all requests, and
it can be used to load balance requests to multiple servers.


## Quick Start

### Locally

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

### In docker compose

1. Run `docker compose up`, it will build and start two container: (a) the
   gateway on port 8080 and vhost api on port 8083, and (b) the test server on
   port 8082, but the testserver is unreachable from the local machine
   directly.  We will set up a virtual host to access it.

   If you want to verify that testserver is not accessible, try running `curl
   localhost:8082` and you'll get an error.

2. It comes with preconfigured vhost "test.localhost:8080" that routes to
   "http://testserver:8082".  You can verify that it works by running `curl
   test.localhost:8080`, and you'll get the "Hello, World!"

3. Run the following command to instruct gateway to route all requests to
   "hello.localhost:8080" to "http://testserver:8082":
   ```sh
   curl -X POST -H "Content-Type: application/json" -d '{"host_prefix": "hello", "target": "http://testserver:8082"}' http://localhost:8083/vhost/
   ```

   This will instruct the gateway to route all requests that arrive at
   "test.localhost:8080" to "http://testserver:8082".

3. Now, you can access the test server by running `curl hello.localhost:8080`,
   and you'll get the "Hello, World!"

   At this point, both "test.localhost:8080" and "hello.localhost:8080" will
   route to the test server.

4. To list existing routes, and verify that, run:
   ```sh
   curl -X GET localhost:8083/vhost/
   ```
   (`-X GET` is optional and instructs curl to use GET method, which is the default
   method for curl, so you can omit it: `curl localhost:8083/vhost/`)

   The API server will respond on this GET request with JSON that lists all
   existing vhosts.

4. You can delete the route now, by running:
   ```sh
   curl -X DELETE localhost:8083/vhost/hello
   ```
