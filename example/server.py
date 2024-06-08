import socket


def start_server(host="10.0.2.2", port=1234):
    # Create a TCP/IP socket
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        # Bind the socket to the address and port
        server_socket.bind((host, port))
        print(f"Server started at {host}:{port}")

        # Listen for incoming connections
        server_socket.listen(5)
        print("Waiting for a connection...")

        while True:
            # Accept a connection
            client_socket, client_address = server_socket.accept()
            try:
                print(f"Connection from {client_address}")

                # Receive the data in small chunks
                while True:
                    data = client_socket.recv(1024)
                    if data:
                        print(f"Received: {data}")
                        # Echo the data back to the client
                        client_socket.sendall(data)
                    else:
                        break
            finally:
                # Clean up the connection
                client_socket.close()
                print(f"Connection from {client_address} closed")

    except Exception as e:
        print(f"Error: {e}")

    finally:
        server_socket.close()


if __name__ == "__main__":
    start_server()
