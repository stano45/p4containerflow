#!/bin/bash

# See more examples at:
# - https://distribution.github.io/distribution/about/deploying/
# - https://www.redhat.com/sysadmin/simple-container-registry

sudo podman run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry \
  -v "$(pwd)"/auth:/auth:z \
  -v "$(pwd)"/certs:/certs:z \
  -v "$(pwd)"/data:/var/lib/registry:z \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/container-registry-local.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/container-registry-local.key \
  -e REGISTRY_COMPATIBILITY_SCHEMA1_ENABLED=true \
  registry:2
