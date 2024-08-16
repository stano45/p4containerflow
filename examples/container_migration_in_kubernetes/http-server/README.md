# Simple HTTP Server with Request Counter

This Python program implements a simple HTTP server that keeps track of the
number of GET requests it has received. Each client receives an initial
response that includes the request count, server hostname, and IP address.
After the initial response, the server sends a dot (`.`) every second for 10
seconds.

## Usage

1. Run server

By default, the server listens on port 12345. To specify a different port, use
the `-p` option:

```
python3 main.py -p <port>
```

2. Send an HTTP request

You can use any HTTP client to make a GET request to the server. The following
script uses `curl` with the default port number:

```
./send_request.sh
```
