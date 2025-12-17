/*
Tesla-http-proxy-insecure is an HTTP server that exposes a REST API for sending end-to-end
authenticated commands to vehicles without TLS encryption.

WARNING: This proxy does NOT encrypt client traffic. Use only behind TLS-terminating
infrastructure (Cloud Run, nginx, Traefik, K8s ingress) or in local development environments.
The proxy still uses HTTPS for outbound Tesla API calls.

This is a thin wrapper around the pkg/proxy package. See the README.md file in the repository
root directory for instructions on using this application.
*/
package main
