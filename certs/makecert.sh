#!/bin/bash

# Generate key and self-signed certificate
openssl req -new -newkey rsa:2048 -x509 -sha256 -days 3650 -nodes -out tls.crt -keyout tls.key