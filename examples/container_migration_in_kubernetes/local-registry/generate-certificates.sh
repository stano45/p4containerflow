#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage:  $0 <hostname>"
    exit 0
fi

# Generate a private key for the CAPermalink
openssl ecparam \
	-out certs/container-registry-root.key \
	-name prime256v1 \
	-genkey

# Generate a certificate signing request for the CA
openssl req -new \
	-sha256 \
	-subj "/CN=$1" \
	-addext "subjectAltName = DNS:$1" \
	-key certs/container-registry-root.key \
	-out certs/container-registry-root.csr

# Generate a root certificate
openssl x509 -req \
	-sha256 \
	-days 3650 \
	-in		certs/container-registry-root.csr \
	-signkey	certs/container-registry-root.key \
	-out		certs/container-registry-root-CA.crt

# Create a private key for the certificate
openssl ecparam \
	-out certs/container-registry-local.key \
	-name prime256v1 \
	-genkey

# Create a certificate signing request for the server SSL
openssl req -new \
	-sha256 \
	-subj "/CN=$1" \
	-addext "subjectAltName = DNS:$1" \
	-key certs/container-registry-local.key \
	-out certs/container-registry-local.csr

# Create a certificate and sign it with the CA private key
openssl x509 -req \
	-in	certs/container-registry-local.csr \
	-CA	certs/container-registry-root-CA.crt \
	-CAkey	certs/container-registry-root.key \
	-out	certs/container-registry-local.crt \
	-CAcreateserial \
	-days 3650 \
	-sha256

