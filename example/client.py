import socket


def tcp_client(server_ip, server_port):
    # Create a TCP/IP socket
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        # Connect to the server
        client_socket.connect((server_ip, server_port))
        print(f"Connected to {server_ip}:{server_port}")

        # Send data
        message = "Hello, Server!"
        print(f"Sending: {message}")
        client_socket.sendall(message.encode())

        # Receive response
        data = client_socket.recv(1024)
        print(f"Received: {data.decode()}")

    except Exception as e:
        print(f"Error: {e}")

    finally:
        client_socket.close()
        print("Connection closed")


if __name__ == "__main__":
    server_ip = "10.0.0.1"  # Change to the server's IP address
    server_port = 1234  # Change to the server's port
    tcp_client(server_ip, server_port)
