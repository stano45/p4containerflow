#!/usr/bin/env python3

import argparse
import logging
import socket
import time
from http.server import BaseHTTPRequestHandler, HTTPServer

# Initialize counter
counter = 0

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)


class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global counter
        # Increment the counter
        counter += 1
        # Log the current count
        logging.info(f"Request count: {counter}")

        # Determine the hostname and IP address
        hostname = socket.gethostname()
        ip_address = socket.gethostbyname(hostname)

        # Respond with the initial headers
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        # Indicate that the connection should be kept alive
        self.send_header("Connection", "keep-alive")
        # Use chunked transfer encoding
        self.send_header("Transfer-Encoding", "chunked")
        self.end_headers()

        # Send the initial message with the counter, hostname, and IP address
        message_str = (
            f"Request [{counter}]: "
            f"Hostname: {hostname}, "
            f"IP Address: {ip_address}\n"
        )
        initial_message = message_str.encode('utf-8')

        self.send_chunk(initial_message)

        # Send a dot every second for 10 seconds
        self.send_dots_for_duration(10)

    def send_chunk(self, data):
        """Send a chunk of data."""
        try:
            self.wfile.write(f"{len(data):X}\r\n".encode('utf-8'))
            self.wfile.write(data)
            # End of the chunk
            self.wfile.write(b"\r\n")
            # Ensure the data is sent immediately
            self.wfile.flush()
        except ConnectionResetError as e:
            logging.warning("Connection reset by peer: %s", e)
            raise
        except BrokenPipeError as e:
            logging.warning("Error while sending data: %s", e)
            raise
        except Exception as e:
            logging.error("Unexpected error: %s", e)
            raise

    def send_dots_for_duration(self, duration):
        """Send a dot every second for the given duration."""
        end_time = time.time() + duration
        while time.time() < end_time:
            # Send a dot as a separate chunk
            dot = b"."
            try:
                self.send_chunk(dot)
            except (ConnectionResetError, BrokenPipeError) as e:
                logging.warning("Connection error: %s", e)
                break
            except Exception as e:
                logging.error("Unexpected error: %s", e)
                break
            # Wait for 1 second before sending another chunk
            time.sleep(1)
        else:
            # Send an empty chunk to indicate the end of the response
            try:
                self.send_chunk(b"")
            except (ConnectionResetError, BrokenPipeError) as e:
                logging.warning("Error while sending end of response: %s", e)
            except Exception as e:
                logging.error("Unexpected error: %s", e)


def run(
    server_class=HTTPServer,
    handler_class=SimpleHTTPRequestHandler,
    port=12345
):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    logging.info(f'Starting HTTP server on port {port}...')
    httpd.serve_forever()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="A simple HTTP server with a request counter"
    )
    parser.add_argument(
        '-p', '--port', type=int, default=12345,
        help='Port number to run the server on (default: 12345)'
    )
    args = parser.parse_args()
    run(port=args.port)
