#!/usr/bin/env python3

import os
import time

import requests


def main():
    server_ip = os.getenv("SERVER_IP", "localhost")
    server_port = os.getenv("SERVER_PORT", "12345")

    server_url = f"http://{server_ip}:{server_port}"
    print(f"Connecting to server at: {server_url}")
    while True:
        try:
            response = requests.get(server_url)
            print("Server Response:")
            print(response.text)
        except requests.exceptions.RequestException as e:
            print(f"Error connecting to server: {e}")
        time.sleep(5)


if __name__ == "__main__":
    main()
