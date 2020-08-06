#!/bin/sh

grpcwebproxy --backend_addr=localhost:8080 \
 --backend_tls=false \
 --server_http_debug_port=9998 \
 --run_tls_server=false \
 --allow_all_origins
