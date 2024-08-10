#!/bin/bash

# Function to escape special characters
escape_string() {
    printf '%q' "$1"
}

if [ "$#" -ne 1 ]; then
    echo "Usage:  $0 <user>"
    exit 1
fi

user=$1

# Read password securely
read -e -s -p "Enter password: " password
echo

# Escape password to handle special characters
escaped_password=$(escape_string "$password")

# Generate the password file using podman and htpasswd
podman pull httpd:2
podman run --entrypoint htpasswd httpd:2 -Bbn "$user" "$escaped_password" > auth/htpasswd

