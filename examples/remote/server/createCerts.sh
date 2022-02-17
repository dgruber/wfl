#!/bin/bash

# private key
openssl genrsa -out server.key 4096

# public key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

chmod 0400 server.key