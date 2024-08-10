#!/usr/bin/sh

# The -N option ensures that curl does not buffer the output and
# you should see the data as it is received from the server in real-time.
curl -N http://localhost:12345

